# Pull request 7 review

Date: 2026-09-29

## Identity

- PR: 7
- Work area: 1.3 iPhone web platform spike
- Milestone: Report the iPhone device checklist with a result for each item and a startup time.
- Target: `main`
- Base: `d6e3c5473ab6222c25412d303892f0277b579cfb`
- Merge base: `d6e3c5473ab6222c25412d303892f0277b579cfb`
- Head: `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659`
- Branch: `docs/pr-6-iphone-platform-report`

## Provider gate

The hand-off session record names Claude Code as the author provider. The PR has one substantive commit, and its changed paths and commit body match that session. This review runs as Codex. The providers differ, so the gate passes.

## Intended behavior and scope

The milestone reports each iPhone checklist result and a React startup time. Its acceptance story requires results for each item and a startup time on the owner's iPhone. D-118 also requires five cold starts in standalone mode, with a median first contentful paint no greater than 2500 ms. The report, decisions, questions, roadmap, README, stage line, and hand-off all fit that scope.

Inspected every changed path: `AGENTS.md`, `README.md`, `docs/decisions.md`, `docs/questions.md`, `docs/research/iphone-platform-spike.md`, `docs/research/platform-cloud-and-ai.md`, `docs/roadmaps/phase-1-risk-spikes.md`, and `docs/session-handoff.md`. The documents agree on the three Phase 1 reports and the PR-6 result. I reviewed the `AGENTS.md` change as content. It does not change the rules of this review.

## Findings

### P2-1: The startup result does not establish five cold starts

Status: open.

Open at: `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659`.

File: `docs/research/iphone-platform-spike.md:112-130`.

Trigger: Treat five page loads 11 seconds apart as the five cold starts required by D-118.

Expected: The go decision uses five demonstrated cold starts in standalone mode, as D-118 requires.

Actual: The report records five launches with navigation type `navigate`, but does not show that iOS terminated the app or web view between launches. The probe assigns a new random launch id per page load (`tools/spikes/iphone_probe/src/lib/launch.ts:1-4`), and records navigation timing (`tools/spikes/iphone_probe/src/lib/startup.ts:45-60`). Neither proves a cold process start. The report also says the five launches took 11 seconds and that iOS can keep parts of the web view in memory (`docs/research/iphone-platform-spike.md:142-144`).

Consequence: The 33 ms median does not yet prove the startup part of the D-118 go bar. The report's overall go claim is unsupported by this measurement.

Correction: Repeat five measurements. Document how each procedure starts the app from cold. Or ask the owner to change the bar, then update the decision register and report.

Regression check: Confirm each sample follows the documented cold-start procedure and runs in standalone mode. Confirm the report calculates the median from these samples. Not run. No new device evidence exists at this head.

## Out of scope

None.

## PR comments

No PR comments, review threads, or reviews exist.

## Description edits

None.

## Verification

- `git diff --stat d6e3c5473ab6222c25412d303892f0277b579cfb..61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659` — inspected all eight changed paths.
- `make verify` at `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659` — passed. STE, reference, lifecycle, context-budget, and unit tests passed. Counts: 302 tool tests, 12 iPhone probe tests, 44 Luna plan tests, 19 recognition tests (1 skipped), and 45 recognition-set tests (2 skipped).
- GitHub checks at `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659` — `pr-contract`, `verify:lint`, `verify:probe`, and `verify:test` passed. `review-gate` failed because the record was absent before this review.
- Device cold-start reproduction — not run. This environment has no access to the owner's iPhone or its launch evidence.
- Push: pending.

## Open questions and accepted risks

None.

## Verdict

**Changes required.** This verdict applies to head `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659`.
The device report does not establish that its five startup samples meet D-118's cold-start condition. The owner must receive revised evidence or change the bar before the report can claim a go.
