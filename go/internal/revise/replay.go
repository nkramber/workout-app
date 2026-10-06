package revise

import (
	"context"
	"fmt"
	"slices"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// Origin names the step that wrote a decision record of the plan.
type Origin string

const (
	// OriginPlan is a record of a new plan, from policy.Decide.
	OriginPlan Origin = "plan"
	// OriginRevision is a record of a revision, from policy.Revise.
	OriginRevision Origin = "revision"
)

// Replayed is one decision record of the active plan and its replay
// under the current policy version (work area 8.3). Rebuilt is false
// when the replay could not make the input of the record again: the
// exercise is not in the catalog, its machine is not confirmed, or the
// policy refuses the input. Now is then the zero record.
type Replayed struct {
	Session int
	Origin  Origin
	Stored  policy.Record
	Now     policy.Record
	Rebuilt bool
}

// CopyReplayed is one target copy of a finished workout (D-291) and
// the target of the rules of the current policy version on the date of
// the workout, with each workout that started before it. Override tells
// that the copy holds an override of the owner (D-293). Rebuilt is false
// when the replay could not make the input of the copy.
//
// The store keeps no time of a revision, so the replay does not know
// which workouts of the plan the server had read when it gave the copy.
// A phone can start a workout before the revision of the last workout,
// and then it shows the target before that revision. So the replay
// checks the copy with each history from the workouts of the earlier
// plans alone up to each workout that started before it. Lag is the
// count of the newest workouts of its plan that the first history that
// passes leaves out. Violations holds each bound of the current version
// that the copy breaks with the full history, when no history passes.
//
// The policy checks an override against its recommendation with
// policy.CheckOverride (D-69, D-293). The workout keeps the working sets
// of the recommendation, and the copy gives its other fields. So
// Violations of an override holds each violation of that check, and its
// Lag is 0.
type CopyReplayed struct {
	Copy       domain.PlannedExercise
	Override   bool
	Rules      policy.Decision
	Lag        int
	Violations []policy.Violation
	Rebuilt    bool
}

// Replay is the replay of the stored data of one user under the current
// policy version. HasPlan is false when the user has no plan. The
// replay calls no model, and it writes nothing.
//
// Layout holds each violation of the rule rotation.no-repeat by the
// sessions of the active plan (D-328). The store keeps no request of a
// plan, so the replay reads the exercises and the cardio of the plan in
// place of the request, and selects no group. So it does not read the
// rule rotation.cover, and LayoutApplies is false when the exercises of
// the plan can make no split.
type Replay struct {
	HasPlan       bool
	Records       []Replayed
	Copies        []CopyReplayed
	LayoutApplies bool
	Layout        []policy.Violation
}

// Replay reads the finished workouts, their target copies, the
// inventory, and the active plan of uid, and replays each decision
// record and each target copy under the current policy version (D-176).
//
// A record of a new plan replays through policy.Decide with its stored
// proposal, and with the history of the workouts of the earlier plans.
// A record of a revision replays through policy.Revise, with the
// history up to the last workout that logged its exercise. A target
// copy replays through policy.Next and policy.Check, with the history
// before its workout.
//
// The stores hold no copy of an input, so the replay makes each input
// again from the stored data. A record holds the hash of its input. A
// hash that differs tells that the replay did not make the same input,
// or that a new version changed the form of the input.
func (r *Reviser) Replay(ctx context.Context, uid string) (Replay, error) {
	history, err := r.finished(ctx, uid, AllHistory)
	if err != nil {
		return Replay{}, err
	}
	inv, err := r.Inventory.Get(ctx, uid)
	if err != nil {
		return Replay{}, fmt.Errorf("revise: the inventory: %w", err)
	}
	pi := inventory.ForPlan(inv)
	var out Replay
	for i, w := range history {
		for _, t := range w.Targets {
			out.Copies = append(out.Copies, r.replayCopy(uid, history[:i], w, t, pi))
		}
	}
	p, ok, err := r.Plans.Get(ctx, uid)
	if err != nil {
		return Replay{}, fmt.Errorf("revise: the plan: %w", err)
	}
	if !ok {
		return out, nil
	}
	out.HasPlan = true
	out.LayoutApplies, out.Layout = layout(p)
	for i, s := range p.Sessions {
		for _, e := range s.Exercises {
			out.Records = append(out.Records, r.replayRecord(uid, p, i, e.Record, history, pi))
		}
	}
	return out, nil
}

// layout replays the sessions of a plan under the rotation rules of the
// current policy version.
func layout(p plan.Plan) (bool, []policy.Violation) {
	t := domain.DefaultBodyTables()
	r := policy.Rotation{Sessions: len(p.Sessions), Groups: map[domain.ExerciseID][]domain.MuscleGroup{}}
	sessions := make([][]domain.ExerciseID, len(p.Sessions))
	for i, s := range p.Sessions {
		r.Filler = r.Filler || s.Cardio != nil
		for _, e := range s.Exercises {
			id := e.Target.Exercise
			sessions[i] = append(sessions[i], id)
			if _, ok := r.Groups[id]; ok {
				continue
			}
			r.Order = append(r.Order, id)
			r.Groups[id] = t.Groups[id]
			r.Filler = r.Filler || len(t.Groups[id]) == 0
		}
	}
	return r.Applies(), policy.CheckRotation(r, sessions)
}

// originOf gives the step that wrote a record. policy.Decide gives a
// Luna source or the rules fallback, and policy.Revise gives neither.
func originOf(rec policy.Record) Origin {
	if rec.Source == policy.SourceLuna || slices.Contains(rec.Rules, policy.RuleFallback) {
		return OriginPlan
	}
	return OriginRevision
}

func (r *Reviser) replayRecord(uid string, p plan.Plan, session int, rec policy.Record, history []workout.Workout, pi inventory.PlanInput) Replayed {
	out := Replayed{Session: session, Origin: originOf(rec), Stored: rec}
	switch out.Origin {
	case OriginPlan:
		// A new plan reads each workout before it, and no workout links
		// to a plan before its save (D-248). An input that the policy
		// refuses gets no history, as in plan.Make.
		before := slices.DeleteFunc(slices.Clone(history), func(w workout.Workout) bool { return w.Plan.PlanCreatedAt.Equal(p.CreatedAt) })
		e, ok := domain.DefaultCatalog().Exercise(rec.Exercise)
		if !ok {
			return out
		}
		entry, ok := pi.Inventory.Entry(e.Machine)
		if !ok {
			return out
		}
		in := policy.Input{Exercise: e, Entry: entry, Today: p.Today, Estimate: pi.Estimates[e.ID], History: outcomes(plan.Plan{}, before, e.ID), Deloads: r.deloads(uid, plan.Plan{}, before)}
		if _, err := policy.Next(in); err != nil {
			in.History, in.Deloads = nil, nil
		}
		now, err := policy.Decide(in, policy.Proposal{Model: rec.Model, Effort: rec.Effort, PromptHash: rec.PromptHash, Target: rec.Proposal})
		if err != nil {
			return out
		}
		out.Now, out.Rebuilt = now, true
	case OriginRevision:
		// The last workout that logged the exercise wrote the record.
		last := -1
		for i, w := range history {
			if _, ok := seen(p, w, rec.Exercise); ok && slices.ContainsFunc(w.Exercises(), func(l workout.ExerciseLog) bool { return l.Exercise == rec.Exercise }) {
				last = i
			}
		}
		if last < 0 {
			return out
		}
		upTo := history[:last+1]
		in, ok := r.input(uid, p, upTo, pi, rec.Exercise, r.deloads(uid, p, upTo))
		if !ok {
			return out
		}
		now, err := policy.Revise(in)
		if err != nil {
			return out
		}
		now.Model, now.Effort, now.PromptHash = rec.Model, rec.Effort, rec.PromptHash
		out.Now, out.Rebuilt = now, true
	}
	return out
}

func (r *Reviser) replayCopy(uid string, before []workout.Workout, w workout.Workout, t domain.PlannedExercise, pi inventory.PlanInput) CopyReplayed {
	oi := slices.IndexFunc(w.Overrides, func(o workout.SeenOverride) bool { return o.Exercise == t.Exercise })
	out := CopyReplayed{Copy: t, Override: oi >= 0}
	e, ok := domain.DefaultCatalog().Exercise(t.Exercise)
	if !ok {
		return out
	}
	entry, ok := pi.Inventory.Entry(e.Machine)
	if !ok {
		return out
	}
	input := func(hist []workout.Workout) policy.Input {
		return policy.Input{Exercise: e, Entry: entry, Today: w.Date, Estimate: pi.Estimates[e.ID], History: outcomes(plan.Plan{}, hist, e.ID), Deloads: r.deloads(uid, plan.Plan{}, hist)}
	}
	in := input(before)
	d, err := policy.Next(in)
	if err != nil {
		return out
	}
	out.Rules, out.Rebuilt = d, true
	if out.Override {
		rec := t
		rec.Working = slices.Clone(w.Overrides[oi].Recommended)
		if out.Violations, err = policy.CheckOverride(t, rec, in); err != nil {
			out.Rebuilt = false
		}
		return out
	}
	full, err := policy.Check(t, in)
	if err != nil {
		out.Rebuilt = false
		return out
	}
	if len(full) == 0 {
		return out
	}
	// The workouts of the plan of the copy, the oldest first. Each
	// shorter history leaves out one more of the newest of them.
	var mine []int
	for i, b := range before {
		if b.Plan.PlanCreatedAt.Equal(w.Plan.PlanCreatedAt) {
			mine = append(mine, i)
		}
	}
	for lag := 1; lag <= len(mine); lag++ {
		left := mine[len(mine)-lag:]
		hist := slices.Clone(before)
		for _, i := range slices.Backward(left) {
			hist = slices.Delete(hist, i, i+1)
		}
		v, err := policy.Check(t, input(hist))
		if err == nil && len(v) == 0 {
			out.Lag = lag
			return out
		}
	}
	out.Violations = full
	return out
}
