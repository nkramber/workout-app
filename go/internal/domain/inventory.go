package domain

// DumbbellSet holds the dumbbells as their lightest weight, heaviest
// weight, and step (D-155). Each weight is the load of one dumbbell.
// The set includes the adjustable bench (D-163).
type DumbbellSet struct {
	Lightest Load
	Heaviest Load
	Step     Load
}

// Check refuses a set with a weight of 0 or less, a heaviest weight
// below the lightest, or a step that does not divide the range.
func (d DumbbellSet) Check() error {
	if err := d.Lightest.Check(); err != nil {
		return invalid("dumbbells lightest: %v", err)
	}
	if err := d.Step.Check(); err != nil {
		return invalid("dumbbells step: %v", err)
	}
	if d.Heaviest < d.Lightest {
		return invalid("dumbbells heaviest %s below lightest %s", d.Heaviest, d.Lightest)
	}
	if (d.Heaviest-d.Lightest)%d.Step != 0 {
		return invalid("dumbbells step %s does not divide %s to %s", d.Step, d.Lightest, d.Heaviest)
	}
	return nil
}

// Weights gives each weight of a set that passes its check, lightest
// first.
func (d DumbbellSet) Weights() []Load {
	if d.Check() != nil {
		return nil
	}
	var out []Load
	for w := d.Lightest; w <= d.Heaviest; w += d.Step {
		out = append(out, w)
	}
	return out
}

// InventoryEntry is one machine of the inventory of the owner. It holds
// the identity and the available weights alone (D-54). The identity is
// the catalog id. A machine or the cable station gives its weights as a
// list, so a stack with an irregular step, such as 12.5 lb, fits. The
// dumbbells give a DumbbellSet. A cardio machine has no weights.
type InventoryEntry struct {
	Machine   MachineID
	Weights   []Load
	Dumbbells *DumbbellSet
}

// Check reads an entry against the catalog: a known machine, and the
// weights of its kind. A list of weights is strictly ascending, and each
// weight is more than 0.
func (e InventoryEntry) Check(c Catalog) error {
	m, ok := c.Machine(e.Machine)
	if !ok {
		return invalid("inventory machine %q: not in the catalog", e.Machine)
	}
	switch m.Kind {
	case KindMachine, KindCable:
		if e.Dumbbells != nil {
			return invalid("inventory machine %q: a dumbbell set on kind %q", e.Machine, m.Kind)
		}
		if len(e.Weights) == 0 {
			return invalid("inventory machine %q: no weights", e.Machine)
		}
		for i, w := range e.Weights {
			if err := w.Check(); err != nil {
				return invalid("inventory machine %q weights[%d]: %v", e.Machine, i, err)
			}
			if i > 0 && w <= e.Weights[i-1] {
				return invalid("inventory machine %q weights[%d]: %s not above %s", e.Machine, i, w, e.Weights[i-1])
			}
		}
	case KindDumbbell:
		if len(e.Weights) != 0 {
			return invalid("inventory machine %q: a list of weights on kind %q", e.Machine, m.Kind)
		}
		if e.Dumbbells == nil {
			return invalid("inventory machine %q: no dumbbell set", e.Machine)
		}
		if err := e.Dumbbells.Check(); err != nil {
			return invalid("inventory machine %q: %v", e.Machine, err)
		}
	case KindCardio:
		if len(e.Weights) != 0 || e.Dumbbells != nil {
			return invalid("inventory machine %q: weights on kind %q", e.Machine, m.Kind)
		}
	}
	return nil
}

// Available gives the available weights of the entry, lightest first.
func (e InventoryEntry) Available() []Load {
	if e.Dumbbells != nil {
		return e.Dumbbells.Weights()
	}
	return append([]Load(nil), e.Weights...)
}

// Inventory is the one active inventory of the owner (D-46). It holds
// each machine one time.
type Inventory struct {
	Entries []InventoryEntry
}

// Check reads each entry, and refuses a machine that is in the
// inventory two times.
func (inv Inventory) Check(c Catalog) error {
	seen := map[MachineID]bool{}
	for _, e := range inv.Entries {
		if err := e.Check(c); err != nil {
			return err
		}
		if seen[e.Machine] {
			return invalid("inventory machine %q: two entries", e.Machine)
		}
		seen[e.Machine] = true
	}
	return nil
}

// Entry gives the entry of a machine.
func (inv Inventory) Entry(id MachineID) (InventoryEntry, bool) {
	for _, e := range inv.Entries {
		if e.Machine == id {
			return e, true
		}
	}
	return InventoryEntry{}, false
}
