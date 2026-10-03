//go:build emulator

// The Firestore store over the emulator. `make emulator-test` starts the
// emulators and runs this file with the tag emulator.
package inventory

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

func emulatorStore(t *testing.T) (*Firestore, *firestore.Client) {
	t.Helper()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" || os.Getenv("GOOGLE_CLOUD_PROJECT") != "demo-workout-app" {
		t.Fatal("run this test with `make emulator-test`, which sets FIRESTORE_EMULATOR_HOST and the demo project")
	}
	client, err := firestore.NewClient(context.Background(), "demo-workout-app")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return FromFirestore(client), client
}

func TestFirestoreRoundTrip(t *testing.T) {
	s, client := emulatorStore(t)
	ctx := context.Background()
	uid := fmt.Sprintf("store-%d", time.Now().UnixNano())

	empty, err := s.Get(ctx, uid)
	if err != nil || len(empty.Machines)+len(empty.Notes) != 0 {
		t.Fatalf("Get of a new uid = %+v, %v, want an empty inventory", empty, err)
	}

	want, err := s.Update(ctx, uid, func(inv Inventory) (Inventory, error) {
		inv, err := inv.SaveMachine(catalog, legPress(lb(100)+5*domain.Tenth))
		if err != nil {
			return inv, err
		}
		if inv, err = inv.SaveMachine(catalog, dumbbells()); err != nil {
			return inv, err
		}
		if inv, err = inv.SaveMachine(catalog, Machine{Entry: domain.InventoryEntry{Machine: "treadmill"}}); err != nil {
			return inv, err
		}
		if inv, err = inv.ConfirmMachine(catalog, dumbbells().Entry); err != nil {
			return inv, err
		}
		inv, _, err = inv.SaveNote(catalog, "", "Hack squat", "n1")
		return inv, err
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, uid)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Get = %+v, %v, want %+v", got, err, want)
	}

	// The document sits at the path of D-197.
	snap, err := client.Doc("users/" + uid + "/inventory/active").Get(ctx)
	if err != nil || !snap.Exists() {
		t.Fatalf("the document of D-197: %v", err)
	}

	// A refused change writes nothing.
	refused := errors.New("refused")
	if _, err := s.Update(ctx, uid, func(inv Inventory) (Inventory, error) { return inv.RemoveMachine("leg_press"), refused }); !errors.Is(err, refused) {
		t.Fatalf("Update = %v, want the error of the change", err)
	}
	if got, _ := s.Get(ctx, uid); !reflect.DeepEqual(got, want) {
		t.Fatalf("a refused change changed the document: %+v", got)
	}
	if _, err := s.Get(ctx, "a/b"); !errors.Is(err, errUID) {
		t.Fatalf("Get of a uid with a slash = %v", err)
	}
}

// TestFirestoreConcurrentSaves proves that the transaction keeps each
// change: eight saves of eight machines at the same time give eight
// machines, each one time.
func TestFirestoreConcurrentSaves(t *testing.T) {
	s, _ := emulatorStore(t)
	ctx := context.Background()
	uid := fmt.Sprintf("race-%d", time.Now().UnixNano())
	ids := []domain.MachineID{"leg_press", "leg_extension", "seated_leg_curl", "lying_leg_curl", "calf_raise", "chest_press", "shoulder_press", "seated_row"}
	var wg sync.WaitGroup
	errs := make(chan error, len(ids))
	for _, id := range ids {
		wg.Go(func() {
			_, err := s.Update(ctx, uid, func(inv Inventory) (Inventory, error) {
				return inv.SaveMachine(catalog, Machine{Entry: domain.InventoryEntry{Machine: id, Weights: stack(10, 100, 10)}})
			})
			errs <- err
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Get(ctx, uid)
	if err != nil || len(got.Machines) != len(ids) {
		t.Fatalf("after %d saves: %d machines, %v", len(ids), len(got.Machines), err)
	}
	if err := got.Check(catalog); err != nil {
		t.Fatal(err)
	}
}

// TestFirestoreApplyOp: five calls at the same time with one op id apply
// one change. Each change adds a new note, so a second apply would show
// as a second note. A refused change stores no op id, and a replay
// changes nothing (D-272).
func TestFirestoreApplyOp(t *testing.T) {
	s, client := emulatorStore(t)
	ctx := context.Background()
	uid := fmt.Sprintf("op-%d", time.Now().UnixNano())
	op := Op{ID: "01920000-0000-7000-8000-000000000001", Entity: "note", EntityID: "n", At: time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)}
	var wg sync.WaitGroup
	replays := make(chan bool, 5)
	errs := make(chan error, 5)
	for k := range 5 {
		wg.Go(func() {
			replayed, err := s.ApplyOp(ctx, uid, op, func(inv Inventory) (Inventory, error) {
				out, _, err := inv.SaveNote(catalog, "", fmt.Sprintf("note %d", k), fmt.Sprintf("n%d", k))
				return out, err
			})
			replays <- replayed
			errs <- err
		})
	}
	wg.Wait()
	close(replays)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var fresh int
	for r := range replays {
		if !r {
			fresh++
		}
	}
	got, err := s.Get(ctx, uid)
	if err != nil || len(got.Notes) != 1 || fresh != 1 {
		t.Fatalf("after 5 calls with one op id: %d notes, %d applies, %v", len(got.Notes), fresh, err)
	}
	snap, err := client.Collection(UsersCollection).Doc(uid).Collection(OpsCollection).Doc(op.ID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var d opDoc
	if err := snap.DataTo(&d); err != nil || d.Entity != "note" || d.EntityID != "n" || !d.At.Equal(op.At) || d.Version != 0 {
		t.Fatalf("op document %+v, %v", d, err)
	}

	refused := Op{ID: "01920000-0000-7000-8000-000000000002", Entity: "machine", EntityID: "leg_press", At: op.At}
	confirm := func(inv Inventory) (Inventory, error) { return inv.ConfirmMachine(catalog, legPress(lb(100)).Entry) }
	if _, err := s.ApplyOp(ctx, uid, refused, confirm); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ApplyOp of a refused change = %v, want ErrNotFound", err)
	}
	if _, err := client.Collection(UsersCollection).Doc(uid).Collection(OpsCollection).Doc(refused.ID).Get(ctx); err == nil {
		t.Fatal("a refused change stored its op id")
	}
	if replayed, err := s.ApplyOp(ctx, uid, op, func(Inventory) (Inventory, error) { return Inventory{}, nil }); err != nil || !replayed {
		t.Fatalf("replay = %v, %v, want replayed", replayed, err)
	}
	if again, _ := s.Get(ctx, uid); len(again.Notes) != 1 {
		t.Fatalf("the replay changed the inventory: %+v", again)
	}
}
