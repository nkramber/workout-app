// Package inventory holds the one active inventory of the owner (D-46)
// and its store (work area 4.2). A machine holds its identity, its
// weights, the load estimates of its exercises, and its state alone
// (D-54, D-192, D-193). A note holds a text that matched no catalog name,
// and no plan reads it (D-191).
//
// Each change gives a new Inventory and leaves the old one as it was, so
// a store can run a change again in a retry of its transaction. A check
// error matches domain.ErrInvalid. An error names ids and numbers alone,
// and never the text of a note (D-80).
package inventory

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// State is the state of a machine (D-193).
type State string

const (
	// Draft is a machine that the owner saved and did not confirm.
	Draft State = "draft"
	// Confirmed is a machine that the owner confirmed. A plan reads it.
	Confirmed State = "confirmed"
)

// The bounds of D-199. Dumbbells keep the bound of domain.DumbbellMax.
const (
	// MaxWeight is the heaviest weight of a machine or of the cable
	// station.
	MaxWeight = 1000 * domain.Pound
	// MaxWeights is the length limit of the weight list of one machine.
	MaxWeights = 200
	// MaxNoteRunes is the length limit of the text of a note, in
	// characters.
	MaxNoteRunes = 200
	// MaxNotes is the limit of notes in the inventory.
	MaxNotes = 50
)

var (
	// ErrNotFound tells that the inventory holds no such machine or
	// note.
	ErrNotFound = errors.New("not found")
	// ErrWeightsChanged tells that the weights of a confirmation are
	// not the stored weights (D-201).
	ErrWeightsChanged = errors.New("the weights are not the stored weights")
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalid, fmt.Sprintf(format, args...))
}

// Machine is one machine of the inventory. Entry holds the catalog id
// and the weights. Estimates holds the optional load estimate of each
// exercise of the machine (D-192).
type Machine struct {
	Entry     domain.InventoryEntry
	Estimates map[domain.ExerciseID]domain.Load
	State     State
}

// Note is a text of the owner with no catalog match (D-191).
type Note struct {
	ID   string
	Text string
}

// Inventory is the one active inventory of the owner (D-46). Machines
// are in catalog order, and notes are in the order of their addition.
type Inventory struct {
	Machines []Machine
	Notes    []Note
}

// Check reads a machine against the catalog: the entry and its bounds,
// a known state, and each estimate. An estimate is for an exercise of
// the machine, and its load is from the lightest to the heaviest weight
// of the machine (D-198). So a cardio machine takes no estimate.
func (m Machine) Check(c domain.Catalog) error {
	id := m.Entry.Machine
	if err := m.Entry.Check(c); err != nil {
		return err
	}
	if len(m.Entry.Weights) > MaxWeights {
		return invalid("inventory machine %q: %d weights, want %d or fewer", id, len(m.Entry.Weights), MaxWeights)
	}
	if n := len(m.Entry.Weights); n > 0 && m.Entry.Weights[n-1] > MaxWeight {
		return invalid("inventory machine %q: heaviest weight %s, want %s or less", id, m.Entry.Weights[n-1], MaxWeight)
	}
	if m.State != Draft && m.State != Confirmed {
		return invalid("inventory machine %q: unknown state %q", id, m.State)
	}
	available := m.Entry.Available()
	for _, ex := range slices.Sorted(maps.Keys(m.Estimates)) {
		load := m.Estimates[ex]
		e, ok := c.Exercise(ex)
		if !ok || e.Machine != id {
			return invalid("inventory machine %q: estimate for exercise %q, which is not an exercise of the machine", id, ex)
		}
		if err := load.Check(); err != nil {
			return invalid("inventory machine %q estimate %q: %v", id, ex, err)
		}
		if len(available) == 0 || load < available[0] || load > available[len(available)-1] {
			return invalid("inventory machine %q estimate %q: %s is outside the weights of the machine", id, ex, load)
		}
	}
	return nil
}

// Check reads a note: an id, and a text of 1 to MaxNoteRunes characters
// with no space at either end.
func (n Note) Check() error {
	if n.ID == "" {
		return invalid("inventory note: empty id")
	}
	if !utf8.ValidString(n.Text) {
		return invalid("inventory note %q: the text is not UTF-8", n.ID)
	}
	if strings.TrimSpace(n.Text) != n.Text {
		return invalid("inventory note %q: a space at an end of the text", n.ID)
	}
	if k := utf8.RuneCountInString(n.Text); k == 0 || k > MaxNoteRunes {
		return invalid("inventory note %q: %d characters, want 1 to %d", n.ID, k, MaxNoteRunes)
	}
	return nil
}

// Check reads each machine and each note. It refuses a machine that is
// in the inventory two times, two notes with one id, and more than
// MaxNotes notes.
func (inv Inventory) Check(c domain.Catalog) error {
	seen := map[domain.MachineID]bool{}
	for _, m := range inv.Machines {
		if err := m.Check(c); err != nil {
			return err
		}
		if seen[m.Entry.Machine] {
			return invalid("inventory machine %q: two entries", m.Entry.Machine)
		}
		seen[m.Entry.Machine] = true
	}
	if len(inv.Notes) > MaxNotes {
		return invalid("inventory: %d notes, want %d or fewer", len(inv.Notes), MaxNotes)
	}
	ids := map[string]bool{}
	for _, n := range inv.Notes {
		if err := n.Check(); err != nil {
			return err
		}
		if ids[n.ID] {
			return invalid("inventory note %q: two notes", n.ID)
		}
		ids[n.ID] = true
	}
	return nil
}

// Machine gives the machine with the catalog id.
func (inv Inventory) Machine(id domain.MachineID) (Machine, bool) {
	i := inv.machineIndex(id)
	if i < 0 {
		return Machine{}, false
	}
	return inv.Machines[i].clone(), true
}

func (inv Inventory) machineIndex(id domain.MachineID) int {
	return slices.IndexFunc(inv.Machines, func(m Machine) bool { return m.Entry.Machine == id })
}

func (inv Inventory) noteIndex(id string) int {
	return slices.IndexFunc(inv.Notes, func(n Note) bool { return n.ID == id })
}

// SaveMachine adds the machine as a draft, or replaces the one entry of
// the machine (D-200). A confirmed machine stays confirmed when its
// weights do not change, and becomes a draft when they change (D-193).
// The estimates of m replace the stored estimates, and the state of m
// is not read.
func (inv Inventory) SaveMachine(c domain.Catalog, m Machine) (Inventory, error) {
	out := inv.clone()
	m = m.clone()
	m.State = Draft
	if i := out.machineIndex(m.Entry.Machine); i >= 0 {
		if old := out.Machines[i]; old.State == Confirmed && SameWeights(old.Entry, m.Entry) {
			m.State = Confirmed
		}
		out.Machines[i] = m
	} else {
		out.Machines = append(out.Machines, m)
	}
	out.sortMachines(c)
	if err := out.Check(c); err != nil {
		return Inventory{}, err
	}
	return out, nil
}

// ConfirmMachine confirms the machine with the weights that the review
// screen showed. It gives ErrNotFound for an unknown machine, and
// ErrWeightsChanged when the shown weights are not the stored weights
// (D-201).
func (inv Inventory) ConfirmMachine(c domain.Catalog, shown domain.InventoryEntry) (Inventory, error) {
	out := inv.clone()
	i := out.machineIndex(shown.Machine)
	if i < 0 {
		return Inventory{}, fmt.Errorf("inventory machine %q: %w", shown.Machine, ErrNotFound)
	}
	if !SameWeights(out.Machines[i].Entry, shown) {
		return Inventory{}, fmt.Errorf("inventory machine %q: %w", shown.Machine, ErrWeightsChanged)
	}
	out.Machines[i].State = Confirmed
	if err := out.Check(c); err != nil {
		return Inventory{}, err
	}
	return out, nil
}

// RemoveMachine removes the machine. An unknown machine changes nothing.
func (inv Inventory) RemoveMachine(id domain.MachineID) Inventory {
	out := inv.clone()
	out.Machines = slices.DeleteFunc(out.Machines, func(m Machine) bool { return m.Entry.Machine == id })
	return out
}

// SaveNote adds a note with the id newID when id is empty, or replaces
// the text of the note id. It removes the spaces at each end of the text.
// It gives ErrNotFound for an unknown id.
func (inv Inventory) SaveNote(c domain.Catalog, id, text, newID string) (Inventory, string, error) {
	out := inv.clone()
	text = strings.TrimSpace(text)
	if id == "" {
		id = newID
		out.Notes = append(out.Notes, Note{ID: id, Text: text})
	} else {
		i := out.noteIndex(id)
		if i < 0 {
			return Inventory{}, "", fmt.Errorf("inventory note %q: %w", id, ErrNotFound)
		}
		out.Notes[i].Text = text
	}
	if err := out.Check(c); err != nil {
		return Inventory{}, "", err
	}
	return out, id, nil
}

// RemoveNote removes the note. An unknown note changes nothing.
func (inv Inventory) RemoveNote(id string) Inventory {
	out := inv.clone()
	out.Notes = slices.DeleteFunc(out.Notes, func(n Note) bool { return n.ID == id })
	return out
}

// SameWeights tells whether two entries give the same weights: the same
// list, or the same dumbbell set, or none.
func SameWeights(a, b domain.InventoryEntry) bool {
	if (a.Dumbbells == nil) != (b.Dumbbells == nil) {
		return false
	}
	if a.Dumbbells != nil && *a.Dumbbells != *b.Dumbbells {
		return false
	}
	return slices.Equal(a.Weights, b.Weights)
}

// sortMachines puts the machines in catalog order. A machine outside the
// catalog goes last, and the check refuses it.
func (inv Inventory) sortMachines(c domain.Catalog) {
	order := map[domain.MachineID]int{}
	for i, m := range c.Machines {
		order[m.ID] = i
	}
	pos := func(id domain.MachineID) int {
		if i, ok := order[id]; ok {
			return i
		}
		return len(c.Machines)
	}
	slices.SortStableFunc(inv.Machines, func(a, b Machine) int { return pos(a.Entry.Machine) - pos(b.Entry.Machine) })
}

func (m Machine) clone() Machine {
	out := m
	out.Entry.Weights = slices.Clone(m.Entry.Weights)
	if m.Entry.Dumbbells != nil {
		d := *m.Entry.Dumbbells
		out.Entry.Dumbbells = &d
	}
	if m.Estimates != nil {
		out.Estimates = maps.Clone(m.Estimates)
	}
	return out
}

func (inv Inventory) clone() Inventory {
	out := Inventory{Notes: slices.Clone(inv.Notes)}
	for _, m := range inv.Machines {
		out.Machines = append(out.Machines, m.clone())
	}
	return out
}
