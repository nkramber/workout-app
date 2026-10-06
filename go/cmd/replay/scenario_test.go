package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/revise"
	"github.com/nkramber/workout-app/go/internal/workout"
)

const uid = "uid-replay-1"

var created = time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)

// stores are the stores of one scenario.
type stores struct {
	plans     plan.Store
	workouts  workout.Store
	inventory inventory.Store
}

// weights are the weights of each machine of a scenario.
func weights() []domain.Load {
	var out []domain.Load
	for w := int64(10); w <= 300; w += 5 {
		out = append(out, domain.Pounds(w))
	}
	return out
}

// seedInventory confirms each machine with the weights of weights, and
// the estimates of the owner.
func seedInventory(t *testing.T, s stores, estimates map[domain.ExerciseID]domain.Load, machines ...domain.MachineID) {
	t.Helper()
	if estimates == nil {
		estimates = map[domain.ExerciseID]domain.Load{}
	}
	if _, err := s.inventory.Update(context.Background(), uid, func(inventory.Inventory) (inventory.Inventory, error) {
		var inv inventory.Inventory
		for _, m := range machines {
			inv.Machines = append(inv.Machines, inventory.Machine{Entry: domain.InventoryEntry{Machine: m, Weights: weights()}, Estimates: estimates, State: inventory.Confirmed})
		}
		return inv, nil
	}); err != nil {
		t.Fatal(err)
	}
}

// newPlan saves a plan of the rules alone for the exercises of each
// session, for a user with no history.
func newPlan(t *testing.T, s stores, estimates map[domain.ExerciseID]domain.Load, sessions ...[]domain.ExerciseID) plan.Plan {
	t.Helper()
	p := plan.Plan{CreatedAt: created, Today: "2026-10-01", PolicyVersion: policy.Version}
	for i, ids := range sessions {
		sess := plan.Session{Title: fmt.Sprintf("Session %d", i+1)}
		for _, id := range ids {
			e, _ := domain.DefaultCatalog().Exercise(id)
			rec, err := policy.Decide(policy.Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Weights: weights()}, Today: p.Today, Estimate: estimates[id]}, policy.Proposal{})
			if err != nil {
				t.Fatal(err)
			}
			sess.Exercises = append(sess.Exercises, plan.Exercise{Target: rec.Target, Record: rec, Reason: rec.Reason, ReasonSource: rec.Source})
		}
		p.Sessions = append(p.Sessions, sess)
	}
	if err := s.plans.Save(context.Background(), uid, p, plan.Exclusions{}, false); err != nil {
		t.Fatal(err)
	}
	return p
}

// logWorkout stores finished workout n of session 0 on 2026-10-01 with
// the target copies. Each copy gets one set log for each working set,
// at the weight of load, or at the load of the set when load is 0, and
// at 3 reps in reserve.
func logWorkout(t *testing.T, s stores, n int, load domain.Load, targets ...domain.PlannedExercise) string {
	t.Helper()
	return logOverrides(t, s, n, load, nil, targets...)
}

// logOverrides is logWorkout with the record of each override of the
// owner among the targets (D-293).
func logOverrides(t *testing.T, s stores, n int, load domain.Load, overrides []workout.SeenOverride, targets ...domain.PlannedExercise) string {
	t.Helper()
	ctx := context.Background()
	wid := fmt.Sprintf("01920000-0000-7000-a000-%012x", n)
	h := workout.Header{Date: "2026-10-01", Plan: workout.PlanLink{PlanCreatedAt: created}, Targets: targets, Overrides: overrides}
	op := n * 1000
	apply := func(e workout.Entry) {
		t.Helper()
		op++
		e.OpID = fmt.Sprintf("01920000-0000-7000-8000-%012x", op)
		e.At = time.Date(2026, 10, 1, 10, n, op%60, 0, time.UTC)
		e.SchemaVersion = workout.SchemaVersion
		if _, err := s.workouts.Apply(ctx, uid, e); err != nil {
			t.Fatal(err)
		}
	}
	start := h
	apply(workout.Entry{Entity: workout.EntityWorkout, EntityID: wid, Header: &start})
	for i, tg := range targets {
		for j, ws := range tg.Working {
			w := ws.Load
			if load != 0 {
				w = load
			}
			set := domain.SetLog{Kind: domain.SetWorking, Reps: ws.Reps, Weight: w, RIR: 3}
			apply(workout.Entry{Entity: workout.EntitySet, EntityID: fmt.Sprintf("01920000-0000-7000-b%03x-%012x", n, i*10+j), WorkoutID: wid, Set: &set, SetExercise: tg.Exercise})
		}
	}
	h.Finished = true
	apply(workout.Entry{Entity: workout.EntityWorkout, EntityID: wid, Header: &h})
	return wid
}

// reviseWorkout revises the plan after a workout, with the fake
// provider.
func reviseWorkout(t *testing.T, s stores, wid string) {
	t.Helper()
	r := &revise.Reviser{
		AI:    &ai.Client{Provider: &ai.Fake{}, Cap: ai.NewMemoryCap(ai.Caps{User: ai.USD, Project: ai.USD})},
		Plans: s.plans, Workouts: s.workouts, Inventory: s.inventory,
		Now: func() time.Time { return time.Date(2026, 10, 1, 11, 0, 0, 0, time.UTC) },
	}
	res, err := r.Revise(context.Background(), uid, wid)
	if err != nil || !res.Revised {
		t.Fatalf("Revise = %+v, %v, want a revision", res, err)
	}
}

// seedV7 stores the data of a user whose decision records have policy
// version 7. A plan of policy version 8 gets one workout and its
// revision. Then each record and each target of the plan loses the
// limit of D-306 and gets the version 7, as policy version 7 gave
// them. The plan holds the chest press in sessions 0 and 1, and the
// seated row and the leg press in session 0. The workout logs the
// chest press and the seated row at the loads of their targets, so the
// rules add reps, and the leg press keeps the record of the new plan.
func seedV7(t *testing.T, s stores) {
	t.Helper()
	seedInventory(t, s, nil, "chest_press", "seated_row", "leg_press")
	p := newPlan(t, s, nil, []domain.ExerciseID{"chest_press", "seated_row", "leg_press"}, []domain.ExerciseID{"chest_press"})
	wid := logWorkout(t, s, 1, 0, p.Sessions[0].Exercises[0].Target, p.Sessions[0].Exercises[1].Target)
	reviseWorkout(t, s, wid)
	if err := s.plans.Update(context.Background(), uid, created, func(q *plan.Plan) error {
		q.PolicyVersion = 7
		for i := range q.Sessions {
			for j := range q.Sessions[i].Exercises {
				e := &q.Sessions[i].Exercises[j]
				e.Record.PolicyVersion = 7
				e.Record.Target.FollowMax = 0
				e.Target.FollowMax = 0
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// checkV7Report checks the report of the replay of seedV7 under the
// current policy version: the three records of the revision change
// their limit under the rule follow.first-set, and the record of the
// new plan and both target copies stay the same.
func checkV7Report(t *testing.T, rep Report) {
	t.Helper()
	want := Report{
		PolicyVersion: policy.Version, Users: 1, UsersWithPlan: 1,
		Records: RecordReport{
			Total: 4, ByVersion: map[string]int{"7": 4}, ByOrigin: map[string]int{"plan": 1, "revision": 3},
			SameInputHash: 4, Unchanged: 1, Changed: 3,
			ChangedByRule: map[string]int{"follow.first-set": 3}, ChangedFields: map[string]int{"follow_max": 3},
		},
		Copies: CopyReport{Total: 2, SameAsRules: 2, ViolationsByRule: map[string]int{}},
	}
	if fmt.Sprintf("%+v", rep) != fmt.Sprintf("%+v", want) {
		t.Fatalf("report:\n got %+v\nwant %+v", rep, want)
	}
}
