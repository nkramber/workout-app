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

// Follow gives the heaviest load that the other working sets of a
// target can use after the owner logs its first set at a heavier weight
// (D-306, D-307). It is the next heavier weight of the machine above the
// load of the first working set, or that load at the top of the list.
// A lighter first set has no limit (D-308). A target with the first-set
// calibration or a calibration set gives 0, because its calibration
// gives the load of the other working sets (D-299). A target with more
// than one load gives 0 too, because the first set then gives no load
// for the other sets.
func Follow(t domain.PlannedExercise, available []domain.Load) domain.Load {
	if t.FirstSetCalibration || len(t.Calibration) > 0 || len(t.Working) == 0 || len(available) == 0 {
		return 0
	}
	base := t.Working[0].Load
	if slices.ContainsFunc(t.Working, func(s domain.WorkingSet) bool { return s.Load != base }) {
		return 0
	}
	if i := slices.IndexFunc(available, func(l domain.Load) bool { return l > base }); i >= 0 {
		return available[i]
	}
	return base
}

// Followed gives the load of the other working sets of a target after
// the owner logged its first set at weight (D-306 to D-308). A target
// with no limit, or a first set at the load of the target, gives the
// load of the target.
func Followed(t domain.PlannedExercise, weight domain.Load) domain.Load {
	if len(t.Working) == 0 {
		return 0
	}
	if base := t.Working[0].Load; t.FollowMax <= 0 || weight == base {
		return base
	}
	return min(weight, t.FollowMax)
}

// FollowReps gives the limits of the reps that the later working sets
// of a target can mirror from a set that went up: the rep range of the
// exercise (D-341, D-343, D-344). A target with a calibration set gives
// none, because the calibration table gives its working sets (D-267).
func FollowReps(t domain.PlannedExercise, e domain.Exercise) Range {
	if len(t.Calibration) > 0 || len(t.Working) == 0 {
		return Range{}
	}
	return RepRange(e)
}

// Limits sets the limits of the later working sets of t: the load of
// Follow and the reps of FollowReps (D-23, D-306, D-343).
func Limits(t *domain.PlannedExercise, e domain.Exercise, available []domain.Load) {
	t.FollowMax = Follow(*t, available)
	r := FollowReps(*t, e)
	t.FollowMinReps, t.FollowMaxReps = r.Min, r.Max
}

// Change is the change of the later working sets of a session after a
// logged set: a load, reps, or both. A value of 0 changes nothing.
type Change struct {
	Load domain.Load
	Reps int
}

// Apply gives s with the change.
func (c Change) Apply(s domain.WorkingSet) domain.WorkingSet {
	if c.Load > 0 {
		s.Load = c.Load
	}
	if c.Reps > 0 {
		s.Reps = c.Reps
	}
	return s
}

// Mirror gives the change of the later working sets after the owner
// logged s against the target shown, from the change c before it. n
// counts the working sets from 0.
//
// From version 10, a set that goes up against the target that it
// showed, with more reps or a heavier weight, at 1 rep in reserve or
// more, with no pain rating of 1 or more, gives the later sets its
// weight up to FollowMax and its reps inside the rep limits (D-341,
// D-343, D-346). A lighter set gives its weight
// at any reps in reserve (D-342). A target with no weight limit, the
// first-set calibration included, mirrors the reps alone (D-344). In
// the first-set calibration, the first set gives the load (D-299).
//
// A target of version 9 or earlier has no rep limits. Then the first
// set alone gives the load (D-299, D-306 to D-308).
func Mirror(t domain.PlannedExercise, n int, shown domain.WorkingSet, s domain.SetLog, c Change) Change {
	if t.FollowMaxReps <= 0 {
		switch {
		case n != 0 || len(t.Working) == 0:
		case t.FirstSetCalibration:
			c.Load = s.Weight
		case t.FollowMax > 0 && s.Weight != t.Working[0].Load:
			c.Load = Followed(t, s.Weight)
		}
		return c
	}
	up := s.Reps > 0 && s.RIR >= 1 && (s.Pain == nil || *s.Pain < 1)
	reps := Range{t.FollowMinReps, t.FollowMaxReps}.clamp(s.Reps)
	switch {
	case t.FirstSetCalibration || t.FollowMax <= 0:
		if t.FirstSetCalibration && n == 0 {
			c.Load = s.Weight
		}
		if up && s.Reps > shown.Reps {
			c.Reps = reps
		}
	case up && (s.Weight > shown.Load || s.Reps > shown.Reps):
		c.Load, c.Reps = min(s.Weight, t.FollowMax), reps
	case s.Weight < shown.Load:
		c.Load = s.Weight
	}
	return c
}

// Mirrored gives the targets of the working sets of a session from its
// logged working sets, in order (D-341 to D-345). shown holds the target
// that each set showed before its log. read holds the target that the
// rules read for each logged set: the values that it gave the later
// sets, or else the target that it showed. next holds the last change,
// applied to each working set, and it is the start of the next session.
// The phone gives the same shown targets (`web/src/lib/workout.ts`).
func Mirrored(t domain.PlannedExercise, logs []domain.SetLog) (shown, read, next []domain.WorkingSet) {
	shown, read, next = slices.Clone(t.Working), slices.Clone(t.Working), slices.Clone(t.Working)
	var c Change
	for j := range shown {
		shown[j] = c.Apply(t.Working[j])
		if j < len(logs) {
			c = Mirror(t, j, shown[j], logs[j], c)
		}
		read[j] = c.Apply(t.Working[j])
	}
	for j := range next {
		next[j] = c.Apply(t.Working[j])
	}
	return shown, read, next
}

// effective gives a copy of the history in which each session has the
// working targets that the rules read. In a session of policy version 6
// or earlier, the table applies to the weight of its first calibration
// set (D-150, D-267). The table applies to the weight that the owner
// logged, because the owner can change the weight before the log
// (D-249). The owner logs the working sets at that load, so the rules
// read the logs against it.
//
// Each other session has the targets of Mirrored: the first-set
// calibration (D-299), the load that the later sets followed (D-309),
// and from version 10 the mirror (D-341 to D-345). Its next targets
// start the next session. A load that the machine does not have goes to
// a weight of the machine.
func (in Input) effective() []Outcome {
	out := slices.Clone(in.History)
	available := in.Entry.Available()
	for i, o := range out {
		cal, logs := calibrationSets(o.Log), o.working()
		if len(o.Target.Calibration) > 0 {
			if len(cal) == 0 {
				continue
			}
			load := calibrationLoads(firstLoad(cal[0].Weight, available), available).For(cal[0])
			w := slices.Clone(o.Target.Working)
			for j := range w {
				w[j].Load = load
			}
			out[i].Target.Working = w
			continue
		}
		_, read, next := Mirrored(o.Target, logs)
		for j, s := range o.Target.Working {
			if read[j].Load != s.Load {
				read[j].Load = firstLoad(read[j].Load, available)
			}
			if next[j].Load != s.Load {
				next[j].Load = firstLoad(next[j].Load, available)
			}
		}
		out[i].Target.Working = read
		out[i].next = next
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
