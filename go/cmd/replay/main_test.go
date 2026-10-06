package main

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/revise"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// TestReplayV7 replays the records of policy version 7 under the
// current version over the memory stores.
func TestReplayV7(t *testing.T) {
	s := stores{plans: plan.NewMemory(), workouts: workout.NewMemory(), inventory: inventory.NewMemory()}
	seedV7(t, s)
	r := &revise.Reviser{Plans: s.plans, Workouts: s.workouts, Inventory: s.inventory}
	rp, err := r.Replay(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	rep := newReport()
	rep.add(rp)
	checkV7Report(t, rep)
}

// TestReplayCurrent: a replay of the records of the current version
// changes nothing.
func TestReplayCurrent(t *testing.T) {
	s := stores{plans: plan.NewMemory(), workouts: workout.NewMemory(), inventory: inventory.NewMemory()}
	seedV7(t, s)
	r := &revise.Reviser{Plans: s.plans, Workouts: s.workouts, Inventory: s.inventory}
	ctx := context.Background()
	rp, err := r.Replay(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	// Give the plan the records of the replay, as the current version
	// gives them.
	if err := s.plans.Update(ctx, uid, created, func(q *plan.Plan) error {
		k := 0
		for i := range q.Sessions {
			for j := range q.Sessions[i].Exercises {
				q.Sessions[i].Exercises[j].Record = rp.Records[k].Now
				k++
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if rp, err = r.Replay(ctx, uid); err != nil {
		t.Fatal(err)
	}
	rep := newReport()
	rep.add(rp)
	if rep.Records.Changed != 0 || rep.Records.Unchanged != 4 || rep.Records.ByVersion[strconv.Itoa(policy.Version)] != 4 {
		t.Fatalf("records = %+v, want 4 unchanged records of version %d", rep.Records, policy.Version)
	}
}

// TestReplayNoPlan: a user with no plan and no workout gives a user and
// no record.
func TestReplayNoPlan(t *testing.T) {
	r := &revise.Reviser{Plans: plan.NewMemory(), Workouts: workout.NewMemory(), Inventory: inventory.NewMemory()}
	rp, err := r.Replay(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	rep := newReport()
	rep.add(rp)
	if rep.Users != 1 || rep.UsersWithPlan != 0 || rep.Records.Total != 0 || rep.Copies.Total != 0 {
		t.Fatalf("report = %+v, want one user and nothing else", rep)
	}
}

func TestChangedFields(t *testing.T) {
	a := domain.PlannedExercise{Exercise: "chest_press", RestSeconds: 60, Working: []domain.WorkingSet{{Reps: 10, Load: domain.Pounds(25), RIR: 2}, {Reps: 10, Load: domain.Pounds(25), RIR: 2}}}
	if got := changedFields(a, a); got != nil {
		t.Fatalf("same = %v, want nil", got)
	}
	b := a
	b.Working = slices.Clone(a.Working)
	b.Working[1].Load = domain.Pounds(30)
	b.Working[0].RIR = 3
	b.FollowMax = domain.Pounds(30)
	b.Calibration = []domain.CalibrationSet{}
	if got, want := changedFields(a, b), []string{"load", "rir", "follow_max"}; !slices.Equal(got, want) {
		t.Fatalf("changedFields = %v, want %v (a nil and an empty list are the same)", got, want)
	}
	c := a
	c.Working = a.Working[:1]
	if got, want := changedFields(a, c), []string{"sets"}; !slices.Equal(got, want) {
		t.Fatalf("changedFields = %v, want %v", got, want)
	}
}

func TestChangeRules(t *testing.T) {
	cases := []struct {
		name        string
		stored, now policy.Record
		fields      []string
		want        []string
	}{
		{"a field rule", policy.Record{Rules: []policy.RuleID{policy.RuleAddReps}}, policy.Record{Rules: []policy.RuleID{policy.RuleAddReps}}, []string{"follow_max"}, []string{"follow.first-set"}},
		{"a new rule and a dropped rule", policy.Record{Rules: []policy.RuleID{policy.RuleAddReps}}, policy.Record{Rules: []policy.RuleID{policy.RuleLoadStep}}, []string{"load", "reps"}, []string{"progress.add-reps", "progress.load-step"}},
		{"the same rules", policy.Record{Rules: []policy.RuleID{policy.RuleEffortHold}}, policy.Record{Rules: []policy.RuleID{policy.RuleEffortHold}}, []string{"rir"}, []string{"progress.effort-hold"}},
		{"a field rule that a record names", policy.Record{}, policy.Record{Rules: []policy.RuleID{policy.RuleCalibrationFirstSet}}, []string{"first_set_calibration"}, []string{"calibration.first-set"}},
		{"a proposal of Luna", policy.Record{Source: policy.SourceLuna}, policy.Record{Source: policy.SourceLuna}, []string{"reps"}, []string{RuleLuna}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := changeRules(c.stored, c.now, c.fields); !slices.Equal(got, c.want) {
				t.Fatalf("changeRules = %v, want %v", got, c.want)
			}
		})
	}
}

func TestCopyViolations(t *testing.T) {
	var c CopyReport
	c.ViolationsByRule = map[string]int{}
	v := []policy.Violation{{Rule: policy.RuleRepBounds, Where: "working[0]"}, {Rule: policy.RuleRepBounds, Where: "working[1]"}, {Rule: policy.RuleLoadCeiling, Where: "working[0]"}}
	c.add(revise.CopyReplayed{Rebuilt: true, Violations: v})
	c.add(revise.CopyReplayed{Override: true})
	want := CopyReport{Total: 2, Overrides: 1, NotRebuilt: 1, SameAsRules: 1, OutsideBounds: 1, ViolationsByRule: map[string]int{"reps.bounds": 1, "load.ceiling": 1}}
	if b, w := mustJSON(t, c), mustJSON(t, want); b != w {
		t.Fatalf("copies = %s, want %s", b, w)
	}
}

func TestCheckTarget(t *testing.T) {
	cases := []struct {
		project  string
		live     bool
		emulator string
		refuse   string
	}{
		{"demo-workout-app", false, "127.0.0.1:8381", ""},
		{"nk-workout-app-prod", true, "", ""},
		{"", false, "127.0.0.1:8381", "-project"},
		{"nk-workout-app-prod", false, "", "-live"},
		{"nk-workout-app-prod", true, "127.0.0.1:8381", "unset one"},
	}
	for _, c := range cases {
		err := checkTarget(c.project, c.live, c.emulator)
		if c.refuse == "" && err != nil || c.refuse != "" && (err == nil || !strings.Contains(err.Error(), c.refuse)) {
			t.Errorf("checkTarget(%q, %v, %q) = %v, want %q", c.project, c.live, c.emulator, err, c.refuse)
		}
	}
}

// TestRunRefusesLive: with no emulator and no -live flag, the command
// opens no client.
func TestRunRefusesLive(t *testing.T) {
	var out bytes.Buffer
	err := run(context.Background(), []string{"-project", "nk-workout-app-prod"}, func(string) string { return "" }, &out)
	if err == nil || !strings.Contains(err.Error(), "-live") || out.Len() != 0 {
		t.Fatalf("run = %v, %q, want a refusal", err, out.String())
	}
}

// TestReportHoldsNoData: the JSON keys of the report are counts and
// rule ids alone (D-80).
func TestReportHoldsNoData(t *testing.T) {
	s := stores{plans: plan.NewMemory(), workouts: workout.NewMemory(), inventory: inventory.NewMemory()}
	seedV7(t, s)
	r := &revise.Reviser{Plans: s.plans, Workouts: s.workouts, Inventory: s.inventory}
	rp, err := r.Replay(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	rep := newReport()
	rep.add(rp)
	var out bytes.Buffer
	if err := write(rep, "", &out); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{uid, "chest_press", "seated_row", "lb", "Reason"} {
		if strings.Contains(out.String(), s) {
			t.Fatalf("the report holds %q:\n%s", s, out.String())
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestReplayStaleHistory: two workouts start from the same target
// before the revision of the first one, as in the live check of
// 2026-10-05. The copy of the second one passes the bounds without the
// first workout, so the report counts it as a stale history and not as
// a copy outside the bounds. A copy above each ceiling counts as
// outside the bounds.
func TestReplayStaleHistory(t *testing.T) {
	s := stores{plans: plan.NewMemory(), workouts: workout.NewMemory(), inventory: inventory.NewMemory()}
	est := map[domain.ExerciseID]domain.Load{"leg_press": domain.Pounds(50)}
	seedInventory(t, s, est, "leg_press")
	p := newPlan(t, s, est, []domain.ExerciseID{"leg_press"})
	start := p.Sessions[0].Exercises[0].Target
	if start.Working[0].Load <= domain.Pounds(20) {
		t.Fatalf("start = %+v, want a load above 20 lb", start)
	}
	reviseWorkout(t, s, logWorkout(t, s, 1, domain.Pounds(20), start))
	logWorkout(t, s, 2, 0, start)
	heavy := start
	heavy.Working = slices.Clone(start.Working)
	for i := range heavy.Working {
		heavy.Working[i].Load = domain.Pounds(200)
	}
	logWorkout(t, s, 3, 0, heavy)

	r := &revise.Reviser{Plans: s.plans, Workouts: s.workouts, Inventory: s.inventory}
	rp, err := r.Replay(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	rep := newReport()
	rep.add(rp)
	want := CopyReport{Total: 3, SameAsRules: 1, DifferentFromRules: 2, StaleHistory: 1, OutsideBounds: 1, ViolationsByRule: map[string]int{"load.ceiling": 1}}
	if got, w := mustJSON(t, rep.Copies), mustJSON(t, want); got != w {
		t.Fatalf("copies = %s, want %s", got, w)
	}
}
