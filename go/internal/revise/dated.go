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
// the reason of the rules, with the cause CauseDate, and it keeps its
// override. Each other exercise keeps its target and its reason. The
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
			if sameTarget(e.Target, now.Target) {
				continue
			}
			x := plan.Exercise{Target: now.Target, Record: now, Reason: now.Reason, ReasonSource: policy.SourceRules, ReasonCause: CauseDate, Override: e.Override}
			if len(now.Target.Calibration) > 0 {
				if x.Calibration, err = policy.CalibrationTable(in); err != nil {
					continue
				}
			}
			changed[id] = x
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

// sameTarget tells whether two targets have the same sets and rest. A
// stored target and a target of the rules can differ in an empty list
// alone.
func sameTarget(a, b domain.PlannedExercise) bool {
	return a.Exercise == b.Exercise && a.RestSeconds == b.RestSeconds &&
		slices.Equal(a.Calibration, b.Calibration) && slices.Equal(a.Working, b.Working)
}
