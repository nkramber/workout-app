package domain

import (
	"errors"
	"maps"
	"slices"
	"testing"
)

// TestEachExerciseIsInBothTables finds each exercise of the catalog in
// the area table of D-208 and in the group table of D-210, and no other
// id.
func TestEachExerciseIsInBothTables(t *testing.T) {
	c, b := DefaultCatalog(), DefaultBodyTables()
	for _, e := range c.Exercises {
		if _, ok := b.Areas[e.ID]; !ok {
			t.Errorf("exercise %q: no row in the area table", e.ID)
		}
		if _, ok := b.Groups[e.ID]; !ok {
			t.Errorf("exercise %q: no row in the group table", e.ID)
		}
	}
	if len(b.Areas) != len(c.Exercises) || len(b.Groups) != len(c.Exercises) {
		t.Errorf("rows = %d areas and %d groups, want %d each", len(b.Areas), len(b.Groups), len(c.Exercises))
	}
	if err := b.Check(c); err != nil {
		t.Fatal(err)
	}
}

// TestEmptyRows reads the rows with no value. Each exercise loads one
// area or more (D-218). Hip adduction and the cardio exercises alone
// have no primary group (D-219).
func TestEmptyRows(t *testing.T) {
	c, b := DefaultCatalog(), DefaultBodyTables()
	for _, e := range c.Exercises {
		if len(b.Areas[e.ID]) == 0 {
			t.Errorf("exercise %q loads no area", e.ID)
		}
		if want := e.Kind == KindCardio || e.ID == "hip_adduction"; (len(b.Groups[e.ID]) == 0) != want {
			t.Errorf("exercise %q: primary groups %v", e.ID, b.Groups[e.ID])
		}
	}
	if g, _ := b.Template(TemplateStrength); !slices.Equal(g.Groups, []MuscleGroup{GroupChest, GroupBack, GroupShoulders, GroupQuadriceps, GroupHamstrings, GroupGlutes}) {
		t.Errorf("Strength = %v, want the six groups of D-220", g.Groups)
	}
}

func TestDefaultBodyTablesAreCopies(t *testing.T) {
	a := DefaultBodyTables()
	a.Areas["leg_press"][0] = AreaWrist
	a.Groups["leg_press"] = nil
	a.Templates[1].Groups[0] = GroupCore
	if b := DefaultBodyTables(); b.Areas["leg_press"][0] == AreaWrist || b.Groups["leg_press"] == nil || b.Templates[1].Groups[0] == GroupCore {
		t.Fatal("DefaultBodyTables shares memory between callers")
	}
}

func TestTemplates(t *testing.T) {
	b := DefaultBodyTables()
	if ids := []TemplateID{b.Templates[0].ID, b.Templates[1].ID}; len(b.Templates) != 2 || ids[0] != TemplateGeneralFitness || ids[1] != TemplateStrength {
		t.Fatalf("templates = %+v", b.Templates)
	}
	if g, _ := b.Template(TemplateGeneralFitness); !slices.Equal(g.Groups, MuscleGroups()) {
		t.Fatalf("General fitness = %v, want each group (D-210)", g.Groups)
	}
	if _, ok := b.Template("bulk"); ok {
		t.Fatal("an unknown template exists")
	}
}

func TestFixedLists(t *testing.T) {
	if got := Areas(); len(got) != 7 || !AreaLowerBack.Known() || Area("neck").Known() || AreaLowerBack.Name() != "Lower back" {
		t.Fatalf("areas = %v", got)
	}
	if got := MuscleGroups(); len(got) != 10 || !GroupCalves.Known() || MuscleGroup("neck").Known() || GroupQuadriceps.Name() != "Quadriceps" {
		t.Fatalf("groups = %v", got)
	}
	for _, a := range Areas() {
		if a.Name() == "" {
			t.Errorf("area %q: no name", a)
		}
	}
	for _, g := range MuscleGroups() {
		if g.Name() == "" {
			t.Errorf("group %q: no name", g)
		}
	}
}

func TestLoads(t *testing.T) {
	b := BodyTables{Areas: map[ExerciseID][]Area{"leg_extension": {AreaKnee}, "abdominal_crunch": nil}}
	cases := []struct {
		id    ExerciseID
		areas []Area
		want  bool
	}{
		{"leg_extension", []Area{AreaKnee}, true},
		{"leg_extension", []Area{AreaShoulder, AreaKnee}, true},
		{"leg_extension", []Area{AreaShoulder}, false},
		{"leg_extension", nil, false},
		{"abdominal_crunch", Areas(), false},
		{"no_row", []Area{AreaAnkle}, true},
		{"no_row", nil, false},
	}
	for _, c := range cases {
		if got := b.Loads(c.id, c.areas); got != c.want {
			t.Errorf("Loads(%q, %v) = %v, want %v", c.id, c.areas, got, c.want)
		}
	}
}

func TestBodyTablesCheckRefuses(t *testing.T) {
	c := DefaultCatalog()
	cases := map[string]func(*BodyTables){
		"version 0":            func(b *BodyTables) { b.Version = 0 },
		"no area row":          func(b *BodyTables) { delete(b.Areas, "chest_press") },
		"no group row":         func(b *BodyTables) { delete(b.Groups, "chest_press") },
		"area row of no id":    func(b *BodyTables) { b.Areas["squat"] = nil },
		"group row of no id":   func(b *BodyTables) { b.Groups["squat"] = []MuscleGroup{GroupQuadriceps} },
		"unknown area":         func(b *BodyTables) { b.Areas["chest_press"] = []Area{"neck"} },
		"area two times":       func(b *BodyTables) { b.Areas["leg_extension"] = []Area{AreaKnee, AreaKnee} },
		"areas out of order":   func(b *BodyTables) { b.Areas["leg_press"] = []Area{AreaKnee, AreaHip} },
		"unknown group":        func(b *BodyTables) { b.Groups["chest_press"] = []MuscleGroup{"pecs"} },
		"template of no group": func(b *BodyTables) { b.Templates[1].Groups = nil },
		"template two times":   func(b *BodyTables) { b.Templates[1].ID = b.Templates[0].ID },
		"template of no name":  func(b *BodyTables) { b.Templates[0].Name = "" },
	}
	for _, name := range slices.Sorted(maps.Keys(cases)) {
		b := DefaultBodyTables()
		cases[name](&b)
		if err := b.Check(c); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: Check = %v, want ErrInvalid", name, err)
		}
	}
}
