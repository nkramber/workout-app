# Pull request 38 response

Date: 2026-10-04

This file answers the findings of round 1 of `docs/reviews/pr-38.md`, at `99e984dd3eb23a440b614091c31328c3c6953f8c`.

## P2-1: New plans lose exercise history beyond the 100-workout window

Result: full merit.

Evidence: the trigger reproduced. `Reviser.History` read the newest `MaxHistory` workouts alone, 100 at most. An exercise that those workouts omit got no history. So a new plan gave it the start of a new exercise at the estimate. It did not give the return of D-179 at 70 percent of the last load. That breaks D-301: an exercise with history gets its target from that history.

Correction: a new plan reads `MaxPlanHistory` workouts at most, 2000, in `go/internal/revise/revise.go`. At 4 sessions each week, that holds more than 9 years. A revision and `ForDate` keep the limit of 100, because they read the history of the exercises of one plan. Each workout is one document, and a new plan is rare, so the larger read stays small.

Regression check: `TestHistoryBeyondRevisionWindow` in `go/internal/revise/revise_test.go` logs the chest press in one old workout and 119 newer workouts with the seated row alone. `History` gives the old chest press outcome, and the policy gives `break.recalibrate` at 70 lb with the first set as the calibration. With the old limit of 100, the test fails. `go test ./...` passed.

## Round 2

This part answers the findings of round 2, at `d9bfc6e1abcbb2966fc4c8379e721759fdda4e6a`.

### P2-2: New plans still lose exercise history after 2,000 workouts

Result: full merit.

Evidence: the trigger reproduced. With 2100 newer workouts that omit the chest press, `History` gave no chest press outcome at the cap of 2000. D-301 sets no limit.

Correction: `History` reads each workout of the user, with the limit `AllHistory` of `finished` in `go/internal/revise/revise.go`. A revision and `ForDate` keep the limit of 100. Each workout is one document, and a new plan is rare, so the full read stays small.

Regression check: `TestHistoryBeyondRevisionWindow` now logs 2100 newer workouts that omit the chest press. `History` gives the old chest press outcome, and the policy gives the return at 70 lb. With the cap of 2000, the test fails. `go test ./...` passed.

### P2-3: The design gives a superseded starting load

Result: full merit.

Evidence: `docs/design.md` said that a plan starts each new exercise at 70 percent of its estimate (D-238). D-300 amends D-238.

Correction: the plan section of `docs/design.md` says that a plan starts each new exercise at its estimate, with the first set as the calibration (D-297, D-300). An exercise with history gets its target from that history (D-301).

Regression check: `make ste-check` and `make ref-check` passed.
