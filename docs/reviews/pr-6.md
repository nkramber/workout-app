# Pull request 6 review

Date: 2026-09-29

## Identity

- PR: 6
- Work area: 1.3, iPhone web platform spike
- Milestone: the iPhone probe, its device pages and checks, and the development project with Firebase Hosting and Firebase Authentication only
- Target: `main`
- Base: `abf6a0933bc471cd4827b87b9ff6bfa7227593e0`
- Merge base: `abf6a0933bc471cd4827b87b9ff6bfa7227593e0`
- Head: `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`
- Branch: `feat/pr-5-iphone-platform-probe`

## Provider gate

The hand-off session 6 names Claude Code as the author provider. The PR has two substantive commits, and the hand-off attributes both rounds to Claude Code. This session is Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone adds a React probe for five required iPhone platform items. It adds browser tests in WebKit and Chromium, a separate `make probe` target, and an optional CI job. The approved development project has Firebase Hosting and Firebase Authentication. The acceptance story requires the probe build and browser tests to pass. It also requires those two Firebase products alone.

The full PR changes 46 paths: the workflow, `AGENTS.md`, `Makefile`, `README.md`, five decision or review documents, and 36 probe paths. The earlier review inspected each path. This review inspected the complete change since `a1b6bff`, including the key configuration, project notes, tests, and the changes to `AGENTS.md`.

## Findings

### P2-1: AGENTS.md says no service exists while Firebase services are active

Status: open.

Open at: `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`.

File: `AGENTS.md:9`.

Trigger: a session reads the stage after creation of the `gym-route-dev` project.

Expected: the stage statement describes the project state consistently with D-99 and D-116.

Actual: it says no service exists, then states that Firebase Hosting and Firebase Authentication exist.

Consequence: the top-level project status gives conflicting instructions about whether services exist.

Correction: qualify the first statement as no workout app or product backend exists yet.

Regression check: read the stage statement with D-99 and D-116, then run `make ref-check`.

## Out of scope

None.

## PR comments

GitHub has no review threads or PR comments.

## Description edits

None.

## Verification

- `make verify` at `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`: passed. It includes STE, reference, lifecycle, context-budget, and all unit tests.
- `make pr-check` failed because this review worktree uses detached `HEAD`. The check requires a matching branch name in the session block.
- GitHub `verify:probe`, `verify:lint`, `verify:test`, and `pr-contract` at `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`: passed.
- GitHub `review-gate` at `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`: failed because the existing record named the earlier head. The updated record must pass after publication.
- Local `make probe` did not run. This machine has Node `v20.17.0`. The target requires Node 22, and GitHub `verify:probe` passed.
- The gcloud API key list command did not run. The shell reported expired credentials. Non-interactive mode blocks a prompt. The hand-off records the owner check from 2026-09-29.
- `git diff --check a1b6bff31a3eed46c917300c10ee97c5febd08fb..2c20c66229ccb21bde5eb29010b4af89fa74b5cf`: passed.
- Push: `4e2b3aa7e738e2913daa31bbce199573395e77fe` is the head of `origin/feat/pr-5-iphone-platform-probe`, verified with `gh pr view`.

## Open questions and accepted risks

- D-117: this session did not reread the external API-key and sign-up settings because gcloud credentials expired. The hand-off records the owner's settings check from 2026-09-29.

## Earlier verdicts

- **Ready for owner merge.** This verdict applied to head `a1b6bff31a3eed46c917300c10ee97c5febd08fb`.

## Verdict

**Changes required.** This verdict applies to head `2c20c66229ccb21bde5eb29010b4af89fa74b5cf`.
The project status in `AGENTS.md` contradicts the Firebase services that D-99 and D-116 say exist.
