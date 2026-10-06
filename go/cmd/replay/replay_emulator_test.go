//go:build emulator

// The acceptance story of work area 8.3 over the Firestore emulator.
// `make emulator-test` starts the emulators and runs this file with the
// tag emulator. The test uses its own database of the emulator, because
// the tests of the other packages write users to the default database at
// the same time.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/workout"
)

const (
	emulatorProject = "demo-workout-app"
	replayDatabase  = "replay-test"
)

// TestReplayEmulator stores the records of policy version 7 with their
// workouts, and replays them under the current version through the
// command. The diff report gives the count of each changed target and
// its rule id.
func TestReplayEmulator(t *testing.T) {
	host := os.Getenv(EnvEmulator)
	if host == "" || os.Getenv("GOOGLE_CLOUD_PROJECT") != emulatorProject {
		t.Fatal("run this test with `make emulator-test`, which sets FIRESTORE_EMULATOR_HOST and the demo project")
	}
	// Clear the database of this test, so a second run starts empty.
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/emulator/v1/projects/%s/databases/%s/documents", host, emulatorProject, replayDatabase), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	ctx := context.Background()
	client, err := firestore.NewClientWithDatabase(ctx, emulatorProject, replayDatabase)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	seedV7(t, stores{plans: plan.FromFirestore(client), workouts: workout.FromFirestore(client), inventory: inventory.FromFirestore(client)})

	rep, err := replayAll(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	checkV7Report(t, rep)
}
