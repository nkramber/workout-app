# Pull request 6 review

Date: 2026-09-29

## Identity

- PR: 6
- Work area: 1.3, iPhone web platform spike
- Milestone: the iPhone probe, its device pages and checks, and the development project with Firebase Hosting and Firebase Authentication only
- Target: `main`
- Base: `abf6a0933bc471cd4827b87b9ff6bfa7227593e0`
- Merge base: `abf6a0933bc471cd4827b87b9ff6bfa7227593e0`
- Head: `50094a139a1e8a461fabc2d6578de1e5020ea14e`
- Branch: `feat/pr-5-iphone-platform-probe`

## Provider gate

The hand-off session 6 names Claude Code as the author provider. The PR has two substantive commits, and the hand-off attributes both rounds to Claude Code. This session is Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone adds a React probe for five required iPhone platform items. It adds browser tests in WebKit and Chromium, a separate `make probe` target, and an optional CI job. The approved development project has Firebase Hosting and Firebase Authentication. The acceptance story requires the probe build and browser tests to pass. It also requires those two Firebase products alone.

The full PR changes 47 paths: the workflow, `AGENTS.md`, `Makefile`, `README.md`, six decision or review documents, and 36 probe paths. The earlier reviews inspected the initial change. This review inspected the correction to `AGENTS.md` and `README.md`, the response, the updated hand-off, and the full earlier review evidence.

## Findings

### P2-1: AGENTS.md says no service exists while Firebase services are active

Status: fixed in `50094a139a1e8a461fabc2d6578de1e5020ea14e`.

Open at: `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`.

File: `AGENTS.md:9`.

Trigger: a session reads the stage after creation of the `gym-route-dev` project.

Expected: the stage statement describes the project state consistently with D-99 and D-116.

Actual: at `2c20c66`, it said no service exists, then stated that Firebase Hosting and Firebase Authentication exist. The current text says that the workout app and backend do not exist, and that the Firebase services exist.

Consequence: the top-level project status gives conflicting instructions about whether services exist.

Correction: the author qualified the first statement as no workout app or backend exists yet. The same correction fixed the status line in `README.md`.

Regression check: the changed statements agree with D-99 and D-116. `make verify` passed, including `ref-check`.

## Out of scope

None.

## PR comments

GitHub has no review threads or PR comments.

## Description edits

None.

## Verification

- `make verify` at `50094a139a1e8a461fabc2d6578de1e5020ea14e`: passed. It includes STE, reference, lifecycle, context-budget, and all unit tests (302 core tests, 3 expected skips across spike tests).
- GitHub `verify:probe`, `verify:lint`, `verify:test`, and `pr-contract` at `50094a139a1e8a461fabc2d6578de1e5020ea14e`: passed.
- GitHub `review-gate` at `50094a139a1e8a461fabc2d6578de1e5020ea14e`: failed because the published record named `2c20c66` and retained `Changes required`. The failure log confirms RG 4 and RG 5 only.
- `make probe` did not run locally because this machine has Node `v20.17.0`. CI passed the build and browser tests.
- This session did not reread the external Firebase key settings because gcloud credentials expired. The hand-off records the owner's check from 2026-09-29.
- `git diff --check abf6a0933bc471cd4827b87b9ff6bfa7227593e0..50094a139a1e8a461fabc2d6578de1e5020ea14e`: passed.
- Push: pending.

## Open questions and accepted risks

- D-117: this session did not reread the external API-key and sign-up settings because gcloud credentials expired. The hand-off records the owner's settings check from 2026-09-29. This review accepts that dated owner evidence.

## Earlier verdicts

- **Ready for owner merge.** This verdict applied to head `a1b6bff31a3eed46c917300c10ee97c5febd08fb`.
- **Changes required.** This verdict applied to head `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`.

## Verdict

**Ready for owner merge.** This verdict applies to head `50094a139a1e8a461fabc2d6578de1e5020ea14e`.
The earlier finding is fixed, and the required build, browser, lint, test, and contract checks pass. The recorded D-117 risk relies on the owner's dated settings check.
