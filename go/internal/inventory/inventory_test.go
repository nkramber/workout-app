package inventory

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

var catalog = domain.DefaultCatalog()

func lb(n int64) domain.Load { return domain.Pounds(n) }

func stack(lo, hi, step int64) []domain.Load {
	var out []domain.Load
	for w := lo; w <= hi; w += step {
		out = append(out, lb(w))
	}
	return out
}

func legPress(estimate domain.Load) Machine {
	m := Machine{Entry: domain.InventoryEntry{Machine: "leg_press", Weights: stack(20, 200, 10)}}
	if estimate > 0 {
		m.Estimates = map[domain.ExerciseID]domain.Load{"leg_press": estimate}
	}
	return m
}

func dumbbells() Machine {
	return Machine{Entry: domain.InventoryEntry{Machine: "dumbbells", Dumbbells: &domain.DumbbellSet{Lightest: lb(5), Heaviest: lb(50), Step: lb(5)}}}
}

func mustSave(t *testing.T, inv Inventory, m Machine) Inventory {
	t.Helper()
	out, err := inv.SaveMachine(catalog, m)
	if err != nil {
		t.Fatalf("SaveMachine(%s): %v", m.Entry.Machine, err)
	}
	return out
}

func mustConfirm(t *testing.T, inv Inventory, e domain.InventoryEntry) Inventory {
	t.Helper()
	out, err := inv.ConfirmMachine(catalog, e)
	if err != nil {
		t.Fatalf("ConfirmMachine(%s): %v", e.Machine, err)
	}
	return out
}

func stateOf(t *testing.T, inv Inventory, id domain.MachineID) State {
	t.Helper()
	m, ok := inv.Machine(id)
	if !ok {
		t.Fatalf("machine %s is not in the inventory", id)
	}
	return m.State
}

func TestMachineCheckRefuses(t *testing.T) {
	cases := []struct {
		name string
		m    Machine
		want string
	}{
		{"unknown machine", Machine{Entry: domain.InventoryEntry{Machine: "barbell"}, State: Draft}, "not in the catalog"},
		{"weight above the bound", Machine{Entry: domain.InventoryEntry{Machine: "leg_press", Weights: []domain.Load{lb(500), MaxWeight + domain.Tenth}}, State: Draft}, "heaviest weight 1000.1 lb"},
		{"too many weights", Machine{Entry: domain.InventoryEntry{Machine: "leg_press", Weights: stack(1, MaxWeights+1, 1)}, State: Draft}, "201 weights"},
		{"unknown state", Machine{Entry: legPress(0).Entry, State: "approved"}, "unknown state"},
		{"estimate for another machine", Machine{Entry: legPress(0).Entry, State: Draft, Estimates: map[domain.ExerciseID]domain.Load{"chest_press": lb(50)}}, `exercise "chest_press"`},
		{"estimate for an unknown exercise", Machine{Entry: legPress(0).Entry, State: Draft, Estimates: map[domain.ExerciseID]domain.Load{"squat": lb(50)}}, `exercise "squat"`},
		{"estimate of 0", Machine{Entry: legPress(0).Entry, State: Draft, Estimates: map[domain.ExerciseID]domain.Load{"leg_press": 0}}, "want more than 0"},
		{"estimate above the heaviest weight", legPressState(lb(210), Draft), "210 lb is outside"},
		{"estimate below the lightest weight", legPressState(lb(10), Draft), "10 lb is outside"},
		{"estimate on a cardio machine", Machine{Entry: domain.InventoryEntry{Machine: "treadmill"}, State: Draft, Estimates: map[domain.ExerciseID]domain.Load{"treadmill": lb(1)}}, "outside the weights"},
		{"dumbbell estimate above the set", Machine{Entry: dumbbells().Entry, State: Draft, Estimates: map[domain.ExerciseID]domain.Load{"db_flat_bench_press": lb(55)}}, "outside the weights"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.m.Check(catalog)
			if !errors.Is(err, domain.ErrInvalid) || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Check = %v, want an ErrInvalid with %q", err, c.want)
			}
		})
	}
}

func legPressState(estimate domain.Load, s State) Machine {
	m := legPress(estimate)
	m.State = s
	return m
}

func TestMachineCheckAccepts(t *testing.T) {
	for _, m := range []Machine{
		legPressState(lb(20), Draft),
		legPressState(lb(200), Confirmed),
		legPressState(lb(125), Draft),
		{Entry: domain.InventoryEntry{Machine: "leg_press", Weights: []domain.Load{MaxWeight}}, State: Draft},
		{Entry: domain.InventoryEntry{Machine: "leg_press", Weights: stack(1, MaxWeights, 1)}, State: Draft},
		{Entry: dumbbells().Entry, State: Draft, Estimates: map[domain.ExerciseID]domain.Load{"db_flat_bench_press": lb(30), "db_goblet_squat": lb(50)}},
		{Entry: domain.InventoryEntry{Machine: "treadmill"}, State: Confirmed},
	} {
		if err := m.Check(catalog); err != nil {
			t.Errorf("Check(%s) = %v", m.Entry.Machine, err)
		}
	}
}

func TestNoteCheck(t *testing.T) {
	long := strings.Repeat("é", MaxNoteRunes)
	if err := (Note{ID: "n1", Text: long}).Check(); err != nil {
		t.Fatalf("a note of %d characters: %v", MaxNoteRunes, err)
	}
	for name, n := range map[string]Note{
		"empty id":        {Text: "Hack squat"},
		"empty text":      {ID: "n1"},
		"long text":       {ID: "n1", Text: long + "x"},
		"space at an end": {ID: "n1", Text: " Hack squat"},
		"not UTF-8":       {ID: "n1", Text: "\xff"},
	} {
		if err := n.Check(); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: Check = %v, want ErrInvalid", name, err)
		}
	}
}

// TestCheckErrorHoldsNoNoteText keeps the text of a note out of each
// error, so an error can go into a log (D-80).
func TestCheckErrorHoldsNoNoteText(t *testing.T) {
	secret := "my private gym note"
	for _, n := range []Note{{ID: "n1", Text: " " + secret}, {ID: "n1", Text: secret + strings.Repeat("x", MaxNoteRunes)}} {
		err := n.Check()
		if err == nil || strings.Contains(err.Error(), "private") {
			t.Fatalf("Check = %v, want an error with no note text", err)
		}
	}
	inv := Inventory{Notes: []Note{{ID: "n1", Text: secret}}}
	_, _, err := inv.SaveNote(catalog, "n1", strings.Repeat(secret, 20), "")
	if err == nil || strings.Contains(err.Error(), "private") {
		t.Fatalf("SaveNote = %v, want an error with no note text", err)
	}
}

func TestInventoryCheck(t *testing.T) {
	two := Inventory{Machines: []Machine{legPressState(0, Draft), legPressState(0, Confirmed)}}
	if err := two.Check(catalog); err == nil || !strings.Contains(err.Error(), "two entries") {
		t.Fatalf("two entries of one machine: Check = %v", err)
	}
	notes := Inventory{Notes: []Note{{ID: "n1", Text: "a"}, {ID: "n1", Text: "b"}}}
	if err := notes.Check(catalog); err == nil || !strings.Contains(err.Error(), "two notes") {
		t.Fatalf("two notes with one id: Check = %v", err)
	}
}

func TestSaveMachineStates(t *testing.T) {
	var inv Inventory
	inv = mustSave(t, inv, legPress(lb(100)))
	if s := stateOf(t, inv, "leg_press"); s != Draft {
		t.Fatalf("a new machine is %s, want draft", s)
	}

	// The state of the request is not read.
	m := legPress(lb(100))
	m.State = Confirmed
	if s := stateOf(t, mustSave(t, inv, m), "leg_press"); s != Draft {
		t.Fatalf("a save with the state confirmed gave %s, want draft", s)
	}

	inv = mustConfirm(t, inv, legPress(0).Entry)
	if s := stateOf(t, inv, "leg_press"); s != Confirmed {
		t.Fatalf("after the confirmation the machine is %s", s)
	}

	// A change of the estimates alone keeps the confirmation (D-200).
	inv = mustSave(t, inv, legPress(lb(120)))
	if s := stateOf(t, inv, "leg_press"); s != Confirmed {
		t.Fatalf("after a change of the estimate the machine is %s, want confirmed", s)
	}
	if got, _ := inv.Machine("leg_press"); got.Estimates["leg_press"] != lb(120) {
		t.Fatalf("estimate = %s, want 120 lb", got.Estimates["leg_press"])
	}

	// A change of the weights makes a draft again (D-193).
	changed := legPress(lb(120))
	changed.Entry.Weights = append(changed.Entry.Weights, lb(210))
	inv = mustSave(t, inv, changed)
	if s := stateOf(t, inv, "leg_press"); s != Draft {
		t.Fatalf("after a change of the weights the machine is %s, want draft", s)
	}
	if len(inv.Machines) != 1 {
		t.Fatalf("the inventory holds %d entries, want one entry for the machine", len(inv.Machines))
	}

	// A cardio machine has no weights, so the save confirms it (D-246).
	treadmill := Machine{Entry: domain.InventoryEntry{Machine: "treadmill"}}
	inv = mustSave(t, inv, treadmill)
	if s := stateOf(t, inv, "treadmill"); s != Confirmed {
		t.Fatalf("a saved cardio machine is %s, want confirmed", s)
	}
	if s := stateOf(t, mustSave(t, inv, treadmill), "treadmill"); s != Confirmed {
		t.Fatalf("a second save of a cardio machine gave %s, want confirmed", s)
	}

	// A change of the dumbbell set makes a draft again.
	inv = mustConfirm(t, mustSave(t, inv, dumbbells()), dumbbells().Entry)
	more := dumbbells()
	more.Entry.Dumbbells.Heaviest = lb(60)
	if s := stateOf(t, mustSave(t, inv, more), "dumbbells"); s != Draft {
		t.Fatalf("after a change of the dumbbell set the machine is %s, want draft", s)
	}
}

func TestSaveMachineRefusesAndKeepsTheInventory(t *testing.T) {
	inv := mustSave(t, Inventory{}, legPress(lb(100)))
	before := inv.clone()
	if _, err := inv.SaveMachine(catalog, legPress(lb(300))); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("an estimate above the weights: %v, want ErrInvalid", err)
	}
	// A change of the weights that moves an estimate out of the range
	// is refused (D-198).
	low := legPress(lb(150))
	low.Entry.Weights = stack(20, 100, 10)
	if _, err := inv.SaveMachine(catalog, low); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("weights below the estimate: %v, want ErrInvalid", err)
	}
	if !reflect.DeepEqual(inv, before) {
		t.Fatalf("a refused change changed the inventory")
	}
}

func TestSaveMachineCatalogOrder(t *testing.T) {
	var inv Inventory
	for _, m := range []Machine{{Entry: domain.InventoryEntry{Machine: "treadmill"}}, dumbbells(), legPress(0)} {
		inv = mustSave(t, inv, m)
	}
	var got []domain.MachineID
	for _, m := range inv.Machines {
		got = append(got, m.Entry.Machine)
	}
	if want := []domain.MachineID{"leg_press", "dumbbells", "treadmill"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestConfirmMachine(t *testing.T) {
	inv := mustSave(t, Inventory{}, legPress(0))
	if _, err := inv.ConfirmMachine(catalog, domain.InventoryEntry{Machine: "chest_press"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("an unknown machine: %v, want ErrNotFound", err)
	}
	other := legPress(0).Entry
	other.Weights = stack(20, 190, 10)
	if _, err := inv.ConfirmMachine(catalog, other); !errors.Is(err, ErrWeightsChanged) {
		t.Fatalf("other weights: %v, want ErrWeightsChanged", err)
	}
	if s := stateOf(t, inv, "leg_press"); s != Draft {
		t.Fatalf("after a refused confirmation the machine is %s", s)
	}
	set := domain.InventoryEntry{Machine: "leg_press", Dumbbells: &domain.DumbbellSet{Lightest: lb(5), Heaviest: lb(50), Step: lb(5)}}
	if _, err := inv.ConfirmMachine(catalog, set); !errors.Is(err, ErrWeightsChanged) {
		t.Fatalf("a dumbbell set for a stack: %v, want ErrWeightsChanged", err)
	}
	cardio := mustSave(t, Inventory{}, Machine{Entry: domain.InventoryEntry{Machine: "treadmill"}})
	if s := stateOf(t, mustConfirm(t, cardio, domain.InventoryEntry{Machine: "treadmill"}), "treadmill"); s != Confirmed {
		t.Fatalf("a cardio machine is %s after the confirmation", s)
	}
}

func TestRemove(t *testing.T) {
	inv := mustSave(t, mustSave(t, Inventory{}, legPress(0)), dumbbells())
	inv, _, err := inv.SaveNote(catalog, "", "Hack squat", "n1")
	if err != nil {
		t.Fatal(err)
	}
	inv = inv.RemoveMachine("leg_press").RemoveMachine("leg_press").RemoveNote("n1").RemoveNote("n9")
	if len(inv.Machines) != 1 || inv.Machines[0].Entry.Machine != "dumbbells" || len(inv.Notes) != 0 {
		t.Fatalf("after the removals: %+v", inv)
	}
}

func TestSaveNote(t *testing.T) {
	inv, id, err := Inventory{}.SaveNote(catalog, "", "  Hack squat  ", "n1")
	if err != nil || id != "n1" {
		t.Fatalf("SaveNote = %q, %v", id, err)
	}
	if inv.Notes[0].Text != "Hack squat" {
		t.Fatalf("text = %q, want the trimmed text", inv.Notes[0].Text)
	}
	inv, id, err = inv.SaveNote(catalog, "n1", "Belt squat", "n2")
	if err != nil || id != "n1" || len(inv.Notes) != 1 || inv.Notes[0].Text != "Belt squat" {
		t.Fatalf("change of a note: %+v, %q, %v", inv.Notes, id, err)
	}
	if _, _, err := inv.SaveNote(catalog, "n9", "x", "n3"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown note: %v, want ErrNotFound", err)
	}
	if _, _, err := inv.SaveNote(catalog, "", "   ", "n3"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("blank note: %v, want ErrInvalid", err)
	}
	for i := len(inv.Notes); i < MaxNotes; i++ {
		if inv, _, err = inv.SaveNote(catalog, "", "x", string(rune('a'+i%26))+strings.Repeat("z", i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := inv.SaveNote(catalog, "", "x", "last"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("note %d: %v, want ErrInvalid", MaxNotes+1, err)
	}
}

// TestChangesShareNoMemory keeps each change free of side effects, so a
// store can run it again in a retry.
func TestChangesShareNoMemory(t *testing.T) {
	inv := mustSave(t, Inventory{}, legPress(lb(100)))
	before := inv.clone()
	out := mustConfirm(t, inv, legPress(0).Entry)
	out.Machines[0].Entry.Weights[0] = lb(1)
	out.Machines[0].Estimates["leg_press"] = lb(1)
	if !reflect.DeepEqual(inv, before) {
		t.Fatalf("a change of the result changed the input")
	}
	got, _ := inv.Machine("leg_press")
	got.Entry.Weights[0] = lb(2)
	if !reflect.DeepEqual(inv, before) {
		t.Fatalf("a change of the result of Machine changed the inventory")
	}
}

func TestForPlan(t *testing.T) {
	var inv Inventory
	inv = mustSave(t, inv, legPress(lb(100)))
	inv = mustSave(t, inv, dumbbells())
	inv = mustSave(t, inv, Machine{Entry: domain.InventoryEntry{Machine: "chest_press", Weights: stack(10, 150, 10)}, Estimates: map[domain.ExerciseID]domain.Load{"chest_press": lb(60)}})
	inv = mustConfirm(t, inv, legPress(0).Entry)
	inv = mustConfirm(t, inv, dumbbells().Entry)
	inv, _, err := inv.SaveNote(catalog, "", "Hack squat", "n1")
	if err != nil {
		t.Fatal(err)
	}

	got := ForPlan(inv)
	var ids []domain.MachineID
	for _, e := range got.Inventory.Entries {
		ids = append(ids, e.Machine)
	}
	if want := []domain.MachineID{"leg_press", "dumbbells"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("ForPlan machines = %v, want the confirmed machines %v alone", ids, want)
	}
	if want := map[domain.ExerciseID]domain.Load{"leg_press": lb(100)}; !reflect.DeepEqual(got.Estimates, want) {
		t.Fatalf("ForPlan estimates = %v, want %v, with no estimate of a draft", got.Estimates, want)
	}
	if err := got.Inventory.Check(catalog); err != nil {
		t.Fatalf("the plan inventory fails the domain check: %v", err)
	}

	// A change of the weights of a confirmed machine takes it out of the
	// plan input (D-193).
	changed := legPress(lb(100))
	changed.Entry.Weights = stack(20, 300, 10)
	inv = mustSave(t, inv, changed)
	after := ForPlan(inv)
	if len(after.Inventory.Entries) != 1 || after.Inventory.Entries[0].Machine != "dumbbells" || len(after.Estimates) != 0 {
		t.Fatalf("ForPlan after the change = %+v, want the dumbbells alone", after)
	}

	after.Inventory.Entries[0].Dumbbells.Heaviest = lb(99)
	if m, _ := inv.Machine("dumbbells"); m.Entry.Dumbbells.Heaviest != lb(50) {
		t.Fatalf("a change of the plan input changed the inventory")
	}
}
