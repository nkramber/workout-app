// Package revise revises the plan after a finished workout (work area
// 7.1). The sync of the outbox calls it when it applies a finished
// workout (D-292). The rules of the policy give the next target of each
// exercise that the workout logged, in each session of the plan that
// holds it (D-290). The reviser of Luna writes the reason alone, and a
// check refuses a reason that names no logged set. The reason of the
// rules then shows (D-288).
//
// The history of an exercise comes from the finished workouts of the
// owner. Each workout holds the target that the owner saw at its start
// (D-291). A workout of an older phone holds no such copy. Its target
// comes from the linked session of the active plan only when the plan
// has no revision yet, because the plan then still holds the target
// that the owner saw.
//
// A plan, a log, and a reason are data of the owner. A log line and an
// error hold ids and counts alone (D-80).
package revise

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// MaxHistory is the largest count of workouts that a revision reads,
// the newest first. At 4 sessions each week, it holds about 6 months.
const MaxHistory = 100

// AllHistory is the limit of finished that reads each workout. A new
// plan reads each workout, because it gives an exercise with history its
// target from that history (D-301). So an exercise that the newest
// MaxHistory workouts omit still gets the return of D-179. Each workout
// is one document, and a new plan is rare, so the read stays small.
const AllHistory = 0

// pageSize is the page of each read of the workouts.
const pageSize = 50

// Timeout limits one revision: the reads, the reviser call, and the
// write. The revision continues when the sync request ends first, so a
// dropped connection does not lose it.
const Timeout = ai.ReviserTimeout + 30*time.Second

// Reviser revises the plan of a user. AI holds the provider and the cap
// hook. Now gives the time, and Log gets one line for each revision. A
// nil Log writes no line.
type Reviser struct {
	AI        *ai.Client
	Plans     plan.Store
	Workouts  workout.Store
	Inventory inventory.Store
	Now       func() time.Time
	Log       *slog.Logger
}

// Result is the outcome of one revision. Revised is false when nothing
// changed: the user has no plan, the plan has this revision already, the
// workout is not a finished workout of the user, or a new plan replaced
// the plan during the revision. Exercises names each exercise that got a
// new target. Status is the status of the reviser call, or "" when no
// call occurred. Luna counts the reasons of Luna that passed the check.
type Result struct {
	Revised   bool
	Exercises []domain.ExerciseID
	Status    ai.Status
	Luna      int
}

// errSeen stops an update when another call revised the plan for the
// same workout first.
var errSeen = errors.New("revise: the plan has this revision")

// Revise revises the plan of uid after the finished workout workoutID.
// It gives an error for a failed read or write of a store alone. A
// failed reviser call gives the reasons of the rules (D-292).
func (r *Reviser) Revise(ctx context.Context, uid, workoutID string) (Result, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), Timeout)
	defer cancel()
	p, ok, err := r.Plans.Get(ctx, uid)
	if err != nil {
		return Result{}, fmt.Errorf("revise: the plan: %w", err)
	}
	if !ok || p.Revised(workoutID) {
		return Result{}, nil
	}
	history, err := r.finished(ctx, uid, MaxHistory)
	if err != nil {
		return Result{}, err
	}
	i := slices.IndexFunc(history, func(w workout.Workout) bool { return w.ID == workoutID })
	if i < 0 {
		return Result{}, nil
	}
	inv, err := r.Inventory.Get(ctx, uid)
	if err != nil {
		return Result{}, fmt.Errorf("revise: the inventory: %w", err)
	}
	inputs := r.inputs(uid, p, history, history[i], inventory.ForPlan(inv))

	res := Result{}
	var reply ai.Result
	if len(inputs) > 0 {
		req := ai.Request{User: uid, Today: history[i].Date, Sessions: 1}
		for _, in := range inputs {
			req.Exercises = append(req.Exercises, in)
			req.Today = max(req.Today, in.Today)
		}
		reply, err = r.AI.Revise(ctx, req)
		if err != nil {
			// A bad request is a fault of this code. The reasons of the
			// rules show.
			r.warn("revise: the reviser request", "uid", uid, "err", err.Error())
			reply = ai.Result{}
		}
		res.Status = reply.Status
	}

	changed := map[domain.ExerciseID]plan.Exercise{}
	for _, in := range inputs {
		rec, err := policy.Revise(in)
		if err != nil {
			return Result{}, fmt.Errorf("revise: the policy: %w", err)
		}
		if reply.Status != "" && reply.Status != ai.StatusCapped {
			rec.Model, rec.Effort, rec.PromptHash = reply.Model, reply.Effort, reply.PromptHash
		}
		ex := plan.Exercise{Target: rec.Target, Record: rec}
		ex.Reason, ex.ReasonSource, ex.ReasonCause = Reason(in, rec, reply)
		if ex.ReasonSource == policy.SourceLuna {
			res.Luna++
		}
		if len(rec.Target.Calibration) > 0 {
			if ex.Calibration, err = policy.CalibrationTable(in); err != nil {
				return Result{}, fmt.Errorf("revise: the calibration loads: %w", err)
			}
		}
		changed[in.Exercise.ID] = ex
	}

	now := r.now().UTC()
	err = r.Plans.Update(ctx, uid, p.CreatedAt, func(q *plan.Plan) error {
		if q.Revised(workoutID) {
			return errSeen
		}
		res.Exercises = apply(q, changed)
		q.Revisions++
		q.LastRevision = &plan.Revision{WorkoutID: workoutID, At: now, Exercises: res.Exercises}
		q.RevisedWorkouts = append(q.RevisedWorkouts, workoutID)
		if n := len(q.RevisedWorkouts); n > plan.MaxRevisedWorkouts {
			q.RevisedWorkouts = q.RevisedWorkouts[n-plan.MaxRevisedWorkouts:]
		}
		return nil
	})
	switch {
	case errors.Is(err, errSeen), errors.Is(err, plan.ErrPlanReplaced):
		return Result{}, nil
	case err != nil:
		return Result{}, fmt.Errorf("revise: the save: %w", err)
	}
	res.Revised = true
	if r.Log != nil {
		// The line holds ids and counts alone (D-80).
		r.Log.Info("plan revised", "uid", uid, "workout_id", workoutID, "exercises", len(res.Exercises),
			"status", string(res.Status), "luna_reasons", res.Luna, "rules_reasons", len(res.Exercises)-res.Luna)
	}
	return res, nil
}

func (r *Reviser) now() time.Time {
	if r.Now == nil {
		return time.Now()
	}
	return r.Now()
}

func (r *Reviser) warn(msg string, args ...any) {
	if r.Log != nil {
		r.Log.Warn(msg, args...)
	}
}

// finished reads the finished workouts of the user, the newest limit
// workouts at most, or each workout for AllHistory. It gives them the
// oldest first. One date orders its workouts by id, and a workout id is
// a UUIDv7 of its start, so the order is the order of the start.
func (r *Reviser) finished(ctx context.Context, uid string, limit int) ([]workout.Workout, error) {
	var all []workout.Workout
	after := ""
	for limit == AllHistory || len(all) < limit {
		n := pageSize
		if limit != AllHistory {
			n = min(n, limit-len(all))
		}
		page, next, err := r.Workouts.List(ctx, uid, n, after)
		if err != nil {
			return nil, fmt.Errorf("revise: the workouts: %w", err)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		after = next
	}
	out := slices.DeleteFunc(all, func(w workout.Workout) bool { return !w.Finished })
	slices.SortFunc(out, func(a, b workout.Workout) int {
		return cmp.Or(strings.Compare(a.Date, b.Date), strings.Compare(a.ID, b.ID))
	})
	return out, nil
}

// inputs gives the policy input of each exercise that the workout
// logged and the plan holds, in the order of the workout (D-290). An
// exercise whose machine is not confirmed, or whose input the policy
// refuses, keeps its target, and a log line names it. Each input holds
// the deloads of the history of each exercise (D-295).
func (r *Reviser) inputs(uid string, p plan.Plan, history []workout.Workout, w workout.Workout, pi inventory.PlanInput) []policy.Input {
	dl := r.deloads(uid, p, history)
	var out []policy.Input
	for _, l := range w.Exercises() {
		if !holds(p, l.Exercise) {
			continue
		}
		if in, ok := r.input(uid, p, history, pi, l.Exercise, dl); ok {
			out = append(out, in)
		}
	}
	return out
}

// input gives the policy input of one exercise, with the date of its
// last session as the date of the next session. It is false for an
// exercise with no history, with no confirmed machine, or with an input
// that the policy refuses.
func (r *Reviser) input(uid string, p plan.Plan, history []workout.Workout, pi inventory.PlanInput, id domain.ExerciseID, deloads []string) (policy.Input, bool) {
	e, ok := domain.DefaultCatalog().Exercise(id)
	if !ok {
		return policy.Input{}, false
	}
	entry, ok := pi.Inventory.Entry(e.Machine)
	if !ok {
		r.warn("revise: no confirmed machine", "uid", uid, "exercise_id", string(e.ID))
		return policy.Input{}, false
	}
	h := outcomes(p, history, e.ID)
	if len(h) == 0 {
		return policy.Input{}, false
	}
	// A new exercise is not a return after a long break (D-300), so the
	// normal rules apply from its second session (D-301).
	in := policy.Input{Exercise: e, Entry: entry, History: h, Today: h[len(h)-1].Date, Estimate: pi.Estimates[e.ID], Deloads: deloads}
	if _, err := policy.Next(in); err != nil {
		r.warn("revise: the policy refused the input", "uid", uid, "exercise_id", string(e.ID), "err", err.Error())
		return policy.Input{}, false
	}
	return in, true
}

// deloads gives the start date of each reactive deload, from the
// history of each exercise that a finished workout logged (D-295). A
// history that the policy refuses gives no deload, and a log line.
func (r *Reviser) deloads(uid string, p plan.Plan, history []workout.Workout) []string {
	var ids []domain.ExerciseID
	for _, w := range history {
		for _, l := range w.Exercises() {
			if !slices.Contains(ids, l.Exercise) {
				ids = append(ids, l.Exercise)
			}
		}
	}
	var hs [][]policy.Outcome
	for _, id := range ids {
		if h := outcomes(p, history, id); len(h) > 0 {
			hs = append(hs, h)
		}
	}
	out, err := policy.Deloads(hs)
	if err != nil {
		r.warn("revise: the policy refused the deload input", "uid", uid, "err", err.Error())
		return nil
	}
	return out
}

// History gives the logged history of uid for a new plan: the outcomes
// of each exercise that a finished workout logged, and the deloads
// (D-301). It reads each workout. A workout of an older phone with no
// target copy gives no outcome, because no plan links it to the new
// plan.
func (r *Reviser) History(ctx context.Context, uid string) (plan.History, error) {
	history, err := r.finished(ctx, uid, AllHistory)
	if err != nil {
		return plan.History{}, err
	}
	out := plan.History{Outcomes: map[domain.ExerciseID][]policy.Outcome{}, Deloads: r.deloads(uid, plan.Plan{}, history)}
	for _, w := range history {
		for _, l := range w.Exercises() {
			if _, done := out.Outcomes[l.Exercise]; done {
				continue
			}
			if h := outcomes(plan.Plan{}, history, l.Exercise); len(h) > 0 {
				out.Outcomes[l.Exercise] = h
			}
		}
	}
	return out, nil
}

// holds tells whether a session of the plan holds the exercise.
func holds(p plan.Plan, id domain.ExerciseID) bool {
	for _, s := range p.Sessions {
		for _, e := range s.Exercises {
			if e.Target.Exercise == id {
				return true
			}
		}
	}
	return false
}

// outcomes gives the history of one exercise, the oldest first: each
// finished workout that logged it and holds the target that the owner
// saw (D-291).
func outcomes(p plan.Plan, history []workout.Workout, id domain.ExerciseID) []policy.Outcome {
	var out []policy.Outcome
	for _, w := range history {
		i := slices.IndexFunc(w.Exercises(), func(l workout.ExerciseLog) bool { return l.Exercise == id })
		if i < 0 {
			continue
		}
		target, ok := seen(p, w, id)
		if !ok {
			continue
		}
		l := w.Exercises()[i]
		log := domain.ExerciseLog{Exercise: id, Skipped: l.Skipped}
		for _, s := range l.Sets {
			log.Sets = append(log.Sets, s.Log)
		}
		out = append(out, policy.Outcome{Date: w.Date, Target: target, Log: log, EndedEarly: w.EndedEarly})
	}
	return out
}

// seen gives the target that the owner saw for an exercise of a
// workout: its copy (D-291), or for a workout of an older phone, the
// target of the linked session of a plan with no revision.
func seen(p plan.Plan, w workout.Workout, id domain.ExerciseID) (domain.PlannedExercise, bool) {
	if len(w.Targets) > 0 {
		return w.Target(id)
	}
	if p.Revisions > 0 || !p.CreatedAt.Equal(w.Plan.PlanCreatedAt) || w.Plan.SessionIndex >= len(p.Sessions) {
		return domain.PlannedExercise{}, false
	}
	for _, e := range p.Sessions[w.Plan.SessionIndex].Exercises {
		if e.Target.Exercise == id {
			return e.Target, true
		}
	}
	return domain.PlannedExercise{}, false
}

// apply gives each changed exercise its new target in each session of
// the plan that holds it (D-290). It gives the changed exercises in the
// order of the plan.
func apply(p *plan.Plan, changed map[domain.ExerciseID]plan.Exercise) []domain.ExerciseID {
	var order []domain.ExerciseID
	for i := range p.Sessions {
		for j, e := range p.Sessions[i].Exercises {
			c, ok := changed[e.Target.Exercise]
			if !ok {
				continue
			}
			p.Sessions[i].Exercises[j] = c
			if !slices.Contains(order, e.Target.Exercise) {
				order = append(order, e.Target.Exercise)
			}
		}
	}
	return order
}
