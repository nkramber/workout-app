//go:build emulator

// The acceptance story of PR-37 over the Firebase emulators, with the
// fake provider of Luna (D-24). A new plan has no calibration set, and
// the first set of each new exercise is the calibration (D-297). The
// owner changes the weight during the first set, and the revision reads
// that weight as the load of the session (D-299). From the second
// session the normal rules apply, and a later plan gives an exercise
// with history no calibration (D-301).
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
)

// planExercise gives the first exercise of the plan with the id.
func planExercise(t *testing.T, p *workoutappv1.Plan, id string) *workoutappv1.PlannedExercise {
	t.Helper()
	for _, s := range p.GetSessions() {
		for _, e := range s.GetExercises() {
			if e.GetExerciseId() == id {
				return e
			}
		}
	}
	t.Fatalf("the plan has no %s", id)
	return nil
}

func TestFirstSetCalibrationAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	uid, token := signUp(t, authHost, fmt.Sprintf("calibration-%d@example.test", time.Now().UnixNano()))
	if _, err := fs.Collection(allowlist.Collection).Doc(uid).Set(ctx, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	base := startAPI(t)
	profiles := workoutappv1connect.NewProfileServiceClient(http.DefaultClient, base)
	plans := workoutappv1connect.NewPlanServiceClient(http.DefaultClient, base)
	if _, err := profiles.SaveProfile(ctx, signed(token, &workoutappv1.SaveProfileRequest{Profile: &workoutappv1.Profile{
		Experience: "intermediate", GoalTemplate: "general_fitness", MuscleGroups: []string{"chest", "back"},
		AgeYears: 32, HeightIn: 70, WeightLb: 180, TrainingDays: 3,
	}})); err != nil {
		t.Fatal(err)
	}
	var w []domain.Load
	for lb := int64(10); lb <= 200; lb += 10 {
		w = append(w, domain.Pounds(lb))
	}
	inv := inventory.Inventory{}
	for _, id := range []domain.MachineID{"chest_press", "seated_row"} {
		inv.Machines = append(inv.Machines, inventory.Machine{Entry: domain.InventoryEntry{Machine: id, Weights: w}, Estimates: map[domain.ExerciseID]domain.Load{}, State: inventory.Confirmed})
	}
	if _, err := inventory.FromFirestore(fs).Update(ctx, uid, func(inventory.Inventory) (inventory.Inventory, error) { return inv, nil }); err != nil {
		t.Fatal(err)
	}
	today := time.Now().UTC().Format(domain.DateLayout)
	request := func() *workoutappv1.Plan {
		t.Helper()
		s, err := plans.RequestPlan(ctx, signed(token, &workoutappv1.RequestPlanRequest{Today: today}))
		if err != nil {
			t.Fatal(err)
		}
		_, p, err := drainPlan(s)
		if err != nil || p == nil {
			t.Fatalf("RequestPlan = %v, %v", p, err)
		}
		return p
	}

	// A new plan gives no calibration set. The first set of each new
	// exercise is the calibration, at the lightest weight with no
	// estimate (D-297, D-300).
	p := request()
	for _, s := range p.GetSessions() {
		for _, e := range s.GetExercises() {
			if !e.GetFirstSetCalibration() || len(e.GetCalibrationSets()) != 0 || len(e.GetCalibrationLoads()) != 0 {
				t.Fatalf("%s: first-set calibration %v, %d calibration sets, %d table rows", e.GetExerciseId(), e.GetFirstSetCalibration(), len(e.GetCalibrationSets()), len(e.GetCalibrationLoads()))
			}
		}
	}
	press := planExercise(t, p, "chest_press")
	first := targetOf(press)
	first.FirstSetCalibration = true
	if first.Working[0].Load != domain.Pounds(10) {
		t.Fatalf("chest press %+v: want the lightest weight", first.Working)
	}

	// The owner changes the weight to 30 lb during the first set, and
	// logs each set at 30 lb. The workout sends the target copy with the
	// flag (D-291).
	created, err := time.Parse(time.RFC3339Nano, p.GetCreatedAt())
	if err != nil {
		t.Fatal(err)
	}
	e := &revEnv{t: t, ctx: ctx, fs: fs, uid: uid, token: token, created: created,
		workouts: workoutappv1connect.NewWorkoutServiceClient(http.DefaultClient, base), plans: plans}
	var sets []domain.SetLog
	for _, s := range first.Working {
		sets = append(sets, revSet(s.Reps, 30, 3))
	}
	c := revCase{id: "first_set", exercise: "chest_press"}
	e.session(c, logged{target: first, sets: sets}, today)

	// The revision reads 30 lb as the load of the session, and the next
	// session uses the normal rules: the reps go up at 30 lb, with no
	// calibration (D-299, D-301).
	next := targetOf(planExercise(t, e.planned(), "chest_press"))
	if next.FirstSetCalibration || next.Working[0].Load != domain.Pounds(30) || next.Working[0].Reps != first.Working[0].Reps+2 {
		t.Fatalf("next target %+v: want %d reps at 30 lb with no calibration", next, first.Working[0].Reps+2)
	}

	// A new plan reads the history: the chest press gets no calibration
	// and starts from 30 lb, and the seated row is still new (D-301).
	again := request()
	if x := planExercise(t, again, "chest_press"); x.GetFirstSetCalibration() || x.GetWorkingSets()[0].GetLoadTenthLb() != 300 {
		t.Fatalf("chest press of the new plan %v: want 30 lb with no calibration", x)
	}
	if x := planExercise(t, again, "seated_row"); !x.GetFirstSetCalibration() {
		t.Fatalf("seated row of the new plan %v: want the first-set calibration", x)
	}
}
