// Package plansvc serves workoutapp.v1.PlanService (work area 5.2). It
// reads the uid that the auth interceptor stored, and keeps the plan of
// each uid apart. The maker of "go/internal/plan" makes each plan.
package plansvc

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/plan"
)

// Server implements workoutappv1connect.PlanServiceHandler.
type Server struct {
	maker   *plan.Maker
	catalog domain.Catalog
}

var _ workoutappv1connect.PlanServiceHandler = (*Server)(nil)

// New gives a server over the maker, with the product catalog (D-155).
func New(m *plan.Maker) *Server {
	return &Server{maker: m, catalog: domain.DefaultCatalog()}
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
	case errors.Is(err, plan.ErrNoProfile), errors.Is(err, plan.ErrNothingToPlan):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, plan.ErrCapped):
		return connect.NewError(connect.CodeResourceExhausted, err)
	case errors.Is(err, plan.ErrNoValidPlan):
		return connect.NewError(connect.CodeUnavailable, err)
	case errors.Is(err, plan.ErrConflict):
		return connect.NewError(connect.CodeAborted, err)
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, errors.New("the request ended"))
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, errors.New("the request passed its time limit"))
	}
	return connect.NewError(connect.CodeInternal, errors.New("the plan request failed"))
}

// GetPlan gives the plan and the exclusions of the caller.
func (s *Server) GetPlan(ctx context.Context, _ *connect.Request[workoutappv1.GetPlanRequest]) (*connect.Response[workoutappv1.GetPlanResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	p, ok, err := s.maker.Plans.Get(ctx, id)
	if err != nil {
		return nil, fail(err)
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
			}
			for _, c := range t.Calibration {
				pe.CalibrationSets = append(pe.CalibrationSets, &workoutappv1.PlannedSet{Reps: int32(c.Reps), LoadTenthLb: int32(c.Load)})
			}
			for _, w := range t.Working {
				pe.WorkingSets = append(pe.WorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
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
