//go:build emulator

// The inventory entries of the outbox over the Firebase emulators (D-272).
// One batch mixes the inventory entries and a workout entry. The server
// applies them in the order of the batch, and GetInventory then gives the
// result. The same batch again changes nothing, and a confirmation with
// other weights gets "failed_precondition" (D-273).
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
	"github.com/nkramber/workout-app/go/internal/workout"
	"github.com/nkramber/workout-app/go/internal/workoutsvc"
)

func TestOutboxInventory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	uid, token := signUp(t, authHost, fmt.Sprintf("outbox-inventory-%d@example.test", time.Now().UnixNano()))
	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	if _, err := fs.Collection(allowlist.Collection).Doc(uid).Set(ctx, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	api := startAPI(t)
	workouts := workoutappv1connect.NewWorkoutServiceClient(http.DefaultClient, api)
	inventories := workoutappv1connect.NewInventoryServiceClient(http.DefaultClient, api)

	n := 0
	entry := func(entity, id string) *workoutappv1.OutboxEntry {
		n++
		return &workoutappv1.OutboxEntry{
			OpId: fmt.Sprintf("01920000-0000-7000-8000-%012x", n), Entity: entity, EntityId: id,
			At: time.Date(2026, 10, 3, 10, n, 0, 0, time.UTC).Format(time.RFC3339), SchemaVersion: workout.SchemaVersion,
		}
	}
	const noteID = "01920000-0000-7000-a000-000000000001"
	save := entry(workoutsvc.EntityMachine, "leg_press")
	save.Payload = &workoutappv1.OutboxEntry_SaveMachine{SaveMachine: &workoutappv1.MachineSave{WeightsTenthLb: []int32{100, 200, 300}}}
	start := entry(workout.EntityWorkout, "01920000-0000-7000-8000-0000000000a1")
	start.Payload = &workoutappv1.OutboxEntry_Workout{Workout: &workoutappv1.WorkoutHeader{
		Date: "2026-10-03", Plan: &workoutappv1.PlanLink{PlanCreatedAt: "2026-10-03T06:11:03.61545Z"},
	}}
	confirm := entry(workoutsvc.EntityMachine, "leg_press")
	confirm.Payload = &workoutappv1.OutboxEntry_ConfirmMachine{ConfirmMachine: &workoutappv1.MachineConfirm{WeightsTenthLb: []int32{100, 200, 300}}}
	note := entry(workoutsvc.EntityNote, noteID)
	note.Payload = &workoutappv1.OutboxEntry_SaveNote{SaveNote: &workoutappv1.NoteSave{Text: "rope handle"}}
	stale := entry(workoutsvc.EntityMachine, "leg_press")
	stale.Payload = &workoutappv1.OutboxEntry_ConfirmMachine{ConfirmMachine: &workoutappv1.MachineConfirm{WeightsTenthLb: []int32{100, 200}}}
	batch := []*workoutappv1.OutboxEntry{save, start, confirm, note, stale}

	sync := func() []*workoutappv1.EntryResult {
		t.Helper()
		res, err := workouts.SyncOutbox(ctx, signed(token, &workoutappv1.SyncOutboxRequest{Entries: batch}))
		if err != nil || len(res.Msg.GetResults()) != len(batch) {
			t.Fatalf("SyncOutbox = %v, %v", res, err)
		}
		return res.Msg.GetResults()
	}
	read := func() *workoutappv1.Inventory {
		t.Helper()
		res, err := inventories.GetInventory(ctx, signed(token, &workoutappv1.GetInventoryRequest{}))
		if err != nil {
			t.Fatalf("GetInventory = %v", err)
		}
		return res.Msg.GetInventory()
	}
	check := func(results []*workoutappv1.EntryResult) {
		t.Helper()
		for i, r := range results[:4] {
			if r.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
				t.Fatalf("result %d = %v", i, r)
			}
		}
		if r := results[4]; r.GetStatus() != workoutappv1.EntryResult_STATUS_REFUSED || r.GetCode() != workoutsvc.CodeFailedPrecondition {
			t.Fatalf("the confirmation with other weights = %v", r)
		}
		inv := read()
		if m := inv.GetMachines(); len(m) != 1 || m[0].GetState() != workoutappv1.MachineState_MACHINE_STATE_CONFIRMED {
			t.Fatalf("machines %v", m)
		}
		if notes := inv.GetNotes(); len(notes) != 1 || notes[0].GetId() != noteID || notes[0].GetText() != "rope handle" {
			t.Fatalf("notes %v", notes)
		}
	}
	check(sync())
	check(sync())
	snaps, err := fs.Collection("users").Doc(uid).Collection("ops").Documents(ctx).GetAll()
	if err != nil || len(snaps) != 4 {
		t.Fatalf("%d op documents, %v, want 4: the refused entry stores none", len(snaps), err)
	}
}
