//go:build emulator

// The acceptance story of PR-29 over the Firebase emulators. The test
// sends a batch of entries for one session through the API, and the
// server holds each set one time. The same batch again changes nothing.
// The server refuses a set outside the bounds of D-164, and no other
// entry of the batch changes. Then it reads the stored documents at the
// paths of D-256.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/workout"
)

func TestWorkoutAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	uid, token := signUp(t, authHost, fmt.Sprintf("workout-%d@example.test", suffix))
	otherUID, otherToken := signUp(t, authHost, fmt.Sprintf("workout-other-%d@example.test", suffix))

	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	for _, id := range []string{uid, otherUID} {
		if _, err := fs.Collection(allowlist.Collection).Doc(id).Set(ctx, map[string]any{}); err != nil {
			t.Fatal(err)
		}
	}
	client := workoutappv1connect.NewWorkoutServiceClient(http.DefaultClient, startAPI(t))

	const workoutID = "01920000-0000-7000-8000-0000000000a1"
	n := 0
	entry := func(entity, entityID string) *workoutappv1.OutboxEntry {
		n++
		return &workoutappv1.OutboxEntry{
			OpId: fmt.Sprintf("01920000-0000-7000-8000-%012x", n), Entity: entity, EntityId: entityID,
			At: time.Date(2026, 10, 3, 10, n, 0, 0, time.UTC).Format(time.RFC3339), SchemaVersion: workout.SchemaVersion,
		}
	}
	setOf := func(set int, exercise string, reps int32, weight int64) *workoutappv1.OutboxEntry {
		e := entry(workout.EntitySet, fmt.Sprintf("01920000-0000-7000-9000-%012x", set))
		e.Payload = &workoutappv1.OutboxEntry_Set{Set: &workoutappv1.SetEntry{
			WorkoutId: workoutID, ExerciseId: exercise, Kind: "working", Reps: reps, WeightTenthsLb: weight, Rir: 2,
		}}
		return e
	}
	start := entry(workout.EntityWorkout, workoutID)
	start.Payload = &workoutappv1.OutboxEntry_Workout{Workout: &workoutappv1.WorkoutHeader{
		Date: "2026-10-03", Plan: &workoutappv1.PlanLink{PlanCreatedAt: "2026-10-03T06:11:03.61545Z", SessionIndex: 0},
	}}
	cardio := entry(workout.EntityCardio, "01920000-0000-7000-9000-0000000000c1")
	cardio.Payload = &workoutappv1.OutboxEntry_Cardio{Cardio: &workoutappv1.CardioEntry{
		WorkoutId: workoutID, ExerciseId: "recumbent_bike", DurationSeconds: 20 * 60, Effort: 6,
	}}
	finish := entry(workout.EntityWorkout, workoutID)
	finish.BaseVersion = 1
	finish.Payload = &workoutappv1.OutboxEntry_Workout{Workout: &workoutappv1.WorkoutHeader{
		Date: "2026-10-03", Plan: &workoutappv1.PlanLink{PlanCreatedAt: "2026-10-03T06:11:03.61545Z", SessionIndex: 0},
		SkippedExerciseIds: []string{"leg_press"}, Finished: true,
	}}
	batch := []*workoutappv1.OutboxEntry{
		start,
		setOf(1, "chest_press", 10, 500),
		setOf(2, "chest_press", 9, 500),
		setOf(3, "chest_press", -1, 500), // outside the bounds of D-164
		setOf(4, "seated_row", 12, 0),    // a weight of 0, outside the bounds of D-164
		setOf(5, "seated_row", 11, 600),
		cardio,
		finish,
	}
	refused := map[int]bool{3: true, 4: true}

	sync := func(token string) []*workoutappv1.EntryResult {
		t.Helper()
		req := signed(token, &workoutappv1.SyncOutboxRequest{Entries: batch})
		res, err := client.SyncOutbox(ctx, req)
		if err != nil || len(res.Msg.GetResults()) != len(batch) {
			t.Fatalf("SyncOutbox = %v, %v", res, err)
		}
		return res.Msg.GetResults()
	}
	read := func(token string) []*workoutappv1.Workout {
		t.Helper()
		res, err := client.ListWorkouts(ctx, signed(token, &workoutappv1.ListWorkoutsRequest{}))
		if err != nil {
			t.Fatalf("ListWorkouts = %v", err)
		}
		return res.Msg.GetWorkouts()
	}

	// The first sync applies each entry but the two bad sets.
	first := sync(token)
	for i, r := range first {
		switch {
		case refused[i] && (r.GetStatus() != workoutappv1.EntryResult_STATUS_REFUSED || r.GetCode() != "invalid_argument"):
			t.Errorf("entry %d = %v, want a refusal", i, r)
		case !refused[i] && r.GetStatus() != workoutappv1.EntryResult_STATUS_APPLIED:
			t.Errorf("entry %d = %v, want applied", i, r)
		}
	}
	if v := first[7].GetVersion(); v != 2 {
		t.Fatalf("the finish gives version %d, want 2", v)
	}

	// The server holds each good set one time, with the link to the plan
	// session (D-248).
	got := read(token)
	if len(got) != 1 {
		t.Fatalf("ListWorkouts = %d workouts, want 1", len(got))
	}
	w := got[0]
	if w.GetWorkoutId() != workoutID || !w.GetFinished() || w.GetPlan().GetPlanCreatedAt() != "2026-10-03T06:11:03.61545Z" || w.GetPlan().GetSessionIndex() != 0 {
		t.Fatalf("workout %v", w)
	}
	ex := w.GetExercises()
	if len(ex) != 3 || ex[0].GetExerciseId() != "chest_press" || len(ex[0].GetSets()) != 2 ||
		ex[1].GetExerciseId() != "seated_row" || len(ex[1].GetSets()) != 1 || ex[1].GetSets()[0].GetReps() != 11 ||
		ex[2].GetExerciseId() != "leg_press" || !ex[2].GetSkipped() {
		t.Fatalf("exercises %v", ex)
	}
	if c := w.GetCardio(); len(c) != 1 || c[0].GetDurationSeconds() != 1200 {
		t.Fatalf("cardio %v", c)
	}

	// The same batch again gives the same results, and changes nothing.
	again := sync(token)
	for i := range again {
		if !proto.Equal(again[i], first[i]) {
			t.Errorf("replay of entry %d = %v, want %v", i, again[i], first[i])
		}
	}
	if after := read(token); len(after) != 1 || !proto.Equal(after[0], w) {
		t.Fatalf("the replay changed the workouts: %v", after)
	}

	// The stored documents: one workout at users/{uid}/workouts/{id}, and
	// one op id document for each applied entry (D-256, D-257).
	user := fs.Collection(workout.UsersCollection).Doc(uid)
	ops, err := user.Collection(workout.OpsCollection).Documents(ctx).GetAll()
	if err != nil || len(ops) != len(batch)-len(refused) {
		t.Fatalf("%d op documents, %v, want %d", len(ops), err, len(batch)-len(refused))
	}
	stored, _, err := workout.FromFirestore(fs).List(ctx, uid, 10, "")
	if err != nil || len(stored) != 1 {
		t.Fatalf("stored = %v, %v", stored, err)
	}
	if err := stored[0].Session().Check(domain.DefaultCatalog()); err != nil {
		t.Fatalf("the stored workout fails the check of the domain session log: %v", err)
	}

	// The other uid reads no workout, and the sets of the workout of the
	// first uid are not its own: they need a workout of its own.
	if other := read(otherToken); len(other) != 0 {
		t.Fatalf("the other uid reads %v", other)
	}
	res, err := client.SyncOutbox(ctx, signed(otherToken, &workoutappv1.SyncOutboxRequest{Entries: batch[1:2]}))
	if err != nil || res.Msg.GetResults()[0].GetCode() != "failed_precondition" {
		t.Fatalf("a set of the other uid = %v, %v, want failed_precondition", res, err)
	}

	// A batch over the limit changes nothing (D-259).
	big := make([]*workoutappv1.OutboxEntry, workout.MaxBatch+1)
	for i := range big {
		big[i] = setOf(100+i, "chest_press", 5, 500)
	}
	if _, err := client.SyncOutbox(ctx, signed(token, &workoutappv1.SyncOutboxRequest{Entries: big})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a batch of %d = %v, want InvalidArgument", len(big), err)
	}
	if after := read(token); !proto.Equal(after[0], w) {
		t.Fatalf("the refused batch changed the workout: %v", after)
	}
}
