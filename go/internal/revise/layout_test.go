package revise

import (
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/plan"
)

func planOf(sessions ...[]domain.ExerciseID) plan.Plan {
	var p plan.Plan
	for _, s := range sessions {
		var sess plan.Session
		for _, id := range s {
			sess.Exercises = append(sess.Exercises, plan.Exercise{Target: domain.PlannedExercise{Exercise: id}})
		}
		p.Sessions = append(p.Sessions, sess)
	}
	return p
}

// TestLayout: the replay reads the rule rotation.no-repeat of the
// sessions of a plan (D-328).
func TestLayout(t *testing.T) {
	applies, v := layout(planOf([]domain.ExerciseID{"chest_press", "abdominal_crunch"}, []domain.ExerciseID{"seated_row", "leg_press"}))
	if !applies || len(v) != 0 {
		t.Fatalf("an A/B split: applies %v, violations %v", applies, v)
	}
	applies, v = layout(planOf([]domain.ExerciseID{"chest_press", "leg_press"}, []domain.ExerciseID{"seated_row", "leg_press"}))
	if !applies || len(v) != 2 || v[0].String() != "rotation.no-repeat sessions[0] and sessions[1]: group quadriceps in both sessions" {
		t.Fatalf("legs in both sessions: applies %v, violations %v", applies, v)
	}
	// One unit and no filler can make no split.
	applies, v = layout(planOf([]domain.ExerciseID{"leg_press"}, []domain.ExerciseID{"leg_press"}))
	if applies || len(v) != 0 {
		t.Fatalf("one unit: applies %v, violations %v", applies, v)
	}
	// A cardio exercise is a filler.
	p := planOf([]domain.ExerciseID{"leg_press"}, []domain.ExerciseID{"leg_press"})
	p.Sessions[1].Cardio = &domain.PlannedCardio{Exercise: "treadmill", Minutes: 20}
	if applies, v = layout(p); !applies || len(v) != 2 {
		t.Fatalf("one unit and cardio: applies %v, violations %v", applies, v)
	}
}
