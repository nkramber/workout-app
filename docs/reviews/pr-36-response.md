# Pull request 36 - author response

Date: 2026-10-04. Review round 1 recorded effective head `a545132`, with the verdict "Changes required".

## P2-1: A failed plan store operation loses the revision

**Result: full merit.**

The trigger reproduces at `a545132`. When `Reviser.Revise` gave a store error, `revise` of `go/internal/workoutsvc/workoutsvc.go` wrote a log line alone. `SyncOutbox` then gave an applied result for each entry, and the phone removed the entries. No later call ran the revision of that workout, so the plan kept its old targets. D-292 needs a revision when the sync applies a finished workout.

A failed call of Luna is no error of `Revise`, because the reasons of the rules then show. So only a store failure of the plan, the workouts, or the inventory gives the error.

**Correction.**

- `go/internal/workoutsvc/workoutsvc.go`: a store failure of a revision gives `UNAVAILABLE` with a fixed text and no path (D-80). Each entry of the batch stays applied. The phone keeps the entries, and the sync tries again (D-277). The replay applies nothing again, and its finished workout runs the revision again. The plan records each revised workout, so a revision that passed does not run two times.
- `proto/workoutapp/v1/workout_service.proto`: the comment of `SyncOutbox` gives this contract. `make proto` wrote the generated comments.
- `docs/design.md` and `go/README.md` give the same behavior.

**Regression checks.**

- `TestSyncOutboxRevises` of `go/internal/workoutsvc/workoutsvc_test.go` gives the reviser a store error. It asserts `UNAVAILABLE` with no path. Then it replays the batch, and asserts the applied result and one more revision. The old code gave no error, so the assertion fails on it.
- `make go-test`, `make contract`, `make emulator-test`, `make web`, and `make verify` pass.
