// Package workoutsvc serves workoutapp.v1.WorkoutService (work areas
// 6.1 and 6.3). It reads the uid that the auth interceptor stored, and
// keeps the workouts of each uid apart.
package workoutsvc

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/revise"
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

// Server implements workoutappv1connect.WorkoutServiceHandler. The
// inventory store applies the inventory entries of the outbox (D-272).
// The reviser, when it is not nil, revises the plan after each finished
// workout of a batch (D-292).
type Server struct {
	store       workout.Store
	inventories inventory.Store
	catalog     domain.Catalog
	reviser     Reviser
	log         *slog.Logger
}

// Reviser revises the plan of a user after a finished workout.
type Reviser interface {
	Revise(ctx context.Context, uid, workoutID string) (revise.Result, error)
}

var _ workoutappv1connect.WorkoutServiceHandler = (*Server)(nil)

// New gives a server over the stores, with the product catalog.
func New(store workout.Store, inventories inventory.Store) *Server {
	return &Server{store: store, inventories: inventories, catalog: domain.DefaultCatalog()}
}

// WithReviser gives the server the reviser of the plan, and the logger
// of a failed revision. A nil logger writes no line.
func (s *Server) WithReviser(r Reviser, log *slog.Logger) *Server {
	s.reviser, s.log = r, log
	return s
}

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
// the result of each one. The workout entries and the inventory entries
// share one order (D-275). A refused entry changes nothing, and the next
// entry still applies. After the batch, the plan gets a revision for
// each workout that an applied entry finished (D-292). A replayed entry
// counts too, so a revision that a dropped answer stopped runs again.
// The reviser skips a workout that revised the plan already. A store
// failure of a revision gives UNAVAILABLE. Each applied entry stays
// applied, so the phone sends the batch again, and the revision runs
// again.
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
	var finished []string
	for _, in := range entries {
		r := &workoutappv1.EntryResult{OpId: in.GetOpId()}
		res, err := s.apply(ctx, id, in)
		switch {
		case err == nil:
			r.Status, r.Version = workoutappv1.EntryResult_STATUS_APPLIED, res.Version
			if w := in.GetWorkout(); w.GetFinished() && in.GetEntity() == workout.EntityWorkout && !slices.Contains(finished, in.GetEntityId()) {
				finished = append(finished, in.GetEntityId())
			}
		case errors.Is(err, workout.ErrInvalid):
			r.Status, r.Code, r.Message = workoutappv1.EntryResult_STATUS_REFUSED, CodeInvalidArgument, err.Error()
		case errors.Is(err, workout.ErrUnknownWorkout), errors.Is(err, inventory.ErrNotFound), errors.Is(err, inventory.ErrWeightsChanged):
			r.Status, r.Code, r.Message = workoutappv1.EntryResult_STATUS_REFUSED, CodeFailedPrecondition, err.Error()
		default:
			return nil, errStore
		}
		out.Results = append(out.Results, r)
	}
	if err := s.revise(ctx, id, finished); err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}

// errRevision is the error of a store failure of a revision. A store
// error can name a path, so the caller gets a fixed text.
var errRevision = connect.NewError(connect.CodeUnavailable, errors.New("the revision of the plan failed; send the batch again"))

// revise revises the plan for each finished workout. A failed reviser
// call is no error: the reasons of the rules show (D-292). A store
// failure stops the revisions, and gives errRevision. The line holds
// ids alone (D-80).
func (s *Server) revise(ctx context.Context, uid string, workouts []string) error {
	if s.reviser == nil {
		return nil
	}
	for _, w := range workouts {
		if _, err := s.reviser.Revise(ctx, uid, w); err != nil {
			if s.log != nil {
				s.log.Error("revision failed", "uid", uid, "workout_id", w)
			}
			return errRevision
		}
	}
	return nil
}

// apply applies one entry. An inventory entry gives the version 0,
// because the inventory has no version.
func (s *Server) apply(ctx context.Context, uid string, in *workoutappv1.OutboxEntry) (workout.Result, error) {
	if op, ch, ok, err := s.inventoryEntry(in); ok {
		if err != nil {
			return workout.Result{}, err
		}
		replayed, err := s.inventories.ApplyOp(ctx, uid, op, ch)
		return workout.Result{Replayed: replayed}, err
	}
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
		e.Header.Targets, e.Header.Overrides = targetsFrom(h.GetTargets())
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
		Targets:    targetsTo(w.Targets, w.Overrides),
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

// targetsFrom gives the target copies of the contract (D-291), and the
// record of each override among them (D-293). Header.Check reads each
// value.
func targetsFrom(in []*workoutappv1.SeenTarget) ([]domain.PlannedExercise, []workout.SeenOverride) {
	var out []domain.PlannedExercise
	var overrides []workout.SeenOverride
	for _, t := range in {
		p := domain.PlannedExercise{Exercise: domain.ExerciseID(t.GetExerciseId()), RestSeconds: int(t.GetRestSeconds())}
		if t.GetOverrideReason() != "" || len(t.GetRecommendedWorkingSets()) > 0 {
			o := workout.SeenOverride{Exercise: p.Exercise, Reason: t.GetOverrideReason()}
			for _, w := range t.GetRecommendedWorkingSets() {
				o.Recommended = append(o.Recommended, domain.WorkingSet{Reps: int(w.GetReps()), Load: domain.Load(w.GetLoadTenthLb()), RIR: int(w.GetRirTarget())})
			}
			overrides = append(overrides, o)
		}
		for _, c := range t.GetCalibrationSets() {
			p.Calibration = append(p.Calibration, domain.CalibrationSet{Reps: int(c.GetReps()), Load: domain.Load(c.GetLoadTenthLb())})
		}
		for _, w := range t.GetWorkingSets() {
			p.Working = append(p.Working, domain.WorkingSet{Reps: int(w.GetReps()), Load: domain.Load(w.GetLoadTenthLb()), RIR: int(w.GetRirTarget())})
		}
		out = append(out, p)
	}
	return out, overrides
}

// targetsTo gives the target copies in the form of the contract. Each
// stored value came from an int32 of the contract, so it fits again.
func targetsTo(in []domain.PlannedExercise, overrides []workout.SeenOverride) []*workoutappv1.SeenTarget {
	var out []*workoutappv1.SeenTarget
	for _, t := range in {
		p := &workoutappv1.SeenTarget{ExerciseId: string(t.Exercise), RestSeconds: int32(t.RestSeconds)}
		for _, o := range overrides {
			if o.Exercise != t.Exercise {
				continue
			}
			p.OverrideReason = o.Reason
			for _, w := range o.Recommended {
				p.RecommendedWorkingSets = append(p.RecommendedWorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
			}
		}
		for _, c := range t.Calibration {
			p.CalibrationSets = append(p.CalibrationSets, &workoutappv1.PlannedSet{Reps: int32(c.Reps), LoadTenthLb: int32(c.Load)})
		}
		for _, w := range t.Working {
			p.WorkingSets = append(p.WorkingSets, &workoutappv1.PlannedSet{Reps: int32(w.Reps), LoadTenthLb: int32(w.Load), RirTarget: int32(w.RIR)})
		}
		out = append(out, p)
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
