package plansvc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
)

func code(err error) connect.Code {
	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce.Code()
	}
	return 0
}

// TestOverrideTarget: the policy checks an override, the plan keeps the
// recommendation, the override, and the reason as separate records, and
// RemoveOverride removes it (D-69, D-293).
func TestOverrideTarget(t *testing.T) {
	f := defaultFixture(t)
	ctx := context.Background()
	if _, _, err := f.request(t, "uid-a"); err != nil {
		t.Fatal(err)
	}
	stored, _, err := f.maker.Plans.Get(ctx, "uid-a")
	if err != nil {
		t.Fatal(err)
	}
	ex := stored.Sessions[0].Exercises[0]
	rec := ex.Target
	id := string(rec.Exercise)
	var sets []*workoutappv1.PlannedSet
	for _, w := range rec.Working {
		sets = append(sets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load) + 100, RirTarget: 0})
	}
	req := func(m *workoutappv1.OverrideTargetRequest) error {
		_, err := f.client.OverrideTarget(ctx, as("uid-a", m))
		return err
	}

	for _, tc := range []struct {
		name string
		m    *workoutappv1.OverrideTargetRequest
		want connect.Code
	}{
		{"no reason", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, WorkingSets: sets, Reason: "  "}, connect.CodeInvalidArgument},
		{"reason over 200 characters", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, WorkingSets: sets, Reason: strings.Repeat("a", 201)}, connect.CodeInvalidArgument},
		{"bad date", &workoutappv1.OverrideTargetRequest{Today: "02.10.2026", ExerciseId: id, WorkingSets: sets, Reason: "x"}, connect.CodeInvalidArgument},
		{"a set fewer", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, WorkingSets: sets[1:], Reason: "x"}, connect.CodeInvalidArgument},
		{"no such load", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, Reason: "x",
			WorkingSets: append([]*workoutappv1.PlannedSet{{Reps: 10, LoadTenthLb: 155}}, sets[1:]...)}, connect.CodeInvalidArgument},
		{"cardio", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: "treadmill", WorkingSets: sets, Reason: "x"}, connect.CodeInvalidArgument},
		{"not in the plan", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: "leg_press", WorkingSets: sets, Reason: "x"}, connect.CodeFailedPrecondition},
	} {
		if err := req(tc.m); code(err) != tc.want {
			t.Errorf("%s: OverrideTarget = %v, want %v", tc.name, err, tc.want)
		}
	}
	// A refused override names the rule and the place, and no value of
	// the owner.
	bad := append([]*workoutappv1.PlannedSet{{Reps: 10, LoadTenthLb: 155}}, sets[1:]...)
	if err := req(&workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, WorkingSets: bad, Reason: "secret reason"}); !strings.Contains(err.Error(), string(policy.RuleOverride)+" working[0]") || strings.Contains(err.Error(), "secret") {
		t.Errorf("refusal %v: want the rule and the place, with no reason", err)
	}

	res, err := f.client.OverrideTarget(ctx, as("uid-a", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, WorkingSets: sets, Reason: "  Felt easy.  "}))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range res.Msg.GetPlan().GetSessions() {
		for _, e := range s.GetExercises() {
			if e.GetExerciseId() != id {
				if e.GetOverride() != nil {
					t.Fatalf("%s: an override of another exercise", e.GetExerciseId())
				}
				continue
			}
			o := e.GetOverride()
			if o == nil || o.GetReason() != "Felt easy." || len(o.GetWorkingSets()) != len(rec.Working) || len(o.GetRecommendedWorkingSets()) != len(rec.Working) {
				t.Fatalf("override %v", o)
			}
			for i, w := range o.GetWorkingSets() {
				if int(w.GetLoadTenthLb()) != int(rec.Working[i].Load)+100 || int(w.GetRirTarget()) != rec.Working[i].RIR {
					t.Fatalf("override working[%d] %v: want 10 lb more, and the reps in reserve of the recommendation", i, w)
				}
			}
			if len(o.GetCalibrationSets()) != len(rec.Calibration) {
				t.Fatalf("override calibration %v", o.GetCalibrationSets())
			}
			if len(rec.Calibration) > 0 && o.GetCalibrationSets()[0].GetLoadTenthLb() != o.GetWorkingSets()[0].GetLoadTenthLb() {
				t.Fatal("the calibration set does not follow the first working set")
			}
			if e.GetWorkingSets()[0].GetLoadTenthLb() != int32(rec.Working[0].Load) {
				t.Fatal("the recommendation changed")
			}
		}
	}
	after, _, _ := f.maker.Plans.Get(ctx, "uid-a")
	o := after.Sessions[0].Exercises[0].Override
	if o == nil || o.Reason != "Felt easy." || !slices.Equal(o.Recommendation.Working, rec.Working) || !slices.Equal(after.Sessions[0].Exercises[0].Target.Working, rec.Working) || o.At.IsZero() {
		t.Fatalf("stored override %+v", o)
	}

	rm, err := f.client.RemoveOverride(ctx, as("uid-a", &workoutappv1.RemoveOverrideRequest{Today: "2026-10-02", ExerciseId: id}))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range rm.Msg.GetPlan().GetSessions() {
		for _, e := range s.GetExercises() {
			if e.GetOverride() != nil {
				t.Fatal("RemoveOverride kept the override")
			}
		}
	}
	if p, _, _ := f.maker.Plans.Get(ctx, "uid-a"); p.Sessions[0].Exercises[0].Override != nil {
		t.Fatal("the stored plan kept the override")
	}

	// A caller with no plan, and a caller with no uid.
	if _, err := f.client.OverrideTarget(ctx, as("uid-b", &workoutappv1.OverrideTargetRequest{Today: "2026-10-02", ExerciseId: id, WorkingSets: sets, Reason: "x"})); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("no plan: %v", err)
	}
	if _, err := f.client.RemoveOverride(ctx, as("", &workoutappv1.RemoveOverrideRequest{Today: "2026-10-02", ExerciseId: id})); code(err) != connect.CodeUnauthenticated {
		t.Errorf("no uid: %v", err)
	}
}

// TestGetPlanToday: GetPlan refuses a bad date, and a nil Dated gives the
// stored targets.
func TestGetPlanToday(t *testing.T) {
	f := defaultFixture(t)
	if _, _, err := f.request(t, "uid-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.client.GetPlan(context.Background(), as("uid-a", &workoutappv1.GetPlanRequest{Today: "2026-12-25"})); code(err) != connect.CodeInvalidArgument {
		t.Errorf("a date far from UTC: %v, want INVALID_ARGUMENT", err)
	}
	res, err := f.client.GetPlan(context.Background(), as("uid-a", &workoutappv1.GetPlanRequest{Today: "2026-10-02"}))
	if err != nil || res.Msg.GetPlan() == nil {
		t.Fatalf("GetPlan = %v, %v", res, err)
	}
}

// expiring is a Dated that marks each override as expired.
type expiring struct{}

func (expiring) ForDate(_ context.Context, _ string, p plan.Plan, _ string) (plan.Plan, error) {
	for i := range p.Sessions {
		p.Sessions[i].Exercises = slices.Clone(p.Sessions[i].Exercises)
		for j := range p.Sessions[i].Exercises {
			if o := p.Sessions[i].Exercises[j].Override; o != nil {
				c := *o
				c.Expired = true
				p.Sessions[i].Exercises[j].Override = &c
			}
		}
	}
	return p, nil
}

// TestOverrideExpired: GetPlan with a date gives the expiry of ForDate in
// the contract, and the stored override stays (D-294, D-295).
func TestOverrideExpired(t *testing.T) {
	f := defaultFixture(t)
	ctx := context.Background()
	if _, _, err := f.request(t, "uid-a"); err != nil {
		t.Fatal(err)
	}
	stored, _, _ := f.maker.Plans.Get(ctx, "uid-a")
	rec := stored.Sessions[0].Exercises[0].Target
	if err := f.maker.Plans.Update(ctx, "uid-a", stored.CreatedAt, func(p *plan.Plan) error {
		p.SetOverride(rec.Exercise, &plan.Override{Target: rec, Recommendation: rec, Reason: "x", Today: "2026-10-02"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	srv := New(f.maker, expiring{})
	res, err := srv.GetPlan(auth.WithUserID(ctx, "uid-a"), connect.NewRequest(&workoutappv1.GetPlanRequest{Today: "2026-10-02"}))
	if err != nil {
		t.Fatal(err)
	}
	if o := res.Msg.GetPlan().GetSessions()[0].GetExercises()[0].GetOverride(); o == nil || !o.GetExpired() {
		t.Fatalf("override %v, want it expired", o)
	}
	again, _, _ := f.maker.Plans.Get(ctx, "uid-a")
	if o := again.Sessions[0].Exercises[0].Override; o == nil || o.Expired || o.Today != "2026-10-02" {
		t.Fatalf("stored override %+v, want it stored with no expiry", o)
	}
}
