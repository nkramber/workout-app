package workoutsvc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/workout"
)

const noteID = "01920000-0000-7000-a000-000000000001"

func invEntry(n int, entity, id string) *workoutappv1.OutboxEntry {
	return &workoutappv1.OutboxEntry{OpId: opID(n), Entity: entity, EntityId: id, At: at(n), SchemaVersion: workout.SchemaVersion}
}

func saveMachine(n int, id string, weights ...int32) *workoutappv1.OutboxEntry {
	e := invEntry(n, EntityMachine, id)
	e.Payload = &workoutappv1.OutboxEntry_SaveMachine{SaveMachine: &workoutappv1.MachineSave{WeightsTenthLb: weights}}
	return e
}

func confirmMachine(n int, id string, weights ...int32) *workoutappv1.OutboxEntry {
	e := invEntry(n, EntityMachine, id)
	e.Payload = &workoutappv1.OutboxEntry_ConfirmMachine{ConfirmMachine: &workoutappv1.MachineConfirm{WeightsTenthLb: weights}}
	return e
}

func removeMachine(n int, id string) *workoutappv1.OutboxEntry {
	e := invEntry(n, EntityMachine, id)
	e.Payload = &workoutappv1.OutboxEntry_RemoveMachine{RemoveMachine: &workoutappv1.MachineRemove{}}
	return e
}

func saveNote(n int, id, text string) *workoutappv1.OutboxEntry {
	e := invEntry(n, EntityNote, id)
	e.Payload = &workoutappv1.OutboxEntry_SaveNote{SaveNote: &workoutappv1.NoteSave{Text: text}}
	return e
}

func removeNote(n int, id string) *workoutappv1.OutboxEntry {
	e := invEntry(n, EntityNote, id)
	e.Payload = &workoutappv1.OutboxEntry_RemoveNote{RemoveNote: &workoutappv1.NoteRemove{}}
	return e
}

func stored(t *testing.T, invs *inventory.Memory) inventory.Inventory {
	t.Helper()
	inv, err := invs.Get(context.Background(), "uid-a")
	if err != nil {
		t.Fatal(err)
	}
	return inv
}

// TestSyncOutboxInventory: the inventory entries apply in the order of the
// batch, mixed with the workout entries (D-272, D-275). A replay of the
// batch applies no entry again: the note stays one note, and the removed
// machine stays removed.
func TestSyncOutboxInventory(t *testing.T) {
	invs := inventory.NewMemory()
	s := New(workout.NewMemory(), invs)
	batch := []*workoutappv1.OutboxEntry{
		saveMachine(1, "leg_press", 100, 200, 300),
		headerEntry(2),
		confirmMachine(3, "leg_press", 100, 200, 300),
		saveNote(4, noteID, "  rope handle  "),
		saveMachine(5, "chest_press", 100, 200),
		setEntry(6, 1, 10),
		removeMachine(7, "chest_press"),
		saveNote(8, noteID, "rope handle, short"),
	}
	first := sync(t, s, batch...)
	for i, r := range first {
		if r.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED || r.GetOpId() != batch[i].GetOpId() {
			t.Fatalf("result %d = %v", i, r)
		}
	}
	if first[0].GetVersion() != 0 || first[1].GetVersion() != 1 {
		t.Fatalf("versions %d and %d, want 0 for the inventory and 1 for the workout", first[0].GetVersion(), first[1].GetVersion())
	}
	inv := stored(t, invs)
	if len(inv.Machines) != 1 || inv.Machines[0].Entry.Machine != "leg_press" || inv.Machines[0].State != inventory.Confirmed {
		t.Fatalf("machines %+v", inv.Machines)
	}
	if len(inv.Notes) != 1 || inv.Notes[0] != (inventory.Note{ID: noteID, Text: "rope handle, short"}) {
		t.Fatalf("notes %+v", inv.Notes)
	}

	// A new save of the chest press after the batch, then the replay.
	sync(t, s, saveMachine(9, "chest_press", 100, 200))
	again := sync(t, s, batch...)
	for i, r := range again {
		if r.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
			t.Fatalf("replay result %d = %v", i, r)
		}
	}
	inv = stored(t, invs)
	if len(inv.Machines) != 2 || len(inv.Notes) != 1 {
		t.Fatalf("the replay changed the inventory: %+v", inv)
	}
	if w := list(t, s); len(w) != 1 || len(w[0].GetExercises()[0].GetSets()) != 1 {
		t.Fatalf("the replay changed the workouts: %v", w)
	}
}

// TestSyncOutboxConfirmRefused: a confirmation with weights that are not
// the stored weights, and a confirmation of an unknown machine, get
// "failed_precondition". The machine stays a draft, and the next entry
// still applies (D-201, D-273).
func TestSyncOutboxConfirmRefused(t *testing.T) {
	invs := inventory.NewMemory()
	s := New(workout.NewMemory(), invs)
	got := sync(t, s,
		saveMachine(1, "leg_press", 100, 200, 300),
		confirmMachine(2, "leg_press", 100, 200),
		confirmMachine(3, "seated_row", 100),
		saveNote(4, noteID, "rope handle"),
	)
	for _, i := range []int{1, 2} {
		if got[i].GetStatus() != workoutappv1.EntryResult_STATUS_REFUSED || got[i].GetCode() != CodeFailedPrecondition {
			t.Fatalf("result %d = %v", i, got[i])
		}
	}
	if got[3].GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
		t.Fatalf("the note after the refusals = %v", got[3])
	}
	inv := stored(t, invs)
	if len(inv.Machines) != 1 || inv.Machines[0].State != inventory.Draft || len(inv.Notes) != 1 {
		t.Fatalf("inventory %+v", inv)
	}
	// The refused op id stays unapplied, so its replay gets the same check.
	if again := sync(t, s, confirmMachine(2, "leg_press", 100, 200)); again[0].GetCode() != CodeFailedPrecondition {
		t.Fatalf("replay of the refused confirmation = %v", again[0])
	}
}

// TestSyncOutboxInventoryRefuses: each bad inventory entry gets
// "invalid_argument", and changes nothing.
func TestSyncOutboxInventoryRefuses(t *testing.T) {
	cases := map[string]struct {
		entry *workoutappv1.OutboxEntry
		want  string
	}{
		"unknown machine": {saveMachine(1, "jetpack", 100), "not a machine of the catalog"},
		"machine as a note": {func() *workoutappv1.OutboxEntry {
			e := saveMachine(1, "leg_press", 100)
			e.Entity = EntityNote
			return e
		}(), "the payload wants machine"},
		"note id not a UUID": {saveNote(1, "note-1", "rope"), "note id"},
		"note too long":      {saveNote(1, noteID, strings.Repeat("a", 201)), "201 characters"},
		"bad weights":        {saveMachine(1, "leg_press", 300, 100), "leg_press"},
		"base version":       {func() *workoutappv1.OutboxEntry { e := removeNote(1, noteID); e.BaseVersion = 3; return e }(), "base version"},
		"schema version":     {func() *workoutappv1.OutboxEntry { e := removeMachine(1, "leg_press"); e.SchemaVersion = 2; return e }(), "schema version"},
		"bad time":           {func() *workoutappv1.OutboxEntry { e := removeMachine(1, "leg_press"); e.At = "now"; return e }(), "at:"},
		"bad op id":          {func() *workoutappv1.OutboxEntry { e := removeMachine(1, "leg_press"); e.OpId = "op-1"; return e }(), "op id"},
		"two estimates": {func() *workoutappv1.OutboxEntry {
			e := saveMachine(1, "leg_press", 100, 200)
			est := &workoutappv1.Estimate{ExerciseId: "leg_press", LoadTenthLb: 100}
			e.GetSaveMachine().Estimates = []*workoutappv1.Estimate{est, est}
			return e
		}(), "two estimates"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			invs := inventory.NewMemory()
			s := New(workout.NewMemory(), invs)
			r := sync(t, s, c.entry)[0]
			if r.GetStatus() != workoutappv1.EntryResult_STATUS_REFUSED || r.GetCode() != CodeInvalidArgument || !strings.Contains(r.GetMessage(), c.want) {
				t.Fatalf("result = %v, want a refusal that names %q", r, c.want)
			}
			if inv := stored(t, invs); len(inv.Machines) != 0 || len(inv.Notes) != 0 {
				t.Fatalf("the refusal changed the inventory: %+v", inv)
			}
		})
	}
}

// TestSyncOutboxDumbbells: a save of the dumbbells keeps the dumbbell set,
// and the confirmation with the same set applies.
func TestSyncOutboxDumbbells(t *testing.T) {
	invs := inventory.NewMemory()
	s := New(workout.NewMemory(), invs)
	ds := &workoutappv1.DumbbellSet{LightestTenthLb: 50, HeaviestTenthLb: 500, StepTenthLb: 50}
	save := invEntry(1, EntityMachine, "dumbbells")
	save.Payload = &workoutappv1.OutboxEntry_SaveMachine{SaveMachine: &workoutappv1.MachineSave{Dumbbells: ds}}
	confirm := invEntry(2, EntityMachine, "dumbbells")
	confirm.Payload = &workoutappv1.OutboxEntry_ConfirmMachine{ConfirmMachine: &workoutappv1.MachineConfirm{Dumbbells: ds}}
	for i, r := range sync(t, s, save, confirm) {
		if r.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
			t.Fatalf("result %d = %v", i, r)
		}
	}
	m, ok := stored(t, invs).Machine("dumbbells")
	if !ok || m.State != inventory.Confirmed || *m.Entry.Dumbbells != (domain.DumbbellSet{Lightest: 50, Heaviest: 500, Step: 50}) {
		t.Fatalf("dumbbells %+v", m)
	}
}

type brokenInventory struct{ *inventory.Memory }

func (brokenInventory) ApplyOp(context.Context, string, inventory.Op, func(inventory.Inventory) (inventory.Inventory, error)) (bool, error) {
	return false, errors.New("rpc error: projects/p/databases/(default)/documents/users/uid-a/ops")
}

// TestSyncOutboxInventoryStoreError gives the caller a fixed text for a
// store error of an inventory entry, so no path leaves the server.
func TestSyncOutboxInventoryStoreError(t *testing.T) {
	s := New(workout.NewMemory(), brokenInventory{inventory.NewMemory()})
	_, err := s.SyncOutbox(signedIn, connect.NewRequest(&workoutappv1.SyncOutboxRequest{Entries: []*workoutappv1.OutboxEntry{removeNote(1, noteID)}}))
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users") {
		t.Fatalf("SyncOutbox = %v, want Internal with no path", err)
	}
}
