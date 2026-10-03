# Pull request 30 - author response

Date: 2026-10-03. Review round 1 recorded effective head `f75610a`, with the verdict "Changes required".

## P2-1: A missing plan link panics during request parsing

**Result: no merit.** The trigger does not reproduce.

The finding says that `h.GetPlan().GetSessionIndex()` panics when the `plan` field is absent. The getters of the generated code accept a nil receiver. In `go/gen/workoutapp/v1/workout_service.pb.go`, `(*PlanLink).GetSessionIndex` reads the field only when the receiver is not nil, and else gives 0. So the call gives 0, and no panic occurs.

The code then gives `Entry.Check` a plan time of zero. The check refuses it with "plan link: want the RFC 3339 time of the plan", and the service gives the code `invalid_argument` (D-248).

**Evidence.**

- `TestSyncOutboxRefuses` of `go/internal/workoutsvc/workoutsvc_test.go` sent this payload at `f75610a`: the entry `noPlan` sets `Plan` to nil. The test expects `STATUS_REFUSED` with `invalid_argument`, and no stored workout. It passed at `f75610a`, in `make go-test` and in the CI check `verify:go`.
- `go test ./internal/workoutsvc/ -run TestSyncOutboxRefuses -v -count=1` passed again for this answer.

**Change.** No change of the product code. The test now also asserts that the refusal of `noPlan` names the plan link. So the test proves that the entry reaches the check of the plan link, and not another refusal.

**Regression checks.**

- `make go-test` passes.
- `make verify` passes.

## Round 2

Review round 2 recorded effective head `a20f04b`, with the verdict "Changes required". It withdrew P2-1.

### P2-2: Entity versions collide when distinct entity IDs match

**Result: full merit.**

The trigger reproduces at `a20f04b`. `Apply` keyed each version by the entity id alone. The phone makes each id as a UUIDv7, so a collision needs a fault of the phone. But the contract of `proto/workoutapp/v1/workout_service.proto` does not make an id unique across the entities. So a header and a set with the same id shared one version, and the set got version 2.

**Correction.**

- `go/internal/workout/workout.go`: the new function `VersionKey` keys each version by the entity and its id. `Apply` reads and writes the version with it.
- `go/internal/workout/firestore_emulator_test.go`: `TestFirestoreOneApply` reads the version of the set with `VersionKey`.

No stored data changes, because no build of this code ran on the live service.

**Regression checks.**

- `TestVersionKey` of `go/internal/workout/workout_test.go` applies a workout, a set, and a cardio log with the same id. The first apply of each gives version 1, and a second header gives version 2. The test does not build on the code of `a20f04b`, because that code has no `VersionKey`. The reproduction of the reviewer gives the old result: version 2 for the set.
- `make go-test`, `make emulator-test`, and `make verify` pass.
