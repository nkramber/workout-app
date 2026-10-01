# Pull request 16 - author response

Date: 2026-09-30. Review round 1 recorded effective head `f4ebf8f`, with the verdict "Changes required".

## P2-1: Check accepts calibration sets that violate the target

**Result: full merit.**

The trigger reproduces at `f4ebf8f`. D-150 and D-177 give one calibration set at the target reps. `Check` refused only a proposal with no calibration set in a calibration session. Three new cases of `TestCheck` gave no violation on the old code:

- two calibration sets,
- a calibration set at 12 reps with working sets of 10,
- a calibration set at other reps after a history.

**Correction.**

- `go/internal/policy/bounds.go`: `Check` refuses a proposal with more than one calibration set. It also refuses a calibration set with reps that are not the reps of the first working set of the proposal. Both give a violation of `RuleCalibrationSet`. The load ceiling of the calibration set stays.
- `go/internal/policy/rules.go`: the text of `RuleCalibrationSet` states the two limits.
- `go/internal/policy/property_test.go`: the oracle of `TestPropertyRefusal` and `TestPropertyFallback` holds the two limits. The generator now adds a second calibration set and changes the reps of a calibration set.

Each target of `Next` already holds one calibration set at the reps of the first working set. So the targets of the policy do not change.

**Regression checks.**

- `go test ./internal/policy -run TestCheck` on the old code: the three new cases failed with no violation. After the correction, each case gives `RuleCalibrationSet`, and a calibration set at the target reps gives no violation.
- `make go-test` and `make verify` pass.

## P2-2: Check accepts an unexpected calibration set

Review round 2 recorded effective head `e42982b`, with the verdict "Changes required".

**Result: full merit.**

The trigger reproduces at `e42982b`. In a session that is not a calibration session, `Check` accepted a calibration set and applied no load ceiling to it. The case "history heavy calibration set" of `TestCheck` gave no violation for a calibration set at 30 lb, when the policy target holds 25 lb. D-177 gives calibration sets to the calibration sessions alone. The correction of P2-1 kept this gap on purpose, and that choice was wrong.

**Correction.**

- `go/internal/policy/bounds.go`: `Check` refuses a calibration set when the target of the policy has none, with a violation of `RuleCalibrationSet`.
- `go/internal/policy/rules.go`: the text of `RuleCalibrationSet` states that a proposal for another session has no calibration set.
- `go/internal/policy/bounds_test.go`: the cases "history calibration set" and "history heavy calibration set" expect `RuleCalibrationSet`. Two older cases expect the new violation too.
- `go/internal/policy/property_test.go`: the oracle refuses a proposal when the presence of its calibration set differs from the target of the policy.

**Regression checks.**

- `go test ./internal/policy -run TestCheck` on the old code: the two new cases failed with no violation. After the correction, each gives `RuleCalibrationSet`.
- `make go-test` and `make verify` pass.
