package policy

import (
	"fmt"
	"strings"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// Outcome is one past session of one exercise: the target that the
// owner saw, and the log of the exercise. EndedEarly tells that the
// session ended with "finish now" (D-63).
type Outcome struct {
	Target     domain.PlannedExercise
	Log        domain.ExerciseLog
	EndedEarly bool
}

// Input is the history of one exercise, oldest first, with the
// exercise and the inventory entry of its machine.
type Input struct {
	Exercise domain.Exercise
	Entry    domain.InventoryEntry
	History  []Outcome
}

// Decision is the next target of one exercise. Rules names each rule
// that set it, in order. Reason is the concise text for the owner that
// names the logged evidence (D-68). Reason is shown to the owner and
// never goes into a log, because it can hold loads and reps (D-80).
type Decision struct {
	Version int
	Target  domain.PlannedExercise
	Rules   []RuleID
	Reason  string
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
	for i, o := range in.History {
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
// D-64). The rules apply in a fixed order, and the first rule that
// decides gives the target. The target is inside each bound of Check.
func Next(in Input) (Decision, error) {
	if err := in.check(); err != nil {
		return Decision{}, err
	}
	if len(in.History) == 0 {
		return Decision{}, ErrNoHistory
	}
	last := in.History[len(in.History)-1]
	var prev *Outcome
	if len(in.History) > 1 {
		prev = &in.History[len(in.History)-2]
	}

	reps := RepRange(in.Exercise)
	b := newBuilder(in, last.Target)
	logs := last.working()
	total, first := last.shortfall()

	switch {
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
		b.rule(RuleShortfallLowRIR, fmt.Sprintf("Set %d ended %d reps short after hard sets. The load stays at %s, and the target is %s.", first+1, last.Target.Working[first].Reps-logs[first].Reps, b.loadText(), b.repsText()))
	case total > 2:
		b.hold(RuleShortfallRepeat, fmt.Sprintf("Set %d ended %d reps short. The target stays the same.", first+1, last.Target.Working[first].Reps-logs[first].Reps))
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
	case b.stepUp():
		b.setReps(reps.Min)
		b.rule(RuleLoadStep, fmt.Sprintf("You reached the top of the range with reps to spare. The load goes up to %s, and the target is %s.", b.loadText(), b.repsText()))
	default:
		b.hold(RuleNoHeavier, "The machine has no heavier weight within one 5 lb step. The target stays the same.")
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

// builder holds the next target while the rules change it.
type builder struct {
	in        Input
	available []domain.Load
	target    domain.PlannedExercise
	rules     []RuleID
	reason    []string
	note      string
}

// newBuilder starts from the last target, inside each bound: the reps
// in the rep range, the reps in reserve in their range, the rest in its
// limits, and each load a valid load of the machine (RuleLoadRepair).
func newBuilder(in Input, last domain.PlannedExercise) *builder {
	b := &builder{in: in, available: in.Entry.Available()}
	reps, rir := RepRange(in.Exercise), RIRRange(in.Exercise)
	b.target = domain.PlannedExercise{
		Exercise:    in.Exercise.ID,
		RestSeconds: RestLimits.clamp(last.RestSeconds),
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

func (b *builder) decision() Decision {
	if b.note != "" {
		b.rule(RuleLoadAvailable, b.note)
	}
	return Decision{
		Version: Version,
		Target:  b.target,
		Rules:   b.rules,
		Reason:  strings.Join(b.reason, " "),
	}
}
