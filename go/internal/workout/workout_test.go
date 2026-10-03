package workout

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
)

const (
	workoutA = "01920000-0000-7000-8000-0000000000a1"
	workoutB = "01920000-0000-7000-8000-0000000000b1"
)

var t0 = time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)

// opID gives the op id n, a lower-case UUIDv7.
func opID(n int) string { return fmt.Sprintf("01920000-0000-7000-8000-%012x", n) }

// entityID gives the entity id n, a lower-case UUID.
func entityID(n int) string { return fmt.Sprintf("01920000-0000-7000-9000-%012x", n) }

func header(n int, workoutID string) Entry {
	return Entry{
		OpID: opID(n), Entity: EntityWorkout, EntityID: workoutID, At: t0.Add(time.Duration(n) * time.Second),
		SchemaVersion: SchemaVersion,
		Header: &Header{
			Date: "2026-10-03",
			Plan: PlanLink{PlanCreatedAt: time.Date(2026, 10, 3, 6, 11, 3, 0, time.UTC), SessionIndex: 1},
		},
	}
}

func set(n int, workoutID string, setID string, exercise domain.ExerciseID, reps int) Entry {
	return Entry{
		OpID: opID(n), Entity: EntitySet, EntityID: setID, WorkoutID: workoutID, At: t0.Add(time.Duration(n) * time.Second),
		SchemaVersion: SchemaVersion, SetExercise: exercise,
		Set: &domain.SetLog{Kind: domain.SetWorking, Reps: reps, Weight: domain.Pounds(50), RIR: 2},
	}
}

func cardio(n int, workoutID, cardioID string) Entry {
	return Entry{
		OpID: opID(n), Entity: EntityCardio, EntityID: cardioID, WorkoutID: workoutID, At: t0.Add(time.Duration(n) * time.Second),
		SchemaVersion: SchemaVersion,
		Cardio:        &domain.CardioLog{Exercise: "treadmill", Duration: 12 * time.Minute, Effort: 6},
	}
}

func ptr[T any](v T) *T { return &v }

// TestCheckRefuses: the check refuses each bad entry with an error of
// ErrInvalid, and accepts each good one. A cardio log of 12 minutes is
// valid, because the log holds the true duration (D-260).
func TestCheckRefuses(t *testing.T) {
	c := domain.DefaultCatalog()
	for name, e := range map[string]Entry{
		"a workout": header(1, workoutA),
		"a set":     set(2, workoutA, entityID(1), "chest_press", 10),
		"a set of 0 reps, which says that the load was too heavy": set(3, workoutA, entityID(2), "chest_press", 0),
		"12 minutes of cardio": cardio(4, workoutA, entityID(3)),
		"a note of 280 characters": func() Entry {
			e := set(5, workoutA, entityID(4), "chest_press", 8)
			e.Set.Note = strings.Repeat("é", MaxNoteRunes)
			return e
		}(),
	} {
		if err := e.Check(c); err != nil {
			t.Errorf("%s: Check = %v, want nil", name, err)
		}
	}

	bad := map[string]func() Entry{
		"an op id of version 4": func() Entry { e := header(1, workoutA); e.OpID = "01920000-0000-4000-8000-000000000001"; return e },
		"an upper-case op id":   func() Entry { e := header(1, workoutA); e.OpID = "0192ABCD-0000-7000-8000-00000000000A"; return e },
		"schema version 2":      func() Entry { e := header(1, workoutA); e.SchemaVersion = 2; return e },
		"schema version 0":      func() Entry { e := header(1, workoutA); e.SchemaVersion = 0; return e },
		"an unknown entity":     func() Entry { e := header(1, workoutA); e.Entity = "machine"; return e },
		"a set payload on a workout entity": func() Entry {
			e := set(1, workoutA, entityID(1), "chest_press", 10)
			e.Entity = EntityWorkout
			return e
		},
		"two payloads": func() Entry {
			e := header(1, workoutA)
			e.Set = set(2, workoutA, entityID(1), "chest_press", 10).Set
			return e
		},
		"no payload":           func() Entry { e := header(1, workoutA); e.Header = nil; return e },
		"a base version of -1": func() Entry { e := header(1, workoutA); e.BaseVersion = -1; return e },
		"no time":              func() Entry { e := header(1, workoutA); e.At = time.Time{}; return e },
		"a workout id that is not a UUID": func() Entry {
			e := header(1, workoutA)
			e.EntityID = "../plan/active"
			return e
		},
		"a set of a workout id that is not a UUID": func() Entry {
			e := set(1, workoutA, entityID(1), "chest_press", 10)
			e.WorkoutID = "a/b"
			return e
		},
		"a bad date":       func() Entry { e := header(1, workoutA); e.Header.Date = "10/03/2026"; return e },
		"no plan time":     func() Entry { e := header(1, workoutA); e.Header.Plan.PlanCreatedAt = time.Time{}; return e },
		"session index 4":  func() Entry { e := header(1, workoutA); e.Header.Plan.SessionIndex = 4; return e },
		"session index -1": func() Entry { e := header(1, workoutA); e.Header.Plan.SessionIndex = -1; return e },
		"an unknown skip": func() Entry {
			e := header(1, workoutA)
			e.Header.Skipped = []domain.ExerciseID{"bench_press"}
			return e
		},
		"a cardio skip": func() Entry { e := header(1, workoutA); e.Header.Skipped = []domain.ExerciseID{"treadmill"}; return e },
		"a skip two times": func() Entry {
			e := header(1, workoutA)
			e.Header.Skipped = []domain.ExerciseID{"leg_press", "leg_press"}
			return e
		},
		"ended early, open": func() Entry { e := header(1, workoutA); e.Header.EndedEarly = true; return e },
		"reps of -1":        func() Entry { return set(1, workoutA, entityID(1), "chest_press", -1) },
		"a weight of 0":     func() Entry { e := set(1, workoutA, entityID(1), "chest_press", 10); e.Set.Weight = 0; return e },
		"rir of -1":         func() Entry { e := set(1, workoutA, entityID(1), "chest_press", 10); e.Set.RIR = -1; return e },
		"pain of 11": func() Entry {
			e := set(1, workoutA, entityID(1), "chest_press", 10)
			e.Set.Pain = ptr(domain.Pain(11))
			return e
		},
		"an unknown set kind": func() Entry { e := set(1, workoutA, entityID(1), "chest_press", 10); e.Set.Kind = "warmup"; return e },
		"a set of a cardio exercise": func() Entry {
			return set(1, workoutA, entityID(1), "treadmill", 10)
		},
		"a set of an unknown exercise": func() Entry { return set(1, workoutA, entityID(1), "bench_press", 10) },
		"a note of 281 characters": func() Entry {
			e := set(1, workoutA, entityID(1), "chest_press", 10)
			e.Set.Note = strings.Repeat("a", MaxNoteRunes+1)
			return e
		},
		"a note that is not UTF-8": func() Entry {
			e := set(1, workoutA, entityID(1), "chest_press", 10)
			e.Set.Note = "\xff"
			return e
		},
		"cardio of 0 seconds":   func() Entry { e := cardio(1, workoutA, entityID(1)); e.Cardio.Duration = 0; return e },
		"a cardio effort of 0":  func() Entry { e := cardio(1, workoutA, entityID(1)); e.Cardio.Effort = 0; return e },
		"a cardio effort of 11": func() Entry { e := cardio(1, workoutA, entityID(1)); e.Cardio.Effort = 11; return e },
		"cardio on a machine":   func() Entry { e := cardio(1, workoutA, entityID(1)); e.Cardio.Exercise = "chest_press"; return e },
		"a cardio note of 281 characters": func() Entry {
			e := cardio(1, workoutA, entityID(1))
			e.Cardio.Note = strings.Repeat("a", MaxNoteRunes+1)
			return e
		},
	}
	for name, make := range bad {
		err := make().Check(c)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: Check = %v, want ErrInvalid", name, err)
		}
	}
}

// TestCheckHidesTheNote: an error names no note and no date (D-80).
func TestCheckHidesTheNote(t *testing.T) {
	e := set(1, workoutA, entityID(1), "chest_press", -1)
	e.Set.Note = "secret-note"
	h := header(2, workoutA)
	h.Header.Date = "2026-13-45"
	for _, e := range []Entry{e, h} {
		err := e.Check(domain.DefaultCatalog())
		if err == nil || strings.Contains(err.Error(), "secret-note") || strings.Contains(err.Error(), "2026-13-45") {
			t.Fatalf("Check = %v, want an error with no note and no date", err)
		}
	}
}

// TestApply: a set needs its workout, an edit replaces a set and adds 1
// to its version, the sets keep the order of their time, and a set
// makes a skipped exercise not skipped. The result passes the check of
// the domain session log.
func TestApply(t *testing.T) {
	c := domain.DefaultCatalog()
	if _, _, err := Apply(nil, set(1, workoutA, entityID(1), "chest_press", 10), c); !errors.Is(err, ErrUnknownWorkout) {
		t.Fatalf("a set of no workout = %v, want ErrUnknownWorkout", err)
	}
	h := header(1, workoutA)
	h.Header.Skipped = []domain.ExerciseID{"leg_press", "seated_row"}
	w, v, err := Apply(nil, h, c)
	if err != nil || v != 1 || w.ID != workoutA {
		t.Fatalf("Apply(header) = %+v, %d, %v", w, v, err)
	}
	steps := []Entry{
		set(4, workoutA, entityID(2), "chest_press", 9),
		set(3, workoutA, entityID(1), "chest_press", 10), // an earlier time, sent later
		set(5, workoutA, entityID(3), "seated_row", 12),
		cardio(6, workoutA, entityID(4)),
	}
	for _, e := range steps {
		before := w.clone()
		if w, v, err = Apply(&w, e, c); err != nil || v != 1 {
			t.Fatalf("Apply(%s %s) = %d, %v", e.Entity, e.EntityID, v, err)
		}
		if len(before.Sets) == len(w.Sets) && len(before.Cardio) == len(w.Cardio) {
			t.Fatalf("Apply(%s) changed nothing", e.Entity)
		}
	}
	edit := set(7, workoutA, entityID(2), "chest_press", 7)
	original := w.clone()
	if w, v, err = Apply(&w, edit, c); err != nil || v != 2 || len(w.Sets) != 3 {
		t.Fatalf("edit = %d sets, version %d, %v", len(w.Sets), v, err)
	}
	if original.Sets[1].Log.Reps != 9 {
		t.Fatal("Apply changed the workout of its argument")
	}
	ex := w.Exercises()
	if len(ex) != 3 || ex[0].Exercise != "chest_press" || ex[1].Exercise != "seated_row" || ex[2].Exercise != "leg_press" {
		t.Fatalf("exercises %+v", ex)
	}
	if ex[0].Sets[0].ID != entityID(1) || ex[0].Sets[1].Log.Reps != 7 || ex[1].Skipped || !ex[2].Skipped {
		t.Fatalf("exercise logs %+v", ex)
	}
	if err := w.Session().Check(c); err != nil {
		t.Fatalf("Session().Check = %v", err)
	}
}

// TestMemoryIdempotent: the same op id applies one time, and its replay
// gives the same version. A refused entry changes nothing.
func TestMemoryIdempotent(t *testing.T) {
	ctx := context.Background()
	s := NewMemory()
	batch := []Entry{header(1, workoutA), set(2, workoutA, entityID(1), "chest_press", 10), set(3, workoutA, entityID(2), "chest_press", 9)}
	for round := range 2 {
		for _, e := range batch {
			r, err := s.Apply(ctx, "uid-a", e)
			if err != nil || r.Version != 1 || r.Replayed != (round == 1) {
				t.Fatalf("round %d %s: %+v, %v", round, e.Entity, r, err)
			}
		}
	}
	list, _, err := s.List(ctx, "uid-a", 10, "")
	if err != nil || len(list) != 1 || len(list[0].Sets) != 2 {
		t.Fatalf("List = %+v, %v", list, err)
	}
	if _, err := s.Apply(ctx, "uid-a", set(4, workoutA, entityID(3), "chest_press", -1)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("a set of -1 reps = %v", err)
	}
	if _, err := s.Apply(ctx, "uid-a", set(5, workoutB, entityID(4), "chest_press", 10)); !errors.Is(err, ErrUnknownWorkout) {
		t.Fatalf("a set of an unknown workout = %v", err)
	}
	if after, _, _ := s.List(ctx, "uid-a", 10, ""); len(after[0].Sets) != 2 || len(after) != 1 {
		t.Fatalf("a refused entry changed the store: %+v", after)
	}
	// A refused op id stays free, so a correct entry can use it later.
	if r, err := s.Apply(ctx, "uid-a", set(4, workoutA, entityID(3), "chest_press", 8)); err != nil || r.Replayed {
		t.Fatalf("the op id of a refused entry = %+v, %v", r, err)
	}
	if other, _, _ := s.List(ctx, "uid-b", 10, ""); len(other) != 0 {
		t.Fatalf("uid-b reads %+v", other)
	}
}

// TestMemoryList: the newest date comes first, and the pages hold each
// workout one time.
func TestMemoryList(t *testing.T) {
	ctx := context.Background()
	s := NewMemory()
	dates := []string{"2026-10-01", "2026-10-03", "2026-10-02", "2026-10-03"}
	for i, d := range dates {
		h := header(i+1, entityID(100+i))
		h.Header.Date = d
		if _, err := s.Apply(ctx, "uid-a", h); err != nil {
			t.Fatal(err)
		}
	}
	var got []string
	after := ""
	for range 3 {
		list, next, err := s.List(ctx, "uid-a", 3, after)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range list {
			got = append(got, w.Date+" "+w.ID[len(w.ID)-3:])
		}
		if after = next; after == "" {
			break
		}
	}
	want := "2026-10-03 067 2026-10-03 065 2026-10-02 066 2026-10-01 064"
	if strings.Join(got, " ") != want {
		t.Fatalf("pages = %q, want %q", strings.Join(got, " "), want)
	}
	if _, _, err := s.List(ctx, "uid-a", 3, entityID(999)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("an unknown page token = %v, want ErrInvalid", err)
	}
}
