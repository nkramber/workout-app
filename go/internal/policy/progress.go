package policy

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// Outcome is one past session of one exercise: its date in the form
// of domain.DateLayout, the target that the owner saw, and the log of
// the exercise. EndedEarly tells that the session ended with "finish
// now" (D-63).
type Outcome struct {
	Date       string
	Target     domain.PlannedExercise
	Log        domain.ExerciseLog
	EndedEarly bool
}

// Input is the history of one exercise, oldest first, with the
// exercise and the inventory entry of its machine. History holds each
// session of the exercise, because the first sessions after a break
// count from the return (D-151). Today is the date of the next session,
// in the form of domain.DateLayout. An input with a history needs it.
//
// Estimate is the load estimate of the owner for a new exercise (D-41),
// or 0 for no estimate. Returning tells that the owner had a break of
// 91 days or more before the first session of the exercise (D-150).
// The policy reads both only for an exercise with no history.
//
// Deloads holds the start date of each reactive deload of the owner,
// oldest first, from Deloads over the history of each exercise (D-295).
// A deload reads more than one exercise, so the caller gives it.
type Input struct {
	Exercise  domain.Exercise
	Entry     domain.InventoryEntry
	History   []Outcome
	Today     string
	Estimate  domain.Load
	Returning bool
	Deloads   []string
}

// Decision is the next target of one exercise. Rules names each rule
// that set it, in order. Loads holds each load of the target before
// and after the rounding (D-176). Reason is the concise text for the
// owner that names the logged evidence (D-68). Reason is shown to the
// owner and never goes into a log, because it can hold loads and reps
// (D-80).
type Decision struct {
	Version int
	Target  domain.PlannedExercise
	Rules   []RuleID
	Loads   []LoadChange
	Reason  string
}

// LoadChange is one load of a target before and after the rounding of
// D-65 and the weight selection of D-149. Where names the set, such as
// "working[0]" or "calibration[0]". A load that no rule computed has
// the same value before and after.
type LoadChange struct {
	Where  string
	Before domain.Load
	After  domain.Load
}

func inputError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}

func (in Input) check() error {
	e := in.Exercise
	if e.ID == "" {
		return inputError("no exercise")
	}
	if e.Kind == domain.KindCardio {
		return inputError("exercise %q: a cardio exercise", e.ID)
	}
	if in.Entry.Machine != e.Machine {
		return inputError("exercise %q: inventory machine %q, want %q", e.ID, in.Entry.Machine, e.Machine)
	}
	switch e.Kind {
	case domain.KindDumbbell:
		if in.Entry.Dumbbells == nil {
			return inputError("exercise %q: no dumbbell set", e.ID)
		}
		if err := in.Entry.Dumbbells.Check(); err != nil {
			return inputError("exercise %q: %v", e.ID, err)
		}
	default:
		if in.Entry.Dumbbells != nil || len(in.Entry.Weights) == 0 {
			return inputError("exercise %q: want a list of weights", e.ID)
		}
		for i, w := range in.Entry.Weights {
			if w <= 0 || (i > 0 && w <= in.Entry.Weights[i-1]) {
				return inputError("exercise %q: weights[%d] %s: want ascending weights above 0", e.ID, i, w)
			}
		}
	}
	available := in.Entry.Available()
	if in.Estimate < 0 {
		return inputError("exercise %q: estimate below 0", e.ID)
	}
	if len(in.History) == 0 && in.Estimate > 0 && (in.Estimate < available[0] || in.Estimate > available[len(available)-1]) {
		return inputError("exercise %q: estimate outside the weights of machine %q", e.ID, in.Entry.Machine)
	}
	if len(in.History) > 0 {
		if _, err := time.Parse(domain.DateLayout, in.Today); err != nil {
			return inputError("today: want the form %s", domain.DateLayout)
		}
	}
	for i, d := range in.Deloads {
		if _, err := time.Parse(domain.DateLayout, d); err != nil {
			return inputError("deloads[%d]: want the form %s", i, domain.DateLayout)
		}
		if i > 0 && day(d) <= day(in.Deloads[i-1]) {
			return inputError("deloads[%d]: want dates in order", i)
		}
	}
	for i, o := range in.History {
		// An error holds no date, because a date is data of the log
		// (D-80).
		if _, err := time.Parse(domain.DateLayout, o.Date); err != nil {
			return inputError("history[%d] date: want the form %s", i, domain.DateLayout)
		}
		if (i > 0 && day(o.Date) < day(in.History[i-1].Date)) || day(o.Date) > day(in.Today) {
			return inputError("history[%d] date: want dates in order, not after today", i)
		}
		if o.Target.Exercise != e.ID || o.Log.Exercise != e.ID {
			return inputError("history[%d]: exercise %q and %q, want %q", i, o.Target.Exercise, o.Log.Exercise, e.ID)
		}
		if len(o.Target.Working) == 0 {
			return inputError("history[%d]: no working set", i)
		}
		for j, s := range o.Target.Working {
			if err := s.Check(); err != nil {
				return inputError("history[%d] working[%d]: %v", i, j, err)
			}
		}
		for j, s := range o.Target.Calibration {
			if err := s.Check(); err != nil {
				return inputError("history[%d] calibration[%d]: %v", i, j, err)
			}
		}
		if o.Log.Skipped && len(o.Log.Sets) > 0 {
			return inputError("history[%d]: skipped with %d sets", i, len(o.Log.Sets))
		}
		for j, s := range o.Log.Sets {
			if err := s.Check(); err != nil {
				return inputError("history[%d] sets[%d]: %v", i, j, err)
			}
		}
	}
	return nil
}

// working gives the logged working sets of an outcome.
func (o Outcome) working() []domain.SetLog {
	var out []domain.SetLog
	for _, s := range o.Log.Sets {
		if s.Kind == domain.SetWorking {
			out = append(out, s)
		}
	}
	return out
}

// shortfall gives the total missed reps of the logged working sets
// against their targets, and the index of the first short set, or -1.
// A planned set with no log adds nothing (D-170).
func (o Outcome) shortfall() (total, first int) {
	first = -1
	logs := o.working()
	for i := 0; i < len(logs) && i < len(o.Target.Working); i++ {
		if miss := o.Target.Working[i].Reps - logs[i].Reps; miss > 0 {
			total += miss
			if first < 0 {
				first = i
			}
		}
	}
	return total, first
}

// trained tells that the owner logged a working set of the outcome. A
// skipped exercise, or a log of calibration sets alone, does not end a
// break.
func (o Outcome) trained() bool { return len(o.working()) > 0 }

func (o Outcome) pain() bool {
	for _, s := range o.Log.Sets {
		if s.Pain != nil && *s.Pain >= 1 {
			return true
		}
	}
	return false
}

func (o Outcome) failure() bool {
	for _, s := range o.working() {
		if s.RIR == 0 {
			return true
		}
	}
	return false
}

// Next gives the next target of an exercise from its history (D-23,
// D-64). An exercise with no history gets its start (D-150). A gap of
// 14 days or more gives the return of the long-break table (D-151). A
// gap of 7 to 13 days holds the target at 3 reps in reserve (D-294).
// Otherwise the rules apply in a fixed order, and the first rule that
// decides gives the target. On a date of a deload, the target gets the
// sets and the reps in reserve of the deload (D-295). The target is
// inside each bound of Check. Next is also the rules fallback: its
// target needs no proposal.
func Next(in Input) (Decision, error) {
	if err := in.check(); err != nil {
		return Decision{}, err
	}
	if len(in.History) == 0 {
		return start(in), nil
	}
	in.History = in.effective()
	ps := in.pause()
	calibrating := in.calibrating()
	before, wasMissed := in.afterMissed()
	in.History = in.rulesHistory()
	last := in.History[len(in.History)-1]
	var prev *Outcome
	if len(in.History) > 1 {
		prev = &in.History[len(in.History)-2]
	}

	reps := RepRange(in.Exercise)
	b := newBuilder(in, last.Target)
	logs := last.working()
	total, first := last.shortfall()

	if ps.ended {
		b.restore(ps.before)
	}

	switch {
	case ps.gap >= BreakDays:
		b.resume(ps.gap)
	case missed(ps.gap):
		b.hold(RuleMissed, fmt.Sprintf("Your last logged set of this exercise was %d days ago. The target stays the same, at 3 reps in reserve.", ps.gap))
		b.setRIR(3)
	case last.Log.Skipped:
		b.hold(RuleSkipped, "You skipped this exercise. The target stays the same.")
	case last.pain():
		b.hold(RulePainHold, "You reported pain on this exercise. The target stays the same for one session.")
	case total > 0 && total <= 2:
		b.hold(RuleShortfallSmall, fmt.Sprintf("Set %d ended %s short. The target stays the same.", first+1, repText(total)))
	case total > 2 && prev != nil && !prev.Log.Skipped && shortTotal(*prev) > 2:
		if !b.stepDown() {
			b.setReps(reps.Min)
			b.rule(RuleShortfallTwice, fmt.Sprintf("Two sessions in a row ended more than 2 reps short. The load is the lightest weight, so the target is %s.", b.repsText()))
		} else {
			b.rule(RuleShortfallTwice, fmt.Sprintf("Two sessions in a row ended more than 2 reps short. The load goes down to %s.", b.loadText()))
		}
	case total > 2 && first > 0 && allLowRIR(logs[:first]):
		b.setReps(reps.Min)
		b.rule(RuleShortfallLowRIR, fmt.Sprintf("Set %d ended %s short after hard sets. The load stays at %s, and the target is %s.", first+1, repText(last.Target.Working[first].Reps-logs[first].Reps), b.loadText(), b.repsText()))
	case total > 2:
		b.hold(RuleShortfallRepeat, fmt.Sprintf("Set %d ended %s short. The target stays the same.", first+1, repText(last.Target.Working[first].Reps-logs[first].Reps)))
	case len(logs) < len(last.Target.Working):
		early := ""
		if last.EndedEarly {
			early = "The session ended early. "
		}
		b.hold(RuleIncomplete, fmt.Sprintf("%sYou logged %d of %d sets. The target stays the same.", early, len(logs), len(last.Target.Working)))
	case lighter(last):
		b.hold(RuleLighterHold, "You used a lighter weight than the target. The target stays the same.")
	case last.failure() && prev != nil && prev.failure():
		b.hold(RuleFailureHold, "Two sessions in a row had a set at 0 reps in reserve. The target stays the same.")
	case isPress(in.Exercise) && anyRIRAtMost(logs, 1):
		b.hold(RulePressHold, "A set ended at 1 rep in reserve or less. The load stays the same.")
	case anyRIRBelow(logs, 3):
		b.hold(RuleEffortHold, "Not every set had 3 reps in reserve or more. The target stays the same.")
	case !b.allAtTop(reps.Max):
		b.addReps(2, reps.Max)
		b.rule(RuleAddReps, fmt.Sprintf("You completed every set with reps to spare. The target is %s.", b.repsText()))
	case calibrating:
		b.hold(RuleCalibrationHold, "This exercise is still in calibration, so the load stays the same.")
	case ps.first:
		b.hold(RuleBreakFirst, "These are your first sessions after a break, so the load stays the same.")
	case b.stepUp():
		b.setReps(reps.Min)
		b.rule(RuleLoadStep, fmt.Sprintf("You reached the top of the range with reps to spare. The load goes up to %s, and the target is %s.", b.loadText(), b.repsText()))
	default:
		b.hold(RuleNoHeavier, "The machine has no heavier weight within one 5 lb step. The target stays the same.")
	}

	afterBreak := in.firstSessions(ps)
	if !afterBreak && wasMissed && !missed(ps.gap) && b.restoreRIR(before) {
		b.rule(RuleMissedRestored, "The session after a missed week ended, so the reps in reserve go back to the target before it.")
	}
	if afterBreak && b.setRIR(3) && ps.gap < BreakDays && !slices.Contains(b.rules, RuleBreakFirst) {
		b.rule(RuleBreakFirst, "These are your first sessions after a break, so each set stops at 3 reps in reserve.")
	}
	if from := in.deloadOf(in.Today, false); from != "" {
		b.deload(from)
	}
	if ps.gap >= RecalibrateDays || calibrating {
		b.calibration()
	}
	return b.decision(), nil
}

func shortTotal(o Outcome) int {
	t, _ := o.shortfall()
	return t
}

func allLowRIR(logs []domain.SetLog) bool {
	for _, s := range logs {
		if s.RIR > 1 {
			return false
		}
	}
	return len(logs) > 0
}

func anyRIRAtMost(logs []domain.SetLog, n int) bool {
	for _, s := range logs {
		if s.RIR <= n {
			return true
		}
	}
	return false
}

func anyRIRBelow(logs []domain.SetLog, n int) bool { return anyRIRAtMost(logs, n-1) }

func lighter(o Outcome) bool {
	logs := o.working()
	for i := 0; i < len(logs) && i < len(o.Target.Working); i++ {
		if logs[i].Weight < o.Target.Working[i].Load {
			return true
		}
	}
	return false
}

func repText(n int) string {
	if n == 1 {
		return "1 rep"
	}
	return fmt.Sprintf("%d reps", n)
}

// builder holds the next target while the rules change it. before
// holds the load of each working set before the rounding, and
// calBefore the load of the calibration set.
type builder struct {
	in        Input
	available []domain.Load
	target    domain.PlannedExercise
	before    []domain.Load
	calBefore domain.Load
	rules     []RuleID
	reason    []string
	note      string
}

// newBuilder starts from the last target, inside each bound: the reps
// in the rep range, the reps in reserve in their range, and each load a
// valid load of the machine (RuleLoadRepair). Each plan gives the rest of
// D-279 again, so a rest of an older policy does not stay.
func newBuilder(in Input, last domain.PlannedExercise) *builder {
	b := &builder{in: in, available: in.Entry.Available()}
	reps, rir := RepRange(in.Exercise), RIRRange(in.Exercise)
	b.target = domain.PlannedExercise{
		Exercise:    in.Exercise.ID,
		RestSeconds: DefaultRest(in.Exercise),
	}
	repaired := false
	for _, s := range last.Working {
		l := s.Load
		if !Valid(l, b.available) {
			l = Select(floor(l), b.available)
			repaired = true
		}
		b.target.Working = append(b.target.Working, domain.WorkingSet{
			Reps: reps.clamp(s.Reps),
			Load: l,
			RIR:  rir.clamp(s.RIR),
		})
		b.before = append(b.before, s.Load)
	}
	if repaired {
		b.rule(RuleLoadRepair, fmt.Sprintf("The machine does not have the last load, so the load is %s.", b.loadText()))
	}
	return b
}

func (b *builder) rule(r RuleID, reason string) {
	b.rules = append(b.rules, r)
	b.reason = append(b.reason, reason)
}

func (b *builder) hold(r RuleID, reason string) { b.rule(r, reason) }

func (b *builder) setReps(n int) {
	for i := range b.target.Working {
		b.target.Working[i].Reps = n
	}
}

func (b *builder) addReps(n, top int) {
	for i := range b.target.Working {
		b.target.Working[i].Reps = min(b.target.Working[i].Reps+n, top)
	}
}

func (b *builder) allAtTop(top int) bool {
	for _, s := range b.target.Working {
		if s.Reps < top {
			return false
		}
	}
	return true
}

// stepUp adds one 5 lb step to each load (D-147). It rounds to the
// nearest 5 lb with a halfway value down (D-148), but never above the
// load plus 5 lb, and selects the weight of the machine (D-149). It
// tells whether a load rose. So no load rises more than one step.
func (b *builder) stepUp() bool {
	rose := false
	for i, s := range b.target.Working {
		r := Round(s.Load+Step, Increase)
		if r > s.Load+Step {
			r -= Step
		}
		if l := Select(r, b.available); l > s.Load {
			b.before[i] = s.Load + Step
			b.target.Working[i].Load = l
			b.machineWeight(i, r, l)
			rose = true
		}
	}
	return rose
}

// stepDown takes one 5 lb step from each load, with a halfway value up
// (D-148), and selects the weight of the machine (D-149). It tells
// whether a load fell.
func (b *builder) stepDown() bool {
	fell := false
	for i, s := range b.target.Working {
		r := Round(s.Load-Step, Other)
		if l := Select(r, b.available); l < s.Load {
			b.before[i] = s.Load - Step
			b.target.Working[i].Load = l
			b.machineWeight(i, r, l)
			fell = true
		}
	}
	return fell
}

// machineWeight keeps the text that names the weight of the machine
// when it is not the rounded load (D-149). The first set gives the
// text.
func (b *builder) machineWeight(i int, rounded, l domain.Load) {
	if i == 0 && rounded != l {
		b.note = fmt.Sprintf("The machine has no %s weight, so the load is %s.", rounded, l)
	}
}

// loadText gives the load of the first set. The policy gives each set
// of an exercise its load from the same rules, so the sets agree when
// the last target agreed.
func (b *builder) loadText() string { return b.target.Working[0].Load.String() }

func (b *builder) repsText() string {
	ws := b.target.Working
	same := true
	for _, s := range ws {
		same = same && s.Reps == ws[0].Reps
	}
	if same {
		return fmt.Sprintf("%d x %d", len(ws), ws[0].Reps)
	}
	parts := make([]string, len(ws))
	for i, s := range ws {
		parts[i] = fmt.Sprint(s.Reps)
	}
	return strings.Join(parts, ", ") + " reps"
}

// setRIR sets the reps in reserve of each working set, and tells
// whether a value changed.
func (b *builder) setRIR(n int) bool {
	changed := false
	for i := range b.target.Working {
		if b.target.Working[i].RIR != n {
			b.target.Working[i].RIR = n
			changed = true
		}
	}
	return changed
}

func (b *builder) decision() Decision {
	if b.note != "" {
		b.rule(RuleLoadAvailable, b.note)
	}
	var loads []LoadChange
	for i, s := range b.target.Calibration {
		loads = append(loads, LoadChange{fmt.Sprintf("calibration[%d]", i), b.calBefore, s.Load})
	}
	for i, s := range b.target.Working {
		loads = append(loads, LoadChange{fmt.Sprintf("working[%d]", i), b.before[i], s.Load})
	}
	return Decision{
		Version: Version,
		Target:  b.target,
		Rules:   b.rules,
		Loads:   loads,
		Reason:  strings.Join(b.reason, " "),
	}
}
