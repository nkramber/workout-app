package policy

import (
	"fmt"
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// Rotation is the input of the rotation rules of the layout of a plan
// (D-328 to D-331). Groups gives the primary muscle groups of each
// exercise of the request, from the body tables of D-210. Order gives
// the exercises of the request in their order. Selected gives the
// groups of the profile, and Filler tells that the request has a cardio
// exercise or an exercise with no group, so that a session can hold no
// exercise of a group.
type Rotation struct {
	Sessions int
	Order    []domain.ExerciseID
	Groups   map[domain.ExerciseID][]domain.MuscleGroup
	Selected []domain.MuscleGroup
	Filler   bool
}

// Units gives the exercises of the request in units, in the order of
// the request. Two exercises with a group in common are in one unit,
// and so is each exercise that links to them. Each unit of a plan is
// in the same sessions, because no group is in two sessions in a row.
// An exercise with no group is in no unit.
func (r Rotation) Units() [][]domain.ExerciseID {
	var units [][]domain.ExerciseID
	var groups [][]domain.MuscleGroup
	for _, id := range r.Order {
		g := r.Groups[id]
		if len(g) == 0 {
			continue
		}
		unit, gs := []domain.ExerciseID{id}, slices.Clone(g)
		for i := 0; i < len(units); {
			if slices.ContainsFunc(groups[i], func(x domain.MuscleGroup) bool { return slices.Contains(gs, x) }) {
				unit, gs = append(units[i], unit...), append(groups[i], gs...)
				units, groups = slices.Delete(units, i, i+1), slices.Delete(groups, i, i+1)
				continue
			}
			i++
		}
		units, groups = append(units, unit), append(groups, gs)
	}
	for _, u := range units {
		slices.SortStableFunc(u, func(a, b domain.ExerciseID) int {
			return slices.Index(r.Order, a) - slices.Index(r.Order, b)
		})
	}
	slices.SortStableFunc(units, func(a, b []domain.ExerciseID) int {
		return slices.Index(r.Order, a[0]) - slices.Index(r.Order, b[0])
	})
	return units
}

// Classes gives the count of the classes of sessions: 2 for an even
// count of sessions, where sessions 1 and 3 share their groups and so
// do sessions 2 and 4, and the count of sessions for an odd count.
func (r Rotation) Classes() int {
	if r.Sessions%2 == 0 {
		return 2
	}
	return r.Sessions
}

// Applies tells if the policy can split the exercises of the request
// into the sessions with the rotation rules. A plan of 1 session, and a
// request with no exercise of a group, have no rotation. With no filler,
// each class of sessions needs a unit of its own.
func (r Rotation) Applies() bool {
	n := len(r.Units())
	return r.Sessions >= 2 && n > 0 && (r.Filler || n >= r.Classes())
}

// Required gives each selected group that an exercise of the request
// trains, in the order of D-210. A group with no exercise in the
// request, after an injury, an exclusion, or a machine that is not
// confirmed, is not required.
func (r Rotation) Required() []domain.MuscleGroup {
	var out []domain.MuscleGroup
	for _, g := range domain.MuscleGroups() {
		if !slices.Contains(r.Selected, g) {
			continue
		}
		for _, id := range r.Order {
			if slices.Contains(r.Groups[id], g) {
				out = append(out, g)
				break
			}
		}
	}
	return out
}

// CheckRotation reads the exercises of each session of a plan, in the
// order of the sessions, against the rotation rules, and gives each
// violation. It gives none when the rules do not apply. The last
// session and the first session are in a row, because the week repeats.
func CheckRotation(r Rotation, sessions [][]domain.ExerciseID) []Violation {
	if !r.Applies() || len(sessions) != r.Sessions {
		return nil
	}
	trained := make([][]domain.MuscleGroup, len(sessions))
	for i, s := range sessions {
		for _, id := range s {
			for _, g := range r.Groups[id] {
				if !slices.Contains(trained[i], g) {
					trained[i] = append(trained[i], g)
				}
			}
		}
	}
	// Each pair of sessions in a row, one time each. Two sessions make
	// one pair.
	var pairs [][2]int
	for i := range sessions {
		j := (i + 1) % len(sessions)
		if len(sessions) == 2 && i == 1 {
			break
		}
		pairs = append(pairs, [2]int{i, j})
	}
	var out []Violation
	for _, p := range pairs {
		where := fmt.Sprintf("sessions[%d] and sessions[%d]", p[0], p[1])
		for _, g := range domain.MuscleGroups() {
			if slices.Contains(trained[p[0]], g) && slices.Contains(trained[p[1]], g) {
				out = append(out, Violation{RuleRotationRepeat, where, fmt.Sprintf("group %s in both sessions", g)})
			}
		}
	}
	for _, g := range r.Required() {
		if r.Sessions%2 == 1 {
			if !slices.ContainsFunc(trained, func(t []domain.MuscleGroup) bool { return slices.Contains(t, g) }) {
				out = append(out, Violation{RuleRotationCover, "sessions", fmt.Sprintf("group %s in no session", g)})
			}
			continue
		}
		for _, p := range pairs {
			if !slices.Contains(trained[p[0]], g) && !slices.Contains(trained[p[1]], g) {
				where := fmt.Sprintf("sessions[%d] and sessions[%d]", p[0], p[1])
				out = append(out, Violation{RuleRotationCover, where, fmt.Sprintf("group %s in neither session", g)})
			}
		}
	}
	return out
}
