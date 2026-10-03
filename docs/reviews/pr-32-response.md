# Pull request 32 - author response

Date: 2026-10-03. Review round 1 recorded effective head `deb04a7`, with the verdict "Changes required".

## P2-1: A changed calibration load bypasses the calibration table

**Result: full merit.** The trigger reproduces. At `deb04a7`, `calibrationLoad` of `web/src/lib/workout.ts` gives null when the logged weight is not the load of the calibration target. So after a tap of "Heavier" and 6+ reps in reserve, the working sets keep the load of the plan. The weight buttons permit a change before the log (D-249), and the table of D-150 reads the set that the owner did. The policy had the same gap: `effective` of `go/internal/policy/calibrate.go` applied the table to the target load, and not to the logged weight.

**Correction.** The policy applies the table to the weight that the owner logged, in `effective` and on the phone (D-267):

- `policy.CalibrationTable` gives one row for each weight of the machine: the weight, and the load after each result of the table. A weight that the policy can not give takes the repair of `RuleLoadRepair` first.
- The plan stores the rows, and the field `calibration_loads` of `PlannedExercise` in `proto/workoutapp/v1/plan_service.proto` is now a list with the weight in each row. The field is not on `main` yet, so `buf breaking` passes.
- `calibrationLoad` reads the row of the logged weight. A plan of policy version 3 gives the load of the plan. A weight that the machine did not have at the time of the plan does too.
- `effective` reads the weight of the first logged calibration set. So the next plan reads the working sets against the load that the phone showed.

**Regression checks.**

- The unit test "gives each working set the load of the calibration table" of `web/src/lib/workout.test.ts`. A set at 300 with 4 reps in reserve gives 300, and a set at 100 with 2 gives 100. Both gave the plan load of 200 at `deb04a7`. With the old lookup put back, the test fails.
- The browser test "a skip and finish now give the correct session log" of `web/e2e/workout.spec.ts`: "Heavier" to 20 lb, then 6+ reps in reserve, gives working sets at 30 lb. It passes in WebKit and Chromium.
- The Go test `TestEffectiveFirstCalibrationSet`. A calibration set at 60 lb with 5 reps in reserve gives 65 lb, not 55 lb from the target. With the target load of `deb04a7` put back, the test fails.
- `TestCalibrationTable` and `TestPropertyCalibrationTable` check each row of the table.
- `make go-test`, `make emulator-test`, `make contract`, and `make web` pass.

## P2-2: Finishing clears a skip after the exercise has a logged set

**Result: no merit.** The finding asks the header to keep a skip on an exercise with a logged set, and to carry that state to the policy. Three contracts of `main` refuse that state, and D-170 gives the same result for both states.

- The contract of `WorkoutHeader` in `proto/workoutapp/v1/workout_service.proto`, from PR-29, says: "An exercise with a logged set is not skipped, even when this list names it."
- The domain model refuses the state. `ExerciseLog.Check` of `go/internal/domain/log.go` refuses "skipped with N sets", and `Input.check` of `go/internal/policy/progress.go` refuses the same history.
- D-170 says: "A planned set with no log is skipped work, not missed reps." The policy reads such an exercise with `RuleIncomplete`, "You logged 1 of 3 sets. The target stays the same." A skipped exercise gets `RuleSkipped`, and its target stays the same too. So the next target does not change with the extra state.

The finding also reads D-170 too broadly. D-170 names no difference between a skip and the planned sets with no log of a logged exercise. A new state of the domain model, the contract, and the policy needs an owner decision, and this milestone does not hold one. `finishWorkout` keeps each skip of an exercise with no logged set, and the browser test "a skip and finish now give the correct session log" shows it.

No file changes for P2-2.
