//go:build emulator

// The deletion of all data over the emulators (D-314, D-315).
package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/usersvc"
)

// TestDeleteHistoryDuringPlanRequest: a plan request waits in its call of
// Luna while the owner deletes the history. When the call ends, the save
// refuses the plan with ABORTED, so no plan comes back after the
// deletion (D-315). A request after the deletion saves its plan.
func TestDeleteHistoryDuringPlanRequest(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	uid, token := signUp(t, authHost, fmt.Sprintf("history-%d@example.test", time.Now().UnixNano()))
	if _, err := fs.Collection(allowlist.Collection).Doc(uid).Set(ctx, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	// The first call of Luna waits until the test releases it.
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	fake := &ai.Fake{Reply: func(c ai.Call) (ai.Reply, error) {
		once.Do(func() { close(entered) })
		<-release
		return ai.EchoReply(c)
	}}
	base := startAPIWith(t, fake)
	profiles := workoutappv1connect.NewProfileServiceClient(http.DefaultClient, base)
	plans := workoutappv1connect.NewPlanServiceClient(http.DefaultClient, base)
	users := workoutappv1connect.NewUserServiceClient(http.DefaultClient, base)
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

	done := make(chan error, 1)
	go func() {
		s, err := plans.RequestPlan(ctx, signed(token, &workoutappv1.RequestPlanRequest{Today: today}))
		if err != nil {
			done <- err
			return
		}
		_, _, err = drainPlan(s)
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(20 * time.Second):
		t.Fatal("the plan request did not reach its call of Luna")
	}
	del, err := users.DeleteHistory(ctx, signed(token, &workoutappv1.DeleteHistoryRequest{Confirmation: usersvc.Confirmation}))
	if err != nil || del.Msg.GetHistoryGeneration() != 1 {
		t.Fatalf("DeleteHistory = %v, %v, want generation 1", del, err)
	}
	close(release)
	if err := <-done; connect.CodeOf(err) != connect.CodeAborted {
		t.Fatalf("the plan request after the deletion = %v, want ABORTED", err)
	}
	got, err := plans.GetPlan(ctx, signed(token, &workoutappv1.GetPlanRequest{}))
	if err != nil || got.Msg.GetPlan() != nil {
		t.Fatalf("GetPlan after the deletion = %v, %v, want no plan", got, err)
	}

	// A request after the deletion reads the new generation, and saves.
	s, err := plans.RequestPlan(ctx, signed(token, &workoutappv1.RequestPlanRequest{Today: today}))
	if err != nil {
		t.Fatal(err)
	}
	if _, p, err := drainPlan(s); err != nil || p == nil {
		t.Fatalf("RequestPlan after the deletion = %v, %v, want a plan", p, err)
	}
	me, err := users.GetMe(ctx, signed(token, &workoutappv1.GetMeRequest{}))
	if err != nil || me.Msg.GetHistoryGeneration() != 1 {
		t.Fatalf("GetMe = %v, %v, want generation 1", me, err)
	}
}
