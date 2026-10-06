package policy

import (
	"slices"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

type ids = []domain.ExerciseID

// rotationOf gives the rotation input of the exercises of the catalog,
// with the groups of the body tables of D-210 and each group selected.
func rotationOf(sessions int, order ...domain.ExerciseID) Rotation {
	t := domain.DefaultBodyTables()
	r := Rotation{Sessions: sessions, Order: order, Groups: map[domain.ExerciseID][]domain.MuscleGroup{}, Selected: domain.MuscleGroups()}
	for _, id := range order {
		r.Groups[id] = t.Groups[id]
		if len(t.Groups[id]) == 0 {
			r.Filler = true
		}
	}
	return r
}

// The full gym of the owner: push, pull, legs, calves, and core.
var gym = ids{"chest_press", "shoulder_press", "triceps_pulldown", "seated_row", "biceps_curl",
	"leg_press", "seated_leg_curl", "db_romanian_deadlift", "calf_raise", "abdominal_crunch"}

func texts(v []Violation) []string {
	var out []string
	for _, x := range v {
		out = append(out, x.String())
	}
	return out
}

func TestUnitsLinkExercisesWithAGroupInCommon(t *testing.T) {
	r := rotationOf(2, append(slices.Clone(gym), "hip_adduction")...)
	got := r.Units()
	want := [][]domain.ExerciseID{
		{"chest_press", "shoulder_press", "triceps_pulldown"},
		{"seated_row", "biceps_curl"},
		{"leg_press", "seated_leg_curl", "db_romanian_deadlift"},
		{"calf_raise"},
		{"abdominal_crunch"},
	}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("units %v: want %v", got, want)
	}
}

func TestUnitsJoinTwoUnitsThroughALaterExercise(t *testing.T) {
	// The leg curl and the leg press share no group, and the Romanian
	// deadlift links them through hamstrings and glutes.
	got := rotationOf(2, "seated_leg_curl", "leg_press", "db_romanian_deadlift").Units()
	want := [][]domain.ExerciseID{{"seated_leg_curl", "leg_press", "db_romanian_deadlift"}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("units %v: want %v", got, want)
	}
}

func TestRotationAcceptsAnABSplit(t *testing.T) {
	r := rotationOf(2, gym...)
	a := ids{"chest_press", "shoulder_press", "triceps_pulldown", "abdominal_crunch"}
	b := ids{"seated_row", "biceps_curl", "leg_press", "seated_leg_curl", "calf_raise"}
	if v := CheckRotation(r, [][]domain.ExerciseID{a, b}); len(v) != 0 {
		t.Fatalf("violations %v: want none", texts(v))
	}
}

func TestRotationRefusesAGroupInTwoSessionsInARow(t *testing.T) {
	// The acceptance story: legs in both sessions.
	r := rotationOf(2, gym...)
	a := ids{"chest_press", "leg_press", "abdominal_crunch"}
	b := ids{"seated_row", "biceps_curl", "leg_press", "seated_leg_curl", "calf_raise", "triceps_pulldown"}
	got := texts(CheckRotation(r, [][]domain.ExerciseID{a, b}))
	want := []string{
		"rotation.no-repeat sessions[0] and sessions[1]: group triceps in both sessions",
		"rotation.no-repeat sessions[0] and sessions[1]: group quadriceps in both sessions",
		"rotation.no-repeat sessions[0] and sessions[1]: group glutes in both sessions",
		"rotation.cover sessions[0] and sessions[1]: group shoulders in neither session",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("violations %q: want %q", got, want)
	}
}

func TestRotationRefusesTheSameSessionsTwice(t *testing.T) {
	// The layout of policy version 8: each session holds the same
	// exercises.
	r := rotationOf(2, gym...)
	v := CheckRotation(r, [][]domain.ExerciseID{gym[:8], gym[:8]})
	if len(v) == 0 || v[0].Rule != RuleRotationRepeat {
		t.Fatalf("violations %v: want rotation.no-repeat first", texts(v))
	}
}

func TestRotationOfFourSessionsAlternates(t *testing.T) {
	r := rotationOf(4, gym...)
	a := ids{"chest_press", "shoulder_press", "triceps_pulldown", "abdominal_crunch"}
	b := ids{"seated_row", "biceps_curl", "leg_press", "seated_leg_curl", "calf_raise"}
	got := texts(CheckRotation(r, [][]domain.ExerciseID{a, b, a[:3], b}))
	want := []string{
		"rotation.cover sessions[1] and sessions[2]: group core in neither session",
		"rotation.cover sessions[2] and sessions[3]: group core in neither session",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("violations %q: want %q", got, want)
	}
	// Session 4 and session 1 are in a row.
	got = texts(CheckRotation(r, [][]domain.ExerciseID{a, b, b, a}))
	if len(got) == 0 || !slices.Contains(got, "rotation.no-repeat sessions[1] and sessions[2]: group back in both sessions") ||
		!slices.Contains(got, "rotation.no-repeat sessions[3] and sessions[0]: group chest in both sessions") {
		t.Fatalf("violations %q: want the repeats of sessions 2 and 3, and of sessions 4 and 1", got)
	}
}

func TestRotationOfThreeSessionsTrainsEachGroupOnce(t *testing.T) {
	r := rotationOf(3, gym...)
	push := ids{"chest_press", "shoulder_press", "triceps_pulldown", "abdominal_crunch"}
	pull := ids{"seated_row", "biceps_curl", "calf_raise"}
	legs := ids{"leg_press", "seated_leg_curl", "db_romanian_deadlift"}
	if v := CheckRotation(r, [][]domain.ExerciseID{push, pull, legs}); len(v) != 0 {
		t.Fatalf("violations %v: want none", texts(v))
	}
	got := texts(CheckRotation(r, [][]domain.ExerciseID{push, pull, legs[:1]}))
	want := []string{"rotation.cover sessions: group hamstrings in no session"}
	if !slices.Equal(got, want) {
		t.Fatalf("violations %q: want %q", got, want)
	}
	// An A-B-A week repeats the groups of session 3 in session 1.
	got = texts(CheckRotation(r, [][]domain.ExerciseID{push, append(slices.Clone(pull), legs...), push}))
	if len(got) == 0 || got[0] != "rotation.no-repeat sessions[2] and sessions[0]: group chest in both sessions" {
		t.Fatalf("violations %q: want the repeat of sessions 3 and 1", got)
	}
}

func TestRotationRequiresSelectedGroupsOfTheRequestAlone(t *testing.T) {
	r := rotationOf(2, gym...)
	// The strength template of D-210 selects no arm, calf, or core.
	r.Selected = []domain.MuscleGroup{domain.GroupChest, domain.GroupBack, domain.GroupShoulders, domain.GroupQuadriceps, domain.GroupHamstrings, domain.GroupGlutes}
	a := ids{"chest_press", "shoulder_press"}
	b := ids{"seated_row", "leg_press", "seated_leg_curl"}
	if v := CheckRotation(r, [][]domain.ExerciseID{a, b}); len(v) != 0 {
		t.Fatalf("violations %v: want none", texts(v))
	}
	// A group with no exercise in the request is not required.
	r = rotationOf(2, "chest_press", "seated_row")
	if got := r.Required(); !slices.Equal(got, []domain.MuscleGroup{domain.GroupChest, domain.GroupBack, domain.GroupBiceps, domain.GroupTriceps}) {
		t.Fatalf("required %v", got)
	}
	if v := CheckRotation(r, [][]domain.ExerciseID{{"chest_press"}, {"seated_row"}}); len(v) != 0 {
		t.Fatalf("violations %v: want none", texts(v))
	}
}

func TestRotationDoesNotApplyWhenNoSplitIsPossible(t *testing.T) {
	cases := []struct {
		name string
		r    Rotation
		want bool
	}{
		{"one unit and no filler", rotationOf(2, "leg_press", "db_romanian_deadlift"), false},
		{"one unit and a cardio exercise", func() Rotation { r := rotationOf(2, "leg_press"); r.Filler = true; return r }(), true},
		{"one unit and an exercise with no group", rotationOf(2, "leg_press", "hip_adduction"), true},
		{"two units", rotationOf(2, "leg_press", "chest_press"), true},
		{"two units and three sessions", rotationOf(3, "leg_press", "chest_press"), false},
		{"three units and three sessions", rotationOf(3, "leg_press", "chest_press", "seated_row"), true},
		{"two units and four sessions", rotationOf(4, "leg_press", "chest_press"), true},
		{"one session", rotationOf(1, gym...), false},
		{"no exercise of a group", rotationOf(2, "hip_adduction"), false},
	}
	for _, c := range cases {
		if got := c.r.Applies(); got != c.want {
			t.Errorf("%s: applies %v: want %v", c.name, got, c.want)
		}
	}
	// With no rotation, the same exercise in each session passes.
	r := rotationOf(2, "leg_press", "db_romanian_deadlift")
	if v := CheckRotation(r, [][]domain.ExerciseID{{"leg_press"}, {"leg_press"}}); len(v) != 0 {
		t.Fatalf("violations %v: want none", texts(v))
	}
}

func TestRotationRulesAreInTheRegistry(t *testing.T) {
	for _, id := range []RuleID{RuleRotationRepeat, RuleRotationCover} {
		if !slices.ContainsFunc(Rules(), func(r Rule) bool { return r.ID == id && r.Text != "" && len(r.Sources) > 0 }) {
			t.Errorf("rule %s: not in the registry", id)
		}
	}
}
