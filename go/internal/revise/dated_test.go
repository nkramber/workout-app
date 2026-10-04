package revise

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
)

func (f *fixture) forDate(today string) plan.Plan {
	f.t.Helper()
	p, err := f.reviser.ForDate(context.Background(), uid, f.plan(), today)
	if err != nil {
		f.t.Fatalf("ForDate: %v", err)
	}
	return p
}

// The targets on the date of the next session: a missed week and a
// break change the target of each session that holds the exercise, and
// ForDate saves nothing (D-151, D-179, D-294).
func TestForDate(t *testing.T) {
	f := newFixture(t)
	press := target("chest_press", 3, 12, lb(25))
	row := target("seated_row", 3, 12, lb(40))
	done := func(p domain.PlannedExercise) []domain.SetLog {
		var out []domain.SetLog
		for _, s := range p.Working {
			out = append(out, set(s.Reps, s.Load, 3))
		}
		return out
	}
	// A plan with no revision reads no store, so a store failure does
	// not reach it.
	f.reviser.Workouts = nil
	if p, err := f.reviser.ForDate(context.Background(), uid, f.plan(), "2026-10-30"); err != nil || !sameTarget(p.Sessions[0].Exercises[0].Target, press) {
		t.Fatalf("ForDate of a plan with no revision = %v, want the plan", err)
	}
	f.reviser.Workouts = f.workouts

	// Each exercise of a plan starts as a return (D-238). The fourth
	// session, 15 days after the first, ends the first sessions, so the
	// rules of a missed session and of a break apply after it.
	for n, date := range []string{"2026-10-01", "2026-10-06", "2026-10-11", "2026-10-16"} {
		cur := f.plan().Sessions[0].Exercises
		press, row = cur[0].Target, cur[1].Target
		f.revise(f.workout(n+1, date, true, log{press, done(press)}, log{row, done(row)}))
	}
	stored := f.plan()

	// 2 days later, nothing changes, and the reason of the revision stays.
	same := f.forDate("2026-10-18")
	for i, s := range same.Sessions {
		for j, e := range s.Exercises {
			if old := stored.Sessions[i].Exercises[j]; !sameTarget(e.Target, old.Target) || e.Reason != old.Reason || e.ReasonCause != old.ReasonCause {
				t.Fatalf("session %d exercise %d: %+v, want the stored exercise", i, j, e)
			}
		}
	}

	for _, tc := range []struct {
		today string
		rule  policy.RuleID
	}{
		{"2026-10-23", policy.RuleMissed},
		{"2026-10-29", policy.RuleMissed},
		{"2026-10-30", policy.RuleBreakShort},
	} {
		p := f.forDate(tc.today)
		for i, s := range p.Sessions {
			for _, e := range s.Exercises {
				if !slices.Contains(e.Record.Rules, tc.rule) || e.ReasonCause != CauseDate || e.ReasonSource != policy.SourceRules || e.Reason != e.Record.Reason {
					t.Fatalf("%s session %d %s: rules %v, cause %q, want %s with the reason of the rules", tc.today, i, e.Target.Exercise, e.Record.Rules, e.ReasonCause, tc.rule)
				}
				for _, w := range e.Target.Working {
					if w.RIR != 3 {
						t.Errorf("%s %s: rir %d, want 3", tc.today, e.Target.Exercise, w.RIR)
					}
				}
			}
		}
	}
	if after := f.plan(); after.Revisions != stored.Revisions || !sameTarget(after.Sessions[0].Exercises[0].Target, stored.Sessions[0].Exercises[0].Target) {
		t.Fatal("ForDate changed the stored plan")
	}

	// A date before the last session, and a bad date, change nothing.
	early := f.forDate("2026-10-15")
	if !sameTarget(early.Sessions[0].Exercises[0].Target, stored.Sessions[0].Exercises[0].Target) {
		t.Error("a date before the last session changed the target")
	}
	if _, err := f.reviser.ForDate(context.Background(), uid, f.plan(), "2026-10-23T00"); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("ForDate of a bad date = %v, want ErrInvalid", err)
	}
}

// ForDate keeps the override of an exercise, and a revision removes it,
// because an override is for one session (D-293).
func TestForDateOverride(t *testing.T) {
	f := newFixture(t)
	press := target("chest_press", 3, 12, lb(25))
	id := f.workout(1, "2026-10-01", true, log{press, []domain.SetLog{set(12, lb(25), 3), set(12, lb(25), 3), set(12, lb(25), 3)}})
	f.revise(id)
	over := target("chest_press", 3, 10, lb(30))
	o := &plan.Override{Target: over, Recommendation: f.plan().Sessions[0].Exercises[0].Target, Reason: "Felt easy.", Today: "2026-10-02"}
	if err := f.plans.Update(context.Background(), uid, created, func(p *plan.Plan) error {
		p.SetOverride("chest_press", o)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// The override stays on a date with the rules of the date of its save.
	// After a missed week or a break, the rules of the date changed, so the
	// override expires and the recommendation applies (D-293, D-294).
	// The stored override does not change.
	for _, tc := range []struct {
		today   string
		expired bool
	}{{"2026-10-03", false}, {"2026-10-07", false}, {"2026-10-08", true}, {"2026-10-20", true}} {
		p := f.forDate(tc.today)
		for i := range p.Sessions {
			got := p.Sessions[i].Exercises[0].Override
			if got == nil || got.Reason != "Felt easy." || !sameTarget(got.Target, over) || got.Expired != tc.expired {
				t.Fatalf("%s session %d: override %+v, want it with expired %v", tc.today, i, got, tc.expired)
			}
		}
	}
	if f.plan().Sessions[0].Exercises[0].Override.Expired {
		t.Fatal("ForDate stored the expiry")
	}
	// An override that the owner saved after the missed week, against the
	// held recommendation, stays.
	late := *o
	late.Today = "2026-10-09"
	if err := f.plans.Update(context.Background(), uid, created, func(p *plan.Plan) error {
		p.SetOverride("chest_press", &late)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := f.forDate("2026-10-10").Sessions[0].Exercises[0].Override; got == nil || got.Expired {
		t.Fatalf("an override of the same rules of the date: %+v, want it kept", got)
	}
	if err := f.plans.Update(context.Background(), uid, created, func(p *plan.Plan) error {
		p.SetOverride("chest_press", o)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// The next workout logs the override, and the revision starts from it.
	id = f.workout(2, "2026-10-03", true, log{over, []domain.SetLog{set(10, lb(30), 3), set(10, lb(30), 3), set(10, lb(30), 3)}})
	f.revise(id)
	p := f.plan()
	for i, s := range p.Sessions {
		e := s.Exercises[0]
		if e.Override != nil {
			t.Fatalf("session %d: the revision kept the override", i)
		}
		if e.Target.Working[0].Load < lb(30) {
			t.Fatalf("session %d: target %+v, want a target from the override at 30 lb", i, e.Target.Working)
		}
	}
}
