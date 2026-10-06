package ai

import (
	"fmt"
	"strings"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// RotationOf gives the input of the rotation rules of a request (D-328
// to D-331): the count of sessions, the exercises of the request with
// their primary groups of the body tables, and the groups of the
// profile. A request with no profile selects no group. A cardio
// exercise or an exercise with no group is a filler.
func RotationOf(req Request) policy.Rotation {
	t := domain.DefaultBodyTables()
	r := policy.Rotation{Sessions: req.Sessions, Groups: map[domain.ExerciseID][]domain.MuscleGroup{}, Filler: len(req.Cardio) > 0}
	for _, x := range req.Exercises {
		id := x.Exercise.ID
		r.Order = append(r.Order, id)
		r.Groups[id] = append([]domain.MuscleGroup(nil), t.Groups[id]...)
		if len(t.Groups[id]) == 0 {
			r.Filler = true
		}
	}
	if p := req.Profile; p != nil {
		for _, g := range p.MuscleGroups {
			r.Selected = append(r.Selected, domain.MuscleGroup(g))
		}
	}
	return r
}

// checkRotation refuses a plan that breaks a rotation rule of the
// policy. The error names the rule ids, the sessions, and the group ids
// alone, so it can go into a log (D-80). The plan API sends it to Luna
// with the retry (D-230).
func checkRotation(p Plan, req Request) error {
	sessions := make([][]domain.ExerciseID, len(p.Sessions))
	for i, s := range p.Sessions {
		for _, e := range s.Exercises {
			sessions[i] = append(sessions[i], e.Target.Exercise)
		}
	}
	v := policy.CheckRotation(RotationOf(req), sessions)
	if len(v) == 0 {
		return nil
	}
	list := make([]string, 0, len(v))
	for _, x := range v {
		list = append(list, x.String())
	}
	return malformed("%s", strings.Join(list, ", "))
}

// rotationNote is the line of the task that tells Luna the rotation.
var rotationNote = fmt.Sprintf(`- When rotation is true in the input, split the muscle groups between the sessions, as the policy rules %s and %s below give them. Each exercise gives its muscle_groups, and an exercise with none trains no group of these rules. When rotation is false, these rules do not apply.`,
	policy.RuleRotationRepeat, policy.RuleRotationCover)
