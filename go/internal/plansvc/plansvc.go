// Package plansvc serves workoutapp.v1.PlanService (work area 5.2). It
// reads the uid that the auth interceptor stored, and keeps the plan of
// each uid apart. The maker of "go/internal/plan" makes each plan.
package plansvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// Server implements workoutappv1connect.PlanServiceHandler.
type Server struct {
	maker   *plan.Maker
	dated   Dated
	catalog domain.Catalog
}

// Dated gives a plan with the targets of the rules on a date, the date
// of the next session. "go/internal/revise" implements it.
type Dated interface {
	ForDate(ctx context.Context, uid string, p plan.Plan, today string) (plan.Plan, error)
}

var _ workoutappv1connect.PlanServiceHandler = (*Server)(nil)

// New gives a server over the maker, with the product catalog (D-155).
// Dated gives the targets on a date. A nil Dated gives the targets that
// the plan holds.
func New(m *plan.Maker, d Dated) *Server {
	return &Server{maker: m, dated: d, catalog: domain.DefaultCatalog()}
}

func uid(ctx context.Context) (string, error) {
	id := auth.UserID(ctx)
	if id == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))
	}
	return id, nil
}

// fail gives the Connect error of an error of the maker or the store.
// An error of the maker names ids and numbers alone, so its text goes
// to the caller. Another error can name a path, so the caller gets a
// fixed text.
func fail(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, plan.ErrNoProfile), errors.Is(err, plan.ErrNothingToPlan), errors.Is(err, errNoPlan), errors.Is(err, errNoExercise):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, plan.ErrCapped):
		return connect.NewError(connect.CodeResourceExhausted, err)
	case errors.Is(err, plan.ErrNoValidPlan):
		return connect.NewError(connect.CodeUnavailable, err)
	case errors.Is(err, plan.ErrConflict), errors.Is(err, plan.ErrPlanReplaced), errors.Is(err, plan.ErrHistoryDeleted):
		return connect.NewError(connect.CodeAborted, err)
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, errors.New("the request ended"))
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, errors.New("the request passed its time limit"))
	}
	return connect.NewError(connect.CodeInternal, errors.New("the plan request failed"))
}

// GetPlan gives the plan and the exclusions of the caller. With a date,
// each target is the target on that date.
func (s *Server) GetPlan(ctx context.Context, req *connect.Request[workoutappv1.GetPlanRequest]) (*connect.Response[workoutappv1.GetPlanResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	today := req.Msg.GetToday()
	if today != "" {
		if err := s.maker.CheckToday(today); err != nil {
			return nil, fail(err)
		}
	}
	p, ok, err := s.maker.Plans.Get(ctx, id)
	if err != nil {
		return nil, fail(err)
	}
	if ok {
		if p, err = s.onDate(ctx, id, p, today); err != nil {
			return nil, fail(err)
		}
	}
	ex, err := s.maker.Plans.Exclusions(ctx, id)
	if err != nil {
		return nil, fail(err)
	}
	out := &workoutappv1.GetPlanResponse{}
	if ok {
		out.Plan = s.toProto(p)
	}
	for _, e := range ex.Items {
		out.Exclusions = append(out.Exclusions, &workoutappv1.Exclusion{ExerciseId: string(e.Exercise), Name: s.name(e.Exercise), Reason: e.Reason})
	}
	return connect.NewResponse(out), nil
}

// onDate gives the plan with the targets on a date, or the plan as it is
// for no date or no Dated.
func (s *Server) onDate(ctx context.Context, uid string, p plan.Plan, today string) (plan.Plan, error) {
	if today == "" || s.dated == nil {
		return p, nil
	}
	return s.dated.ForDate(ctx, uid, p, today)
}

func (s *Server) now() time.Time {
	if s.maker.Now == nil {
		return time.Now()
	}
	return s.maker.Now()
}

// errNoPlan is an override for a caller with no plan.
var errNoPlan = errors.New("plan: request a plan first")

// errNoExercise is an override of an exercise that no session of the
// plan holds.
var errNoExercise = errors.New("plan: no session of the plan holds the exercise")

// OverrideTarget saves an override of the owner (D-69, D-293). The
// recommendation is the target on the date of the request. The policy
// checks the override before the save (D-23).
func (s *Server) OverrideTarget(ctx context.Context, req *connect.Request[workoutappv1.OverrideTargetRequest]) (*connect.Response[workoutappv1.OverrideTargetResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	m := req.Msg
	reason, err := plan.OverrideReason(m.GetReason())
	if err != nil {
		return nil, fail(err)
	}
	p, err := s.change(ctx, id, m.GetToday(), domain.ExerciseID(m.GetExerciseId()), func(rec domain.PlannedExercise, in policy.Input) (*plan.Override, error) {
		o := domain.PlannedExercise{Exercise: rec.Exercise, RestSeconds: rec.RestSeconds, FirstSetCalibration: rec.FirstSetCalibration}
		for i, w := range m.GetWorkingSets() {
			rir := 0
			if i < len(rec.Working) {
				rir = rec.Working[i].RIR
			}
			o.Working = append(o.Working, domain.WorkingSet{Reps: int(w.GetReps()), Load: domain.Load(w.GetLoadTenthLb()), RIR: rir})
		}
		if len(rec.Calibration) > 0 && len(o.Working) > 0 {
			o.Calibration = []domain.CalibrationSet{{Reps: o.Working[0].Reps, Load: o.Working[0].Load}}
		}
		v, err := policy.CheckOverride(o, rec, in)
		if err != nil {
			return nil, err
		}
		if len(v) > 0 {
			parts := make([]string, len(v))
			for i, x := range v {
				parts[i] = x.String()
			}
			return nil, fmt.Errorf("%w: override: %s", domain.ErrInvalid, strings.Join(parts, "; "))
		}
		// The policy gives the limit of the other working sets of the
		// override too (D-306, D-307).
		o.FollowMax = policy.Follow(o, in.Entry.Available())
		return &plan.Override{Target: o, Recommendation: rec, Reason: reason, At: s.now().UTC(), Today: m.GetToday()}, nil
	})
	if err != nil {
		return nil, fail(err)
	}
	return connect.NewResponse(&workoutappv1.OverrideTargetResponse{Plan: s.toProto(p)}), nil
}

// RemoveOverride removes the override of an exercise.
func (s *Server) RemoveOverride(ctx context.Context, req *connect.Request[workoutappv1.RemoveOverrideRequest]) (*connect.Response[workoutappv1.RemoveOverrideResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	m := req.Msg
	p, err := s.change(ctx, id, m.GetToday(), domain.ExerciseID(m.GetExerciseId()), func(domain.PlannedExercise, policy.Input) (*plan.Override, error) {
		return nil, nil
	})
	if err != nil {
		return nil, fail(err)
	}
	return connect.NewResponse(&workoutappv1.RemoveOverrideResponse{Plan: s.toProto(p)}), nil
}

// change saves the override that f gives for an exercise, in one
// transaction with the read of the plan, and gives the plan on the date.
// f gets the recommendation on the date and the policy input of the
// exercise. A nil override removes the override.
func (s *Server) change(ctx context.Context, uid, today string, ex domain.ExerciseID, f func(domain.PlannedExercise, policy.Input) (*plan.Override, error)) (plan.Plan, error) {
	if err := s.maker.CheckToday(today); err != nil {
		return plan.Plan{}, err
	}
	e, ok := s.catalog.Exercise(ex)
	if !ok || e.Kind == domain.KindCardio {
		return plan.Plan{}, fmt.Errorf("%w: override: the exercise is not a resistance exercise of the catalog", domain.ErrInvalid)
	}
	p, ok, err := s.maker.Plans.Get(ctx, uid)
	if err != nil {
		return plan.Plan{}, err
	}
	if !ok {
		return plan.Plan{}, errNoPlan
	}
	dated, err := s.onDate(ctx, uid, p, today)
	if err != nil {
		return plan.Plan{}, err
	}
	rec, ok := target(dated, ex)
	if !ok {
		return plan.Plan{}, errNoExercise
	}
	inv, err := s.maker.Inventory.Get(ctx, uid)
	if err != nil {
		return plan.Plan{}, err
	}
	entry, ok := inventory.ForPlan(inv).Inventory.Entry(e.Machine)
	if !ok {
		return plan.Plan{}, fmt.Errorf("%w: override: the machine of the exercise is not confirmed", domain.ErrInvalid)
	}
	o, err := f(rec, policy.Input{Exercise: e, Entry: entry})
	if err != nil {
		return plan.Plan{}, err
	}
	err = s.maker.Plans.Update(ctx, uid, p.CreatedAt, func(q *plan.Plan) error {
		if !q.SetOverride(ex, o) {
			return errNoExercise
		}
		return nil
	})
	if err != nil {
		return plan.Plan{}, err
	}
	dated.SetOverride(ex, o)
	return dated, nil
}

// target gives the target of an exercise in the first session that
// holds it.
func target(p plan.Plan, id domain.ExerciseID) (domain.PlannedExercise, bool) {
	for _, s := range p.Sessions {
		for _, e := range s.Exercises {
			if e.Target.Exercise == id {
				return e.Target, true
			}
		}
	}
	return domain.PlannedExercise{}, false
}

// RequestPlan makes a new plan, and streams each step and the plan.
func (s *Server) RequestPlan(ctx context.Context, req *connect.Request[workoutappv1.RequestPlanRequest], stream *connect.ServerStream[workoutappv1.RequestPlanResponse]) error {
	return run(ctx, s, req.Msg.GetToday(), nil, func(ev *workoutappv1.PlanProgress, p *workoutappv1.Plan) error {
		if p != nil {
			return stream.Send(&workoutappv1.RequestPlanResponse{Event: &workoutappv1.RequestPlanResponse_Plan{Plan: p}})
		}
		return stream.Send(&workoutappv1.RequestPlanResponse{Event: &workoutappv1.RequestPlanResponse_Progress{Progress: ev}})
	})
}

// ExcludeExercise adds an exclusion, makes a new plan, and streams each
// step and the plan.
func (s *Server) ExcludeExercise(ctx context.Context, req *connect.Request[workoutappv1.ExcludeExerciseRequest], stream *connect.ServerStream[workoutappv1.ExcludeExerciseResponse]) error {
	ex := &plan.Exclusion{Exercise: domain.ExerciseID(req.Msg.GetExerciseId()), Reason: req.Msg.GetReason()}
	return run(ctx, s, req.Msg.GetToday(), ex, func(ev *workoutappv1.PlanProgress, p *workoutappv1.Plan) error {
		if p != nil {
			return stream.Send(&workoutappv1.ExcludeExerciseResponse{Event: &workoutappv1.ExcludeExerciseResponse_Plan{Plan: p}})
		}
		return stream.Send(&workoutappv1.ExcludeExerciseResponse{Event: &workoutappv1.ExcludeExerciseResponse_Progress{Progress: ev}})
	})
}

// run makes a plan and sends each event. When a send fails, the stream
// is closed, so run ends the request: the maker then stops at its next
// step and saves nothing (D-237).
func run(ctx context.Context, s *Server, today string, ex *plan.Exclusion, send func(*workoutappv1.PlanProgress, *workoutappv1.Plan) error) error {
	id, err := uid(ctx)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p, err := s.maker.Make(ctx, id, today, ex, func(pr plan.Progress) {
		ev := &workoutappv1.PlanProgress{
			Step: string(pr.Step), Attempt: int32(pr.Attempt), MaxAttempts: int32(pr.MaxAttempts), PreviousStatus: string(pr.Previous),
		}
		if send(ev, nil) != nil {
			cancel()
		}
	})
	if err != nil {
		return fail(err)
	}
	return send(nil, s.toProto(p))
}

func (s *Server) name(id domain.ExerciseID) string {
	if e, ok := s.catalog.Exercise(id); ok {
		return e.Name
	}
	return string(id)
}

// toProto gives the plan of the contract. Each load and count of a
// stored plan passed the policy, so it fits an int32.
func (s *Server) toProto(p plan.Plan) *workoutappv1.Plan {
	out := &workoutappv1.Plan{
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339), Today: p.Today, Summary: p.Summary,
		PromptVersion: p.PromptVersion, PolicyVersion: int32(p.PolicyVersion), Attempts: int32(p.Attempts),
	}
	if r := p.LastRevision; r != nil {
		rev := &workoutappv1.PlanRevision{WorkoutId: r.WorkoutID, RevisedAt: r.At.UTC().Format(time.RFC3339)}
		for _, id := range r.Exercises {
			rev.ExerciseIds = append(rev.ExerciseIds, string(id))
		}
		out.LastRevision = rev
	}
	for _, g := range p.Guidance {
		out.Guidance = append(out.Guidance, guidance(g))
	}
	for _, sess := range p.Sessions {
		ps := &workoutappv1.PlanSession{Title: sess.Title, WarmUp: guidance(sess.WarmUp), CoolDown: guidance(sess.CoolDown)}
		for _, e := range sess.Exercises {
			t := e.Target
			pe := &workoutappv1.PlannedExercise{
				ExerciseId: string(t.Exercise), Name: s.name(t.Exercise), RestSeconds: int32(t.RestSeconds),
				Reason: e.Reason, Source: string(e.Record.Source), ReasonSource: string(e.ReasonSourceOf()),
				FirstSetCalibration: t.FirstSetCalibration, FollowMaxTenthLb: int32(t.FollowMax),
			}
			for _, c := range t.Calibration {
				pe.CalibrationSets = append(pe.CalibrationSets, &workoutappv1.PlannedSet{Reps: int32(c.Reps), LoadTenthLb: int32(c.Load)})
			}
			for _, w := range t.Working {
				pe.WorkingSets = append(pe.WorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
			}
			if o := e.Override; o != nil {
				pe.Override = &workoutappv1.TargetOverride{Reason: o.Reason, CreatedAt: o.At.UTC().Format(time.RFC3339), Expired: o.Expired, FollowMaxTenthLb: int32(o.Target.FollowMax)}
				for _, c := range o.Target.Calibration {
					pe.Override.CalibrationSets = append(pe.Override.CalibrationSets, &workoutappv1.PlannedSet{Reps: int32(c.Reps), LoadTenthLb: int32(c.Load)})
				}
				for _, w := range o.Target.Working {
					pe.Override.WorkingSets = append(pe.Override.WorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
				}
				for _, w := range o.Recommendation.Working {
					pe.Override.RecommendedWorkingSets = append(pe.Override.RecommendedWorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
				}
			}
			for _, c := range e.Calibration {
				pe.CalibrationLoads = append(pe.CalibrationLoads, &workoutappv1.CalibrationLoads{
					WeightTenthLb: int32(c.Weight), DownTenthLb: int32(c.Down), KeepTenthLb: int32(c.Keep), UpOneTenthLb: int32(c.UpOne), UpTwoTenthLb: int32(c.UpTwo),
				})
			}
			ps.Exercises = append(ps.Exercises, pe)
		}
		if c := sess.Cardio; c != nil {
			ps.Cardio = &workoutappv1.PlannedCardio{ExerciseId: string(c.Exercise), Name: s.name(c.Exercise), Minutes: int32(c.Minutes)}
		}
		out.Sessions = append(out.Sessions, ps)
	}
	return out
}

// guidance gives the item of an id. A stored id that the catalog no
// longer holds gives the id with no text.
func guidance(id ai.GuidanceID) *workoutappv1.GuidanceItem {
	g, ok := ai.GuidanceItemOf(id)
	if !ok {
		return &workoutappv1.GuidanceItem{Id: string(id)}
	}
	return &workoutappv1.GuidanceItem{Id: string(g.ID), Kind: string(g.Kind), Text: g.Text}
}
