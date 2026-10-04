# Pull request 37 response

Date: 2026-10-04

This file answers the findings of round 1 of `docs/reviews/pr-37.md`, at `35cf4efd195fa2c5290c1886acff4fda1eef93a4`.

## P2-1: The deload lasts eight calendar dates

Result: partial merit.

Evidence: the trigger reproduced. `deloadOf` gave the deload targets from the date of the start to 7 dates after it, so the targets had 8 dates. A session counted as a deload session only after the date of the start. So the two ranges did not agree. A second session on the date of the start then got the deload targets, but the rules read it as evidence.

The correction of the review ends the deload 6 dates after the start. The session that started the deload is a normal session. So that correction gives the deload targets to 6 dates of sessions alone. It also keeps the disagreement on the date of the start. D-295 says that the decline starts a deload of 7 days. So the deload covers the 7 dates after the session that started it, for the targets and for the sessions.

Correction: `deloadOf` in `go/internal/policy/disrupt.go` gives one range, from 1 to 7 dates after the start, for each use (D-295). The rule text of `deload.reactive` and `docs/design.md` say "the 7 dates after that session". The revision of the workout that starts a deload gives the target of its own date. `GetPlan` with a date gives the deload from the next date. `TestDisruptionAcceptanceStory` reads the plan on the first date of the deload.

Regression check: `TestScenarioDeload` reads each date from the start to 8 dates after it, and only dates 1 to 7 give the deload. It also proves that a session on the date of the start is evidence, and a session in the deload is not. The golden file `h_deload_trigger_day` holds the target of the date of the start. `go test ./internal/policy` passed, and the old code fails the date loop on the date of the start.

## P2-2: A dated plan read keeps a stale override

Result: full merit.

Evidence: the trigger reproduced. `ForDate` copied the override with no change, and the phone used its sets after a missed week or a break. So the workout did not get the hold of D-294 or the table of D-179. No check of the policy covered the override on the new date (D-23).

Correction: an override keeps the local date of its save. `ForDate` reads the date rules of the save date and of the date of the read: `missed.hold`, the three `break` rules, and `deload.reactive`. When they differ, the dated plan marks the override as expired, and the stored override does not change. The contract gives `TargetOverride.expired`.

The phone uses the recommendation for an expired override, and the plan screen tells the owner that the change no longer applies. The owner can override the new recommendation again, and the policy checks it on the new date. The correction does not change the sets of the override. A rule of the date can apply two times to the sets of the owner. That occurs when the owner saved the override on a date of the same rule.

Regression check: `TestForDateOverride` in `go/internal/revise` keeps the override on the dates of the same rules. It marks the override as expired after a missed week and after a break. An override that the owner saved after the missed week stays. `TestOverrideExpired` in `go/internal/plansvc` proves the field of the contract, and the stored override. The unit test "uses the recommendation for an expired override" in `web/src/lib/workout.test.ts` proves the sets of the workout. `go test ./...` and the web unit tests passed.
