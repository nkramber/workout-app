// Package workout applies the outbox entries of the phone to the logged
// sessions of the owner (work areas 6.1 and 6.3). The phone keeps each
// log and its outbox entry in one local transaction, and sends the
// outbox later (D-77, D-132). The server applies each entry one time
// alone, keyed by its client op id, and keeps each applied op id with
// no end date (D-257).
//
// A logged session is a Workout: the header, the sets, and the cardio
// logs, in the form of the session log of "go/internal/domain" (D-256).
// An entry holds the whole new state of its entity, and the phone wins
// for a workout entry (D-258). The check of a set uses the bounds of
// D-164, the check of a cardio log uses D-123 and D-260, and a note has
// 280 characters or fewer (D-261).
//
// A log holds the notes and the dates of the owner, so no error and no
// log line holds them (D-80). An error names ids and numbers alone.
package workout

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/profile"
)

// The entities of an outbox entry.
const (
	EntityWorkout = "workout"
	EntitySet     = "set"
	EntityCardio  = "cardio"
)

// SchemaVersion is the one version of the entry form that the server
// knows (D-132). The web client names it OUTBOX_SCHEMA_VERSION.
const SchemaVersion = 1

// MaxBatch is the largest count of entries in one sync call (D-259).
const MaxBatch = 100

// MaxTargets is the largest count of target copies in one header. A
// session of the plan has 8 exercises or fewer (D-233), so the limit
// leaves room and keeps the document small.
const MaxTargets = 20

// MaxNoteRunes is the length limit of the note of a set or of a cardio
// log, in characters (D-261).
const MaxNoteRunes = 280

// ErrInvalid is the error of a bad entry. It matches domain.ErrInvalid,
// so each check error of the domain model matches it too.
var ErrInvalid = domain.ErrInvalid

// ErrUnknownWorkout is the error of a set or a cardio log of a workout
// that the server does not hold. The entry of the workout comes first in
// the outbox, so a later sync can apply it.
var ErrUnknownWorkout = errors.New("workout: unknown workout")

type checkError string

func (e checkError) Error() string { return string(e) }
func (checkError) Is(t error) bool { return t == ErrInvalid }

func invalid(format string, args ...any) error {
	return checkError(fmt.Sprintf(format, args...))
}

var (
	opIDForm = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	idForm   = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// CheckOpID refuses an op id that is not a lower-case UUIDv7 (D-132).
// The store keys each op with it, so the check comes before each read.
func CheckOpID(id string) error {
	if !opIDForm.MatchString(id) {
		return invalid("op id: want a lower-case UUIDv7")
	}
	return nil
}

// CheckID refuses an entity id that is not a lower-case UUID. A
// workout id is a document id, and a set id and a cardio id are keys in
// the document, so the form keeps each one safe.
func CheckID(what, id string) error {
	if !idForm.MatchString(id) {
		return invalid("%s: want a lower-case UUID", what)
	}
	return nil
}

// PlanLink names the session of the plan that a workout started from
// (D-248). A plan has no id, so its creation time names it.
type PlanLink struct {
	PlanCreatedAt time.Time
	SessionIndex  int
}

// Header is the state of a workout with no set and no cardio log.
// Skipped names the skipped exercises (D-47, D-63). EndedEarly records
// "finish now", which ends the workout. Targets holds the target of each
// exercise that the owner saw at the start (D-291). A header of an older
// phone has none. Overrides holds the record of each override of the
// owner among the targets (D-69, D-293).
type Header struct {
	Date       string
	Plan       PlanLink
	Skipped    []domain.ExerciseID
	EndedEarly bool
	Finished   bool
	Targets    []domain.PlannedExercise
	Overrides  []SeenOverride
}

// MaxOverrideReasonRunes is the length limit of the reason of an
// override, in characters (D-293).
const MaxOverrideReasonRunes = 200

// SeenOverride is the record of an override of the owner in a workout
// (D-69, D-293): the recommendation that the override replaced, and the
// reason of the owner. The target copy of the exercise holds the sets
// of the override, so the three stay separate records. The reason is
// data of the owner, so it never goes into a log (D-80).
type SeenOverride struct {
	Exercise    domain.ExerciseID
	Recommended []domain.WorkingSet
	Reason      string
}

// Target gives the target copy of an exercise.
func (h Header) Target(id domain.ExerciseID) (domain.PlannedExercise, bool) {
	for _, t := range h.Targets {
		if t.Exercise == id {
			return t, true
		}
	}
	return domain.PlannedExercise{}, false
}

// Set is one logged set of a workout. At is the time of the change on
// the phone, and the sets of an exercise keep its order.
type Set struct {
	ID       string
	Exercise domain.ExerciseID
	At       time.Time
	Log      domain.SetLog
}

// Cardio is one cardio log of a workout.
type Cardio struct {
	ID  string
	At  time.Time
	Log domain.CardioLog
}

// Entry is one outbox entry (D-132). One of Header, Set, and Cardio
// holds the payload, and it must agree with Entity. WorkoutID names the
// workout of a set or of a cardio log. For a workout entry, EntityID is
// the workout id.
type Entry struct {
	OpID          string
	Entity        string
	EntityID      string
	BaseVersion   int64
	At            time.Time
	SchemaVersion int
	WorkoutID     string
	Header        *Header
	Set           *domain.SetLog
	SetExercise   domain.ExerciseID
	Cardio        *domain.CardioLog
}

// Workout is one logged session. Versions holds the server version of
// each entity of the workout, by the key of VersionKey.
type Workout struct {
	ID string
	Header
	Sets     []Set
	Cardio   []Cardio
	Versions map[string]int64
}

// ExerciseLog is the log of one exercise of a workout, with the id and
// the time of each set.
type ExerciseLog struct {
	Exercise domain.ExerciseID
	Skipped  bool
	Sets     []Set
}

// Workout gives the id of the workout that the entry changes.
func (e Entry) Workout() string {
	if e.Entity == EntityWorkout {
		return e.EntityID
	}
	return e.WorkoutID
}

// Check refuses a bad entry: an unknown schema version or entity, a
// payload that does not agree with the entity, a bad id or time, and
// each log outside its bounds. It reads no stored state.
func (e Entry) Check(c domain.Catalog) error {
	if err := CheckOpID(e.OpID); err != nil {
		return err
	}
	if e.SchemaVersion != SchemaVersion {
		return invalid("schema version %d: want %d", e.SchemaVersion, SchemaVersion)
	}
	if e.BaseVersion < 0 {
		return invalid("base version %d: want 0 or more", e.BaseVersion)
	}
	if e.At.IsZero() {
		return invalid("at: want an RFC 3339 time")
	}
	var payloads int
	for _, set := range []bool{e.Header != nil, e.Set != nil, e.Cardio != nil} {
		if set {
			payloads++
		}
	}
	if payloads != 1 {
		return invalid("entity %q: want one payload, got %d", e.Entity, payloads)
	}
	switch e.Entity {
	case EntityWorkout:
		if e.Header == nil {
			return invalid("entity %q: the payload is not a workout", e.Entity)
		}
		if err := CheckID("workout id", e.EntityID); err != nil {
			return err
		}
		return e.Header.check(c)
	case EntitySet:
		if e.Set == nil {
			return invalid("entity %q: the payload is not a set", e.Entity)
		}
		if err := e.checkIDs("set id"); err != nil {
			return err
		}
		l := domain.ExerciseLog{Exercise: e.SetExercise, Sets: []domain.SetLog{*e.Set}}
		if err := l.Check(c); err != nil {
			return err
		}
		return checkNote(e.Set.Note)
	case EntityCardio:
		if e.Cardio == nil {
			return invalid("entity %q: the payload is not a cardio log", e.Entity)
		}
		if err := e.checkIDs("cardio id"); err != nil {
			return err
		}
		if err := e.Cardio.Check(c); err != nil {
			return err
		}
		return checkNote(e.Cardio.Note)
	default:
		return invalid("entity %q: want %q, %q, or %q", e.Entity, EntityWorkout, EntitySet, EntityCardio)
	}
}

func (e Entry) checkIDs(what string) error {
	if err := CheckID(what, e.EntityID); err != nil {
		return err
	}
	return CheckID("workout id", e.WorkoutID)
}

func checkNote(note string) error {
	if !utf8.ValidString(note) {
		return invalid("note: not UTF-8")
	}
	if n := utf8.RuneCountInString(note); n > MaxNoteRunes {
		return invalid("note: %d characters: want %d or fewer", n, MaxNoteRunes)
	}
	return nil
}

func (h Header) check(c domain.Catalog) error {
	if _, err := time.Parse(domain.DateLayout, h.Date); err != nil {
		return invalid("workout date: want the form %s", domain.DateLayout)
	}
	if h.Plan.PlanCreatedAt.IsZero() {
		return invalid("plan link: want the RFC 3339 time of the plan")
	}
	if i := h.Plan.SessionIndex; i < 0 || i >= profile.MaxTrainingDays {
		return invalid("plan link session index %d: want 0 to %d", i, profile.MaxTrainingDays-1)
	}
	if len(h.Targets) > MaxTargets {
		return invalid("targets: %d: want %d or fewer", len(h.Targets), MaxTargets)
	}
	copies := map[domain.ExerciseID]bool{}
	for i, t := range h.Targets {
		if err := t.Check(c); err != nil {
			return invalid("targets[%d]: %v", i, err)
		}
		if copies[t.Exercise] {
			return invalid("targets[%d]: exercise %q: in the list two times", i, t.Exercise)
		}
		copies[t.Exercise] = true
	}
	overrides := map[domain.ExerciseID]bool{}
	for i, o := range h.Overrides {
		switch {
		case !copies[o.Exercise]:
			return invalid("overrides[%d]: exercise %q: no target copy", i, o.Exercise)
		case overrides[o.Exercise]:
			return invalid("overrides[%d]: exercise %q: in the list two times", i, o.Exercise)
		case len(o.Recommended) == 0:
			return invalid("overrides[%d]: no recommended working set", i)
		case !utf8.ValidString(o.Reason) || strings.TrimSpace(o.Reason) == "":
			return invalid("overrides[%d]: want a reason of UTF-8 text", i)
		case utf8.RuneCountInString(o.Reason) > MaxOverrideReasonRunes:
			return invalid("overrides[%d]: reason: want %d characters or fewer", i, MaxOverrideReasonRunes)
		}
		for j, w := range o.Recommended {
			if err := w.Check(); err != nil {
				return invalid("overrides[%d] recommended[%d]: %v", i, j, err)
			}
		}
		overrides[o.Exercise] = true
	}
	seen := map[domain.ExerciseID]bool{}
	for _, id := range h.Skipped {
		e, ok := c.Exercise(id)
		if !ok {
			return invalid("skipped exercise %q: not in the catalog", id)
		}
		if e.Kind == domain.KindCardio {
			return invalid("skipped exercise %q: a cardio exercise", id)
		}
		if seen[id] {
			return invalid("skipped exercise %q: in the list two times", id)
		}
		if len(copies) > 0 && !copies[id] {
			return invalid("skipped exercise %q: the header has no target for it", id)
		}
		seen[id] = true
	}
	if h.EndedEarly && !h.Finished {
		return invalid("workout: ended early but not finished")
	}
	return nil
}

// Apply gives the workout after the entry, and the new version of the
// entity of the entry. w is nil when the server holds no such workout.
// The entry holds the whole new state of its entity, and the server
// version of the entity does not refuse it (D-258). Apply checks the
// entry first, and changes no field of w.
func Apply(w *Workout, e Entry, c domain.Catalog) (Workout, int64, error) {
	if err := e.Check(c); err != nil {
		return Workout{}, 0, err
	}
	if w == nil && e.Entity != EntityWorkout {
		return Workout{}, 0, fmt.Errorf("%w: %s %s of workout %s", ErrUnknownWorkout, e.Entity, e.EntityID, e.WorkoutID)
	}
	if w != nil && e.Entity == EntitySet && len(w.Targets) > 0 {
		if _, ok := w.Target(e.SetExercise); !ok {
			return Workout{}, 0, invalid("set of exercise %q: the workout has no target for it", e.SetExercise)
		}
	}
	var out Workout
	if w == nil {
		out = Workout{ID: e.EntityID}
	} else {
		out = w.clone()
	}
	switch e.Entity {
	case EntityWorkout:
		h := *e.Header
		h.Skipped = slices.Clone(h.Skipped)
		h.Targets = cloneTargets(h.Targets)
		h.Overrides = cloneOverrides(h.Overrides)
		out.Header = h
	case EntitySet:
		s := Set{ID: e.EntityID, Exercise: e.SetExercise, At: e.At.UTC(), Log: *e.Set}
		out.Sets = upsert(out.Sets, s, func(x Set) string { return x.ID })
		slices.SortStableFunc(out.Sets, func(a, b Set) int { return cmp.Or(a.At.Compare(b.At), cmp.Compare(a.ID, b.ID)) })
	case EntityCardio:
		l := Cardio{ID: e.EntityID, At: e.At.UTC(), Log: *e.Cardio}
		out.Cardio = upsert(out.Cardio, l, func(x Cardio) string { return x.ID })
		slices.SortStableFunc(out.Cardio, func(a, b Cardio) int { return cmp.Or(a.At.Compare(b.At), cmp.Compare(a.ID, b.ID)) })
	}
	if out.Versions == nil {
		out.Versions = map[string]int64{}
	}
	key := VersionKey(e.Entity, e.EntityID)
	out.Versions[key]++
	return out, out.Versions[key], nil
}

// VersionKey gives the key of the version of an entity: the entity and
// its id. The contract does not make an id unique across the entities,
// so a set and a workout with the same id keep two versions.
func VersionKey(entity, id string) string { return entity + ":" + id }

func upsert[T any](list []T, v T, id func(T) string) []T {
	for i, x := range list {
		if id(x) == id(v) {
			list[i] = v
			return list
		}
	}
	return append(list, v)
}

// Exercises gives the log of each exercise: each exercise with a set,
// in the order of its first set, then each skipped exercise with no
// set, in the order of the skip list. An exercise with a set is not
// skipped, so each log passes domain.ExerciseLog.Check.
func (w Workout) Exercises() []ExerciseLog {
	var out []ExerciseLog
	index := map[domain.ExerciseID]int{}
	for _, s := range w.Sets {
		i, ok := index[s.Exercise]
		if !ok {
			i = len(out)
			index[s.Exercise] = i
			out = append(out, ExerciseLog{Exercise: s.Exercise})
		}
		out[i].Sets = append(out[i].Sets, s)
	}
	for _, id := range w.Skipped {
		if _, ok := index[id]; !ok {
			index[id] = len(out)
			out = append(out, ExerciseLog{Exercise: id, Skipped: true})
		}
	}
	return out
}

// Session gives the workout in the form of the session log of the
// domain model, which the policy reads (D-168).
func (w Workout) Session() domain.Session {
	s := domain.Session{Date: w.Date, EndedEarly: w.EndedEarly}
	for _, l := range w.Exercises() {
		el := domain.ExerciseLog{Exercise: l.Exercise, Skipped: l.Skipped}
		for _, set := range l.Sets {
			el.Sets = append(el.Sets, set.Log)
		}
		s.Exercises = append(s.Exercises, el)
	}
	for _, c := range w.Cardio {
		s.Cardio = append(s.Cardio, c.Log)
	}
	return s
}

func cloneTargets(in []domain.PlannedExercise) []domain.PlannedExercise {
	if in == nil {
		return nil
	}
	out := make([]domain.PlannedExercise, len(in))
	for i, t := range in {
		t.Calibration = slices.Clone(t.Calibration)
		t.Working = slices.Clone(t.Working)
		out[i] = t
	}
	return out
}

func cloneOverrides(in []SeenOverride) []SeenOverride {
	if in == nil {
		return nil
	}
	out := make([]SeenOverride, len(in))
	for i, o := range in {
		o.Recommended = slices.Clone(o.Recommended)
		out[i] = o
	}
	return out
}

func (w Workout) clone() Workout {
	out := w
	out.Skipped = slices.Clone(w.Skipped)
	out.Targets = cloneTargets(w.Targets)
	out.Overrides = cloneOverrides(w.Overrides)
	out.Sets = slices.Clone(w.Sets)
	out.Cardio = slices.Clone(w.Cardio)
	out.Versions = make(map[string]int64, len(w.Versions))
	for k, v := range w.Versions {
		out.Versions[k] = v
	}
	return out
}
