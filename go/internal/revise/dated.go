package revise

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// ForDate gives the plan with the targets of the rules on a date, the
// date of the next session. A revision gives each target on the date of
// the finished workout, so a target can not know the gap to the next
// session. ForDate applies the long-break table (D-151, D-179), the rule
// of a missed session (D-294), and the start and the end of a deload
// (D-295) on that date.
//
// ForDate changes the target of an exercise that the owner logged under
// this plan, when the rules give a different target on the date. A
// revision gave the target of such an exercise, so the target is a
// target of the rules (D-288). A deload that a later workout of another
// exercise started also changes it. Such an exercise gets the target and
// the reason of the rules, with the cause CauseDate. Its override stays,
// and the override expires when the rules of the date changed after its
// save. Each other exercise keeps its target and its reason. The
// result shares no memory with p, and ForDate saves nothing: the same
// history and the same date give the same targets.
func (r *Reviser) ForDate(ctx context.Context, uid string, p plan.Plan, today string) (plan.Plan, error) {
	if _, err := time.Parse(domain.DateLayout, today); err != nil {
		return plan.Plan{}, fmt.Errorf("%w: today: want the form %s", domain.ErrInvalid, domain.DateLayout)
	}
	// Each finished workout revises the plan, and a deload starts at a
	// finished workout. So a plan with no revision has no exercise that
	// the owner logged under it, and ForDate reads no store for it.
	if p.Revisions == 0 {
		return copyPlan(p), nil
	}
	history, err := r.finished(ctx, uid)
	if err != nil {
		return plan.Plan{}, err
	}
	inv, err := r.Inventory.Get(ctx, uid)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("revise: the inventory: %w", err)
	}
	pi := inventory.ForPlan(inv)
	dl := r.deloads(uid, p, history)

	changed := map[domain.ExerciseID]plan.Exercise{}
	for _, s := range p.Sessions {
		for _, e := range s.Exercises {
			id := e.Target.Exercise
			if _, done := changed[id]; done {
				continue
			}
			in, ok := r.input(uid, p, history, pi, id, dl)
			if !ok {
				continue
			}
			// A plan made after the last session of the exercise gave
			// its target with no history (D-238), and a date before
			// that session is a clock of the phone that is behind.
			last := in.History[len(in.History)-1].Date
			if last < p.Today || today < last {
				continue
			}
			in.Today = today
			now, err := policy.Revise(in)
			if err != nil {
				r.warn("revise: the policy refused the date", "uid", uid, "exercise_id", string(id), "err", err.Error())
				continue
			}
			x, change := e, false
			if !sameTarget(e.Target, now.Target) {
				x = plan.Exercise{Target: now.Target, Record: now, Reason: now.Reason, ReasonSource: policy.SourceRules, ReasonCause: CauseDate, Override: e.Override}
				if len(now.Target.Calibration) > 0 {
					if x.Calibration, err = policy.CalibrationTable(in); err != nil {
						continue
					}
				}
				change = true
			}
			if o := e.Override; o != nil && stale(in, *o, now) {
				c := *o
				c.Expired = true
				x.Override, change = &c, true
			}
			if change {
				changed[id] = x
			}
		}
	}

	q := copyPlan(p)
	apply(&q, changed)
	return q, nil
}

// copyPlan gives a copy of the sessions of a plan, so a change of the
// copy does not change p.
func copyPlan(p plan.Plan) plan.Plan {
	q := p
	q.Sessions = slices.Clone(p.Sessions)
	for i := range q.Sessions {
		q.Sessions[i].Exercises = slices.Clone(p.Sessions[i].Exercises)
	}
	return q
}

// dateRules are the rules that read the date of the next session.
var dateRules = []policy.RuleID{policy.RuleMissed, policy.RuleBreakShort, policy.RuleBreakLong, policy.RuleBreakRecalibrate, policy.RuleDeload}

// stale tells whether the rules of the date of the next session changed
// after the save of an override (D-293 to D-295). The owner chose the
// override against the recommendation of the date of the save. After a
// missed session, a break, or the start or the end of a deload, the
// override no longer has the check of the policy for the new date, so
// the recommendation applies. An override with a date that the policy
// refuses is stale too.
func stale(in policy.Input, o plan.Override, now policy.Record) bool {
	saved := o.Today
	if saved == "" {
		saved = o.At.UTC().Format(domain.DateLayout)
	}
	in.Today = saved
	then, err := policy.Revise(in)
	if err != nil {
		return true
	}
	of := func(rs []policy.RuleID) []policy.RuleID {
		return slices.DeleteFunc(slices.Clone(rs), func(r policy.RuleID) bool { return !slices.Contains(dateRules, r) })
	}
	return !slices.Equal(of(then.Rules), of(now.Rules))
}

// sameTarget tells whether two targets have the same sets and rest. A
// stored target and a target of the rules can differ in an empty list
// alone.
func sameTarget(a, b domain.PlannedExercise) bool {
	return a.Exercise == b.Exercise && a.RestSeconds == b.RestSeconds &&
		slices.Equal(a.Calibration, b.Calibration) && slices.Equal(a.Working, b.Working)
}
