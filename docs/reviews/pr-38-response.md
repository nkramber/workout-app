# Pull request 38 response

Date: 2026-10-04

This file answers the findings of round 1 of `docs/reviews/pr-38.md`, at `99e984dd3eb23a440b614091c31328c3c6953f8c`.

## P2-1: New plans lose exercise history beyond the 100-workout window

Result: full merit.

Evidence: the trigger reproduced. `Reviser.History` read the newest `MaxHistory` workouts alone, 100 at most. An exercise that those workouts omit got no history. So a new plan gave it the start of a new exercise at the estimate. It did not give the return of D-179 at 70 percent of the last load. That breaks D-301: an exercise with history gets its target from that history.

Correction: a new plan reads `MaxPlanHistory` workouts at most, 2000, in `go/internal/revise/revise.go`. At 4 sessions each week, that holds more than 9 years. A revision and `ForDate` keep the limit of 100, because they read the history of the exercises of one plan. Each workout is one document, and a new plan is rare, so the larger read stays small.

Regression check: `TestHistoryBeyondRevisionWindow` in `go/internal/revise/revise_test.go` logs the chest press in one old workout and 119 newer workouts with the seated row alone. `History` gives the old chest press outcome, and the policy gives `break.recalibrate` at 70 lb with the first set as the calibration. With the old limit of 100, the test fails. `go test ./...` passed.
