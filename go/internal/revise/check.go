package revise

import (
	"regexp"
	"slices"
	"strconv"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// The causes of a reason of the rules after a revision (D-288). Each
// one is an id with no text of Luna, so a record and a log line can
// hold it (D-80). A failed call gives "call-" and its status, such as
// "call-timeout", and a blocked text gives "blocked-" and the rule of
// the filter.
const (
	// CauseNoCall: no reviser call occurred.
	CauseNoCall = "no-call"
	// CauseCapped: the monthly AI cap refused the call (D-25).
	CauseCapped = "capped"
	// CauseNoReason: the output has no reason for the exercise.
	CauseNoReason = "no-reason"
	// CauseNoSet: the reason names no logged set.
	CauseNoSet = "no-logged-set"
	// CauseUnknownSet: the reason names a set that the last session did
	// not log.
	CauseUnknownSet = "unknown-set"
	// CauseUnnamedSet: the text names none of the sets that the reason
	// names.
	CauseUnnamedSet = "unnamed-set"
	// CauseNumber: the text holds a number that is not in the evidence.
	CauseNumber = "unknown-number"
)

// Reason gives the reason that the owner sees for a revised exercise,
// its source, and the cause when the reason of the rules shows (D-288).
// res is the result of the reviser call, or a zero result when no call
// occurred.
func Reason(in policy.Input, rec policy.Record, res ai.Result) (string, policy.Source, string) {
	rules := func(cause string) (string, policy.Source, string) { return rec.Reason, policy.SourceRules, cause }
	switch res.Status {
	case "":
		return rules(CauseNoCall)
	case ai.StatusCapped:
		return rules(CauseCapped)
	case ai.StatusOK:
	default:
		return rules("call-" + string(res.Status))
	}
	i := slices.IndexFunc(res.Reasons, func(r ai.Reason) bool { return r.Exercise == in.Exercise.ID })
	if i < 0 {
		return rules(CauseNoReason)
	}
	r := res.Reasons[i]
	if cause := Check(in, rec, r); cause != "" {
		return rules(cause)
	}
	return r.Text, policy.SourceLuna, ""
}

var (
	numberForm = regexp.MustCompile(`\d+(?:\.\d+)?`)
	setForm    = regexp.MustCompile(`(?i)\bset (\d+)\b`)
)

// Check reads a reason of Luna against the logged sets of the last
// session of the exercise, and gives the cause of a refusal, or "" for
// a reason that passes (D-68, D-288). A reason passes when:
//   - the filter of blocked claims did not block it (D-183),
//   - it names one or more sets, and the last session logged each one,
//   - its text names a named set, as "set 2" or with the reps or the
//     weight of that set, and
//   - each number of its text is a number of the evidence: a named set
//     and its values, the missed reps of a named working set, a value of
//     the last target or of the next target, the count of logged working
//     sets, or a number of the reason of the rules.
func Check(in policy.Input, rec policy.Record, r ai.Reason) string {
	if r.Blocked != "" {
		return "blocked-" + string(r.Blocked)
	}
	if len(r.Sets) == 0 {
		return CauseNoSet
	}
	if len(in.History) == 0 {
		return CauseUnknownSet
	}
	last := in.History[len(in.History)-1]
	logged := numbered(last.Log.Sets)
	allowed := map[float64]bool{}
	add := func(v ...float64) {
		for _, x := range v {
			allowed[x] = true
		}
	}
	named := map[float64]bool{}
	for _, ref := range r.Sets {
		s, ok := logged[ref]
		if !ok {
			return CauseUnknownSet
		}
		add(float64(ref.Number), float64(s.Reps), pounds(s.Weight), float64(s.RIR))
		named[float64(ref.Number)] = true
		// The missed reps of a working set against its target.
		if ref.Kind == domain.SetWorking && ref.Number <= len(last.Target.Working) {
			if miss := last.Target.Working[ref.Number-1].Reps - s.Reps; miss > 0 {
				add(float64(miss))
			}
		}
		if s.Pain != nil {
			add(float64(*s.Pain))
		}
	}
	if !names(r, logged, named) {
		return CauseUnnamedSet
	}
	for _, t := range []domain.PlannedExercise{last.Target, rec.Target} {
		add(float64(len(t.Working)), float64(len(t.Calibration)))
		for _, s := range t.Working {
			add(float64(s.Reps), pounds(s.Load), float64(s.RIR))
		}
		for _, s := range t.Calibration {
			add(float64(s.Reps), pounds(s.Load))
		}
	}
	working := 0
	for _, s := range last.Log.Sets {
		if s.Kind == domain.SetWorking {
			working++
		}
	}
	add(float64(working))
	for _, n := range numbers(rec.Reason) {
		add(n)
	}
	for _, n := range numbers(r.Text) {
		if !allowed[n] {
			return CauseNumber
		}
	}
	return ""
}

// names tells whether the text names a set of the reason: "set N" of a
// named number, or the reps or the weight of a named set.
func names(r ai.Reason, logged map[ai.SetRef]domain.SetLog, named map[float64]bool) bool {
	for _, m := range setForm.FindAllStringSubmatch(r.Text, -1) {
		if n, err := strconv.ParseFloat(m[1], 64); err == nil && named[n] {
			return true
		}
	}
	values := map[float64]bool{}
	for _, ref := range r.Sets {
		s := logged[ref]
		values[float64(s.Reps)] = true
		values[pounds(s.Weight)] = true
	}
	for _, n := range numbers(r.Text) {
		if values[n] {
			return true
		}
	}
	return false
}

// numbered gives each logged set by its kind and its number inside its
// kind, from 1, as the reviser input numbers them.
func numbered(sets []domain.SetLog) map[ai.SetRef]domain.SetLog {
	out := map[ai.SetRef]domain.SetLog{}
	count := map[domain.SetKind]int{}
	for _, s := range sets {
		count[s.Kind]++
		out[ai.SetRef{Kind: s.Kind, Number: count[s.Kind]}] = s
	}
	return out
}

func numbers(text string) []float64 {
	var out []float64
	for _, m := range numberForm.FindAllString(text, -1) {
		if n, err := strconv.ParseFloat(m, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// pounds gives a load in pounds. A division of tenths gives the nearest
// float, as the parse of the same decimal text does.
func pounds(l domain.Load) float64 { return float64(l) / float64(domain.Pound) }
