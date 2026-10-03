//go:build emulator

// The Firestore store over the emulator. `make emulator-test` starts the
// emulators and runs this file with the tag emulator.
package workout

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/domain"
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

// TestFirestoreRoundTrip: the store keeps each field of a workout at the
// paths of D-256, with the optional values and the plan link, and keeps
// each applied op id with no end date (D-257).
func TestFirestoreRoundTrip(t *testing.T) {
	client := emulatorClient(t)
	s := FromFirestore(client)
	ctx := context.Background()
	uid := fmt.Sprintf("workout-%d", time.Now().UnixNano())

	h := header(1, workoutA)
	h.Header.Plan.PlanCreatedAt = time.Date(2026, 10, 3, 6, 11, 3, 615450000, time.UTC)
	h.Header.Skipped = []domain.ExerciseID{"leg_press"}
	h.Header.EndedEarly, h.Header.Finished = true, true
	s1 := set(2, workoutA, entityID(1), "chest_press", 10)
	s1.Set.Pain, s1.Set.Note = ptr(domain.Pain(0)), "seat 4"
	s2 := set(3, workoutA, entityID(2), "chest_press", 9)
	s2.Set.Kind = domain.SetCalibration
	c1 := cardio(4, workoutA, entityID(3))
	c1.Cardio.Distance, c1.Cardio.Resistance, c1.Cardio.Pain, c1.Cardio.Note = ptr(domain.Distance(21)), ptr(0), ptr(domain.Pain(2)), "hills"
	for _, e := range []Entry{h, s1, s2, c1} {
		if r, err := s.Apply(ctx, uid, e); err != nil || r.Version != 1 || r.Replayed {
			t.Fatalf("Apply(%s) = %+v, %v", e.Entity, r, err)
		}
	}
	mem := NewMemory()
	for _, e := range []Entry{h, s1, s2, c1} {
		if _, err := mem.Apply(ctx, uid, e); err != nil {
			t.Fatal(err)
		}
	}
	got, next, err := s.List(ctx, uid, 10, "")
	if err != nil || next != "" || len(got) != 1 {
		t.Fatalf("List = %+v, %q, %v", got, next, err)
	}
	want, _, _ := mem.List(ctx, uid, 10, "")
	if !reflect.DeepEqual(got[0], want[0]) {
		t.Fatalf("stored workout\n got %+v\nwant %+v", got[0], want[0])
	}

	op, err := client.Collection(UsersCollection).Doc(uid).Collection(OpsCollection).Doc(opID(2)).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var d opDoc
	if err := op.DataTo(&d); err != nil || d.Entity != EntitySet || d.EntityID != entityID(1) || d.WorkoutID != workoutA || d.Version != 1 {
		t.Fatalf("op document %+v, %v", d, err)
	}
	if _, ok := op.Data()["expire_at"]; ok {
		t.Fatal("the op document has an end date, want none (D-257)")
	}
	doc, err := client.Collection(UsersCollection).Doc(uid).Collection(WorkoutsCollection).Doc(workoutA).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ex := doc.Data()["exercises"].([]any); len(ex) != 2 {
		t.Fatalf("the document holds %d exercise logs, want the chest press and the skipped leg press", len(ex))
	}
}

// TestFirestoreOneApply: ten calls with the same op id at the same time
// apply it one time. The other nine read the applied op id.
func TestFirestoreOneApply(t *testing.T) {
	client := emulatorClient(t)
	s := FromFirestore(client)
	ctx := context.Background()
	uid := fmt.Sprintf("workout-once-%d", time.Now().UnixNano())
	if _, err := s.Apply(ctx, uid, header(1, workoutA)); err != nil {
		t.Fatal(err)
	}
	e := set(2, workoutA, entityID(1), "chest_press", 10)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var applied, replayed int
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := s.Apply(ctx, uid, e)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				t.Errorf("Apply = %v", err)
			case r.Version != 1:
				t.Errorf("version %d, want 1", r.Version)
			case r.Replayed:
				replayed++
			default:
				applied++
			}
		}()
	}
	wg.Wait()
	if applied != 1 || replayed != 9 {
		t.Fatalf("applied %d, replayed %d, want 1 and 9", applied, replayed)
	}
	got, _, err := s.List(ctx, uid, 10, "")
	if err != nil || len(got) != 1 || len(got[0].Sets) != 1 || got[0].Versions[VersionKey(EntitySet, entityID(1))] != 1 {
		t.Fatalf("List = %+v, %v", got, err)
	}
}

// TestFirestoreList: the query gives the newest date first, the larger
// id first for one date, and each workout one time over the pages. A
// refused entry and an unknown page token change nothing.
func TestFirestoreList(t *testing.T) {
	client := emulatorClient(t)
	s := FromFirestore(client)
	ctx := context.Background()
	uid := fmt.Sprintf("workout-list-%d", time.Now().UnixNano())
	mem := NewMemory()
	for i, d := range []string{"2026-10-01", "2026-10-03", "2026-10-02", "2026-10-03", "2026-09-30"} {
		h := header(i+1, entityID(100+i))
		h.Header.Date = d
		for _, st := range []Store{s, mem} {
			if _, err := st.Apply(ctx, uid, h); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := s.Apply(ctx, uid, set(50, entityID(999), entityID(1), "chest_press", 10)); !errors.Is(err, ErrUnknownWorkout) {
		t.Fatalf("a set of an unknown workout = %v", err)
	}
	read := func(st Store) []string {
		var ids []string
		after := ""
		for range 5 {
			list, next, err := st.List(ctx, uid, 2, after)
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range list {
				ids = append(ids, w.ID)
			}
			if after = next; after == "" {
				break
			}
		}
		return ids
	}
	if got, want := read(s), read(mem); !reflect.DeepEqual(got, want) || len(got) != 5 {
		t.Fatalf("Firestore pages %v, want %v", got, want)
	}
	if _, _, err := s.List(ctx, uid, 2, entityID(998)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("an unknown page token = %v, want ErrInvalid", err)
	}
}
