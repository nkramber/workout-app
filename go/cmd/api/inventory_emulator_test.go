//go:build emulator

// The acceptance story of work area 4.2 over the Firebase emulators. The
// test saves, confirms, changes, and removes machines and notes through
// the API. Then it reads the stored document, and the plan input of
// inventory.ForPlan.
package main

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
)

func signed[T any](token string, msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+token)
	return req
}

func TestInventoryAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	uid, token := signUp(t, authHost, fmt.Sprintf("inventory-%d@example.test", suffix))
	otherUID, otherToken := signUp(t, authHost, fmt.Sprintf("inventory-other-%d@example.test", suffix))
	_, outsideToken := signUp(t, authHost, fmt.Sprintf("inventory-outside-%d@example.test", suffix))

	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	for _, id := range []string{uid, otherUID} {
		if _, err := fs.Collection(allowlist.Collection).Doc(id).Set(ctx, map[string]any{}); err != nil {
			t.Fatal(err)
		}
	}

	client := workoutappv1connect.NewInventoryServiceClient(http.DefaultClient, startAPI(t))
	machineState := func(inv *workoutappv1.Inventory, id string) workoutappv1.MachineState {
		t.Helper()
		for _, m := range inv.GetMachines() {
			if m.GetMachineId() == id {
				return m.GetState()
			}
		}
		t.Fatalf("machine %s is not in the inventory", id)
		return 0
	}
	const (
		draft     = workoutappv1.MachineState_MACHINE_STATE_DRAFT
		confirmed = workoutappv1.MachineState_MACHINE_STATE_CONFIRMED
	)

	// The catalog comes from the server.
	cat, err := client.GetCatalog(ctx, signed(token, &workoutappv1.GetCatalogRequest{}))
	if err != nil || len(cat.Msg.GetMachines()) != len(domain.DefaultCatalog().Machines) {
		t.Fatalf("GetCatalog = %v, %v", cat, err)
	}

	// Save three machines and a note. Each new machine with weights is a
	// draft. A cardio machine has no weights, so its save confirms it
	// (D-246).
	legPress := &workoutappv1.SaveMachineRequest{
		MachineId:      "leg_press",
		WeightsTenthLb: []int32{200, 300, 400, 500, 600},
		Estimates:      []*workoutappv1.Estimate{{ExerciseId: "leg_press", LoadTenthLb: 400}},
	}
	dumbbells := &workoutappv1.SaveMachineRequest{
		MachineId: "dumbbells",
		Dumbbells: &workoutappv1.DumbbellSet{LightestTenthLb: 50, HeaviestTenthLb: 500, StepTenthLb: 50},
		Estimates: []*workoutappv1.Estimate{{ExerciseId: "db_flat_bench_press", LoadTenthLb: 300}},
	}
	treadmill := &workoutappv1.SaveMachineRequest{MachineId: "treadmill"}
	for _, req := range []*workoutappv1.SaveMachineRequest{legPress, dumbbells, treadmill} {
		res, err := client.SaveMachine(ctx, signed(token, req))
		if err != nil {
			t.Fatalf("SaveMachine(%s): %v", req.GetMachineId(), err)
		}
		want := draft
		if req == treadmill {
			want = confirmed
		}
		if s := machineState(res.Msg.GetInventory(), req.GetMachineId()); s != want {
			t.Fatalf("new machine %s is %v, want %v", req.GetMachineId(), s, want)
		}
	}
	note, err := client.SaveNote(ctx, signed(token, &workoutappv1.SaveNoteRequest{Text: "Hack squat"}))
	if err != nil {
		t.Fatal(err)
	}
	noteID := note.Msg.GetNoteId()

	// The server refuses a bad entry.
	bad := &workoutappv1.SaveMachineRequest{MachineId: "leg_press", WeightsTenthLb: []int32{200, 300}, Estimates: legPress.GetEstimates()}
	if _, err := client.SaveMachine(ctx, signed(token, bad)); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("an estimate above the weights = %v, want InvalidArgument", err)
	}
	if _, err := client.SaveMachine(ctx, signed(token, &workoutappv1.SaveMachineRequest{MachineId: "barbell"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a machine outside the catalog = %v, want InvalidArgument", err)
	}

	// Confirm with the weights that the review screen showed (D-201).
	confirm := func(req *workoutappv1.SaveMachineRequest) (*connect.Response[workoutappv1.ConfirmMachineResponse], error) {
		return client.ConfirmMachine(ctx, signed(token, &workoutappv1.ConfirmMachineRequest{
			MachineId: req.GetMachineId(), WeightsTenthLb: req.GetWeightsTenthLb(), Dumbbells: req.GetDumbbells(),
		}))
	}
	if _, err := confirm(bad); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("a confirmation of other weights = %v, want FailedPrecondition", err)
	}
	for _, req := range []*workoutappv1.SaveMachineRequest{legPress, dumbbells} {
		res, err := confirm(req)
		if err != nil || machineState(res.Msg.GetInventory(), req.GetMachineId()) != confirmed {
			t.Fatalf("ConfirmMachine(%s) = %v, %v", req.GetMachineId(), res, err)
		}
	}

	// A second save of a machine replaces its one entry. A change of the
	// estimates alone keeps the confirmation (D-200). A change of the
	// weights makes a draft again (D-193).
	legPress.Estimates[0].LoadTenthLb = 450
	res, err := client.SaveMachine(ctx, signed(token, legPress))
	if err != nil || machineState(res.Msg.GetInventory(), "leg_press") != confirmed || len(res.Msg.GetInventory().GetMachines()) != 3 {
		t.Fatalf("a change of the estimate = %v, %v, want leg_press confirmed and three machines", res, err)
	}
	legPress.WeightsTenthLb = append(legPress.WeightsTenthLb, 700)
	if res, err = client.SaveMachine(ctx, signed(token, legPress)); err != nil || machineState(res.Msg.GetInventory(), "leg_press") != draft {
		t.Fatalf("a change of the weights = %v, %v, want leg_press draft", res, err)
	}

	// Change and remove.
	if _, err := client.SaveNote(ctx, signed(token, &workoutappv1.SaveNoteRequest{Id: noteID, Text: "Belt squat"})); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := client.RemoveMachine(ctx, signed(token, &workoutappv1.RemoveMachineRequest{MachineId: "treadmill"})); err != nil {
			t.Fatal(err)
		}
	}

	got, err := client.GetInventory(ctx, signed(token, &workoutappv1.GetInventoryRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	inv := got.Msg.GetInventory()
	if len(inv.GetMachines()) != 2 || machineState(inv, "leg_press") != draft || machineState(inv, "dumbbells") != confirmed {
		t.Fatalf("inventory = %v, want leg_press draft and dumbbells confirmed", inv)
	}
	if len(inv.GetNotes()) != 1 || inv.GetNotes()[0].GetText() != "Belt squat" {
		t.Fatalf("notes = %v", inv.GetNotes())
	}

	// A stored machine holds its identity, its weights, its estimates,
	// and its state alone (D-54, D-192, D-193), at the path of D-197.
	snap, err := fs.Collection("users").Doc(uid).Collection("inventory").Doc("active").Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if keys := slices.Sorted(maps.Keys(snap.Data())); !reflect.DeepEqual(keys, []string{"machines", "notes"}) {
		t.Fatalf("document fields = %v", keys)
	}
	allowed := map[string]bool{"machine": true, "weights": true, "dumbbells": true, "estimates": true, "state": true}
	for _, raw := range snap.Data()["machines"].([]any) {
		for k := range raw.(map[string]any) {
			if !allowed[k] {
				t.Fatalf("a stored machine holds the field %q", k)
			}
		}
	}

	// The plan input gives the confirmed machines alone. The draft and
	// the note never reach it (D-49, D-191, D-193).
	stored, err := inventory.FromFirestore(fs).Get(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	plan := inventory.ForPlan(stored)
	if len(plan.Inventory.Entries) != 1 || plan.Inventory.Entries[0].Machine != "dumbbells" {
		t.Fatalf("plan input = %+v, want the dumbbells alone", plan.Inventory)
	}
	if want := map[domain.ExerciseID]domain.Load{"db_flat_bench_press": 300}; !reflect.DeepEqual(plan.Estimates, want) {
		t.Fatalf("plan estimates = %v, want %v", plan.Estimates, want)
	}

	// The confirmation of the new weights puts the machine back.
	if _, err := confirm(legPress); err != nil {
		t.Fatal(err)
	}
	stored, _ = inventory.FromFirestore(fs).Get(ctx, uid)
	if plan = inventory.ForPlan(stored); len(plan.Inventory.Entries) != 2 || plan.Estimates["leg_press"] != 450 {
		t.Fatalf("plan input after the confirmation = %+v", plan)
	}

	// The removal of the note leaves no note.
	rm, err := client.RemoveNote(ctx, signed(token, &workoutappv1.RemoveNoteRequest{Id: noteID}))
	if err != nil || len(rm.Msg.GetInventory().GetNotes()) != 0 {
		t.Fatalf("RemoveNote = %v, %v", rm, err)
	}

	// Each uid reads its own inventory alone, and a uid off the
	// allowlist reads nothing.
	other, err := client.GetInventory(ctx, signed(otherToken, &workoutappv1.GetInventoryRequest{}))
	if err != nil || len(other.Msg.GetInventory().GetMachines()) != 0 {
		t.Fatalf("GetInventory of another uid = %v, %v, want an empty inventory", other, err)
	}
	if _, err := client.SaveMachine(ctx, signed(outsideToken, treadmill)); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("SaveMachine off the allowlist = %v, want PermissionDenied", err)
	}
}
