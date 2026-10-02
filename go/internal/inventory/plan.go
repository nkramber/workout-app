package inventory

import (
	"maps"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// PlanInput is the part of the inventory that a plan reads: the
// confirmed machines and their estimates (D-49, D-193). Estimates holds
// the estimate of each exercise that has one. An exercise with no
// estimate starts at the lightest weight (D-192).
type PlanInput struct {
	Inventory domain.Inventory
	Estimates map[domain.ExerciseID]domain.Load
}

// ForPlan gives the confirmed machines alone, so a plan never reads a
// draft or a note (D-49, D-191, D-193). The result shares no memory with
// inv.
func ForPlan(inv Inventory) PlanInput {
	out := PlanInput{Estimates: map[domain.ExerciseID]domain.Load{}}
	for _, m := range inv.Machines {
		if m.State != Confirmed {
			continue
		}
		m = m.clone()
		out.Inventory.Entries = append(out.Inventory.Entries, m.Entry)
		maps.Copy(out.Estimates, m.Estimates)
	}
	return out
}
