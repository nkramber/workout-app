package policy

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The values of a missed session (D-294) and of the reactive deload
// (D-289, D-295).
const (
	// MissedDays is the shortest gap of a missed session. A gap from
	// MissedDays to BreakDays-1 days holds the target at 3 reps in
	// reserve for one session.
	MissedDays = 7
	// DeloadDays is the length of a deload, from the date of the
	// session that started it.
	DeloadDays = 7
	// DeclineSessions is the count of declines in a row that an
	// exercise needs, and DeclineExercises is the count of such
	// exercises that starts a deload.
	DeclineSessions  = 2
	DeclineExercises = 2
)

// missed tells that a gap is a missed session: 7 to 13 days (D-294).
func missed(gap int) bool { return gap >= MissedDays && gap < BreakDays }

// afterMissed gives the target of the session before the last trained
// session, when the last trained session followed a missed session. The
// missed session holds its target at 3 reps in reserve for that one
// session (D-294), so the next target gets back the reps in reserve of
// the target before it.
func (in Input) afterMissed() (domain.PlannedExercise, bool) {
	var trained []Outcome
	for _, o := range in.History {
		if o.trained() {
			trained = append(trained, o)
		}
	}
	n := len(trained)
	if n < 2 || !missed(day(trained[n-1].Date)-day(trained[n-2].Date)) {
		return domain.PlannedExercise{}, false
	}
	return trained[n-2].Target, true
}

// restoreRIR gives each working set the reps in reserve of the set of
// the same position of a target, inside the bounds of the exercise. It
// tells whether a value changed.
func (b *builder) restoreRIR(p domain.PlannedExercise) bool {
	rir := RIRRange(b.in.Exercise)
	changed := false
	for i := range b.target.Working {
		if len(p.Working) == 0 {
			break
		}
		r := rir.clamp(p.Working[min(i, len(p.Working)-1)].RIR)
		if b.target.Working[i].RIR != r {
			b.target.Working[i].RIR = r
			changed = true
		}
	}
	return changed
}

// deloadOf gives the date of the session that started the deload of a
// date, or "" when the date is in no deload. A deload covers the
// DeloadDays dates after the session that started it (D-295). The same
// dates give the deload targets and the deload sessions, so a session in
// the deload is no evidence, and a session on the date of the start is
// evidence.
func (in Input) deloadOf(date string) string {
	d := day(date)
	for _, s := range in.Deloads {
		if from := day(s); d > from && d <= from+DeloadDays {
			return s
		}
	}
	return ""
}

// rulesHistory gives the history that the rules read: the history with
// no deload session. After a deload, the targets of before the deload
// return (D-295), so a deload session is no evidence for the next
// target. A history of deload sessions alone stays as it is.
func (in Input) rulesHistory() []Outcome {
	out := slices.DeleteFunc(slices.Clone(in.History), func(o Outcome) bool { return in.deloadOf(o.Date) != "" })
	if len(out) == 0 {
		return in.History
	}
	return out
}

// deload gives the target of a deload session (D-295): 0.6 times the
// sets, rounded to the nearest whole set and at least 1, at the same
// load and 3 reps in reserve.
//
// The reason of the deload replaces the reason of the other rules, and
// it names the target after the deload. The rules stay in the record.
func (b *builder) deload(from string) {
	after := b.repsText()
	n := len(b.target.Working)
	k := max(1, (6*n+5)/10)
	b.target.Working = b.target.Working[:k]
	b.before = b.before[:k]
	b.setRIR(3)
	end := day(from) + DeloadDays
	b.reason = nil
	b.rule(RuleDeload, fmt.Sprintf("Your reps went down in %d sessions in a row on %d or more exercises, so this week is a deload. Until %s, the target is %s at %s, at 3 reps in reserve. Then it goes back to %s.",
		DeclineSessions, DeclineExercises, dateText(end), b.repsText(), b.loadText(), after))
}

// dateText gives the date of a count of days from 1970-01-01.
func dateText(d int) string {
	return time.Unix(int64(d)*86400, 0).UTC().Format(domain.DateLayout)
}

// decline tells that session b has fewer total reps of working sets
// than session a, at the same load or a heavier load (D-295). The load
// is the logged weight of the first working set.
func decline(a, b Outcome) bool {
	la, lb := a.working(), b.working()
	if len(la) == 0 || len(lb) == 0 {
		return false
	}
	return totalReps(lb) < totalReps(la) && lb[0].Weight >= la[0].Weight
}

func totalReps(sets []domain.SetLog) int {
	t := 0
	for _, s := range sets {
		t += s.Reps
	}
	return t
}

// Deloads gives the start date of each reactive deload of the owner,
// oldest first, from the history of each exercise (D-289, D-295). Each
// history is the history of one exercise, oldest first, as in Input.
//
// A deload starts at the date of a session when 2 or more exercises
// declined in 2 sessions in a row. An exercise counts when its last 3
// trained sessions up to that date show 2 declines, and its last one is
// less than BreakDays days before that date (D-296). The sessions of a deload
// and the sessions before its end start no new deload. So after a
// deload, each exercise needs 3 new sessions for the next one.
//
// The caller gives the result to the Deloads field of each input. The
// same histories give the same dates.
func Deloads(histories [][]Outcome) ([]string, error) {
	var days []int
	trained := make([][]Outcome, len(histories))
	for i, h := range histories {
		for j, o := range h {
			// An error holds no date, because a date is data of the log
			// (D-80).
			if _, err := time.Parse(domain.DateLayout, o.Date); err != nil {
				return nil, inputError("histories[%d][%d] date: want the form %s", i, j, domain.DateLayout)
			}
			if j > 0 && day(o.Date) < day(h[j-1].Date) {
				return nil, inputError("histories[%d][%d] date: want dates in order", i, j)
			}
			if o.trained() {
				trained[i] = append(trained[i], o)
				days = append(days, day(o.Date))
			}
		}
	}
	sort.Ints(days)
	days = slices.Compact(days)
	var out []string
	end := -1 << 31
	for _, d := range days {
		if d <= end {
			continue
		}
		count := 0
		for _, h := range trained {
			var seq []Outcome
			for _, o := range h {
				if od := day(o.Date); od > end && od <= d {
					seq = append(seq, o)
				}
			}
			n := len(seq)
			if n < DeclineSessions+1 || day(seq[n-1].Date) <= d-BreakDays {
				continue
			}
			ok := true
			for k := n - DeclineSessions; k < n; k++ {
				ok = ok && decline(seq[k-1], seq[k])
			}
			if ok {
				count++
			}
		}
		if count >= DeclineExercises {
			out = append(out, dateText(d))
			end = d + DeloadDays
		}
	}
	return out, nil
}

// overrideCheck holds the violations of an override (D-293).
type overrideCheck struct {
	in  Input
	out []Violation
}

// CheckOverride reads an override of the owner against the
// recommendation that it replaces, and gives each violation (D-69,
// D-293). An override changes the load and the reps of each working set
// alone. It keeps the count of sets, the reps in reserve, and the rest
// of the recommendation. Each set has 6 to 20 reps, each load is a
// valid load of the machine, and a calibration set has the reps and the
// load of the first working set. No violation means that the policy
// accepts the override (D-23).
func CheckOverride(o, rec domain.PlannedExercise, in Input) ([]Violation, error) {
	if err := (Input{Exercise: in.Exercise, Entry: in.Entry}).check(); err != nil {
		return nil, err
	}
	if rec.Exercise != in.Exercise.ID {
		return nil, inputError("recommendation exercise %q: want %q", rec.Exercise, in.Exercise.ID)
	}
	c := overrideCheck{in: in}
	if o.Exercise != rec.Exercise {
		c.add("exercise", "exercise %q: want %q", o.Exercise, rec.Exercise)
	}
	if o.RestSeconds != rec.RestSeconds {
		c.add("rest", "rest %d s: want %d s, the rest of the recommendation", o.RestSeconds, rec.RestSeconds)
	}
	if len(o.Working) != len(rec.Working) {
		c.add("working", "%d sets: want %d, the sets of the recommendation", len(o.Working), len(rec.Working))
	}
	if len(o.Calibration) != len(rec.Calibration) {
		c.add("calibration", "%d calibration sets: want %d", len(o.Calibration), len(rec.Calibration))
	}
	for i, s := range o.Working {
		where := fmt.Sprintf("working[%d]", i)
		c.reps(where, s.Reps)
		c.load(where, s.Load)
		if i < len(rec.Working) && s.RIR != rec.Working[i].RIR {
			c.add(where, "rir %d: want %d, the reps in reserve of the recommendation", s.RIR, rec.Working[i].RIR)
		}
	}
	for i, s := range o.Calibration {
		where := fmt.Sprintf("calibration[%d]", i)
		if len(o.Working) > 0 && (s.Reps != o.Working[0].Reps || s.Load != o.Working[0].Load) {
			c.add(where, "%d reps at %s: want the reps and the load of the first working set", s.Reps, s.Load)
		}
	}
	return c.out, nil
}

func (c *overrideCheck) add(where, format string, args ...any) {
	c.out = append(c.out, Violation{RuleOverride, where, fmt.Sprintf(format, args...)})
}

func (c *overrideCheck) reps(where string, n int) {
	if !RepLimits.Has(n) {
		c.add(where, "reps %d: want %d to %d", n, RepLimits.Min, RepLimits.Max)
	}
}

func (c *overrideCheck) load(where string, l domain.Load) {
	available := c.in.Entry.Available()
	switch {
	case !slices.Contains(available, l):
		c.add(where, "load %s: not an available weight of machine %q", l, c.in.Entry.Machine)
	case !Valid(l, available):
		c.add(where, "load %s: not a multiple of 5 lb or the weight that D-149 selects", l)
	}
}
