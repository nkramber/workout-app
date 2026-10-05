package workoutsvc

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/revise"
	"github.com/nkramber/workout-app/go/internal/workout"
)

var signedIn = auth.WithUserID(context.Background(), "uid-a")

const workoutID = "01920000-0000-7000-8000-0000000000a1"

func opID(n int) string { return fmt.Sprintf("01920000-0000-7000-8000-%012x", n) }

func setID(n int) string { return fmt.Sprintf("01920000-0000-7000-9000-%012x", n) }

func at(n int) string {
	return time.Date(2026, 10, 3, 10, 0, n, 0, time.UTC).Format(time.RFC3339)
}

func headerEntry(n int) *workoutappv1.OutboxEntry {
	return &workoutappv1.OutboxEntry{
		OpId: opID(n), Entity: workout.EntityWorkout, EntityId: workoutID, At: at(n), SchemaVersion: workout.SchemaVersion,
		Payload: &workoutappv1.OutboxEntry_Workout{Workout: &workoutappv1.WorkoutHeader{
			Date: "2026-10-03", Plan: &workoutappv1.PlanLink{PlanCreatedAt: "2026-10-03T06:11:03.6Z", SessionIndex: 1},
		}},
	}
}

func setEntry(n, set int, reps int32) *workoutappv1.OutboxEntry {
	return &workoutappv1.OutboxEntry{
		OpId: opID(n), Entity: workout.EntitySet, EntityId: setID(set), At: at(n), SchemaVersion: workout.SchemaVersion,
		Payload: &workoutappv1.OutboxEntry_Set{Set: &workoutappv1.SetEntry{
			WorkoutId: workoutID, ExerciseId: "chest_press", Kind: "working", Reps: reps, WeightTenthsLb: 500, Rir: 2,
			Pain: proto.Int32(1), Note: "seat 4",
		}},
	}
}

func cardioEntry(n int) *workoutappv1.OutboxEntry {
	return &workoutappv1.OutboxEntry{
		OpId: opID(n), Entity: workout.EntityCardio, EntityId: setID(900 + n), At: at(n), SchemaVersion: workout.SchemaVersion,
		Payload: &workoutappv1.OutboxEntry_Cardio{Cardio: &workoutappv1.CardioEntry{
			WorkoutId: workoutID, ExerciseId: "treadmill", DurationSeconds: 1500, Effort: 7, DistanceTenthsMi: proto.Int32(21),
		}},
	}
}

func sync(t *testing.T, s *Server, entries ...*workoutappv1.OutboxEntry) []*workoutappv1.EntryResult {
	t.Helper()
	res, err := s.SyncOutbox(signedIn, connect.NewRequest(&workoutappv1.SyncOutboxRequest{Entries: entries}))
	if err != nil {
		t.Fatalf("SyncOutbox = %v", err)
	}
	if len(res.Msg.GetResults()) != len(entries) {
		t.Fatalf("%d results for %d entries", len(res.Msg.GetResults()), len(entries))
	}
	return res.Msg.GetResults()
}

func list(t *testing.T, s *Server) []*workoutappv1.Workout {
	t.Helper()
	res, err := s.ListWorkouts(signedIn, connect.NewRequest(&workoutappv1.ListWorkoutsRequest{}))
	if err != nil {
		t.Fatalf("ListWorkouts = %v", err)
	}
	return res.Msg.GetWorkouts()
}

// TestSyncOutbox: a batch with one bad set applies each other entry, and
// gives a result for each entry in the order of the request. The same
// batch again gives the same results and changes nothing.
func TestSyncOutbox(t *testing.T) {
	s := New(workout.NewMemory(), inventory.NewMemory())
	batch := []*workoutappv1.OutboxEntry{headerEntry(1), setEntry(2, 1, 10), setEntry(3, 2, -1), setEntry(4, 3, 8), cardioEntry(5)}
	first := sync(t, s, batch...)
	for i, r := range first {
		wantStatus := workoutappv1.EntryResult_STATUS_APPLIED
		if i == 2 {
			wantStatus = workoutappv1.EntryResult_STATUS_REFUSED
		}
		if r.GetOpId() != batch[i].GetOpId() || r.GetStatus() != wantStatus {
			t.Fatalf("result %d = %v", i, r)
		}
	}
	if r := first[2]; r.GetCode() != CodeInvalidArgument || !strings.Contains(r.GetMessage(), "reps -1") || r.GetVersion() != 0 {
		t.Fatalf("the refused set = %v", r)
	}
	before := list(t, s)
	again := sync(t, s, batch...)
	for i := range again {
		if !proto.Equal(again[i], first[i]) {
			t.Fatalf("replay result %d = %v, want %v", i, again[i], first[i])
		}
	}
	after := list(t, s)
	if len(after) != 1 || !proto.Equal(after[0], before[0]) {
		t.Fatalf("the replay changed the workouts: %v", after)
	}
	w := after[0]
	if w.GetPlan().GetPlanCreatedAt() != "2026-10-03T06:11:03.6Z" || w.GetPlan().GetSessionIndex() != 1 || w.GetDate() != "2026-10-03" {
		t.Fatalf("workout header %v", w)
	}
	if len(w.GetExercises()) != 1 || len(w.GetExercises()[0].GetSets()) != 2 || w.GetExercises()[0].GetSets()[0].GetPain() != 1 {
		t.Fatalf("exercises %v", w.GetExercises())
	}
	if c := w.GetCardio(); len(c) != 1 || c[0].GetDurationSeconds() != 1500 || c[0].GetDistanceTenthsMi() != 21 || c[0].Resistance != nil {
		t.Fatalf("cardio %v", c)
	}
}

// TestSyncOutboxRefuses: each bad entry gets its code, and a batch over
// the limit gets INVALID_ARGUMENT with no change (D-259).
func TestSyncOutboxRefuses(t *testing.T) {
	s := New(workout.NewMemory(), inventory.NewMemory())
	unknownSchema := headerEntry(1)
	unknownSchema.SchemaVersion = 2
	unknownEntity := headerEntry(2)
	unknownEntity.Entity = "machine"
	badTime := headerEntry(3)
	badTime.At = "yesterday"
	noPayload := headerEntry(4)
	noPayload.Payload = nil
	badPlanTime := headerEntry(5)
	badPlanTime.GetWorkout().Plan.PlanCreatedAt = "today"
	noPlan := headerEntry(6)
	noPlan.GetWorkout().Plan = nil
	results := sync(t, s, unknownSchema, unknownEntity, badTime, noPayload, badPlanTime, noPlan, setEntry(7, 1, 10))
	for i, r := range results {
		want := CodeInvalidArgument
		if i == 6 {
			want = CodeFailedPrecondition
		}
		if r.GetStatus() != workoutappv1.EntryResult_STATUS_REFUSED || r.GetCode() != want {
			t.Errorf("result %d = %v, want %s", i, r, want)
		}
	}
	// A header with no plan link reaches the check of the entry: the
	// getters of the generated code read a nil plan as empty (D-248).
	if m := results[5].GetMessage(); !strings.Contains(m, "plan link") {
		t.Fatalf("the header with no plan link gives %q, want the plan link check", m)
	}
	if got := list(t, s); len(got) != 0 {
		t.Fatalf("the refused entries stored %v", got)
	}

	var big []*workoutappv1.OutboxEntry
	for i := range workout.MaxBatch + 1 {
		big = append(big, setEntry(100+i, 100+i, 10))
	}
	big[0] = headerEntry(99)
	_, err := s.SyncOutbox(signedIn, connect.NewRequest(&workoutappv1.SyncOutboxRequest{Entries: big}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a batch of %d = %v, want InvalidArgument", len(big), err)
	}
	if got := list(t, s); len(got) != 0 {
		t.Fatalf("the refused batch stored %v", got)
	}
	if got := sync(t, s, big[:workout.MaxBatch]...); got[workout.MaxBatch-1].GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
		t.Fatalf("a batch of %d = %v", workout.MaxBatch, got[workout.MaxBatch-1])
	}
}

type failStore struct{ workout.Store }

func (failStore) Apply(context.Context, string, workout.Entry) (workout.Result, error) {
	return workout.Result{}, errors.New("rpc error: users/uid-a/ops/x: unavailable")
}

func (failStore) List(context.Context, string, int, string) ([]workout.Workout, string, error) {
	return nil, "", errors.New("rpc error: users/uid-a/workouts: unavailable")
}

// TestStoreFailure: a store error gives INTERNAL with a fixed text, and
// no path.
func TestStoreFailure(t *testing.T) {
	s := New(failStore{}, inventory.NewMemory())
	_, err := s.SyncOutbox(signedIn, connect.NewRequest(&workoutappv1.SyncOutboxRequest{Entries: []*workoutappv1.OutboxEntry{headerEntry(1)}}))
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users/") {
		t.Fatalf("SyncOutbox = %v", err)
	}
	_, err = s.ListWorkouts(signedIn, connect.NewRequest(&workoutappv1.ListWorkoutsRequest{}))
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users/") {
		t.Fatalf("ListWorkouts = %v", err)
	}
}

// TestListWorkouts: the limit and the page token.
func TestListWorkouts(t *testing.T) {
	s := New(workout.NewMemory(), inventory.NewMemory())
	for i, d := range []string{"2026-10-01", "2026-10-02", "2026-10-03"} {
		e := headerEntry(i + 1)
		e.EntityId = setID(i + 1)
		e.GetWorkout().Date = d
		sync(t, s, e)
	}
	res, err := s.ListWorkouts(signedIn, connect.NewRequest(&workoutappv1.ListWorkoutsRequest{Limit: 2}))
	if err != nil || len(res.Msg.GetWorkouts()) != 2 || res.Msg.GetWorkouts()[0].GetDate() != "2026-10-03" || res.Msg.GetNextPageToken() != setID(2) {
		t.Fatalf("page 1 = %v, %v", res, err)
	}
	res, err = s.ListWorkouts(signedIn, connect.NewRequest(&workoutappv1.ListWorkoutsRequest{Limit: 2, PageToken: res.Msg.GetNextPageToken()}))
	if err != nil || len(res.Msg.GetWorkouts()) != 1 || res.Msg.GetWorkouts()[0].GetDate() != "2026-10-01" || res.Msg.GetNextPageToken() != "" {
		t.Fatalf("page 2 = %v, %v", res, err)
	}
	for name, req := range map[string]*workoutappv1.ListWorkoutsRequest{
		"limit -1":              {Limit: -1},
		"limit 51":              {Limit: MaxLimit + 1},
		"a token of a path":     {PageToken: "../plan"},
		"a token of no workout": {PageToken: setID(99)},
	} {
		if _, err := s.ListWorkouts(signedIn, connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: ListWorkouts = %v, want InvalidArgument", name, err)
		}
	}
	if _, err := s.ListWorkouts(context.Background(), connect.NewRequest(&workoutappv1.ListWorkoutsRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("no uid = %v, want Unauthenticated", err)
	}
}

// fakeReviser records each revision.
type fakeReviser struct {
	calls []string
	err   error
}

func (f *fakeReviser) Revise(_ context.Context, uid, w string) (revise.Result, error) {
	f.calls = append(f.calls, uid+" "+w)
	return revise.Result{}, f.err
}

// TestSyncOutboxRevises: the sync revises the plan one time for each
// finished workout of the batch, after each entry of the batch (D-292).
// A replayed finish revises again, and the reviser skips a done
// revision. A store failure of a revision gives UNAVAILABLE, and the
// entries stay applied, so a replay of the batch runs the revision
// again. The header keeps the target copies (D-291).
func TestSyncOutboxRevises(t *testing.T) {
	r := &fakeReviser{}
	s := New(workout.NewMemory(), inventory.NewMemory()).WithReviser(r, nil)
	start := headerEntry(1)
	start.GetWorkout().Targets = []*workoutappv1.SeenTarget{{
		ExerciseId: "chest_press", RestSeconds: 60,
		CalibrationSets:     []*workoutappv1.PlannedSet{{Reps: 8, LoadTenthLb: 400}},
		WorkingSets:         []*workoutappv1.PlannedSet{{Reps: 8, LoadTenthLb: 500, RirTarget: 2}},
		FirstSetCalibration: true,
		FollowMaxTenthLb:    550,
	}}
	finish := proto.Clone(start).(*workoutappv1.OutboxEntry)
	finish.OpId, finish.At = opID(3), at(3)
	finish.GetWorkout().Finished = true
	twice := proto.Clone(finish).(*workoutappv1.OutboxEntry)
	twice.OpId, twice.At = opID(4), at(4)

	sync(t, s, start, setEntry(2, 1, 8))
	if len(r.calls) != 0 {
		t.Fatalf("an open workout revised the plan: %v", r.calls)
	}
	sync(t, s, finish, twice)
	if want := []string{"uid-a " + workoutID}; !slices.Equal(r.calls, want) {
		t.Fatalf("revisions %v, want %v", r.calls, want)
	}
	r.err = errors.New("rpc error: users/uid-a/plan/active: unavailable")
	_, err := s.SyncOutbox(signedIn, connect.NewRequest(&workoutappv1.SyncOutboxRequest{Entries: []*workoutappv1.OutboxEntry{finish}}))
	if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "users/") {
		t.Fatalf("a failed revision: %v, want UNAVAILABLE with no path", err)
	}
	r.err = nil
	for _, res := range sync(t, s, finish) {
		if res.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED {
			t.Fatalf("the replay after a failed revision: %v", res)
		}
	}
	if len(r.calls) != 3 {
		t.Fatalf("revisions %d, want 3: the failed one and its replay", len(r.calls))
	}
	w := list(t, s)[0]
	if got := w.GetTargets(); len(got) != 1 || !proto.Equal(got[0], start.GetWorkout().GetTargets()[0]) {
		t.Fatalf("copies %v", got)
	}
}
