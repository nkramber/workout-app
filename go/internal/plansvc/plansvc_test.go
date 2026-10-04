package plansvc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/profile"
)

const uidHeader = "X-Test-Uid"

var now = time.Date(2026, 10, 2, 15, 0, 0, 0, time.UTC)

type fixture struct {
	client workoutappv1connect.PlanServiceClient
	fake   *ai.Fake
	maker  *plan.Maker
}

// newFixture serves the service over HTTP. The header X-Test-Uid stands
// in for the auth interceptor. The user "uid-a" has a profile and a
// confirmed chest press and seated row.
func newFixture(t *testing.T, caps ai.Caps) *fixture {
	t.Helper()
	ctx := context.Background()
	profiles, inventories := profile.NewMemory(), inventory.NewMemory()
	p := profile.Profile{Experience: profile.Intermediate, Template: domain.TemplateStrength, Groups: []domain.MuscleGroup{domain.GroupChest, domain.GroupBack},
		AgeYears: 32, HeightIn: 70, WeightLb: 180, TrainingDays: 2}
	if err := profiles.Save(ctx, "uid-a", p); err != nil {
		t.Fatal(err)
	}
	var w []domain.Load
	for lb := int64(10); lb <= 200; lb += 10 {
		w = append(w, domain.Pounds(lb))
	}
	inv := inventory.Inventory{Machines: []inventory.Machine{
		{Entry: domain.InventoryEntry{Machine: "chest_press", Weights: w}, State: inventory.Confirmed},
		{Entry: domain.InventoryEntry{Machine: "seated_row", Weights: w}, State: inventory.Confirmed},
	}}
	if _, err := inventories.Update(ctx, "uid-a", func(inventory.Inventory) (inventory.Inventory, error) { return inv, nil }); err != nil {
		t.Fatal(err)
	}
	f := &fixture{fake: &ai.Fake{}}
	f.maker = &plan.Maker{
		AI:       &ai.Client{Provider: f.fake, Cap: ai.NewMemoryCap(caps)},
		Profiles: profiles, Inventory: inventories, Plans: plan.NewMemory(), Errors: &plan.MemoryErrors{},
		Now: func() time.Time { return now },
	}
	_, h := workoutappv1connect.NewPlanServiceHandler(New(f.maker, nil))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get(uidHeader); id != "" {
			r = r.WithContext(auth.WithUserID(r.Context(), id))
		}
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	f.client = workoutappv1connect.NewPlanServiceClient(srv.Client(), srv.URL)
	return f
}

func defaultFixture(t *testing.T) *fixture {
	return newFixture(t, ai.Caps{User: ai.USD, Project: ai.USD})
}

func as[T any](id string, msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	if id != "" {
		req.Header().Set(uidHeader, id)
	}
	return req
}

type event[T any] interface {
	*T
	GetProgress() *workoutappv1.PlanProgress
	GetPlan() *workoutappv1.Plan
}

// drain reads each event of a stream. It gives the progress events, the
// plan, and the error of the stream.
func drain[T any, P event[T]](s *connect.ServerStreamForClient[T]) ([]*workoutappv1.PlanProgress, *workoutappv1.Plan, error) {
	var steps []*workoutappv1.PlanProgress
	var p *workoutappv1.Plan
	for s.Receive() {
		ev := P(s.Msg())
		if ev.GetPlan() != nil {
			p = ev.GetPlan()
		} else {
			steps = append(steps, ev.GetProgress())
		}
	}
	return steps, p, s.Err()
}

func (f *fixture) request(t *testing.T, id string) ([]*workoutappv1.PlanProgress, *workoutappv1.Plan, error) {
	t.Helper()
	s, err := f.client.RequestPlan(context.Background(), as(id, &workoutappv1.RequestPlanRequest{Today: "2026-10-02"}))
	if err != nil {
		return nil, nil, err
	}
	return drain(s)
}

// TestRequestPlan: the stream sends the progress, then the plan, and
// GetPlan reads the plan again.
func TestRequestPlan(t *testing.T) {
	f := defaultFixture(t)
	steps, p, err := f.request(t, "uid-a")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range steps {
		got = append(got, s.GetStep())
		if s.GetMaxAttempts() != plan.MaxAttempts {
			t.Fatalf("max attempts %d", s.GetMaxAttempts())
		}
	}
	if strings.Join(got, ",") != "call,check,save" || steps[0].GetAttempt() != 1 {
		t.Fatalf("steps %v", steps)
	}
	if p == nil || len(p.GetSessions()) != 2 || p.GetAttempts() != 1 || p.GetCreatedAt() != "2026-10-02T15:00:00Z" || p.GetPromptVersion() != ai.PromptVersion {
		t.Fatalf("plan %v", p)
	}
	s := p.GetSessions()[0]
	if s.GetWarmUp().GetText() == "" || s.GetCoolDown().GetKind() != "cool_down" || len(s.GetExercises()) != 2 {
		t.Fatalf("session %v", s)
	}
	e := s.GetExercises()[0]
	if e.GetExerciseId() != "chest_press" || e.GetName() != "Chest press" || e.GetSource() != "luna" || e.GetReason() == "" ||
		len(e.GetCalibrationSets()) != 0 || len(e.GetWorkingSets()) != 3 || e.GetWorkingSets()[0].GetLoadTenthLb() != 100 || e.GetWorkingSets()[0].GetRirTarget() != 3 {
		t.Fatalf("exercise %v", e)
	}
	// The first set is the calibration, so the plan has no table of
	// D-267 (D-297).
	if !e.GetFirstSetCalibration() || len(e.GetCalibrationLoads()) != 0 {
		t.Fatalf("first-set calibration %v, calibration loads %v", e.GetFirstSetCalibration(), e.GetCalibrationLoads())
	}

	if e.GetReasonSource() != "luna" || p.GetLastRevision() != nil {
		t.Fatalf("reason source %q, last revision %v", e.GetReasonSource(), p.GetLastRevision())
	}

	res, err := f.client.GetPlan(context.Background(), as("uid-a", &workoutappv1.GetPlanRequest{}))
	if err != nil || res.Msg.GetPlan().GetCreatedAt() != p.GetCreatedAt() || len(res.Msg.GetExclusions()) != 0 {
		t.Fatalf("GetPlan = %v, %v", res, err)
	}

	// A revision gives the last revision and the reason source of each
	// revised exercise (D-288, D-290).
	at := time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)
	created := time.Date(2026, 10, 2, 15, 0, 0, 0, time.UTC)
	if err := f.maker.Plans.Update(context.Background(), "uid-a", created, func(q *plan.Plan) error {
		q.LastRevision = &plan.Revision{WorkoutID: "w1", At: at, Exercises: []domain.ExerciseID{"chest_press"}}
		q.Sessions[0].Exercises[0].ReasonSource = policy.SourceRules
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	res, err = f.client.GetPlan(context.Background(), as("uid-a", &workoutappv1.GetPlanRequest{}))
	r := res.Msg.GetPlan().GetLastRevision()
	if err != nil || r.GetWorkoutId() != "w1" || r.GetRevisedAt() != "2026-10-03T09:00:00Z" || !slices.Equal(r.GetExerciseIds(), []string{"chest_press"}) ||
		res.Msg.GetPlan().GetSessions()[0].GetExercises()[0].GetReasonSource() != "rules" {
		t.Fatalf("GetPlan after a revision = %v, %v", res, err)
	}
	other, err := f.client.GetPlan(context.Background(), as("uid-b", &workoutappv1.GetPlanRequest{}))
	if err != nil || other.Msg.GetPlan() != nil {
		t.Fatalf("GetPlan of another uid = %v, %v", other, err)
	}
}

// TestExcludeExercise: the new plan lacks the exercise, and GetPlan
// gives the exclusion with its name and reason.
func TestExcludeExercise(t *testing.T) {
	f := defaultFixture(t)
	s, err := f.client.ExcludeExercise(context.Background(), as("uid-a", &workoutappv1.ExcludeExerciseRequest{
		Today: "2026-10-02", ExerciseId: "seated_row", Reason: "It is always busy.",
	}))
	if err != nil {
		t.Fatal(err)
	}
	steps, p, err := drain(s)
	if err != nil || len(steps) != 3 || p == nil {
		t.Fatalf("steps %v, plan %v, err %v", steps, p, err)
	}
	for _, sess := range p.GetSessions() {
		for _, e := range sess.GetExercises() {
			if e.GetExerciseId() == "seated_row" {
				t.Fatal("the plan holds the excluded exercise")
			}
		}
	}
	res, err := f.client.GetPlan(context.Background(), as("uid-a", &workoutappv1.GetPlanRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	ex := res.Msg.GetExclusions()
	if len(ex) != 1 || ex[0].GetExerciseId() != "seated_row" || ex[0].GetName() != "Seated row" || ex[0].GetReason() != "It is always busy." {
		t.Fatalf("exclusions %v", ex)
	}
}

// TestErrorCodes: each error of the maker gives its Connect code, and
// a stream with an error sends no plan.
func TestErrorCodes(t *testing.T) {
	ctx := context.Background()
	f := defaultFixture(t)
	if _, _, err := f.request(t, ""); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("no uid: %v", err)
	}
	if _, err := f.client.GetPlan(ctx, as("", &workoutappv1.GetPlanRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetPlan with no uid: %v", err)
	}
	if _, _, err := f.request(t, "uid-b"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("no profile: %v", err)
	}
	s, _ := f.client.RequestPlan(ctx, as("uid-a", &workoutappv1.RequestPlanRequest{Today: "02.10.2026"}))
	if _, _, err := drain(s); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("bad date: %v", err)
	}
	x, _ := f.client.ExcludeExercise(ctx, as("uid-a", &workoutappv1.ExcludeExerciseRequest{Today: "2026-10-02", ExerciseId: "chest_press", Reason: strings.Repeat("x", 201)}))
	if _, _, err := drain(x); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("long reason: %v", err)
	}

	f.fake.Reply = func(ai.Call) (ai.Reply, error) { return ai.Reply{}, errors.New("openai: HTTP 500") }
	steps, p, err := f.request(t, "uid-a")
	if connect.CodeOf(err) != connect.CodeUnavailable || p != nil || len(steps) != plan.MaxAttempts {
		t.Errorf("four failures: %v, plan %v, %d steps", err, p, len(steps))
	}
	if strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("the error names the cause of a call: %v", err)
	}

	capped := newFixture(t, ai.Caps{})
	if _, _, err := capped.request(t, "uid-a"); connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Errorf("capped: %v", err)
	}
}

func TestFail(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code connect.Code
	}{
		{plan.ErrConflict, connect.CodeAborted},
		{context.Canceled, connect.CodeCanceled},
		{context.DeadlineExceeded, connect.CodeDeadlineExceeded},
		{errors.New("firestore: projects/p/databases/x"), connect.CodeInternal},
	} {
		err := fail(tc.err)
		if connect.CodeOf(err) != tc.code || strings.Contains(err.Error(), "projects/") {
			t.Errorf("fail(%v) = %v, want %v", tc.err, err, tc.code)
		}
	}
}
