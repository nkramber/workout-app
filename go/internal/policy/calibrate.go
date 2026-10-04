package policy

import (
	"fmt"
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The values of the start and the calibration of a new exercise (D-178,
// D-180, D-300).
const (
	// StartSets is the count of working sets of a start.
	StartSets = 3
	// ReturnPercent is the percent of the estimate or of the last load
	// after a break of 91 days or more.
	ReturnPercent = 70
)

// start gives the target of an exercise with no history (D-178, D-180,
// D-300, D-301): 3 working sets at the bottom of the rep range, at 3
// reps in reserve, and the first set is the calibration. The load is
// the estimate, or the lightest weight when no estimate exists. When
// Returning is true, the load is 70 percent of the estimate (D-179).
func start(in Input) Decision {
	b := &builder{in: in, available: in.Entry.Available()}
	b.target = domain.PlannedExercise{Exercise: in.Exercise.ID, RestSeconds: DefaultRest(in.Exercise)}
	reps := RepRange(in.Exercise).Min
	before, rounded := b.available[0], b.available[0]
	if in.Estimate > 0 {
		dir := Other
		before = in.Estimate
		if in.Returning {
			before, dir = in.Estimate*ReturnPercent/100, Return
		}
		rounded = Round(before, dir)
	}
	load := Select(rounded, b.available)
	for range StartSets {
		b.target.Working = append(b.target.Working, domain.WorkingSet{Reps: reps, Load: load, RIR: 3})
		b.before = append(b.before, before)
	}
	b.machineWeight(0, rounded, load)
	switch {
	case in.Estimate == 0:
		b.rule(RuleStartLightest, fmt.Sprintf("This exercise is new, and it has no load estimate. It starts at the lightest weight, %s.", load))
	case in.Returning:
		b.rule(RuleStartEstimate, fmt.Sprintf("This exercise is new, and you return after a long break. It starts at 70 percent of your estimate of %s: %s.", in.Estimate, load))
	default:
		b.rule(RuleStartEstimate, fmt.Sprintf("This exercise is new. It starts at %s, from your estimate of %s.", load, in.Estimate))
	}
	b.firstSet()
	return b.decision()
}

// firstSet makes the first working set the calibration (D-297, D-299).
// The owner changes the weight during its first reps, and the other
// working sets use the weight that the owner logged for it.
func (b *builder) firstSet() {
	b.target.FirstSetCalibration = true
	b.rule(RuleCalibrationFirstSet, "The first set is the calibration. Change the weight during its first reps when it is too light or too heavy. The other sets use the weight of the first set.")
}

// effective gives a copy of the history in which each calibration
// session has the working load that its calibration gave. In a session
// of the first-set calibration, the load is the weight that the owner
// logged for the first working set (D-299). In a session of policy
// version 6 or earlier, the table applies to the weight of its first
// calibration set (D-150, D-267). The table applies to the weight that
// the owner logged, because the owner can change the weight before the
// log (D-249). The owner logs the working sets at that load, so the
// rules read the logs against it.
func (in Input) effective() []Outcome {
	out := slices.Clone(in.History)
	available := in.Entry.Available()
	for i, o := range out {
		var load domain.Load
		switch cal, logs := calibrationSets(o.Log), o.working(); {
		case len(o.Target.Calibration) > 0 && len(cal) > 0:
			load = calibrationLoads(firstLoad(cal[0].Weight, available), available).For(cal[0])
		case o.Target.FirstSetCalibration && len(logs) > 0:
			load = firstLoad(logs[0].Weight, available)
		default:
			continue
		}
		w := slices.Clone(o.Target.Working)
		for j := range w {
			w[j].Load = load
		}
		out[i].Target.Working = w
	}
	return out
}

func calibrationSets(l domain.ExerciseLog) []domain.SetLog {
	var out []domain.SetLog
	for _, s := range l.Sets {
		if s.Kind == domain.SetCalibration {
			out = append(out, s)
		}
	}
	return out
}

// firstLoad gives a valid load of the machine for a load of a target,
// as RuleLoadRepair does.
func firstLoad(l domain.Load, available []domain.Load) domain.Load {
	if Valid(l, available) {
		return l
	}
	return Select(floor(l), available)
}

// CalibrationLoads gives the load of the working sets of a session
// after its one calibration set at Weight, for each result of the table
// of RuleCalibrationTable (D-150, D-267). A plan of policy version 6 or
// earlier holds one for each weight of the machine, so the phone applies
// the table with no network to the weight that the owner logged (D-23,
// D-249). From version 7, no target has a calibration set (D-297). A machine with no
// weight for a change keeps the load. A weight that the policy can not
// give takes the repair of RuleLoadRepair first.
type CalibrationLoads struct {
	Weight domain.Load
	// Down follows 2 or fewer reps in reserve, or a pain report.
	Down domain.Load
	// Keep follows 3 or 4 reps in reserve.
	Keep domain.Load
	// UpOne follows 5 reps in reserve.
	UpOne domain.Load
	// UpTwo follows 6 or more reps in reserve.
	UpTwo domain.Load
}

// For gives the load of the working sets after a logged calibration
// set.
func (c CalibrationLoads) For(s domain.SetLog) domain.Load {
	switch {
	case s.Pain != nil && *s.Pain >= 1, s.RIR <= 2:
		return c.Down
	case s.RIR >= 6:
		return c.UpTwo
	case s.RIR == 5:
		return c.UpOne
	}
	return c.Keep
}

// CalibrationTable gives the loads of a calibration set at each weight
// of the machine, the lightest weight first.
func CalibrationTable(in Input) ([]CalibrationLoads, error) {
	if err := in.check(); err != nil {
		return nil, err
	}
	available := in.Entry.Available()
	out := make([]CalibrationLoads, 0, len(available))
	for _, w := range available {
		c := calibrationLoads(firstLoad(w, available), available)
		c.Weight = w
		out = append(out, c)
	}
	return out, nil
}

func calibrationLoads(first domain.Load, available []domain.Load) CalibrationLoads {
	return CalibrationLoads{
		Weight: first,
		Down:   move(first, -1, available),
		Keep:   first,
		UpOne:  move(first, 1, available),
		UpTwo:  move(first, 2, available),
	}
}

// move gives the load after a change of steps of 5 lb, with the weight
// that D-149 selects. An increase rounds a halfway value down and never
// passes the load plus the steps, and a decrease rounds it up (D-148).
// It gives the same load when the machine has no weight for the change.
func move(l domain.Load, steps int, available []domain.Load) domain.Load {
	switch {
	case steps > 0:
		up := l + Step*domain.Load(steps)
		r := Round(up, Increase)
		if r > up {
			r -= Step
		}
		if n := Select(r, available); n > l {
			return n
		}
	case steps < 0:
		if n := Select(Round(l+Step*domain.Load(steps), Other), available); n < l {
			return n
		}
	}
	return l
}
