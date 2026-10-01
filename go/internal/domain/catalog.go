package domain

import "slices"

// Kind names the kind of a machine and of its exercises.
type Kind string

const (
	KindMachine  Kind = "machine"  // a selectorized machine with a weight stack
	KindCable    Kind = "cable"    // the cable station of D-154
	KindDumbbell Kind = "dumbbell" // the dumbbell set with its adjustable bench (D-163)
	KindCardio   Kind = "cardio"   // a cardio machine
)

// Kinds lists each kind in catalog order.
var Kinds = []Kind{KindMachine, KindCable, KindDumbbell, KindCardio}

// Region names the body region of an exercise (D-161). The first four
// are the regions of EV-1, and the volume rules count them. Core and
// cardio stay outside those counts.
type Region string

const (
	RegionUpperPush Region = "upper_push"
	RegionUpperPull Region = "upper_pull"
	RegionLowerPush Region = "lower_push"
	RegionLowerPull Region = "lower_pull"
	RegionCore      Region = "core"
	RegionCardio    Region = "cardio"
)

// Regions lists each region in catalog order.
var Regions = []Region{RegionUpperPush, RegionUpperPull, RegionLowerPush, RegionLowerPull, RegionCore, RegionCardio}

// MachineID is the stable id of a machine of the catalog. An id never
// changes, and a removed id is never used again.
type MachineID string

// ExerciseID is the stable id of an exercise of the catalog, with the
// same rule as MachineID.
type ExerciseID string

// Machine is one item of equipment of the catalog: a selectorized
// machine, the cable station, the dumbbell set, or a cardio machine
// (D-155). The dumbbell set includes the adjustable bench (D-163).
type Machine struct {
	ID   MachineID
	Name string
	Kind Kind
}

// Exercise is one movement on one machine. A machine with two
// movements, such as hip abduction and hip adduction, gives two
// exercises (D-159). The kind of an exercise is the kind of its
// machine.
type Exercise struct {
	ID      ExerciseID
	Name    string
	Machine MachineID
	Kind    Kind
	Region  Region
}

// Catalog holds the machines and the exercises, with a version that
// changes with each change of the data.
type Catalog struct {
	Version   int
	Machines  []Machine
	Exercises []Exercise
}

func (k Kind) known() bool   { return slices.Contains(Kinds, k) }
func (r Region) known() bool { return slices.Contains(Regions, r) }

// Check reads the rules of a catalog: unique ids, a name for each item,
// a known kind and region, the kind of the machine on each exercise, the
// cardio region for cardio alone, and one exercise or more on each
// machine.
func (c Catalog) Check() error {
	if c.Version < 1 {
		return invalid("catalog version %d: want 1 or more", c.Version)
	}
	machines := map[MachineID]Machine{}
	for i, m := range c.Machines {
		switch {
		case m.ID == "":
			return invalid("machines[%d]: empty id", i)
		case m.Name == "":
			return invalid("machine %q: empty name", m.ID)
		case !m.Kind.known():
			return invalid("machine %q: unknown kind %q", m.ID, m.Kind)
		}
		if _, dup := machines[m.ID]; dup {
			return invalid("machine %q: duplicate id", m.ID)
		}
		machines[m.ID] = m
	}
	used := map[MachineID]bool{}
	exercises := map[ExerciseID]bool{}
	for i, e := range c.Exercises {
		switch {
		case e.ID == "":
			return invalid("exercises[%d]: empty id", i)
		case e.Name == "":
			return invalid("exercise %q: empty name", e.ID)
		case exercises[e.ID]:
			return invalid("exercise %q: duplicate id", e.ID)
		case !e.Region.known():
			return invalid("exercise %q: unknown region %q", e.ID, e.Region)
		}
		exercises[e.ID] = true
		m, ok := machines[e.Machine]
		if !ok {
			return invalid("exercise %q: unknown machine %q", e.ID, e.Machine)
		}
		if e.Kind != m.Kind {
			return invalid("exercise %q: kind %q, machine %q has kind %q", e.ID, e.Kind, m.ID, m.Kind)
		}
		if (e.Kind == KindCardio) != (e.Region == RegionCardio) {
			return invalid("exercise %q: kind %q with region %q", e.ID, e.Kind, e.Region)
		}
		used[e.Machine] = true
	}
	for _, m := range c.Machines {
		if !used[m.ID] {
			return invalid("machine %q: no exercise", m.ID)
		}
	}
	return nil
}

// Machine gives the machine with the id.
func (c Catalog) Machine(id MachineID) (Machine, bool) {
	i := slices.IndexFunc(c.Machines, func(m Machine) bool { return m.ID == id })
	if i < 0 {
		return Machine{}, false
	}
	return c.Machines[i], true
}

// Exercise gives the exercise with the id.
func (c Catalog) Exercise(id ExerciseID) (Exercise, bool) {
	i := slices.IndexFunc(c.Exercises, func(e Exercise) bool { return e.ID == id })
	if i < 0 {
		return Exercise{}, false
	}
	return c.Exercises[i], true
}

// ExercisesOfKind gives the exercises of one kind, in catalog order.
func (c Catalog) ExercisesOfKind(k Kind) []Exercise {
	return c.filter(func(e Exercise) bool { return e.Kind == k })
}

// ExercisesInRegion gives the exercises of one region, in catalog order.
func (c Catalog) ExercisesInRegion(r Region) []Exercise {
	return c.filter(func(e Exercise) bool { return e.Region == r })
}

// ExercisesOnMachine gives the exercises of one machine, in catalog
// order.
func (c Catalog) ExercisesOnMachine(id MachineID) []Exercise {
	return c.filter(func(e Exercise) bool { return e.Machine == id })
}

func (c Catalog) filter(keep func(Exercise) bool) []Exercise {
	var out []Exercise
	for _, e := range c.Exercises {
		if keep(e) {
			out = append(out, e)
		}
	}
	return out
}
