//go:build emulator

// The Firestore store over the emulator. `make emulator-test` starts the
// emulators and runs this file with the tag emulator.
package profile

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
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
	uid := fmt.Sprintf("profile-%d", time.Now().UnixNano())

	if _, ok, err := s.Get(ctx, uid); ok || err != nil {
		t.Fatalf("Get of a new uid = %v, %v, want no profile", ok, err)
	}
	want := valid()
	if err := s.Save(ctx, uid, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.Get(ctx, uid)
	if err != nil || !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("Get = %+v, %v, %v, want %+v", got, ok, err, want)
	}

	// A second save replaces the whole profile. Empty lists stay empty
	// lists.
	second := valid()
	second.Experience, second.InjuredAreas, second.Cardio, second.InjuryText = Advanced, nil, nil, ""
	if err := s.Save(ctx, uid, second); err != nil {
		t.Fatal(err)
	}
	got, _, err = s.Get(ctx, uid)
	if err != nil || got.Experience != Advanced || len(got.InjuredAreas)+len(got.Cardio) != 0 || got.InjuryText != "" {
		t.Fatalf("after the second save: %+v, %v", got, err)
	}

	// The stored document is at users/{uid}/profile/active, and it holds
	// the fields of the profile alone.
	snap, err := client.Collection("users").Doc(uid).Collection("profile").Doc("active").Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for k := range snap.Data() {
		keys = append(keys, k)
	}
	wantKeys := []string{
		"age_years", "cardio_exercises", "experience", "free_text", "goal_template", "height_in",
		"injured_areas", "injury_text", "muscle_groups", "training_days", "weight_lb",
	}
	if !sameSet(keys, wantKeys) {
		t.Fatalf("stored fields = %v, want %v", keys, wantKeys)
	}
	if areas, _ := snap.Data()["injured_areas"].([]any); areas == nil {
		t.Fatalf("injured_areas = %#v, want an empty array", snap.Data()["injured_areas"])
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]bool{}
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		if !m[v] {
			return false
		}
	}
	return true
}
