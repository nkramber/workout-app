# Pull request 42 - author response

Date: 2026-10-06. Review round 1 recorded effective head `b43060c`, with the verdict "Changes required".

## P1-1: PR-42 contains the PR-41 milestone

**Result: no merit.**

The finding reads the GitHub number 42 as the roadmap id PR-42. In this repository, the roadmap id and the GitHub number are two different numbers (D-86). The roadmap id is in the title and the branch:

- The title is `feat: the policy replay and the incident runbook (PR-41)`, and the branch is `feat/pr-41-version-migration`. Both agree with the PR-41 section of `docs/roadmaps/phase-8-personal-use-operations.md`.
- The base `c712f56` has the subject "fix: the steady workout screen (PR-40) (#41)". So roadmap PR-40 merged as GitHub pull request 41.
- The hand-off records "GitHub PR 41 merged as `c712f56`" for PR-40, and "GitHub PR 40 merged as `5d5df59`" for PR-39 in the earlier record.

So GitHub pull request 42 holds roadmap PR-41, and its milestone is the milestone that D-311 gives PR-41. The owner approved it in this session before the first edit (D-12). PR-42, the four-week check, stays a later pull request.

**Correction.** None.

## P2-1: Replay skips validation of override copies

**Result: full merit.**

The trigger reproduces at `b43060c`. `replayCopy` returned before any check of an override copy, so an override outside its bounds counted as rebuilt with no violation.

**Correction.**

- `go/internal/revise/replay.go`: an override copy now gets `policy.CheckOverride` (D-69, D-293). The recommendation is the copy with the working sets that the workout keeps in its override record. The report counts each violation under its rule, and adds the copy to `outside_bounds`.
- `docs/operations.md` section 3 names the check of an override copy.

**Regression checks.**

- `TestReplayOverrides` of `go/cmd/replay/main_test.go` logs a valid override of 6 reps and an override of 25 reps. The report gives 1 copy outside the bounds under `override.bounds`. On `b43060c`, the same test gives 0 copies outside the bounds.
- `make go-test`, `make emulator-test`, and `make verify` pass.
