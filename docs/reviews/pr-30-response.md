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
