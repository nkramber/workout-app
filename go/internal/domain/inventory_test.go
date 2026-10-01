package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestDumbbellSetCheck(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  DumbbellSet
		ok   bool
	}{
		{"5 to 50 by 5", DumbbellSet{Pounds(5), Pounds(50), Pounds(5)}, true},
		{"one weight", DumbbellSet{Pounds(20), Pounds(20), Pounds(5)}, true},
		{"2.5 lb step", DumbbellSet{Pounds(5), Pounds(15), 25}, true},
		{"lightest 0", DumbbellSet{0, Pounds(50), Pounds(5)}, false},
		{"step 0", DumbbellSet{Pounds(5), Pounds(50), 0}, false},
		{"negative step", DumbbellSet{Pounds(5), Pounds(50), -Pounds(5)}, false},
		{"heaviest below lightest", DumbbellSet{Pounds(50), Pounds(5), Pounds(5)}, false},
		{"step does not divide", DumbbellSet{Pounds(5), Pounds(52), Pounds(5)}, false},
	} {
		err := tc.set.Check()
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestDumbbellSetWeights(t *testing.T) {
	got := DumbbellSet{Pounds(5), Pounds(15), 25}.Weights()
	want := []Load{50, 75, 100, 125, 150}
	if !slices.Equal(got, want) {
		t.Errorf("Weights() = %v, want %v", got, want)
	}
	if got := (DumbbellSet{Pounds(5), Pounds(15), 0}).Weights(); got != nil {
		t.Errorf("Weights() of a bad set = %v, want nil", got)
	}
}

func TestInventoryEntryCheck(t *testing.T) {
	c := DefaultCatalog()
	stack := []Load{Pounds(10), Pounds(20), 325, Pounds(40)}
	set := &DumbbellSet{Pounds(5), Pounds(50), Pounds(5)}
	for _, tc := range []struct {
		name  string
		entry InventoryEntry
		ok    bool
	}{
		{"machine with a 12.5 lb step", InventoryEntry{Machine: "chest_press", Weights: stack}, true},
		{"cable station", InventoryEntry{Machine: "cable_station", Weights: stack}, true},
		{"dumbbells", InventoryEntry{Machine: "dumbbells", Dumbbells: set}, true},
		{"cardio", InventoryEntry{Machine: "treadmill"}, true},
		{"unknown machine", InventoryEntry{Machine: "smith_machine", Weights: stack}, false},
		{"machine with no weights", InventoryEntry{Machine: "chest_press"}, false},
		{"machine with a dumbbell set", InventoryEntry{Machine: "chest_press", Weights: stack, Dumbbells: set}, false},
		{"weight 0", InventoryEntry{Machine: "leg_press", Weights: []Load{0, Pounds(10)}}, false},
		{"weights not ascending", InventoryEntry{Machine: "leg_press", Weights: []Load{Pounds(20), Pounds(10)}}, false},
		{"weight two times", InventoryEntry{Machine: "leg_press", Weights: []Load{Pounds(10), Pounds(10)}}, false},
		{"dumbbells with no set", InventoryEntry{Machine: "dumbbells"}, false},
		{"dumbbells with a list", InventoryEntry{Machine: "dumbbells", Weights: stack, Dumbbells: set}, false},
		{"dumbbells with a bad set", InventoryEntry{Machine: "dumbbells", Dumbbells: &DumbbellSet{Pounds(5), Pounds(52), Pounds(5)}}, false},
		{"cardio with weights", InventoryEntry{Machine: "rowing_machine", Weights: stack}, false},
		{"cardio with a dumbbell set", InventoryEntry{Machine: "rowing_machine", Dumbbells: set}, false},
	} {
		err := tc.entry.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
}

func TestInventoryEntryAvailable(t *testing.T) {
	stack := []Load{Pounds(10), 225}
	e := InventoryEntry{Machine: "leg_press", Weights: stack}
	got := e.Available()
	if !slices.Equal(got, stack) {
		t.Fatalf("Available() = %v, want %v", got, stack)
	}
	got[0] = 0
	if e.Weights[0] != Pounds(10) {
		t.Fatal("a change of Available() changed the entry")
	}
	d := InventoryEntry{Machine: "dumbbells", Dumbbells: &DumbbellSet{Pounds(10), Pounds(20), Pounds(5)}}
	if got := d.Available(); !slices.Equal(got, []Load{Pounds(10), Pounds(15), Pounds(20)}) {
		t.Errorf("Available() of dumbbells = %v", got)
	}
	if got := (InventoryEntry{Machine: "treadmill"}).Available(); len(got) != 0 {
		t.Errorf("Available() of cardio = %v, want none", got)
	}
}

func TestInventoryCheck(t *testing.T) {
	c := DefaultCatalog()
	press := InventoryEntry{Machine: "chest_press", Weights: []Load{Pounds(10), Pounds(20)}}
	bike := InventoryEntry{Machine: "upright_bike"}
	for _, tc := range []struct {
		name string
		inv  Inventory
		ok   bool
	}{
		{"empty", Inventory{}, true},
		{"two machines", Inventory{Entries: []InventoryEntry{press, bike}}, true},
		{"one machine two times", Inventory{Entries: []InventoryEntry{press, bike, press}}, false},
		{"a bad entry", Inventory{Entries: []InventoryEntry{press, {Machine: "chest_press"}}}, false},
	} {
		err := tc.inv.Check(c)
		if (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("%s: Check() = %v, want ok %v", tc.name, err, tc.ok)
		}
	}
	inv := Inventory{Entries: []InventoryEntry{press, bike}}
	if e, ok := inv.Entry("upright_bike"); !ok || e.Machine != "upright_bike" {
		t.Errorf("Entry(upright_bike) = %+v, %v", e, ok)
	}
	if _, ok := inv.Entry("leg_press"); ok {
		t.Error("Entry(leg_press) found an entry that is not in the inventory")
	}
}
