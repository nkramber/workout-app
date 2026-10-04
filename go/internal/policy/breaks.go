package policy

import (
	"fmt"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The values of the long-break table and of the first sessions after a
// break (D-151, D-179). A gap is the count of days from the last
// session of the exercise with a logged working set to the next
// session.
const (
	// BreakDays is the shortest long break. The first sessions after a
	// break also last at least this count of days.
	BreakDays = 14
	// LongBreakDays starts the row of 28 to 90 days.
	LongBreakDays = 28
	// RecalibrateDays starts the row of 91 days or more.
	RecalibrateDays = 91
	// FirstSessions is the least count of first sessions after a break.
	FirstSessions = 3
)

// day gives the count of days from 1970-01-01 to a date. The input
// check proves the form of each date before a rule reads it.
func day(s string) int {
	t, err := time.Parse(domain.DateLayout, s)
	if err != nil {
		return 0
	}
	return int(t.Unix() / 86400)
}

// pause is the break state of an exercise for its next session.
type pause struct {
	// gap is the count of days from the last session with a logged
	// working set to today, or -1 when no session has one.
	gap int
	// first tells that the next session is one of the first sessions
	// after a break, after the return session.
	first bool
	// ended tells that the first sessions ended with the last logged
	// session, and before is the count of sets of the target before
	// the break.
	ended  bool
	before int
}

// pause reads the dates of the history. The return session is the
// latest session with a logged working set that follows a gap of 14
// days or more. When Returning is true, the first such session of the
// exercise is a return too. The first sessions are the first 3 sessions
// from the return, or the sessions in the first 14 days from it, the
// longer of the two (D-151).
func (in Input) pause() pause {
	p := pause{gap: -1}
	var trained []int
	for i, o := range in.History {
		if o.trained() {
			trained = append(trained, i)
		}
	}
	if len(trained) == 0 {
		return p
	}
	today := day(in.Today)
	last := day(in.History[trained[len(trained)-1]].Date)
	p.gap = today - last
	if p.gap >= BreakDays {
		return p
	}
	r := -1
	for k := len(trained) - 1; k >= 0; k-- {
		if k == 0 {
			if in.Returning {
				r = 0
			}
			break
		}
		if day(in.History[trained[k]].Date)-day(in.History[trained[k-1]].Date) >= BreakDays {
			r = k
			break
		}
	}
	if r < 0 {
		return p
	}
	from := day(in.History[trained[r]].Date)
	within := func(session, days int) bool { return session <= FirstSessions || days < BreakDays }
	n := len(trained) - r
	if within(n+1, today-from) {
		p.first = true
		return p
	}
	if within(n, last-from) && trained[r] > 0 {
		p.ended = true
		p.before = len(in.History[trained[r]-1].Target.Working)
	}
	return p
}

// firstSessions tells that the next session is one of the first
// sessions after a break: the return session, a later first session,
// or the start of an exercise after a break of the owner (D-151).
func (in Input) firstSessions(p pause) bool {
	return p.first || p.gap >= BreakDays || (p.gap < 0 && in.Returning)
}

// resume gives the target of the return session after a gap of 14 days
// or more, from the long-break table (D-151, D-179). The load goes down
// 10 percent, 20 percent, or to 70 percent, and a halfway value rounds
// down (D-148). The first two rows remove one set. Each working set
// stops at 3 reps in reserve. After 91 days or more, Next makes the
// first set the calibration (D-301).
func (b *builder) resume(gap int) {
	pct, rule := domain.Load(90), RuleBreakShort
	switch {
	case gap >= RecalibrateDays:
		pct, rule = 70, RuleBreakRecalibrate
	case gap >= LongBreakDays:
		pct, rule = 80, RuleBreakLong
	}
	for i, s := range b.target.Working {
		c := s.Load * pct / 100
		r := Round(c, Return)
		l := Select(r, b.available)
		b.before[i] = c
		b.target.Working[i].Load = l
		b.machineWeight(i, r, l)
	}
	if rule != RuleBreakRecalibrate && len(b.target.Working) > 1 {
		n := len(b.target.Working) - 1
		b.target.Working = b.target.Working[:n]
		b.before = b.before[:n]
	}
	b.setRIR(3)
	if rule == RuleBreakRecalibrate {
		b.rule(rule, fmt.Sprintf("Your last logged set of this exercise was %d days ago. The load goes down to %s at 3 reps in reserve.", gap, b.loadText()))
		return
	}
	b.rule(rule, fmt.Sprintf("Your last logged set of this exercise was %d days ago. The target is %s at %s, at 3 reps in reserve.", gap, b.repsText(), b.loadText()))
}

// restore gives the target the count of sets of the target before the
// break, after the first sessions after the break (D-179). Each new set
// is a copy of the last set.
func (b *builder) restore(n int) {
	if len(b.target.Working) >= n {
		return
	}
	for len(b.target.Working) < n {
		w := b.target.Working[len(b.target.Working)-1]
		b.target.Working = append(b.target.Working, w)
		b.before = append(b.before, w.Load)
	}
	b.rule(RuleBreakSets, fmt.Sprintf("The first sessions after your break ended, so the target is back to %d sets.", n))
}
