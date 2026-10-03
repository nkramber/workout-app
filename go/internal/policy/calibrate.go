package policy

import (
	"fmt"
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The values of the start and the calibration of a new exercise (D-150,
// D-177, D-180).
const (
	// StartSets is the count of working sets of a start.
	StartSets = 3
	// CalibrationSessions is the count of calibration sessions of a new
	// exercise, and after a break of 91 days or more. Each one makes one
	// change at most (D-267).
	CalibrationSessions = 3
	// ReturnPercent is the percent of the estimate or of the last load
	// after a break of 91 days or more.
	ReturnPercent = 70
)

// start gives the target of an exercise with no history (D-150, D-178,
// D-180): one calibration set and 3 working sets at the bottom of the
// rep range, at 3 reps in reserve. The load comes from the estimate, or
// it is the lightest weight when no estimate exists.
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
	b.calibration()
	return b.decision()
}

// calibration adds one calibration set at the reps and the load of the
// first working set (D-150, D-177).
func (b *builder) calibration() {
	w := b.target.Working[0]
	b.target.Calibration = []domain.CalibrationSet{{Reps: w.Reps, Load: w.Load}}
	b.calBefore = b.before[0]
	b.rule(RuleCalibrationSet, fmt.Sprintf("The session starts with one calibration set of %d reps at %s. Stop it at 3 to 4 reps in reserve.", w.Reps, w.Load))
}

// calibrating tells that the next session is a calibration session
// (D-177). A session with a calibration set in its target is a
// calibration session. The last sessions of the history count back to
// a session with no calibration set, or to a break of 91 days or more.
// Fewer than 3 such sessions with a logged working set give one more.
func (in Input) calibrating() bool {
	h := in.History
	if len(h) == 0 || len(h[len(h)-1].Target.Calibration) == 0 {
		return false
	}
	n, later := 0, -1
	for i := len(h) - 1; i >= 0; i-- {
		o := h[i]
		if len(o.Target.Calibration) == 0 {
			break
		}
		if !o.trained() {
			continue
		}
		d := day(o.Date)
		if later >= 0 && later-d >= RecalibrateDays {
			break
		}
		n++
		later = d
	}
	return n < CalibrationSessions
}

// effective gives a copy of the history in which each calibration
// session has the working load that its first calibration set gave
// (D-150, D-267). The table applies to the weight that the owner logged,
// because the owner can change the weight before the log (D-249). The
// owner logs the working sets at that load, so the rules read the logs
// against it.
func (in Input) effective() []Outcome {
	out := slices.Clone(in.History)
	available := in.Entry.Available()
	for i, o := range out {
		cal := calibrationSets(o.Log)
		if len(o.Target.Calibration) == 0 || len(cal) == 0 {
			continue
		}
		load := calibrationLoads(firstLoad(cal[0].Weight, available), available).For(cal[0])
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
// of RuleCalibrationTable (D-150, D-267). The plan holds one for each
// weight of the machine, so the phone applies the table with no network
// to the weight that the owner logged (D-23, D-249). A machine with no
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
