package profile

import (
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// PlannerInput is the part of the profile that a planner call sends to
// Luna (D-209): the experience, the goal template, the muscle groups,
// the free text, and the cardio preference. Sessions is the count of
// sessions that the planner plans (D-211). Exercises holds the resistance
// exercises of the catalog that load no injured area (D-208).
//
// The type has no field for the age, the height, the weight, the injured
// areas, or the injury text, so a planner call can not send them.
type PlannerInput struct {
	Experience Experience
	Template   domain.TemplateID
	Groups     []domain.MuscleGroup
	FreeText   string
	// Cardio is the cardio preference with no exercise that loads an
	// injured area.
	Cardio    []domain.ExerciseID
	Sessions  int
	Exercises []domain.ExerciseID
}

// ForPlan gives the planner input of a profile. The server removes each
// exercise that loads an injured area before the call to Luna (D-208).
// The table gives the areas of each exercise. The result shares no
// memory with p.
func ForPlan(p Profile, c domain.Catalog, t domain.BodyTables) PlannerInput {
	out := PlannerInput{
		Experience: p.Experience,
		Template:   p.Template,
		Groups:     slices.Clone(p.Groups),
		FreeText:   p.FreeText,
		Sessions:   p.TrainingDays,
	}
	for _, id := range p.Cardio {
		if !t.Loads(id, p.InjuredAreas) {
			out.Cardio = append(out.Cardio, id)
		}
	}
	for _, e := range c.Exercises {
		if e.Kind != domain.KindCardio && !t.Loads(e.ID, p.InjuredAreas) {
			out.Exercises = append(out.Exercises, e.ID)
		}
	}
	return out
}
