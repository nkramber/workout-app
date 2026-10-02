// Package profile holds the one profile of the owner and its store (work
// area 5.1). The profile holds the onboarding inputs of D-41, D-42, and
// D-208 to D-211. The load estimates of D-41 live in the inventory
// (D-192).
//
// A check error matches domain.ErrInvalid. An error names the field, the
// ids, and the numbers alone, and never the age, the height, the weight,
// or a text of the profile (D-80).
package profile

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// Experience is the training experience of the owner (D-30). People with
// no weight training are out of scope, so no value names them (D-33).
type Experience string

const (
	Intermediate Experience = "intermediate"
	Advanced     Experience = "advanced"
)

// Experiences gives a new list of each experience value.
func Experiences() []Experience { return []Experience{Intermediate, Advanced} }

// Name gives the display name of a known value.
func (e Experience) Name() string {
	return map[Experience]string{Intermediate: "Intermediate", Advanced: "Advanced"}[e]
}

// The bounds of each field. Pounds and inches are the one unit (D-28).
const (
	MinAge, MaxAge                   = 18, 90
	MinHeightIn, MaxHeightIn         = 48, 96
	MinWeightLb, MaxWeightLb         = 80, 500
	MinTrainingDays, MaxTrainingDays = 2, 4 // D-211
	// MaxTextRunes is the length limit of the free text and of the
	// injury text, in characters.
	MaxTextRunes = 500
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalid, fmt.Sprintf(format, args...))
}

// Profile is the one profile of the owner. Each list holds each value
// one time, in the order of its fixed list or of the catalog.
type Profile struct {
	Experience   Experience
	Template     domain.TemplateID
	Groups       []domain.MuscleGroup
	FreeText     string
	InjuredAreas []domain.Area
	InjuryText   string
	AgeYears     int
	HeightIn     int
	WeightLb     int
	// Cardio is the cardio preference: the cardio exercises that the
	// owner likes. An empty list means no optional cardio.
	Cardio       []domain.ExerciseID
	TrainingDays int
}

// Normalize gives a copy of p with the spaces at each end of each text
// removed, and each list in the order of its fixed list or of the
// catalog. It keeps an unknown value and a duplicate, so Check can
// refuse them.
func (p Profile) Normalize(c domain.Catalog) Profile {
	out := p.clone()
	out.FreeText = strings.TrimSpace(out.FreeText)
	out.InjuryText = strings.TrimSpace(out.InjuryText)
	sortBy(out.Groups, domain.MuscleGroups())
	sortBy(out.InjuredAreas, domain.Areas())
	var ids []domain.ExerciseID
	for _, e := range c.Exercises {
		ids = append(ids, e.ID)
	}
	sortBy(out.Cardio, ids)
	return out
}

// sortBy puts the values of row in the order of list. An unknown value
// goes to the end.
func sortBy[T comparable](row, list []T) {
	rank := func(v T) int {
		if i := slices.Index(list, v); i >= 0 {
			return i
		}
		return len(list)
	}
	slices.SortStableFunc(row, func(a, b T) int { return rank(a) - rank(b) })
}

// Check reads each field against its bound and each value against its
// fixed list: the experience, a template of the tables, one group or
// more, the areas, the cardio exercises of the catalog, the numbers, and
// the texts. A text has no space at either end.
func (p Profile) Check(c domain.Catalog, t domain.BodyTables) error {
	if !slices.Contains(Experiences(), p.Experience) {
		return invalid("profile experience: unknown value %q", p.Experience)
	}
	if _, ok := t.Template(p.Template); !ok {
		return invalid("profile goal template: unknown value %q", p.Template)
	}
	if len(p.Groups) == 0 {
		return invalid("profile muscle groups: want 1 or more")
	}
	if err := unique("muscle group", p.Groups, domain.MuscleGroup.Known); err != nil {
		return err
	}
	if err := unique("injured area", p.InjuredAreas, domain.Area.Known); err != nil {
		return err
	}
	cardio := func(id domain.ExerciseID) bool {
		e, ok := c.Exercise(id)
		return ok && e.Kind == domain.KindCardio
	}
	if err := unique("cardio exercise", p.Cardio, cardio); err != nil {
		return err
	}
	for _, n := range []struct {
		field     string
		v, lo, hi int
	}{
		{"age", p.AgeYears, MinAge, MaxAge},
		{"height", p.HeightIn, MinHeightIn, MaxHeightIn},
		{"weight", p.WeightLb, MinWeightLb, MaxWeightLb},
		{"training days", p.TrainingDays, MinTrainingDays, MaxTrainingDays},
	} {
		// The value can be the age, the height, or the weight, so the
		// error gives the bound alone (D-80).
		if n.v < n.lo || n.v > n.hi {
			return invalid("profile %s: outside %d to %d", n.field, n.lo, n.hi)
		}
	}
	for _, s := range []struct{ field, v string }{{"free text", p.FreeText}, {"injury text", p.InjuryText}} {
		if err := checkText(s.field, s.v); err != nil {
			return err
		}
	}
	return nil
}

// unique reads that each value is known and comes one time. An unknown
// value is an id of a fixed list, so the error can name it.
func unique[T comparable](field string, row []T, known func(T) bool) error {
	seen := map[T]bool{}
	for _, v := range row {
		if !known(v) {
			return invalid("profile %s: unknown value %q", field, fmt.Sprint(v))
		}
		if seen[v] {
			return invalid("profile %s: %q two times", field, fmt.Sprint(v))
		}
		seen[v] = true
	}
	return nil
}

func checkText(field, s string) error {
	if !utf8.ValidString(s) {
		return invalid("profile %s: the text is not UTF-8", field)
	}
	if strings.TrimSpace(s) != s {
		return invalid("profile %s: a space at an end of the text", field)
	}
	if k := utf8.RuneCountInString(s); k > MaxTextRunes {
		return invalid("profile %s: %d characters, want %d or fewer", field, k, MaxTextRunes)
	}
	return nil
}

func (p Profile) clone() Profile {
	p.Groups = slices.Clone(p.Groups)
	p.InjuredAreas = slices.Clone(p.InjuredAreas)
	p.Cardio = slices.Clone(p.Cardio)
	return p
}
