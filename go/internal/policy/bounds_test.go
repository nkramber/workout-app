package policy

import (
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

func TestRanges(t *testing.T) {
	want := map[domain.ExerciseID]Range{
		"chest_press":              {8, 12},
		"leg_extension":            {8, 12},
		"db_lateral_raise":         {10, 20},
		"db_biceps_curl":           {10, 20},
		"db_hammer_curl":           {10, 20},
		"triceps_pulldown":         {10, 20},
		"lat_pulldown":             {8, 15},
		"db_flat_bench_press":      {8, 15},
		"db_one_arm_row":           {8, 15},
		"db_romanian_deadlift":     {8, 15},
		"db_goblet_squat":          {8, 15},
		"db_seated_shoulder_press": {8, 15},
	}
	for id, r := range want {
		if got := RepRange(exercise(t, id)); got != r {
			t.Errorf("RepRange(%s) = %v, want %v", id, got, r)
		}
	}
	// Each id of a rule is in the catalog, so a change of the catalog
	// can not drop a rule.
	for _, m := range []map[domain.ExerciseID]bool{smallFreeWeight, dumbbellPress} {
		for id := range m {
			exercise(t, id)
		}
	}
	for _, e := range domain.DefaultCatalog().Exercises {
		if e.Kind == domain.KindCardio {
			continue
		}
		reps, rir := RepRange(e), RIRRange(e)
		if reps.Min < RepLimits.Min || reps.Max > RepLimits.Max || reps.Min > reps.Max {
			t.Errorf("RepRange(%s) = %v: outside %v", e.ID, reps, RepLimits)
		}
		if rir.Min < 1 || rir.Max > 3 {
			t.Errorf("RIRRange(%s) = %v: outside 1 to 3 (D-37)", e.ID, rir)
		}
		wantRIR := Range{1, 3}
		if dumbbellPress[e.ID] {
			wantRIR = Range{2, 3}
		}
		if rir != wantRIR {
			t.Errorf("RIRRange(%s) = %v, want %v", e.ID, rir, wantRIR)
		}
		if !RestLimits.Has(DefaultRest(e)) {
			t.Errorf("DefaultRest(%s) = %d: outside %v", e.ID, DefaultRest(e), RestLimits)
		}
	}
	if DefaultRest(exercise(t, "leg_press")) != 180 || DefaultRest(exercise(t, "chest_press")) != 120 {
		t.Error("DefaultRest: want 180 s for the leg press and 120 s for the chest press")
	}
}

func TestCheck(t *testing.T) {
	lb := domain.Pounds
	in := machineInput(t, "chest_press", []domain.Load{lb(10), lb(14), lb(20), lb(25), lb(30)})
	in.History = []Outcome{outcome(target("chest_press", 2, 10, lb(25)), set(10, lb(25), 2), set(10, lb(25), 2))}
	in = dated(in)
	good := target("chest_press", 2, 10, lb(25))
	db := dumbbellInput(t, "db_incline_bench_press")
	db.Estimate = lb(20)
	cal := []domain.CalibrationSet{{Reps: 10, Load: lb(20)}}
	for _, tc := range []struct {
		name string
		in   Input
		edit func(p *domain.PlannedExercise)
		want []RuleID
	}{
		{"good", in, func(p *domain.PlannedExercise) {}, nil},
		{"lower load", in, func(p *domain.PlannedExercise) { p.Working[0].Load = lb(20) }, nil},
		{"reps 5", in, func(p *domain.PlannedExercise) { p.Working[0].Reps = 5 }, []RuleID{RuleRepBounds}},
		{"reps 21", in, func(p *domain.PlannedExercise) { p.Working[1].Reps = 21 }, []RuleID{RuleRepBounds}},
		{"rir 0", in, func(p *domain.PlannedExercise) { p.Working[0].RIR = 0 }, []RuleID{RuleRIRBounds}},
		{"rir 4", in, func(p *domain.PlannedExercise) { p.Working[0].RIR = 4 }, []RuleID{RuleRIRBounds}},
		{"rest 59", in, func(p *domain.PlannedExercise) { p.RestSeconds = 59 }, []RuleID{RuleRestBounds}},
		{"rest 181", in, func(p *domain.PlannedExercise) { p.RestSeconds = 181 }, []RuleID{RuleRestBounds}},
		{"no working set", in, func(p *domain.PlannedExercise) { p.Working = nil }, []RuleID{RuleRepBounds}},
		{"not on the stack", in, func(p *domain.PlannedExercise) { p.Working[0].Load = lb(15) }, []RuleID{RuleLoadAvailable}},
		{"zero load", in, func(p *domain.PlannedExercise) { p.Working[0].Load = 0 }, []RuleID{RuleLoadAvailable}},
		{"two steps", in, func(p *domain.PlannedExercise) { p.Working[1].Load = lb(30) }, []RuleID{RuleLoadCeiling}},
		{"calibration", in, func(p *domain.PlannedExercise) {
			p.Calibration = []domain.CalibrationSet{{Reps: 30, Load: lb(11)}}
		}, []RuleID{RuleCalibrationSet, RuleRepBounds, RuleLoadAvailable}},
		{"press rir 1", db, func(p *domain.PlannedExercise) {
			*p = target("db_incline_bench_press", 2, 10, lb(20))
			p.Calibration = cal
			p.Working[0].RIR = 1
		}, []RuleID{RuleRIRPress}},
		{"start", db, func(p *domain.PlannedExercise) {
			*p = target("db_incline_bench_press", 3, 10, lb(15))
			p.Calibration = cal
		}, nil},
		{"start two calibration sets", db, func(p *domain.PlannedExercise) {
			*p = target("db_incline_bench_press", 2, 10, lb(20))
			p.Calibration = []domain.CalibrationSet{{Reps: 10, Load: lb(15)}, {Reps: 10, Load: lb(20)}}
		}, []RuleID{RuleCalibrationSet}},
		{"start calibration reps", db, func(p *domain.PlannedExercise) {
			*p = target("db_incline_bench_press", 2, 10, lb(20))
			p.Calibration = []domain.CalibrationSet{{Reps: 12, Load: lb(20)}}
		}, []RuleID{RuleCalibrationSet}},
		{"history calibration reps", in, func(p *domain.PlannedExercise) {
			p.Calibration = []domain.CalibrationSet{{Reps: 8, Load: lb(25)}}
		}, []RuleID{RuleCalibrationSet, RuleCalibrationSet}},
		{"history calibration set", in, func(p *domain.PlannedExercise) {
			p.Calibration = []domain.CalibrationSet{{Reps: 10, Load: lb(25)}}
		}, []RuleID{RuleCalibrationSet}},
		{"history heavy calibration set", in, func(p *domain.PlannedExercise) {
			p.Calibration = []domain.CalibrationSet{{Reps: 10, Load: lb(30)}}
		}, []RuleID{RuleCalibrationSet}},
		{"start no calibration set", db, func(p *domain.PlannedExercise) {
			*p = target("db_incline_bench_press", 2, 10, lb(20))
		}, []RuleID{RuleCalibrationSet}},
		{"start above the estimate", db, func(p *domain.PlannedExercise) {
			*p = target("db_incline_bench_press", 2, 10, lb(50))
			p.Calibration = []domain.CalibrationSet{{Reps: 10, Load: lb(25)}}
		}, []RuleID{RuleLoadCeiling, RuleLoadCeiling, RuleLoadCeiling}},
	} {
		p := good
		p.Working = append([]domain.WorkingSet(nil), good.Working...)
		tc.edit(&p)
		v, err := Check(p, tc.in)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		var got []RuleID
		for _, x := range v {
			got = append(got, x.Rule)
		}
		if len(got) != len(tc.want) {
			t.Errorf("%s: Check = %v, want %v", tc.name, v, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s: Check = %v, want %v", tc.name, v, tc.want)
			}
		}
	}
	// The rounding rule: 12 lb is on the stack, but D-149 selects it for
	// no multiple of 5 lb.
	odd := machineInput(t, "chest_press", []domain.Load{lb(10), lb(12), lb(14)})
	odd.History = []Outcome{outcome(target("chest_press", 2, 10, lb(14)), set(10, lb(14), 2), set(10, lb(14), 2))}
	v, _ := Check(target("chest_press", 2, 10, lb(12)), dated(odd))
	if len(v) != 2 || v[0].Rule != RuleLoadRounding {
		t.Errorf("Check of 12 lb = %v, want %s", v, RuleLoadRounding)
	}
}

// The input refuses a dumbbell set above 100 lb (D-166), so no load of
// a dumbbell target is above it.
func TestCheckDumbbellMax(t *testing.T) {
	in := dumbbellInput(t, "db_goblet_squat")
	in.Entry.Dumbbells = &domain.DumbbellSet{Lightest: domain.Pounds(100), Heaviest: domain.Pounds(100), Step: domain.Pounds(5)}
	if _, err := Check(target("db_goblet_squat", 1, 10, domain.Pounds(100)), in); err != nil {
		t.Fatalf("Check at 100 lb: %v", err)
	}
	in.Entry.Dumbbells = &domain.DumbbellSet{Lightest: domain.Pounds(105), Heaviest: domain.Pounds(105), Step: domain.Pounds(5)}
	if _, err := Check(target("db_goblet_squat", 1, 10, domain.Pounds(105)), in); !errors.Is(err, ErrInput) {
		t.Fatalf("Check with a 105 lb dumbbell set: %v, want ErrInput", err)
	}
}

func TestInputErrors(t *testing.T) {
	lb := domain.Pounds
	good := machineInput(t, "chest_press", stack(10, 100, 5))
	good.History = []Outcome{outcome(target("chest_press", 2, 10, lb(25)), set(10, lb(25), 2))}
	good = dated(good)
	for _, tc := range []struct {
		name string
		edit func(in *Input)
		want error
	}{
		{"no today", func(in *Input) { in.Today = "" }, ErrInput},
		{"bad today", func(in *Input) { in.Today = "2026-13-01" }, ErrInput},
		{"no date", func(in *Input) { in.History[0].Date = "" }, ErrInput},
		{"date after today", func(in *Input) { in.Today = "2026-08-31" }, ErrInput},
		{"dates out of order", func(in *Input) {
			in.History = append(in.History, in.History[0])
			in.History[0].Date = "2026-09-02"
		}, ErrInput},
		{"bad calibration set", func(in *Input) {
			in.History[0].Target.Calibration = []domain.CalibrationSet{{Reps: 0, Load: lb(25)}}
		}, ErrInput},
		{"estimate below 0", func(in *Input) { in.Estimate = -1 }, ErrInput},
		{"estimate above the stack", func(in *Input) { in.History = nil; in.Estimate = lb(105) }, ErrInput},
		{"estimate below the stack", func(in *Input) { in.History = nil; in.Estimate = lb(5) }, ErrInput},
		{"no exercise", func(in *Input) { in.Exercise = domain.Exercise{} }, ErrInput},
		{"cardio", func(in *Input) { in.Exercise = exercise(t, "treadmill"); in.Entry.Machine = "treadmill" }, ErrInput},
		{"other machine", func(in *Input) { in.Entry.Machine = "leg_press" }, ErrInput},
		{"no weights", func(in *Input) { in.Entry.Weights = nil }, ErrInput},
		{"weights out of order", func(in *Input) { in.Entry.Weights = []domain.Load{lb(20), lb(10)} }, ErrInput},
		{"no dumbbells", func(in *Input) { *in = dumbbellInput(t, "db_hammer_curl"); in.Entry.Dumbbells = nil }, ErrInput},
		{"other exercise", func(in *Input) { in.History[0].Log.Exercise = "leg_press" }, ErrInput},
		{"no working set", func(in *Input) { in.History[0].Target.Working = nil }, ErrInput},
		{"bad set", func(in *Input) { in.History[0].Log.Sets[0].Weight = 0 }, ErrInput},
		{"skipped with sets", func(in *Input) { in.History[0].Log.Skipped = true }, ErrInput},
	} {
		in := good
		in.History = []Outcome{good.History[0]}
		in.History[0].Log.Sets = append([]domain.SetLog(nil), good.History[0].Log.Sets...)
		in.History[0].Target.Working = append([]domain.WorkingSet(nil), good.History[0].Target.Working...)
		tc.edit(&in)
		if _, err := Next(in); !errors.Is(err, tc.want) {
			t.Errorf("%s: Next error %v, want %v", tc.name, err, tc.want)
		}
	}
	if _, err := Check(target("leg_press", 2, 10, lb(25)), good); !errors.Is(err, ErrInput) {
		t.Errorf("Check of another exercise: %v, want ErrInput", err)
	}
}

var (
	decisionID = regexp.MustCompile(`^D-\d+$`)
	evidenceID = regexp.MustCompile(`^EV-\d+$`)
)

// Each rule has a unique id, a text, and sources. Each source is a
// decision of docs/decisions.md or an evidence id of the safety
// research (D-38).
func TestRules(t *testing.T) {
	decisions, err := os.ReadFile("../../../docs/decisions.md")
	if err != nil {
		t.Fatal(err)
	}
	research, err := os.ReadFile("../../../docs/research/exercise-safety.md")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[RuleID]bool{}
	for _, r := range Rules() {
		if r.ID == "" || r.Text == "" || len(r.Sources) == 0 {
			t.Errorf("rule %q: want an id, a text, and sources", r.ID)
		}
		if seen[r.ID] {
			t.Errorf("rule %q: duplicate id", r.ID)
		}
		seen[r.ID] = true
		hasDecision := false
		for _, s := range r.Sources {
			switch {
			case decisionID.MatchString(s):
				hasDecision = true
				if !regexp.MustCompile(`(?m)^\| ` + s + `[ |(]`).Match(decisions) {
					t.Errorf("rule %q: %s is not in docs/decisions.md", r.ID, s)
				}
			case evidenceID.MatchString(s):
				if !regexp.MustCompile(`(?m)^\| ` + s + ` \|`).Match(research) {
					t.Errorf("rule %q: %s is not in the evidence register", r.ID, s)
				}
			default:
				t.Errorf("rule %q: source %q is not a D- or EV- id", r.ID, s)
			}
		}
		if !hasDecision {
			t.Errorf("rule %q: no owner decision", r.ID)
		}
	}
	// Rules gives a copy, so a caller can not change the rules.
	Rules()[0].Sources[0] = "changed"
	if Rules()[0].Sources[0] == "changed" {
		t.Error("Rules gives the table itself, want a copy")
	}
}
