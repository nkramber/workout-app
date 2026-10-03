//go:build emulator

// The acceptance story of PR-26 over the Firebase emulators. The fake
// provider stands in for Luna (D-24). The test saves the profile of the
// core user of D-31 through the API, and a synthetic inventory in the
// store. The plan request streams its progress and returns a valid plan.
// A malformed output gets a retry and an error record, and 4 failures
// give an error with no change. No plan holds an exercise of an injured
// area, an excluded exercise, or a machine that the owner did not
// confirm.
package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/capstore"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
)

type planEvent[T any] interface {
	*T
	GetProgress() *workoutappv1.PlanProgress
	GetPlan() *workoutappv1.Plan
}

func drainPlan[T any, P planEvent[T]](s *connect.ServerStreamForClient[T]) ([]*workoutappv1.PlanProgress, *workoutappv1.Plan, error) {
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

func planIDs(p *workoutappv1.Plan) []string {
	var out []string
	for _, s := range p.GetSessions() {
		for _, e := range s.GetExercises() {
			out = append(out, e.GetExerciseId())
		}
		if c := s.GetCardio(); c != nil {
			out = append(out, c.GetExerciseId())
		}
	}
	return out
}

func TestPlanAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	uid, token := signUp(t, authHost, fmt.Sprintf("plan-%d@example.test", time.Now().UnixNano()))
	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	if _, err := fs.Collection(allowlist.Collection).Doc(uid).Set(ctx, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	// The first call gives an output that is not JSON. Each later call
	// gives the echo of the fake.
	fake := &ai.Fake{}
	fake.Reply = func(c ai.Call) (ai.Reply, error) {
		if len(fake.Calls()) == 1 {
			return ai.Reply{Text: "not the JSON of the schema"}, nil
		}
		return ai.EchoReply(c)
	}
	base := startAPIWith(t, fake)
	profiles := workoutappv1connect.NewProfileServiceClient(http.DefaultClient, base)
	plans := workoutappv1connect.NewPlanServiceClient(http.DefaultClient, base)

	// The core user of D-31, with an injured knee, 3 training days, and
	// the treadmill as the cardio preference.
	_, err = profiles.SaveProfile(ctx, signed(token, &workoutappv1.SaveProfileRequest{Profile: &workoutappv1.Profile{
		Experience: "intermediate", GoalTemplate: "general_fitness",
		MuscleGroups: []string{"chest", "back", "shoulders", "biceps", "triceps", "quadriceps", "hamstrings", "glutes", "calves", "core"},
		InjuredAreas: []string{"knee"}, AgeYears: 32, HeightIn: 70, WeightLb: 180,
		CardioExerciseIds: []string{"treadmill"}, TrainingDays: 3,
	}}))
	if err != nil {
		t.Fatal(err)
	}
	var w []domain.Load
	for lb := int64(10); lb <= 200; lb += 10 {
		w = append(w, domain.Pounds(lb))
	}
	m := func(id domain.MachineID, st inventory.State) inventory.Machine {
		return inventory.Machine{Entry: domain.InventoryEntry{Machine: id, Weights: w}, Estimates: map[domain.ExerciseID]domain.Load{}, State: st}
	}
	inv := inventory.Inventory{Machines: []inventory.Machine{
		m("leg_press", inventory.Confirmed), m("chest_press", inventory.Confirmed), m("seated_row", inventory.Confirmed),
		m("cable_station", inventory.Confirmed), m("shoulder_press", inventory.Draft),
		{Entry: domain.InventoryEntry{Machine: "treadmill"}, Estimates: map[domain.ExerciseID]domain.Load{}, State: inventory.Confirmed},
	}}
	if _, err := inventory.FromFirestore(fs).Update(ctx, uid, func(inventory.Inventory) (inventory.Inventory, error) { return inv, nil }); err != nil {
		t.Fatal(err)
	}
	// The leg press and the treadmill load the knee, and the shoulder
	// press is a draft.
	allowed := []string{"chest_press", "seated_row", "lat_pulldown", "triceps_pulldown"}
	today := time.Now().UTC().Format(domain.DateLayout)

	// The request streams the progress and returns a valid plan.
	s, err := plans.RequestPlan(ctx, signed(token, &workoutappv1.RequestPlanRequest{Today: today}))
	if err != nil {
		t.Fatal(err)
	}
	steps, p, err := drainPlan(s)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, st := range steps {
		got = append(got, fmt.Sprintf("%s %d %s", st.GetStep(), st.GetAttempt(), st.GetPreviousStatus()))
	}
	if !slices.Equal(got, []string{"call 1 ", "call 2 malformed", "check 0 ", "save 0 "}) {
		t.Fatalf("steps %q", got)
	}
	if len(p.GetSessions()) != 3 || p.GetAttempts() != 2 || p.GetPromptVersion() != ai.PromptVersion {
		t.Fatalf("plan %v", p)
	}
	for _, id := range planIDs(p) {
		if !slices.Contains(allowed, id) {
			t.Fatalf("the plan holds %q", id)
		}
	}
	for _, sess := range p.GetSessions() {
		for _, e := range sess.GetExercises() {
			if e.GetSource() != "luna" || len(e.GetWorkingSets()) == 0 || e.GetWorkingSets()[0].GetLoadTenthLb() != 100 {
				t.Fatalf("exercise %v", e)
			}
		}
	}

	// The store holds the plan with each decision record (D-176), the
	// error record of the failed call (D-236), and the charge of each
	// call in the cap store (D-189).
	stored, ok, err := plan.FromFirestore(fs).Get(ctx, uid)
	if err != nil || !ok || stored.Sessions[0].Exercises[0].Record.InputHash == "" || stored.Sessions[0].Exercises[0].Record.PromptHash == "" {
		t.Fatalf("stored plan: %v, %v", ok, err)
	}
	errs, err := fs.Collection(plan.ErrorsCollection).Where("uid", "==", uid).Documents(ctx).GetAll()
	if err != nil || len(errs) != 1 || errs[0].Data()["status"] != "malformed" || errs[0].Data()["output"] != "not the JSON of the schema" {
		t.Fatalf("error records: %d, %v", len(errs), err)
	}
	if exp, ok := errs[0].Data()["expire_at"].(time.Time); !ok || exp.Before(time.Now().Add(89*24*time.Hour)) {
		t.Fatalf("expire_at %v", errs[0].Data()["expire_at"])
	}
	spend, err := fs.Doc(fmt.Sprintf("users/%s/%s/%s", uid, capstore.Collection, capstore.Month(time.Now()))).Get(ctx)
	if err != nil || spend.Data()["charged_nano_usd"].(int64) <= 0 {
		t.Fatalf("the cap store holds no charge: %v", err)
	}

	// An exclusion makes a new plan with no such exercise, and GetPlan
	// gives the exclusion.
	x, err := plans.ExcludeExercise(ctx, signed(token, &workoutappv1.ExcludeExerciseRequest{Today: today, ExerciseId: "chest_press", Reason: "Busy at that hour."}))
	if err != nil {
		t.Fatal(err)
	}
	if _, p, err = drainPlan(x); err != nil || p == nil || slices.Contains(planIDs(p), "chest_press") {
		t.Fatalf("plan after the exclusion: %v, %v", p, err)
	}
	res, err := plans.GetPlan(ctx, signed(token, &workoutappv1.GetPlanRequest{}))
	if err != nil || res.Msg.GetPlan().GetCreatedAt() != p.GetCreatedAt() || len(res.Msg.GetExclusions()) != 1 || res.Msg.GetExclusions()[0].GetReason() != "Busy at that hour." {
		t.Fatalf("GetPlan = %v, %v", res, err)
	}

	// Four failed calls give UNAVAILABLE, 4 error records, and no
	// change of the plan or of the exclusions (D-230, D-234).
	refuse := &ai.Fake{Reply: func(ai.Call) (ai.Reply, error) { return ai.Reply{Refusal: true}, nil }}
	failing := workoutappv1connect.NewPlanServiceClient(http.DefaultClient, startAPIWith(t, refuse))
	y, err := failing.ExcludeExercise(ctx, signed(token, &workoutappv1.ExcludeExerciseRequest{Today: today, ExerciseId: "seated_row"}))
	if err != nil {
		t.Fatal(err)
	}
	if steps, q, err := drainPlan(y); connect.CodeOf(err) != connect.CodeUnavailable || q != nil || len(steps) != plan.MaxAttempts {
		t.Fatalf("four failures: %v, plan %v, %d steps", err, q, len(steps))
	}
	if n := len(refuse.Calls()); n != plan.MaxAttempts {
		t.Fatalf("%d calls, want %d", n, plan.MaxAttempts)
	}
	after, err := plans.GetPlan(ctx, signed(token, &workoutappv1.GetPlanRequest{}))
	if err != nil || after.Msg.GetPlan().GetCreatedAt() != p.GetCreatedAt() || len(after.Msg.GetExclusions()) != 1 {
		t.Fatalf("the failed request changed the store: %v, %v", after, err)
	}
	errs, _ = fs.Collection(plan.ErrorsCollection).Where("uid", "==", uid).Documents(ctx).GetAll()
	if len(errs) != 1+plan.MaxAttempts {
		t.Fatalf("%d error records, want %d", len(errs), 1+plan.MaxAttempts)
	}
}
