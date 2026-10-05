package workout

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The Firestore paths of a user are users/{uid}/workouts/{workoutId},
// one document for each logged session, and users/{uid}/ops/{opId}, one
// document for each applied op id (D-256). An op id document has no end
// date (D-257).
const (
	UsersCollection    = "users"
	WorkoutsCollection = "workouts"
	OpsCollection      = "ops"
)

// Result is the result of one applied entry: the server version of its
// entity. Replayed is true when an earlier call applied the op id, and
// the store changed nothing.
type Result struct {
	Version  int64
	Replayed bool
}

// Store applies the entries and reads the workouts of a user. Apply
// reads the op id first. An applied op id gives its stored result and
// changes nothing. Else Apply applies the entry, and writes the workout
// and the op id together, or writes nothing. It gives an error that
// matches ErrInvalid or ErrUnknownWorkout for a refused entry. List
// gives the workouts, the newest date first, after the workout of the
// id after, or from the first when after is empty. The next token is
// the id of the last workout, or empty when no workout remains.
type Store interface {
	Apply(ctx context.Context, uid string, e Entry) (Result, error)
	List(ctx context.Context, uid string, limit int, after string) ([]Workout, string, error)
}

var errUID = errors.New("workout: a uid of 1 or more characters with no slash is required")

// ErrUnknownPage is the error of a page token that names no workout of
// the user.
var ErrUnknownPage = invalid("page token: not a workout of the user")

func checkUID(uid string) error {
	if uid == "" || strings.Contains(uid, "/") {
		return errUID
	}
	return nil
}

// Memory is a Store in memory, for tests. Catalog is the catalog of the
// checks.
type Memory struct {
	Catalog  domain.Catalog
	mu       sync.Mutex
	workouts map[string]map[string]Workout
	ops      map[string]map[string]Result
}

// NewMemory gives an empty Memory store with the product catalog.
func NewMemory() *Memory {
	return &Memory{Catalog: domain.DefaultCatalog(), workouts: map[string]map[string]Workout{}, ops: map[string]map[string]Result{}}
}

// Apply applies the entry under the lock of the store.
func (s *Memory) Apply(_ context.Context, uid string, e Entry) (Result, error) {
	if err := checkUID(uid); err != nil {
		return Result{}, err
	}
	if err := CheckOpID(e.OpID); err != nil {
		return Result{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.ops[uid][e.OpID]; ok {
		r.Replayed = true
		return r, nil
	}
	var stored *Workout
	if w, ok := s.workouts[uid][e.Workout()]; ok {
		stored = &w
	}
	w, version, err := Apply(stored, e, s.Catalog)
	if err != nil {
		return Result{}, err
	}
	if s.workouts[uid] == nil {
		s.workouts[uid], s.ops[uid] = map[string]Workout{}, map[string]Result{}
	}
	s.workouts[uid][w.ID] = w
	s.ops[uid][e.OpID] = Result{Version: version}
	return Result{Version: version}, nil
}

// List gives the workouts of the uid, the newest date first, and for
// one date the larger id first, as Firestore gives them.
func (s *Memory) List(_ context.Context, uid string, limit int, after string) ([]Workout, string, error) {
	if err := checkUID(uid); err != nil {
		return nil, "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var all []Workout
	for _, w := range s.workouts[uid] {
		all = append(all, w.clone())
	}
	slices.SortFunc(all, func(a, b Workout) int {
		if c := strings.Compare(b.Date, a.Date); c != 0 {
			return c
		}
		return strings.Compare(b.ID, a.ID)
	})
	if after != "" {
		i := slices.IndexFunc(all, func(w Workout) bool { return w.ID == after })
		if i < 0 {
			return nil, "", ErrUnknownPage
		}
		all = all[i+1:]
	}
	return page(all, limit)
}

func page(all []Workout, limit int) ([]Workout, string, error) {
	if len(all) <= limit {
		return all, "", nil
	}
	return all[:limit], all[limit-1].ID, nil
}

// Firestore is the production Store.
type Firestore struct {
	client  *firestore.Client
	catalog domain.Catalog
	now     func() time.Time
}

// FromFirestore gives the Store of the client, with the product
// catalog.
func FromFirestore(client *firestore.Client) *Firestore {
	return &Firestore{client: client, catalog: domain.DefaultCatalog(), now: time.Now}
}

func (s *Firestore) user(uid string) *firestore.DocumentRef {
	return s.client.Collection(UsersCollection).Doc(uid)
}

// Apply runs one transaction: it reads the op id and the workout, then
// writes the workout and the op id. Firestore runs the function again
// when another transaction changes a document that it read, so two
// calls with the same op id apply it one time.
func (s *Firestore) Apply(ctx context.Context, uid string, e Entry) (Result, error) {
	if err := checkUID(uid); err != nil {
		return Result{}, err
	}
	if err := CheckOpID(e.OpID); err != nil {
		return Result{}, err
	}
	opRef := s.user(uid).Collection(OpsCollection).Doc(e.OpID)
	var res Result
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		res = Result{}
		snap, err := tx.Get(opRef)
		switch {
		case err == nil:
			var d opDoc
			if err := snap.DataTo(&d); err != nil {
				return err
			}
			res = Result{Version: d.Version, Replayed: true}
			return nil
		case status.Code(err) != codes.NotFound:
			return err
		}
		// Check the entry before the read of the workout, so a bad id never
		// becomes a document path.
		if err := e.Check(s.catalog); err != nil {
			return err
		}
		ref := s.user(uid).Collection(WorkoutsCollection).Doc(e.Workout())
		var stored *Workout
		snap, err = tx.Get(ref)
		switch {
		case err == nil:
			var d workoutDoc
			if err := snap.DataTo(&d); err != nil {
				return err
			}
			w := d.workout(ref.ID)
			stored = &w
		case status.Code(err) != codes.NotFound:
			return err
		}
		w, version, err := Apply(stored, e, s.catalog)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		if err := tx.Set(ref, encodeWorkout(w, now)); err != nil {
			return err
		}
		res = Result{Version: version}
		return tx.Create(opRef, opDoc{
			Entity: e.Entity, EntityID: e.EntityID, WorkoutID: w.ID,
			BaseVersion: e.BaseVersion, Version: version, At: e.At.UTC(), AppliedAt: now,
		})
	}, firestore.MaxAttempts(5))
	if err != nil {
		return Result{}, err
	}
	return res, nil
}

// List reads the workouts of the uid, the newest date first. The query
// orders by one field, so the single-field index of Firestore serves it,
// and the document id orders the workouts of one date.
func (s *Firestore) List(ctx context.Context, uid string, limit int, after string) ([]Workout, string, error) {
	if err := checkUID(uid); err != nil {
		return nil, "", err
	}
	col := s.user(uid).Collection(WorkoutsCollection)
	q := col.OrderBy("date", firestore.Desc).OrderBy(firestore.DocumentID, firestore.Desc)
	if after != "" {
		if err := CheckID("page token", after); err != nil {
			return nil, "", err
		}
		snap, err := col.Doc(after).Get(ctx)
		if status.Code(err) == codes.NotFound {
			return nil, "", ErrUnknownPage
		}
		if err != nil {
			return nil, "", err
		}
		q = q.StartAfter(snap)
	}
	snaps, err := q.Limit(limit + 1).Documents(ctx).GetAll()
	if err != nil {
		return nil, "", err
	}
	var all []Workout
	for _, snap := range snaps {
		var d workoutDoc
		if err := snap.DataTo(&d); err != nil {
			return nil, "", err
		}
		all = append(all, d.workout(snap.Ref.ID))
	}
	return page(all, limit)
}

// The stored op id. It holds ids, versions, and times alone.
type opDoc struct {
	Entity      string    `firestore:"entity"`
	EntityID    string    `firestore:"entity_id"`
	WorkoutID   string    `firestore:"workout_id"`
	BaseVersion int64     `firestore:"base_version"`
	Version     int64     `firestore:"version"`
	At          time.Time `firestore:"at"`
	AppliedAt   time.Time `firestore:"applied_at"`
}

// The stored workout, in the form of the session log of the domain
// model: the exercises with their sets, and the cardio logs. The skip
// list of the phone stays too, so a later entry can change it.
type workoutDoc struct {
	Date       string           `firestore:"date"`
	Plan       planDoc          `firestore:"plan"`
	Skipped    []string         `firestore:"skipped_exercises"`
	EndedEarly bool             `firestore:"ended_early"`
	Finished   bool             `firestore:"finished"`
	Exercises  []exerciseDoc    `firestore:"exercises"`
	Cardio     []cardioDoc      `firestore:"cardio"`
	Targets    []targetDoc      `firestore:"targets"`
	Versions   map[string]int64 `firestore:"versions"`
	UpdatedAt  time.Time        `firestore:"updated_at"`
}

type planDoc struct {
	CreatedAt    time.Time `firestore:"created_at"`
	SessionIndex int64     `firestore:"session_index"`
}

// targetDoc is the copy of the target that the owner saw (D-291), in
// the form of the plan document. A calibration set stores an RIR of 0
// (D-150). A workout of an older phone has none.
type targetDoc struct {
	Exercise    string      `firestore:"exercise_id"`
	Rest        int64       `firestore:"rest_seconds"`
	Calibration []targetSet `firestore:"calibration_sets"`
	Working     []targetSet `firestore:"working_sets"`
	// The first working set is the calibration (D-297). A workout of an
	// older phone has no such field.
	FirstSet bool `firestore:"first_set_calibration,omitempty"`
	// The limit of the other working sets after the first set (D-306,
	// D-307). A workout of an older phone has no such field.
	FollowMax int64 `firestore:"follow_max_tenth_lb,omitempty"`
	// After an override of the owner: the recommendation that it
	// replaced, and the reason of the owner (D-69, D-293).
	Recommended    []targetSet `firestore:"recommended_working_sets,omitempty"`
	OverrideReason string      `firestore:"override_reason,omitempty"`
}

type targetSet struct {
	Reps int64 `firestore:"reps"`
	Load int64 `firestore:"load_tenth_lb"`
	RIR  int64 `firestore:"rir_target"`
}

func encodeTargets(in []domain.PlannedExercise, overrides []SeenOverride) []targetDoc {
	out := make([]targetDoc, 0, len(in))
	for _, t := range in {
		d := targetDoc{Exercise: string(t.Exercise), Rest: int64(t.RestSeconds), Calibration: []targetSet{}, Working: []targetSet{}, FirstSet: t.FirstSetCalibration, FollowMax: int64(t.FollowMax)}
		for _, o := range overrides {
			if o.Exercise != t.Exercise {
				continue
			}
			d.OverrideReason = o.Reason
			for _, w := range o.Recommended {
				d.Recommended = append(d.Recommended, targetSet{int64(w.Reps), int64(w.Load), int64(w.RIR)})
			}
		}
		for _, c := range t.Calibration {
			d.Calibration = append(d.Calibration, targetSet{Reps: int64(c.Reps), Load: int64(c.Load)})
		}
		for _, w := range t.Working {
			d.Working = append(d.Working, targetSet{int64(w.Reps), int64(w.Load), int64(w.RIR)})
		}
		out = append(out, d)
	}
	return out
}

func decodeTargets(in []targetDoc) ([]domain.PlannedExercise, []SeenOverride) {
	var out []domain.PlannedExercise
	var overrides []SeenOverride
	for _, d := range in {
		t := domain.PlannedExercise{Exercise: domain.ExerciseID(d.Exercise), RestSeconds: int(d.Rest), FirstSetCalibration: d.FirstSet, FollowMax: domain.Load(d.FollowMax)}
		if d.OverrideReason != "" || len(d.Recommended) > 0 {
			o := SeenOverride{Exercise: t.Exercise, Reason: d.OverrideReason}
			for _, w := range d.Recommended {
				o.Recommended = append(o.Recommended, domain.WorkingSet{Reps: int(w.Reps), Load: domain.Load(w.Load), RIR: int(w.RIR)})
			}
			overrides = append(overrides, o)
		}
		for _, c := range d.Calibration {
			t.Calibration = append(t.Calibration, domain.CalibrationSet{Reps: int(c.Reps), Load: domain.Load(c.Load)})
		}
		for _, w := range d.Working {
			t.Working = append(t.Working, domain.WorkingSet{Reps: int(w.Reps), Load: domain.Load(w.Load), RIR: int(w.RIR)})
		}
		out = append(out, t)
	}
	return out, overrides
}

type exerciseDoc struct {
	Exercise string   `firestore:"exercise_id"`
	Skipped  bool     `firestore:"skipped"`
	Sets     []setDoc `firestore:"sets"`
}

type setDoc struct {
	ID     string    `firestore:"set_id"`
	At     time.Time `firestore:"at"`
	Kind   string    `firestore:"kind"`
	Reps   int64     `firestore:"reps"`
	Weight int64     `firestore:"weight_tenths_lb"`
	RIR    int64     `firestore:"rir"`
	Pain   *int64    `firestore:"pain"`
	Note   string    `firestore:"note"`
}

type cardioDoc struct {
	ID         string    `firestore:"cardio_id"`
	At         time.Time `firestore:"at"`
	Exercise   string    `firestore:"exercise_id"`
	Seconds    int64     `firestore:"duration_seconds"`
	Effort     int64     `firestore:"effort"`
	Distance   *int64    `firestore:"distance_tenths_mi"`
	Resistance *int64    `firestore:"resistance"`
	Pain       *int64    `firestore:"pain"`
	Note       string    `firestore:"note"`
}

func encodeWorkout(w Workout, now time.Time) workoutDoc {
	d := workoutDoc{
		Date:       w.Date,
		Plan:       planDoc{CreatedAt: w.Plan.PlanCreatedAt.UTC(), SessionIndex: int64(w.Plan.SessionIndex)},
		Skipped:    make([]string, 0, len(w.Skipped)),
		EndedEarly: w.EndedEarly,
		Finished:   w.Finished,
		Exercises:  []exerciseDoc{},
		Cardio:     []cardioDoc{},
		Targets:    encodeTargets(w.Targets, w.Overrides),
		Versions:   w.Versions,
		UpdatedAt:  now,
	}
	for _, id := range w.Skipped {
		d.Skipped = append(d.Skipped, string(id))
	}
	for _, l := range w.Exercises() {
		e := exerciseDoc{Exercise: string(l.Exercise), Skipped: l.Skipped, Sets: []setDoc{}}
		for _, s := range l.Sets {
			e.Sets = append(e.Sets, setDoc{
				ID: s.ID, At: s.At, Kind: string(s.Log.Kind), Reps: int64(s.Log.Reps), Weight: int64(s.Log.Weight),
				RIR: int64(s.Log.RIR), Pain: ptr64(s.Log.Pain), Note: s.Log.Note,
			})
		}
		d.Exercises = append(d.Exercises, e)
	}
	for _, c := range w.Cardio {
		d.Cardio = append(d.Cardio, cardioDoc{
			ID: c.ID, At: c.At, Exercise: string(c.Log.Exercise), Seconds: int64(c.Log.Duration / time.Second),
			Effort: int64(c.Log.Effort), Distance: ptr64(c.Log.Distance), Resistance: ptr64(c.Log.Resistance),
			Pain: ptr64(c.Log.Pain), Note: c.Log.Note,
		})
	}
	return d
}

func (d workoutDoc) workout(id string) Workout {
	w := Workout{
		ID: id,
		Header: Header{
			Date:       d.Date,
			Plan:       PlanLink{PlanCreatedAt: d.Plan.CreatedAt.UTC(), SessionIndex: int(d.Plan.SessionIndex)},
			EndedEarly: d.EndedEarly,
			Finished:   d.Finished,
		},
		Versions: map[string]int64{},
	}
	w.Targets, w.Overrides = decodeTargets(d.Targets)
	for _, id := range d.Skipped {
		w.Skipped = append(w.Skipped, domain.ExerciseID(id))
	}
	for _, e := range d.Exercises {
		for _, s := range e.Sets {
			w.Sets = append(w.Sets, Set{
				ID: s.ID, Exercise: domain.ExerciseID(e.Exercise), At: s.At.UTC(),
				Log: domain.SetLog{
					Kind: domain.SetKind(s.Kind), Reps: int(s.Reps), Weight: domain.Load(s.Weight),
					RIR: int(s.RIR), Pain: fromPtr[domain.Pain](s.Pain), Note: s.Note,
				},
			})
		}
	}
	// The exercises group the sets, so the order of the time comes back.
	slices.SortStableFunc(w.Sets, func(a, b Set) int {
		if c := a.At.Compare(b.At); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	for _, c := range d.Cardio {
		w.Cardio = append(w.Cardio, Cardio{
			ID: c.ID, At: c.At.UTC(),
			Log: domain.CardioLog{
				Exercise: domain.ExerciseID(c.Exercise), Duration: time.Duration(c.Seconds) * time.Second,
				Effort: int(c.Effort), Distance: fromPtr[domain.Distance](c.Distance), Resistance: fromPtr[int](c.Resistance),
				Pain: fromPtr[domain.Pain](c.Pain), Note: c.Note,
			},
		})
	}
	for k, v := range d.Versions {
		w.Versions[k] = v
	}
	return w
}

func ptr64[T ~int](p *T) *int64 {
	if p == nil {
		return nil
	}
	v := int64(*p)
	return &v
}

func fromPtr[T ~int](p *int64) *T {
	if p == nil {
		return nil
	}
	v := T(*p)
	return &v
}
