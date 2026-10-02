package profile

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

var (
	catalog = domain.DefaultCatalog()
	tables  = domain.DefaultBodyTables()
)

// valid gives a synthetic profile at the low bound of each number.
func valid() Profile {
	return Profile{
		Experience:   Intermediate,
		Template:     domain.TemplateGeneralFitness,
		Groups:       domain.MuscleGroups(),
		FreeText:     "More upper body.",
		InjuredAreas: []domain.Area{domain.AreaKnee},
		InjuryText:   "Old knee strain.",
		AgeYears:     MinAge,
		HeightIn:     MinHeightIn,
		WeightLb:     MinWeightLb,
		Cardio:       []domain.ExerciseID{"upright_bike", "rowing_machine"},
		TrainingDays: MinTrainingDays,
	}
}

func TestValidProfilePasses(t *testing.T) {
	high := valid()
	high.Experience, high.Template = Advanced, domain.TemplateStrength
	high.AgeYears, high.HeightIn, high.WeightLb, high.TrainingDays = MaxAge, MaxHeightIn, MaxWeightLb, MaxTrainingDays
	high.FreeText, high.InjuryText = strings.Repeat("é", MaxTextRunes), strings.Repeat("x", MaxTextRunes)
	empty := valid()
	empty.FreeText, empty.InjuryText, empty.InjuredAreas, empty.Cardio = "", "", nil, nil
	for name, p := range map[string]Profile{"low bounds": valid(), "high bounds": high, "empty lists and texts": empty} {
		if err := p.Check(catalog, tables); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestCheckRefuses gives one field outside its bound or its list in each
// case. The error names the field, and never the value of a number or
// of a text (D-80).
func TestCheckRefuses(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Profile)
		want   string
	}{
		{"beginner", func(p *Profile) { p.Experience = "beginner" }, "experience"},
		{"no experience", func(p *Profile) { p.Experience = "" }, "experience"},
		{"unknown template", func(p *Profile) { p.Template = "bulk" }, "goal template"},
		{"no group", func(p *Profile) { p.Groups = nil }, "muscle groups"},
		{"unknown group", func(p *Profile) { p.Groups = []domain.MuscleGroup{"neck"} }, "muscle group"},
		{"group two times", func(p *Profile) { p.Groups = []domain.MuscleGroup{"chest", "chest"} }, "two times"},
		{"unknown area", func(p *Profile) { p.InjuredAreas = []domain.Area{"neck"} }, "injured area"},
		{"area two times", func(p *Profile) { p.InjuredAreas = []domain.Area{"knee", "knee"} }, "two times"},
		{"resistance exercise as cardio", func(p *Profile) { p.Cardio = []domain.ExerciseID{"leg_press"} }, "cardio exercise"},
		{"unknown cardio", func(p *Profile) { p.Cardio = []domain.ExerciseID{"swim"} }, "cardio exercise"},
		{"cardio two times", func(p *Profile) { p.Cardio = []domain.ExerciseID{"treadmill", "treadmill"} }, "two times"},
		{"age 17", func(p *Profile) { p.AgeYears = MinAge - 1 }, "age"},
		{"age 91", func(p *Profile) { p.AgeYears = MaxAge + 1 }, "age"},
		{"height 47", func(p *Profile) { p.HeightIn = MinHeightIn - 1 }, "height"},
		{"height 97", func(p *Profile) { p.HeightIn = MaxHeightIn + 1 }, "height"},
		{"weight 79", func(p *Profile) { p.WeightLb = MinWeightLb - 1 }, "weight"},
		{"weight 501", func(p *Profile) { p.WeightLb = MaxWeightLb + 1 }, "weight"},
		{"1 training day", func(p *Profile) { p.TrainingDays = MinTrainingDays - 1 }, "training days"},
		{"5 training days", func(p *Profile) { p.TrainingDays = MaxTrainingDays + 1 }, "training days"},
		{"long free text", func(p *Profile) { p.FreeText = strings.Repeat("x", MaxTextRunes+1) }, "free text"},
		{"long injury text", func(p *Profile) { p.InjuryText = strings.Repeat("x", MaxTextRunes+1) }, "injury text"},
		{"space at an end", func(p *Profile) { p.FreeText = "Legs " }, "free text"},
		{"not UTF-8", func(p *Profile) { p.InjuryText = "\xff" }, "injury text"},
	}
	for _, c := range cases {
		p := valid()
		c.change(&p)
		err := p.Check(catalog, tables)
		if !errors.Is(err, domain.ErrInvalid) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Check = %v, want ErrInvalid that names %q", c.name, err, c.want)
			continue
		}
		// A number case names its value, such as "age 17". The error
		// gives the bound alone.
		if v, ok := strings.CutPrefix(c.name, c.want+" "); ok && strings.Contains(err.Error(), v) {
			t.Errorf("%s: the error %q holds the value %q", c.name, err, v)
		}
		if strings.Contains(err.Error(), "Legs") {
			t.Errorf("%s: the error %q holds the text", c.name, err)
		}
	}
}

func TestNormalize(t *testing.T) {
	p := valid()
	p.FreeText, p.InjuryText = "  More upper body. ", "\tOld knee strain.\n"
	p.Groups = []domain.MuscleGroup{"core", "chest", "neck"}
	p.InjuredAreas = []domain.Area{"knee", "shoulder"}
	p.Cardio = []domain.ExerciseID{"stair_climber", "treadmill"}
	in := p.clone()
	got := p.Normalize(catalog)
	if got.FreeText != "More upper body." || got.InjuryText != "Old knee strain." {
		t.Fatalf("texts = %q, %q", got.FreeText, got.InjuryText)
	}
	if !slices.Equal(got.Groups, []domain.MuscleGroup{"chest", "core", "neck"}) ||
		!slices.Equal(got.InjuredAreas, []domain.Area{"shoulder", "knee"}) ||
		!slices.Equal(got.Cardio, []domain.ExerciseID{"treadmill", "stair_climber"}) {
		t.Fatalf("lists = %v, %v, %v", got.Groups, got.InjuredAreas, got.Cardio)
	}
	if !reflect.DeepEqual(p, in) {
		t.Fatalf("Normalize changed its input: %+v", p)
	}
}

// TestPlannerInputHoldsD209Alone reads the fields of PlannerInput, so a
// new field, such as the age or the injury text, fails the test (D-209).
func TestPlannerInputHoldsD209Alone(t *testing.T) {
	want := []string{"Experience", "Template", "Groups", "FreeText", "Cardio", "Sessions", "Exercises"}
	typ := reflect.TypeFor[PlannerInput]()
	var got []string
	for i := range typ.NumField() {
		got = append(got, typ.Field(i).Name)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("PlannerInput fields = %v, want %v", got, want)
	}

	p := valid()
	in := ForPlan(p, catalog, tables)
	if in.Experience != p.Experience || in.Template != p.Template || !slices.Equal(in.Groups, p.Groups) ||
		in.FreeText != p.FreeText || in.Sessions != p.TrainingDays {
		t.Fatalf("ForPlan = %+v", in)
	}
	in.Groups[0] = "back"
	if p.Groups[0] != domain.GroupChest {
		t.Fatal("ForPlan shares memory with the profile")
	}
}

// TestPlannerInputHasNoInjuredArea reads the input for no area, for each
// area alone, and for all areas. No exercise of the input loads an
// injured area, and each removed exercise loads one (D-208).
func TestPlannerInputHasNoInjuredArea(t *testing.T) {
	sets := [][]domain.Area{nil, domain.Areas()}
	for _, a := range domain.Areas() {
		sets = append(sets, []domain.Area{a})
	}
	var cardio []domain.ExerciseID
	for _, e := range catalog.ExercisesOfKind(domain.KindCardio) {
		cardio = append(cardio, e.ID)
	}
	for _, areas := range sets {
		p := valid()
		p.InjuredAreas, p.Cardio = areas, cardio
		in := ForPlan(p, catalog, tables)
		kept := append(slices.Clone(in.Exercises), in.Cardio...)
		for _, e := range catalog.Exercises {
			loads := slices.ContainsFunc(tables.Areas[e.ID], func(a domain.Area) bool { return slices.Contains(areas, a) })
			if loads == slices.Contains(kept, e.ID) {
				t.Errorf("areas %v: exercise %q loads an injured area = %v, kept = %v", areas, e.ID, loads, !loads)
			}
		}
		for _, id := range in.Exercises {
			if e, _ := catalog.Exercise(id); e.Kind == domain.KindCardio {
				t.Errorf("areas %v: cardio exercise %q is in Exercises", areas, id)
			}
		}
		if len(areas) == 0 && len(in.Exercises)+len(in.Cardio) != len(catalog.Exercises) {
			t.Errorf("no injury: %d exercises kept, want %d", len(in.Exercises)+len(in.Cardio), len(catalog.Exercises))
		}
	}
}

// TestPlannerInputKeepsTheCardioPreference keeps only the cardio that
// the owner picked.
func TestPlannerInputKeepsTheCardioPreference(t *testing.T) {
	p := valid()
	p.InjuredAreas = nil
	if in := ForPlan(p, catalog, tables); !slices.Equal(in.Cardio, p.Cardio) {
		t.Fatalf("Cardio = %v, want %v", in.Cardio, p.Cardio)
	}
	p.Cardio = nil
	if in := ForPlan(p, catalog, tables); len(in.Cardio) != 0 {
		t.Fatalf("no cardio preference gives %v", in.Cardio)
	}
}

// TestGapInTheTableRemoves gives an exercise with no row of the area
// table. ForPlan removes it for each injury, so a gap never keeps an
// exercise of an injured area.
func TestGapInTheTableRemoves(t *testing.T) {
	gap := domain.BodyTables{Version: 1, Areas: map[domain.ExerciseID][]domain.Area{}}
	p := valid()
	p.InjuredAreas, p.Cardio = []domain.Area{domain.AreaWrist}, []domain.ExerciseID{"treadmill"}
	if in := ForPlan(p, catalog, gap); len(in.Exercises)+len(in.Cardio) != 0 {
		t.Fatalf("ForPlan with no rows = %+v, want no exercise", in)
	}
	p.InjuredAreas = nil
	if in := ForPlan(p, catalog, gap); len(in.Cardio) != 1 {
		t.Fatalf("ForPlan with no injury = %+v, want each exercise", in)
	}
}
