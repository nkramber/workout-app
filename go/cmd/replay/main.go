// Command replay is the policy replay of work area 8.3. It reads the
// finished workouts, their target copies (D-291), the inventory, and
// the active plan of each user, and replays each decision record and
// each target copy under the current policy version (D-176). It writes
// a JSON diff report with counts and rule ids alone (D-80).
//
// The replay calls no model, so it costs nothing. It reads the store
// and writes nothing to it. A read of the live store needs the approval
// of the owner, so the command reads it only with the -live flag. With
// no -live flag, it needs FIRESTORE_EMULATOR_HOST. `docs/operations.md`
// gives the steps.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/inventory"
	"github.com/nkramber/workout-app/go/internal/plan"
	"github.com/nkramber/workout-app/go/internal/revise"
	"github.com/nkramber/workout-app/go/internal/workout"
)

// EnvEmulator is the environment variable that sends each call of the
// Firestore client to the emulator.
const EnvEmulator = "FIRESTORE_EMULATOR_HOST"

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "replay:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	project := fs.String("project", "", "the Google Cloud project of the store (required)")
	live := fs.Bool("live", false, "read the live store: the owner approves each live read")
	out := fs.String("out", "", "the path of the JSON report (default: standard output)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := checkTarget(*project, *live, getenv(EnvEmulator)); err != nil {
		return err
	}
	client, err := firestore.NewClient(ctx, *project)
	if err != nil {
		return fmt.Errorf("the Firestore client: %w", err)
	}
	defer func() { _ = client.Close() }()
	rep, err := replayAll(ctx, client)
	if err != nil {
		return err
	}
	return write(rep, *out, stdout)
}

// checkTarget refuses a run with no project, a live run over the
// emulator, and a run of the live store with no -live flag.
func checkTarget(project string, live bool, emulator string) error {
	switch {
	case project == "":
		return errors.New("-project is required")
	case live && emulator != "":
		return fmt.Errorf("-live with %s: unset one", EnvEmulator)
	case !live && emulator == "":
		return fmt.Errorf("no %s: a read of the live store needs -live and the approval of the owner", EnvEmulator)
	}
	return nil
}

// replayAll replays the data of each user of the store.
func replayAll(ctx context.Context, client *firestore.Client) (Report, error) {
	refs, err := client.Collection(plan.UsersCollection).DocumentRefs(ctx).GetAll()
	if err != nil {
		return Report{}, fmt.Errorf("the users: %w", err)
	}
	r := &revise.Reviser{Plans: plan.FromFirestore(client), Workouts: workout.FromFirestore(client), Inventory: inventory.FromFirestore(client)}
	rep := newReport()
	for _, ref := range refs {
		rp, err := r.Replay(ctx, ref.ID)
		if err != nil {
			// The error can hold the uid, an id that stays out of the report.
			return Report{}, fmt.Errorf("the replay of a user: %w", err)
		}
		rep.add(rp)
	}
	return rep, nil
}

func write(rep Report, path string, stdout io.Writer) error {
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if strings.TrimSpace(path) == "" {
		_, err = stdout.Write(b)
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
