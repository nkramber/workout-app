// Package workoutsvc serves workoutapp.v1.WorkoutService (work areas
// 6.1 and 6.3). It reads the uid that the auth interceptor stored, and
// keeps the workouts of each uid apart.
package workoutsvc

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// The page size of ListWorkouts: DefaultLimit when the request gives 0,
// and MaxLimit at most.
const (
	DefaultLimit = 20
	MaxLimit     = 50
)

// The codes of a refused entry.
const (
	CodeInvalidArgument    = "invalid_argument"
	CodeFailedPrecondition = "failed_precondition"
)

// Server implements workoutappv1connect.WorkoutServiceHandler.
type Server struct {
	store workout.Store
}

var _ workoutappv1connect.WorkoutServiceHandler = (*Server)(nil)

// New gives a server over the store.
func New(store workout.Store) *Server { return &Server{store: store} }

func uid(ctx context.Context) (string, error) {
	id := auth.UserID(ctx)
	if id == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))
	}
	return id, nil
}

// errStore is the error of a store failure. A store error can name a
// path, so the caller gets a fixed text. Each applied entry stays
// applied, and a replay of the batch gives its result again.
var errStore = connect.NewError(connect.CodeInternal, errors.New("the workout store failed"))

// SyncOutbox applies each entry in the order of the request, and gives
// the result of each one. A refused entry changes nothing, and the next
// entry still applies.
func (s *Server) SyncOutbox(ctx context.Context, req *connect.Request[workoutappv1.SyncOutboxRequest]) (*connect.Response[workoutappv1.SyncOutboxResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	entries := req.Msg.GetEntries()
	if n := len(entries); n > workout.MaxBatch {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("a batch of more than 100 entries"))
	}
	out := &workoutappv1.SyncOutboxResponse{}
	for _, in := range entries {
		r := &workoutappv1.EntryResult{OpId: in.GetOpId()}
		res, err := s.apply(ctx, id, in)
		switch {
		case err == nil:
			r.Status, r.Version = workoutappv1.EntryResult_STATUS_APPLIED, res.Version
		case errors.Is(err, workout.ErrInvalid):
			r.Status, r.Code, r.Message = workoutappv1.EntryResult_STATUS_REFUSED, CodeInvalidArgument, err.Error()
		case errors.Is(err, workout.ErrUnknownWorkout):
			r.Status, r.Code, r.Message = workoutappv1.EntryResult_STATUS_REFUSED, CodeFailedPrecondition, err.Error()
		default:
			return nil, errStore
		}
		out.Results = append(out.Results, r)
	}
	return connect.NewResponse(out), nil
}

func (s *Server) apply(ctx context.Context, uid string, in *workoutappv1.OutboxEntry) (workout.Result, error) {
	e, err := fromProto(in)
	if err != nil {
		return workout.Result{}, err
	}
	return s.store.Apply(ctx, uid, e)
}

// ListWorkouts gives one page of the workouts of the caller.
func (s *Server) ListWorkouts(ctx context.Context, req *connect.Request[workoutappv1.ListWorkoutsRequest]) (*connect.Response[workoutappv1.ListWorkoutsResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	limit := int(req.Msg.GetLimit())
	switch {
	case limit == 0:
		limit = DefaultLimit
	case limit < 0 || limit > MaxLimit:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("limit: want 0 to 50"))
	}
	after := req.Msg.GetPageToken()
	if after != "" {
		if err := workout.CheckID("page token", after); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	list, next, err := s.store.List(ctx, id, limit, after)
	if errors.Is(err, workout.ErrInvalid) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, errStore
	}
	out := &workoutappv1.ListWorkoutsResponse{NextPageToken: next}
	for _, w := range list {
		out.Workouts = append(out.Workouts, toProto(w))
	}
	return connect.NewResponse(out), nil
}

type checkError string

func (e checkError) Error() string { return string(e) }
func (checkError) Is(t error) bool { return t == workout.ErrInvalid }

// fromProto gives the entry of the contract. It refuses a time or a
// plan time that is not RFC 3339. Entry.Check reads each other field.
func fromProto(in *workoutappv1.OutboxEntry) (workout.Entry, error) {
	e := workout.Entry{
		OpID:          in.GetOpId(),
		Entity:        in.GetEntity(),
		EntityID:      in.GetEntityId(),
		BaseVersion:   in.GetBaseVersion(),
		SchemaVersion: int(in.GetSchemaVersion()),
	}
	at, err := time.Parse(time.RFC3339Nano, in.GetAt())
	if err != nil {
		return e, checkError("at: want an RFC 3339 time")
	}
	e.At = at
	switch p := in.GetPayload().(type) {
	case *workoutappv1.OutboxEntry_Workout:
		h := p.Workout
		var created time.Time
		if h.GetPlan() != nil {
			if created, err = time.Parse(time.RFC3339Nano, h.GetPlan().GetPlanCreatedAt()); err != nil {
				return e, checkError("plan link: want the RFC 3339 time of the plan")
			}
		}
		e.Header = &workout.Header{
			Date:       h.GetDate(),
			Plan:       workout.PlanLink{PlanCreatedAt: created, SessionIndex: int(h.GetPlan().GetSessionIndex())},
			Skipped:    ids(h.GetSkippedExerciseIds()),
			EndedEarly: h.GetEndedEarly(),
			Finished:   h.GetFinished(),
		}
	case *workoutappv1.OutboxEntry_Set:
		s := p.Set
		e.WorkoutID, e.SetExercise = s.GetWorkoutId(), domain.ExerciseID(s.GetExerciseId())
		e.Set = &domain.SetLog{
			Kind: domain.SetKind(s.GetKind()), Reps: int(s.GetReps()), Weight: domain.Load(s.GetWeightTenthsLb()),
			RIR: int(s.GetRir()), Pain: opt[domain.Pain](s.Pain), Note: s.GetNote(),
		}
	case *workoutappv1.OutboxEntry_Cardio:
		c := p.Cardio
		e.WorkoutID = c.GetWorkoutId()
		e.Cardio = &domain.CardioLog{
			Exercise: domain.ExerciseID(c.GetExerciseId()), Duration: time.Duration(c.GetDurationSeconds()) * time.Second,
			Effort: int(c.GetEffort()), Distance: opt[domain.Distance](c.DistanceTenthsMi), Resistance: opt[int](c.Resistance),
			Pain: opt[domain.Pain](c.Pain), Note: c.GetNote(),
		}
	}
	return e, nil
}

func toProto(w workout.Workout) *workoutappv1.Workout {
	out := &workoutappv1.Workout{
		WorkoutId: w.ID,
		Date:      w.Date,
		Plan: &workoutappv1.PlanLink{
			PlanCreatedAt: w.Plan.PlanCreatedAt.UTC().Format(time.RFC3339Nano),
			SessionIndex:  int32(w.Plan.SessionIndex),
		},
		EndedEarly: w.EndedEarly,
		Finished:   w.Finished,
	}
	// Each stored number passed its check, and the contract gave it as
	// an int32 or an int64, so it fits again.
	for _, l := range w.Exercises() {
		e := &workoutappv1.LoggedExercise{ExerciseId: string(l.Exercise), Skipped: l.Skipped}
		for _, s := range l.Sets {
			e.Sets = append(e.Sets, &workoutappv1.LoggedSet{
				SetId: s.ID, Kind: string(s.Log.Kind), Reps: int32(s.Log.Reps), WeightTenthsLb: int64(s.Log.Weight),
				Rir: int32(s.Log.RIR), Pain: opt32(s.Log.Pain), Note: s.Log.Note,
			})
		}
		out.Exercises = append(out.Exercises, e)
	}
	for _, c := range w.Cardio {
		out.Cardio = append(out.Cardio, &workoutappv1.LoggedCardio{
			CardioId: c.ID, ExerciseId: string(c.Log.Exercise), DurationSeconds: int32(c.Log.Duration / time.Second),
			Effort: int32(c.Log.Effort), DistanceTenthsMi: opt32(c.Log.Distance), Resistance: opt32(c.Log.Resistance),
			Pain: opt32(c.Log.Pain), Note: c.Log.Note,
		})
	}
	return out
}

func ids(in []string) []domain.ExerciseID {
	var out []domain.ExerciseID
	for _, v := range in {
		out = append(out, domain.ExerciseID(v))
	}
	return out
}

func opt[T ~int](p *int32) *T {
	if p == nil {
		return nil
	}
	v := T(*p)
	return &v
}

func opt32[T ~int](p *T) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}
