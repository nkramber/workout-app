package domain

import (
	"errors"
	"slices"
	"testing"
)

// d155 is each item of D-155 with its kind, and each exercise with its
// region of D-161. The test writes the list again from the decisions, so
// a change of the catalog data must change this list too.
var d155 = []struct {
	machine  MachineID
	kind     Kind
	exercise ExerciseID
	region   Region
}{
	{"leg_press", KindMachine, "leg_press", RegionLowerPush},
	{"leg_extension", KindMachine, "leg_extension", RegionLowerPush},
	{"seated_leg_curl", KindMachine, "seated_leg_curl", RegionLowerPull},
	{"lying_leg_curl", KindMachine, "lying_leg_curl", RegionLowerPull},
	{"hip_abduction_adduction", KindMachine, "hip_abduction", RegionLowerPull},
	{"hip_abduction_adduction", KindMachine, "hip_adduction", RegionLowerPush},
	{"calf_raise", KindMachine, "calf_raise", RegionLowerPush},
	{"chest_press", KindMachine, "chest_press", RegionUpperPush},
	{"shoulder_press", KindMachine, "shoulder_press", RegionUpperPush},
	{"seated_row", KindMachine, "seated_row", RegionUpperPull},
	{"biceps_curl", KindMachine, "biceps_curl", RegionUpperPull},
	{"abdominal_crunch", KindMachine, "abdominal_crunch", RegionCore},
	{"back_extension", KindMachine, "back_extension", RegionLowerPull},
	{"cable_station", KindCable, "lat_pulldown", RegionUpperPull},
	{"cable_station", KindCable, "triceps_pulldown", RegionUpperPush},
	{"dumbbells", KindDumbbell, "db_flat_bench_press", RegionUpperPush},
	{"dumbbells", KindDumbbell, "db_incline_bench_press", RegionUpperPush},
	{"dumbbells", KindDumbbell, "db_seated_shoulder_press", RegionUpperPush},
	{"dumbbells", KindDumbbell, "db_one_arm_row", RegionUpperPull},
	{"dumbbells", KindDumbbell, "db_biceps_curl", RegionUpperPull},
	{"dumbbells", KindDumbbell, "db_hammer_curl", RegionUpperPull},
	{"dumbbells", KindDumbbell, "db_lateral_raise", RegionUpperPush},
	{"dumbbells", KindDumbbell, "db_romanian_deadlift", RegionLowerPull},
	{"dumbbells", KindDumbbell, "db_goblet_squat", RegionLowerPush},
	{"treadmill", KindCardio, "treadmill", RegionCardio},
	{"upright_bike", KindCardio, "upright_bike", RegionCardio},
	{"recumbent_bike", KindCardio, "recumbent_bike", RegionCardio},
	{"rowing_machine", KindCardio, "rowing_machine", RegionCardio},
	{"elliptical", KindCardio, "elliptical", RegionCardio},
	{"stair_climber", KindCardio, "stair_climber", RegionCardio},
}

func TestDefaultCatalogPasses(t *testing.T) {
	if err := DefaultCatalog().Check(); err != nil {
		t.Fatalf("DefaultCatalog().Check() = %v", err)
	}
}

func TestDefaultCatalogHoldsD155(t *testing.T) {
	c := DefaultCatalog()
	machines := map[MachineID]bool{}
	for _, want := range d155 {
		machines[want.machine] = true
		m, ok := c.Machine(want.machine)
		if !ok {
			t.Errorf("machine %q: not in the catalog", want.machine)
			continue
		}
		if m.Kind != want.kind {
			t.Errorf("machine %q: kind %q, want %q", m.ID, m.Kind, want.kind)
		}
		e, ok := c.Exercise(want.exercise)
		if !ok {
			t.Errorf("exercise %q: not in the catalog", want.exercise)
			continue
		}
		if e.Machine != want.machine || e.Kind != want.kind || e.Region != want.region {
			t.Errorf("exercise %q = %+v, want machine %q, kind %q, region %q",
				e.ID, e, want.machine, want.kind, want.region)
		}
	}
	// The catalog holds D-155 and nothing more: 12 machines, the cable
	// station, the dumbbells, and 6 cardio machines.
	if len(c.Machines) != len(machines) || len(machines) != 20 {
		t.Errorf("machines = %d, D-155 = %d, want 20", len(c.Machines), len(machines))
	}
	if len(c.Exercises) != len(d155) {
		t.Errorf("exercises = %d, want %d", len(c.Exercises), len(d155))
	}
}

func TestDefaultCatalogIsACopy(t *testing.T) {
	a := DefaultCatalog()
	a.Exercises[0].Region = RegionCore
	a.Machines[0].Name = "changed"
	b := DefaultCatalog()
	if b.Exercises[0].Region != RegionLowerPush || b.Machines[0].Name != "Leg press" {
		t.Fatal("a change of one catalog changed the next DefaultCatalog")
	}
}

func ids(es []Exercise) []ExerciseID {
	var out []ExerciseID
	for _, e := range es {
		out = append(out, e.ID)
	}
	return out
}

func TestLookupByID(t *testing.T) {
	c := DefaultCatalog()
	for _, tc := range []struct {
		id   ExerciseID
		ok   bool
		want MachineID
	}{
		{"hip_adduction", true, "hip_abduction_adduction"},
		{"lat_pulldown", true, "cable_station"},
		{"db_goblet_squat", true, "dumbbells"},
		{"treadmill", true, "treadmill"},
		{"pec_fly", false, ""},
		{"", false, ""},
	} {
		e, ok := c.Exercise(tc.id)
		if ok != tc.ok || e.Machine != tc.want {
			t.Errorf("Exercise(%q) = %+v, %v, want machine %q, %v", tc.id, e, ok, tc.want, tc.ok)
		}
	}
	for _, tc := range []struct {
		id   MachineID
		ok   bool
		want Kind
	}{
		{"cable_station", true, KindCable},
		{"dumbbells", true, KindDumbbell},
		{"leg_press", true, KindMachine},
		{"smith_machine", false, ""},
	} {
		m, ok := c.Machine(tc.id)
		if ok != tc.ok || m.Kind != tc.want {
			t.Errorf("Machine(%q) = %+v, %v, want kind %q, %v", tc.id, m, ok, tc.want, tc.ok)
		}
	}
}

func TestLookupByKind(t *testing.T) {
	c := DefaultCatalog()
	for _, tc := range []struct {
		kind Kind
		want int
	}{
		{KindMachine, 13},
		{KindCable, 2},
		{KindDumbbell, 9},
		{KindCardio, 6},
		{"barbell", 0},
	} {
		got := c.ExercisesOfKind(tc.kind)
		if len(got) != tc.want {
			t.Errorf("ExercisesOfKind(%q) = %d, want %d: %v", tc.kind, len(got), tc.want, ids(got))
		}
		for _, e := range got {
			if e.Kind != tc.kind {
				t.Errorf("ExercisesOfKind(%q) holds %q of kind %q", tc.kind, e.ID, e.Kind)
			}
		}
	}
	if got := ids(c.ExercisesOfKind(KindCable)); !slices.Equal(got, []ExerciseID{"lat_pulldown", "triceps_pulldown"}) {
		t.Errorf("ExercisesOfKind(cable) = %v, want catalog order", got)
	}
}

func TestLookupByRegion(t *testing.T) {
	c := DefaultCatalog()
	total := 0
	for _, tc := range []struct {
		region Region
		want   []ExerciseID
	}{
		{RegionUpperPush, []ExerciseID{"chest_press", "shoulder_press", "triceps_pulldown", "db_flat_bench_press", "db_incline_bench_press", "db_seated_shoulder_press", "db_lateral_raise"}},
		{RegionUpperPull, []ExerciseID{"seated_row", "biceps_curl", "lat_pulldown", "db_one_arm_row", "db_biceps_curl", "db_hammer_curl"}},
		{RegionLowerPush, []ExerciseID{"leg_press", "leg_extension", "hip_adduction", "calf_raise", "db_goblet_squat"}},
		{RegionLowerPull, []ExerciseID{"seated_leg_curl", "lying_leg_curl", "hip_abduction", "back_extension", "db_romanian_deadlift"}},
		{RegionCore, []ExerciseID{"abdominal_crunch"}},
		{RegionCardio, []ExerciseID{"treadmill", "upright_bike", "recumbent_bike", "rowing_machine", "elliptical", "stair_climber"}},
		{"arms", nil},
	} {
		got := ids(c.ExercisesInRegion(tc.region))
		if !slices.Equal(got, tc.want) {
			t.Errorf("ExercisesInRegion(%q) = %v, want %v", tc.region, got, tc.want)
		}
		total += len(got)
	}
	if total != len(c.Exercises) {
		t.Errorf("the regions hold %d exercises, want each of %d", total, len(c.Exercises))
	}
}

func TestLookupByMachine(t *testing.T) {
	c := DefaultCatalog()
	for _, tc := range []struct {
		machine MachineID
		want    int
	}{
		{"hip_abduction_adduction", 2},
		{"cable_station", 2},
		{"dumbbells", 9},
		{"chest_press", 1},
		{"none", 0},
	} {
		if got := c.ExercisesOnMachine(tc.machine); len(got) != tc.want {
			t.Errorf("ExercisesOnMachine(%q) = %v, want %d", tc.machine, ids(got), tc.want)
		}
	}
}

func TestCatalogCheckRefuses(t *testing.T) {
	m := func(id MachineID, k Kind) Machine { return Machine{ID: id, Name: "n", Kind: k} }
	e := func(id ExerciseID, on MachineID, k Kind, r Region) Exercise {
		return Exercise{ID: id, Name: "n", Machine: on, Kind: k, Region: r}
	}
	good := func() Catalog {
		return Catalog{
			Version:   1,
			Machines:  []Machine{m("a", KindMachine), m("t", KindCardio)},
			Exercises: []Exercise{e("a", "a", KindMachine, RegionCore), e("t", "t", KindCardio, RegionCardio)},
		}
	}
	if err := good().Check(); err != nil {
		t.Fatalf("good catalog: %v", err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*Catalog)
	}{
		{"version 0", func(c *Catalog) { c.Version = 0 }},
		{"empty machine id", func(c *Catalog) { c.Machines[0].ID = "" }},
		{"empty machine name", func(c *Catalog) { c.Machines[0].Name = "" }},
		{"unknown machine kind", func(c *Catalog) { c.Machines[0].Kind = "barbell" }},
		{"duplicate machine", func(c *Catalog) { c.Machines = append(c.Machines, m("a", KindMachine)) }},
		{"empty exercise id", func(c *Catalog) { c.Exercises[0].ID = "" }},
		{"empty exercise name", func(c *Catalog) { c.Exercises[0].Name = "" }},
		{"duplicate exercise", func(c *Catalog) { c.Exercises = append(c.Exercises, e("a", "a", KindMachine, RegionCore)) }},
		{"unknown region", func(c *Catalog) { c.Exercises[0].Region = "arms" }},
		{"unknown machine", func(c *Catalog) { c.Exercises[0].Machine = "z" }},
		{"kind not of the machine", func(c *Catalog) { c.Exercises[0].Kind = KindCable }},
		{"cardio region on a machine", func(c *Catalog) { c.Exercises[0].Region = RegionCardio }},
		{"cardio with a body region", func(c *Catalog) { c.Exercises[1].Region = RegionCore }},
		{"machine with no exercise", func(c *Catalog) { c.Machines = append(c.Machines, m("b", KindMachine)) }},
	} {
		c := good()
		tc.mutate(&c)
		if err := c.Check(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: Check() = %v, want ErrInvalid", tc.name, err)
		}
	}
}
