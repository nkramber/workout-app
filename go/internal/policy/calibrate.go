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
	// exercise, and after a break of 91 days or more.
	CalibrationSessions = 3
	// CalibrationChanges is the most load changes of a calibration in
	// one session.
	CalibrationChanges = 3
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
// session has the working load that its calibration set the owner to
// (D-150). The owner logs the working sets at that load, so the rules
// read the logs against it.
func (in Input) effective() []Outcome {
	out := slices.Clone(in.History)
	available := in.Entry.Available()
	for i, o := range out {
		cal := calibrationSets(o.Log)
		if len(o.Target.Calibration) == 0 || len(cal) == 0 {
			continue
		}
		step := calibrate(firstLoad(o.Target.Calibration[0].Load, available), cal, available)
		w := slices.Clone(o.Target.Working)
		for j := range w {
			w[j].Load = step.Load
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

// CalibrationStep is the next set of the calibration of one session
// (D-150). Again tells that the next set is one more calibration set at
// Load. Otherwise the working sets use Load. Reason is the text for the
// owner, and it never goes into a log (D-80).
type CalibrationStep struct {
	Load    domain.Load
	Again   bool
	Changes int
	Rule    RuleID
	Reason  string
}

// Calibrate gives the next set after the calibration sets that the
// owner logged in a session, in order. first is the load of the
// calibration set of the target. With no logged set, the next set is
// the calibration set at first. The table of RuleCalibrationTable gives
// each change. The calibration ends when the load stays, when the
// machine has no lighter or heavier weight, or after 3 changes. A later
// set does not change an ended calibration.
func Calibrate(in Input, first domain.Load, sets []domain.SetLog) (CalibrationStep, error) {
	if err := in.check(); err != nil {
		return CalibrationStep{}, err
	}
	available := in.Entry.Available()
	if !slices.Contains(available, first) || !Valid(first, available) {
		return CalibrationStep{}, inputError("exercise %q: first calibration load is not a valid weight", in.Exercise.ID)
	}
	for i, s := range sets {
		if err := s.Check(); err != nil {
			return CalibrationStep{}, inputError("sets[%d]: %v", i, err)
		}
		if s.Kind != domain.SetCalibration {
			return CalibrationStep{}, inputError("sets[%d]: kind %q, want %q", i, s.Kind, domain.SetCalibration)
		}
	}
	return calibrate(first, sets, available), nil
}

func calibrate(load domain.Load, sets []domain.SetLog, available []domain.Load) CalibrationStep {
	c := CalibrationStep{Load: load, Again: true, Rule: RuleCalibrationTable}
	for _, s := range sets {
		steps := 0
		switch {
		case s.Pain != nil && *s.Pain >= 1, s.RIR <= 2:
			steps = -1
		case s.RIR >= 6:
			steps = 2
		case s.RIR == 5:
			steps = 1
		}
		next := move(c.Load, steps, available)
		if next == c.Load {
			c.Again = false
			if steps == 0 {
				c.Reason = fmt.Sprintf("The calibration set was on target. The working sets use %s.", c.Load)
			} else {
				c.Reason = fmt.Sprintf("The machine has no weight for the change. The working sets use %s.", c.Load)
			}
			return c
		}
		c.Load = next
		c.Changes++
		if c.Changes == CalibrationChanges {
			c.Again = false
			c.Reason = fmt.Sprintf("The calibration made %d changes, the limit. The working sets use %s.", CalibrationChanges, c.Load)
			return c
		}
	}
	c.Reason = fmt.Sprintf("Do one calibration set at %s. Stop it at 3 to 4 reps in reserve.", c.Load)
	return c
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
