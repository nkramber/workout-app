package workoutsvc

import (
	"time"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/inventorysvc"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// The entities of the inventory entries of the outbox (D-272).
const (
	EntityMachine = "machine"
	EntityNote    = "note"
)

// change is the change of the inventory that one entry gives. A store can
// run it more than one time, so it has no side effect.
type change = func(inventory.Inventory) (inventory.Inventory, error)

// inventoryEntry reads an inventory entry of the outbox. It gives ok
// false for an entry with no inventory payload. For an inventory entry,
// it gives the op and the change, or a check error that matches
// workout.ErrInvalid. The change applies the rules of InventoryService.
func (s *Server) inventoryEntry(in *workoutappv1.OutboxEntry) (op inventory.Op, ch change, ok bool, err error) {
	var entity string
	switch in.GetPayload().(type) {
	case *workoutappv1.OutboxEntry_SaveMachine, *workoutappv1.OutboxEntry_ConfirmMachine, *workoutappv1.OutboxEntry_RemoveMachine:
		entity = EntityMachine
	case *workoutappv1.OutboxEntry_SaveNote, *workoutappv1.OutboxEntry_RemoveNote:
		entity = EntityNote
	default:
		return op, nil, false, nil
	}
	if err := workout.CheckOpID(in.GetOpId()); err != nil {
		return op, nil, true, err
	}
	if v := in.GetSchemaVersion(); v != workout.SchemaVersion {
		return op, nil, true, checkError("schema version: want 1")
	}
	if in.GetBaseVersion() != 0 {
		return op, nil, true, checkError("base version: want 0 for an inventory entry")
	}
	at, err := time.Parse(time.RFC3339Nano, in.GetAt())
	if err != nil {
		return op, nil, true, checkError("at: want an RFC 3339 time")
	}
	if in.GetEntity() != entity {
		return op, nil, true, checkError("entity: the payload wants " + entity)
	}
	id := in.GetEntityId()
	if entity == EntityMachine {
		if _, known := s.catalog.Machine(domain.MachineID(id)); !known {
			return op, nil, true, checkError("entity id: not a machine of the catalog")
		}
	} else if err := workout.CheckID("note id", id); err != nil {
		return op, nil, true, err
	}
	op = inventory.Op{ID: in.GetOpId(), Entity: entity, EntityID: id, At: at}

	switch p := in.GetPayload().(type) {
	case *workoutappv1.OutboxEntry_SaveMachine:
		estimates, err := inventorysvc.Estimates(p.SaveMachine.GetEstimates())
		if err != nil {
			return op, nil, true, err
		}
		m := inventory.Machine{Entry: inventorysvc.Entry(id, p.SaveMachine.GetWeightsTenthLb(), p.SaveMachine.GetDumbbells()), Estimates: estimates}
		ch = func(inv inventory.Inventory) (inventory.Inventory, error) { return inv.SaveMachine(s.catalog, m) }
	case *workoutappv1.OutboxEntry_ConfirmMachine:
		shown := inventorysvc.Entry(id, p.ConfirmMachine.GetWeightsTenthLb(), p.ConfirmMachine.GetDumbbells())
		ch = func(inv inventory.Inventory) (inventory.Inventory, error) {
			return inv.ConfirmMachine(s.catalog, shown)
		}
	case *workoutappv1.OutboxEntry_RemoveMachine:
		ch = func(inv inventory.Inventory) (inventory.Inventory, error) {
			return inv.RemoveMachine(domain.MachineID(id)), nil
		}
	case *workoutappv1.OutboxEntry_SaveNote:
		text := p.SaveNote.GetText()
		// The phone makes the id of a new note, so the entry adds the note
		// when the inventory does not hold it yet (D-272).
		ch = func(inv inventory.Inventory) (inventory.Inventory, error) {
			for _, n := range inv.Notes {
				if n.ID == id {
					out, _, err := inv.SaveNote(s.catalog, id, text, "")
					return out, err
				}
			}
			out, _, err := inv.SaveNote(s.catalog, "", text, id)
			return out, err
		}
	case *workoutappv1.OutboxEntry_RemoveNote:
		ch = func(inv inventory.Inventory) (inventory.Inventory, error) { return inv.RemoveNote(id), nil }
	}
	return op, ch, true, nil
}
