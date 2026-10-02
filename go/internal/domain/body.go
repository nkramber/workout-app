package domain

import "slices"

// Area names a body area of an injury (D-208). The list is fixed, and
// the owner selects areas from it.
type Area string

const (
	AreaShoulder  Area = "shoulder"
	AreaElbow     Area = "elbow"
	AreaWrist     Area = "wrist"
	AreaLowerBack Area = "lower_back"
	AreaHip       Area = "hip"
	AreaKnee      Area = "knee"
	AreaAnkle     Area = "ankle"
)

// Areas gives a new list of each area, from the top of the body down.
func Areas() []Area {
	return []Area{AreaShoulder, AreaElbow, AreaWrist, AreaLowerBack, AreaHip, AreaKnee, AreaAnkle}
}

// Known tells if the area is on the fixed list.
func (a Area) Known() bool { return slices.Contains(Areas(), a) }

// Name gives the display name of a known area.
func (a Area) Name() string {
	return map[Area]string{
		AreaShoulder: "Shoulder", AreaElbow: "Elbow", AreaWrist: "Wrist", AreaLowerBack: "Lower back",
		AreaHip: "Hip", AreaKnee: "Knee", AreaAnkle: "Ankle",
	}[a]
}

// MuscleGroup names one of the ten fixed muscle groups of D-210.
type MuscleGroup string

const (
	GroupChest      MuscleGroup = "chest"
	GroupBack       MuscleGroup = "back"
	GroupShoulders  MuscleGroup = "shoulders"
	GroupBiceps     MuscleGroup = "biceps"
	GroupTriceps    MuscleGroup = "triceps"
	GroupQuadriceps MuscleGroup = "quadriceps"
	GroupHamstrings MuscleGroup = "hamstrings"
	GroupGlutes     MuscleGroup = "glutes"
	GroupCalves     MuscleGroup = "calves"
	GroupCore       MuscleGroup = "core"
)

// MuscleGroups gives a new list of each group, in the order of D-210.
func MuscleGroups() []MuscleGroup {
	return []MuscleGroup{
		GroupChest, GroupBack, GroupShoulders, GroupBiceps, GroupTriceps,
		GroupQuadriceps, GroupHamstrings, GroupGlutes, GroupCalves, GroupCore,
	}
}

// Known tells if the group is on the fixed list.
func (g MuscleGroup) Known() bool { return slices.Contains(MuscleGroups(), g) }

// Name gives the display name of a known group.
func (g MuscleGroup) Name() string {
	return map[MuscleGroup]string{
		GroupChest: "Chest", GroupBack: "Back", GroupShoulders: "Shoulders", GroupBiceps: "Biceps",
		GroupTriceps: "Triceps", GroupQuadriceps: "Quadriceps", GroupHamstrings: "Hamstrings",
		GroupGlutes: "Glutes", GroupCalves: "Calves", GroupCore: "Core",
	}[g]
}

// TemplateID is the stable id of a goal template (D-210).
type TemplateID string

const (
	TemplateGeneralFitness TemplateID = "general_fitness"
	TemplateStrength       TemplateID = "strength"
)

// GoalTemplate is a goal template and the groups that it selects. The
// owner can change the selection after the pick of a template (D-210).
type GoalTemplate struct {
	ID     TemplateID
	Name   string
	Groups []MuscleGroup
}

// BodyTables holds the versioned tables of D-208 and D-210, and the goal
// templates. Areas gives the areas that each exercise of the catalog
// loads, and Groups gives its primary muscle groups. Each exercise has a
// row in both tables. A group row with no value is an exercise that no
// group selects. The research of each row is in
// docs/research/exercise-safety.md (D-38).
type BodyTables struct {
	Version   int
	Areas     map[ExerciseID][]Area
	Groups    map[ExerciseID][]MuscleGroup
	Templates []GoalTemplate
}

// Template gives the template with the id.
func (t BodyTables) Template(id TemplateID) (GoalTemplate, bool) {
	i := slices.IndexFunc(t.Templates, func(g GoalTemplate) bool { return g.ID == id })
	if i < 0 {
		return GoalTemplate{}, false
	}
	return t.Templates[i], true
}

// Loads tells if the exercise loads one or more of the areas. An
// exercise with no row loads each area, so a gap in the table removes
// an exercise and never keeps it (D-208).
func (t BodyTables) Loads(id ExerciseID, areas []Area) bool {
	row, ok := t.Areas[id]
	if !ok {
		return len(areas) > 0
	}
	for _, a := range areas {
		if slices.Contains(row, a) {
			return true
		}
	}
	return false
}

// Check reads the tables against the catalog: a row in both tables for
// each exercise and for no other id, known values with no duplicate in
// a row, values in the order of the fixed list, and templates with a
// unique id, a name, and one known group or more.
func (t BodyTables) Check(c Catalog) error {
	if t.Version < 1 {
		return invalid("body tables version %d: want 1 or more", t.Version)
	}
	for _, e := range c.Exercises {
		areas, okA := t.Areas[e.ID]
		groups, okG := t.Groups[e.ID]
		switch {
		case !okA:
			return invalid("exercise %q: no row in the area table", e.ID)
		case !okG:
			return invalid("exercise %q: no row in the group table", e.ID)
		}
		if err := inOrder(e.ID, "area", areas, Areas()); err != nil {
			return err
		}
		if err := inOrder(e.ID, "group", groups, MuscleGroups()); err != nil {
			return err
		}
	}
	for id := range t.Areas {
		if _, ok := c.Exercise(id); !ok {
			return invalid("area table: exercise %q is not in the catalog", id)
		}
	}
	for id := range t.Groups {
		if _, ok := c.Exercise(id); !ok {
			return invalid("group table: exercise %q is not in the catalog", id)
		}
	}
	seen := map[TemplateID]bool{}
	for i, g := range t.Templates {
		switch {
		case g.ID == "":
			return invalid("templates[%d]: empty id", i)
		case g.Name == "":
			return invalid("template %q: empty name", g.ID)
		case seen[g.ID]:
			return invalid("template %q: duplicate id", g.ID)
		case len(g.Groups) == 0:
			return invalid("template %q: no group", g.ID)
		}
		seen[g.ID] = true
		if err := inOrder(ExerciseID(g.ID), "group", g.Groups, MuscleGroups()); err != nil {
			return err
		}
	}
	return nil
}

// inOrder reads that each value of row is on the list, one time, in the
// order of the list.
func inOrder[T comparable](id ExerciseID, what string, row, list []T) error {
	last := -1
	for _, v := range row {
		i := slices.Index(list, v)
		switch {
		case i < 0:
			return invalid("%q: unknown %s %v", id, what, v)
		case i == last:
			return invalid("%q: %s %v two times", id, what, v)
		case i < last:
			return invalid("%q: %s %v out of order", id, what, v)
		}
		last = i
	}
	return nil
}
