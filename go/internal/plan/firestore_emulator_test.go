//go:build emulator

// The Firestore stores over the emulator. `make emulator-test` starts the
// emulators and runs this file with the tag emulator.
package plan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/ai"
)

func emulatorClient(t *testing.T) *firestore.Client {
	t.Helper()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" || os.Getenv("GOOGLE_CLOUD_PROJECT") != "demo-workout-app" {
		t.Fatal("run this test with `make emulator-test`, which sets FIRESTORE_EMULATOR_HOST and the demo project")
	}
	client, err := firestore.NewClient(context.Background(), "demo-workout-app")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestFirestoreRoundTrip: the store keeps each field of a plan and the
// exclusions at the paths of D-226, and a save with an old revision
// changes nothing (D-234).
func TestFirestoreRoundTrip(t *testing.T) {
	client := emulatorClient(t)
	s := FromFirestore(client)
	ctx := context.Background()
	id := fmt.Sprintf("plan-%d", time.Now().UnixNano())

	if _, ok, err := s.Get(ctx, id); ok || err != nil {
		t.Fatalf("Get of a new uid = %v, %v, want no plan", ok, err)
	}
	if ex, err := s.Exclusions(ctx, id); err != nil || len(ex.Items) != 0 || ex.Revision != 0 {
		t.Fatalf("Exclusions of a new uid = %+v, %v", ex, err)
	}
	want := fullPlan()
	ex := Exclusions{Items: []Exclusion{{"seated_row", "a reason"}}}
	if err := s.Save(ctx, id, want, ex, true); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.Get(ctx, id)
	if err != nil || !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("Get = %+v, %v, %v\nwant %+v", got, ok, err, want)
	}
	gotEx, err := s.Exclusions(ctx, id)
	if err != nil || gotEx.Revision != 1 || !reflect.DeepEqual(gotEx.Items, ex.Items) {
		t.Fatalf("Exclusions = %+v, %v", gotEx, err)
	}
	for _, path := range []string{"users/" + id + "/plan/active", "users/" + id + "/exclusions/active"} {
		if _, err := client.Doc(path).Get(ctx); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
	}

	// A save from revision 0 is a conflict, and the plan stays.
	if err := s.Save(ctx, id, Plan{Today: "other"}, Exclusions{}, false); !errors.Is(err, ErrConflict) {
		t.Fatalf("Save with an old revision = %v, want ErrConflict", err)
	}
	if got, _, _ := s.Get(ctx, id); got.Today != want.Today {
		t.Fatal("a refused save changed the plan")
	}
	// A save that keeps the list keeps its revision.
	if err := s.Save(ctx, id, Plan{Today: "next"}, gotEx, false); err != nil {
		t.Fatal(err)
	}
	if again, _ := s.Exclusions(ctx, id); again.Revision != 1 {
		t.Fatalf("revision %d, want 1", again.Revision)
	}
}

// TestFirestoreUpdate: an update changes the plan of the same CreatedAt
// in one transaction. A new plan, no plan, and an error of the change
// write nothing (D-292).
func TestFirestoreUpdate(t *testing.T) {
	s := FromFirestore(emulatorClient(t))
	ctx := context.Background()
	id := fmt.Sprintf("plan-update-%d", time.Now().UnixNano())
	p := fullPlan()
	if err := s.Update(ctx, id, p.CreatedAt, func(*Plan) error { return nil }); !errors.Is(err, ErrPlanReplaced) {
		t.Fatalf("no plan: %v", err)
	}
	if err := s.Save(ctx, id, p, Exclusions{}, false); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(ctx, id, p.CreatedAt.Add(time.Second), func(*Plan) error { return nil }); !errors.Is(err, ErrPlanReplaced) {
		t.Fatalf("another plan: %v", err)
	}
	boom := errors.New("boom")
	if err := s.Update(ctx, id, p.CreatedAt, func(q *Plan) error { q.Summary = "x"; return boom }); !errors.Is(err, boom) {
		t.Fatalf("a failed change: %v", err)
	}
	if err := s.Update(ctx, id, p.CreatedAt, func(q *Plan) error {
		q.Revisions++
		q.RevisedWorkouts = append(q.RevisedWorkouts, "w3")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	got, _, _ := s.Get(ctx, id)
	if got.Summary != p.Summary || got.Revisions != 3 || !got.Revised("w3") || !got.CreatedAt.Equal(p.CreatedAt) {
		t.Fatalf("plan after the updates: %+v", got)
	}
}

// TestFirestoreErrors: an error record is a new document of the
// top-level collection, with the field of the TTL policy (D-236).
func TestFirestoreErrors(t *testing.T) {
	client := emulatorClient(t)
	ctx := context.Background()
	id := fmt.Sprintf("plan-errors-%d", time.Now().UnixNano())
	r := ErrorRecord{User: id, Time: now, ExpireAt: now.Add(ErrorRetention), Request: KindPlan, Attempt: 1, MaxAttempts: MaxAttempts,
		Status: ai.StatusMalformed, Cause: "cause", Output: "output"}
	for range 2 {
		if err := ErrorsFromFirestore(client).Add(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	docs, err := client.Collection(ErrorsCollection).Where("uid", "==", id).Documents(ctx).GetAll()
	if err != nil || len(docs) != 2 {
		t.Fatalf("%d records, %v, want 2", len(docs), err)
	}
	var d errorDoc
	if err := docs[0].DataTo(&d); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d, encodeError(r)) {
		t.Fatalf("record %+v", d)
	}
}
