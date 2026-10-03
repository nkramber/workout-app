package policy

import (
	"fmt"
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// Range is a closed range of integers.
type Range struct {
	Min, Max int
}

// Has tells whether n is in the range.
func (r Range) Has(n int) bool { return n >= r.Min && n <= r.Max }

func (r Range) clamp(n int) int { return min(max(n, r.Min), r.Max) }

// RepLimits are the hard limits of the reps of a target set (D-167).
var RepLimits = Range{6, 20}

// RestLimits are the limits of the rest of a target, in seconds (D-172).
var RestLimits = Range{60, 180}

// RestDefault is the rest of each exercise, in seconds (D-172, D-279).
const RestDefault = 60

// The exercises with their own rep range or reps in reserve (D-167,
// D-171).
var (
	smallFreeWeight = map[domain.ExerciseID]bool{
		"db_lateral_raise": true,
		"db_biceps_curl":   true,
		"db_hammer_curl":   true,
		"triceps_pulldown": true,
	}
	dumbbellPress = map[domain.ExerciseID]bool{
		"db_flat_bench_press":      true,
		"db_incline_bench_press":   true,
		"db_seated_shoulder_press": true,
	}
)

// RepRange gives the rep range of an exercise (D-167): 8 to 12 on a
// machine, 10 to 20 for a small single-joint dumbbell or cable
// exercise, and 8 to 15 for each other dumbbell or cable exercise
// (REC-19). Progression keeps a target inside this range.
func RepRange(e domain.Exercise) Range {
	switch {
	case e.Kind == domain.KindMachine:
		return Range{8, 12}
	case smallFreeWeight[e.ID]:
		return Range{10, 20}
	default:
		return Range{8, 15}
	}
}

// RIRRange gives the reps in reserve of a target set: 1 to 3 (D-37),
// and 2 to 3 for a dumbbell press (D-171).
func RIRRange(e domain.Exercise) Range {
	if isPress(e) {
		return Range{2, 3}
	}
	return Range{1, 3}
}

func isPress(e domain.Exercise) bool { return dumbbellPress[e.ID] }

// DefaultRest gives the rest of an exercise in seconds. Each exercise
// rests 60 seconds, the leg press too (D-279).
func DefaultRest(domain.Exercise) int { return RestDefault }

// Violation is one rule that a proposal breaks. Where names the place,
// such as "working[2]". Detail holds ids and numbers alone (D-80).
type Violation struct {
	Rule   RuleID
	Where  string
	Detail string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s %s: %s", v.Rule, v.Where, v.Detail)
}

// Check reads a proposed target of the exercise of an input against
// each bound, and gives each violation. No violation means that the
// policy accepts the proposal. A proposal with a violation is refused,
// and the owner never sees it (D-23).
//
// The load of each set is at most the load of the next target of the
// policy (RuleLoadCeiling). Outside a calibration session, a proposal
// has no more working sets than that target, and a set at the load of
// the target set of the same position has no more reps and no fewer
// reps in reserve (RuleEffortCeiling, D-186). In a calibration session,
// the reps can go up and the reps in reserve are practice (D-177). With no history, that target is the start of
// D-150. A proposal has a calibration set only when the target of the
// policy has one, and then it has exactly one, at the reps and the load
// of its first working set (RuleCalibrationSet, D-181). In the
// first sessions
// after a break, each working set stops at 3 reps in reserve, and the
// proposal has no more sets than the target (RuleBreakFirst).
func Check(p domain.PlannedExercise, in Input) ([]Violation, error) {
	if err := in.check(); err != nil {
		return nil, err
	}
	if p.Exercise != in.Exercise.ID {
		return nil, inputError("proposal exercise %q: want %q", p.Exercise, in.Exercise.ID)
	}
	d, err := Next(in)
	if err != nil {
		return nil, err
	}
	ceiling := d.Target.Working
	first := in.firstSessions(in.pause())

	var out []Violation
	add := func(r RuleID, where, format string, args ...any) {
		out = append(out, Violation{r, where, fmt.Sprintf(format, args...)})
	}
	available := in.Entry.Available()
	checkLoad := func(where string, l domain.Load) {
		switch {
		case !slices.Contains(available, l):
			add(RuleLoadAvailable, where, "load %s: not an available weight of machine %q", l, in.Entry.Machine)
		case !Valid(l, available):
			add(RuleLoadRounding, where, "load %s: not a multiple of 5 lb or the weight that D-149 selects", l)
		}
	}

	if !RestLimits.Has(p.RestSeconds) {
		add(RuleRestBounds, "rest", "rest %d s: want %d to %d s", p.RestSeconds, RestLimits.Min, RestLimits.Max)
	}
	if len(p.Working) == 0 {
		add(RuleRepBounds, "working", "no working set")
	}
	effort := len(d.Target.Calibration) == 0
	if effort && !first && len(p.Working) > len(d.Target.Working) {
		add(RuleEffortCeiling, "working", "%d sets: want %d or fewer", len(p.Working), len(d.Target.Working))
	}
	if first && len(p.Working) > len(d.Target.Working) {
		add(RuleBreakFirst, "working", "%d sets: want %d or fewer in the first sessions after a break", len(p.Working), len(d.Target.Working))
	}
	switch {
	case len(d.Target.Calibration) > 0 && len(p.Calibration) == 0:
		add(RuleCalibrationSet, "calibration", "no calibration set in a calibration session")
	case len(d.Target.Calibration) == 0 && len(p.Calibration) > 0:
		add(RuleCalibrationSet, "calibration", "a calibration set outside a calibration session")
	}
	if len(p.Calibration) > 1 {
		add(RuleCalibrationSet, "calibration", "%d calibration sets: want 1", len(p.Calibration))
	}
	rir := RIRRange(in.Exercise)
	for i, s := range p.Calibration {
		where := fmt.Sprintf("calibration[%d]", i)
		switch {
		case !RepLimits.Has(s.Reps):
			add(RuleRepBounds, where, "reps %d: want %d to %d", s.Reps, RepLimits.Min, RepLimits.Max)
		case len(p.Working) > 0 && s.Reps != p.Working[0].Reps:
			add(RuleCalibrationSet, where, "reps %d: want %d, the reps of the first working set", s.Reps, p.Working[0].Reps)
		}
		checkLoad(where, s.Load)
		if len(d.Target.Calibration) > 0 && s.Load > d.Target.Calibration[0].Load {
			add(RuleLoadCeiling, where, "load %s: want %s or less", s.Load, d.Target.Calibration[0].Load)
		}
		if len(p.Working) > 0 && s.Load != p.Working[0].Load {
			add(RuleCalibrationSet, where, "load %s: want %s, the load of the first working set", s.Load, p.Working[0].Load)
		}
	}
	for i, s := range p.Working {
		where := fmt.Sprintf("working[%d]", i)
		bounded := RepLimits.Has(s.Reps) && rir.Has(s.RIR) && !(first && s.RIR < 3)
		if !RepLimits.Has(s.Reps) {
			add(RuleRepBounds, where, "reps %d: want %d to %d", s.Reps, RepLimits.Min, RepLimits.Max)
		}
		switch {
		case !rir.Has(s.RIR):
			r := RuleRIRBounds
			if isPress(in.Exercise) {
				r = RuleRIRPress
			}
			add(r, where, "rir %d: want %d to %d", s.RIR, rir.Min, rir.Max)
		case first && s.RIR < 3:
			add(RuleBreakFirst, where, "rir %d: want 3 in the first sessions after a break", s.RIR)
		}
		checkLoad(where, s.Load)
		c := ceiling[min(i, len(ceiling)-1)]
		switch {
		case s.Load > c.Load:
			add(RuleLoadCeiling, where, "load %s: want %s or less", s.Load, c.Load)
		case !effort || !bounded || s.Load < c.Load:
		case s.Reps > c.Reps:
			add(RuleEffortCeiling, where, "reps %d at %s: want %d or fewer", s.Reps, s.Load, c.Reps)
		case s.RIR < c.RIR:
			add(RuleEffortCeiling, where, "rir %d at %s: want %d or more", s.RIR, s.Load, c.RIR)
		}
	}
	return out, nil
}
