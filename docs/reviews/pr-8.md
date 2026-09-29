# Pull request 8 review

Date: 2026-09-29

## Identity

- PR: 8
- Work area: Phase 2 platform skeleton
- Milestone: the Phase 2 focused roadmap, owner answers, and the end of the D-4 period
- Target: `main`
- Base: `71b5220c6ca2f1c159023ceefc6675135144ebfe`
- Merge base: `71b5220c6ca2f1c159023ceefc6675135144ebfe`
- Head: `6d1db28c486b09000dbebb35b275863e6057682a`
- Branch: `docs/pr-7-phase-2-roadmap`

## Provider gate

The author provider is Claude Code, as stated in the Session 8 record of `docs/session-handoff.md`. The changed files are the work of that session, as the pull request body and hand-off state. The reviewer provider is Codex, from the active environment. The providers differ, so the provider gate passes.

## Intended behavior and scope

The PR adds a focused Phase 2 roadmap and records owner answers. It enables the document review override after the owner ended D-4. The roadmap gives each Phase 2 PR an id, branch, title, concerns, acceptance story, and checks. `make verify` must pass.

I inspected each changed path: `.claude/skills/one-pr-one-session/SKILL.md`, `.claude/skills/pr-review/SKILL.md`, `.claude/skills/pr-review/references/answer-review.md`, `.claude/skills/pr-review/references/review-record.md`, `AGENTS.md`, `README.md`, `docs/decisions.md`, `docs/questions.md`, `docs/research/decktome-patterns.md`, `docs/reviews/README.md`, `docs/roadmaps/README.md`, `docs/roadmaps/high-level-roadmap.md`, `docs/roadmaps/phase-2-platform-skeleton.md`, `docs/session-handoff.md`, `docs/tools/review_gate.py`, and `docs/tools/test_review_gate.py`.

## Findings

### P2-1: The public records disclose owner billing details

Status: fixed in `6d1db28c486b09000dbebb35b275863e6057682a`.

Open at: `f1a9321a02fbfa15017b82da69b9c93195260e56`.

File: `docs/questions.md:235`.

Trigger: A reader of the public repository reads the new Q-142 answer or the matching entry in `docs/session-handoff.md:21`.

Expected: The repository must not contain personal data of the owner. The exception in D-106 applies only to required test-image author credit.

Actual: The Q-142 row and hand-off disclose an owner-specific billing trial status and remaining credit and time. The details are unverified. The context identifies the owner and the information gives private financial data.

Consequence: The pull request publishes private owner billing information in a public repository.

Correction: The author removed the billing status, credit, and remaining-time details from both public records. Q-142 stays open until the account check.

Regression check: The search `rg -n "280 USD|70 days|free trial" docs/questions.md docs/session-handoff.md docs/roadmaps` returned no matches. The author reports `make ref-check`, `make ste-check`, and `make verify` passed at this head. I reran `make verify` at `6d1db28c486b09000dbebb35b275863e6057682a`. All checks passed.

## Earlier verdicts

- `Changes required` at `f1a9321a02fbfa15017b82da69b9c93195260e56` for P2-1.

## Out of scope

None.

## PR comments

No comments or review threads exist on PR #8.

## Description edits

None.

## Verification

- `make verify` at `6d1db28c486b09000dbebb35b275863e6057682a`: passed. STE, references, lifecycle, context budget, and all tests passed.
- `git diff --check 71b5220c6ca2f1c159023ceefc6675135144ebfe...HEAD`: passed.
- GitHub `pr-contract`, `verify:lint`, `verify:probe`, and `verify:test` at `6d1db28c486b09000dbebb35b275863e6057682a`: passed.
- GitHub `review-gate` at `6d1db28c486b09000dbebb35b275863e6057682a`: failed at RG 4 because this record still held `Changes required`.
- PR comments and review threads: none. I read one saved command output at `/tmp/pr8-comments.json` and `/tmp/pr8-review-comments.json`.
- Inspected all 18 changed paths from base `71b5220c6ca2f1c159023ceefc6675135144ebfe` to effective head `6d1db28c486b09000dbebb35b275863e6057682a`.
- `make ste-check` and `make ref-check`: pending for this record and hand-off update.
- Push: pending.

## Open questions and accepted risks

Q-142 remains open. Later work area 2.3 must verify the billing account and budget alert. This does not block the Phase 2 roadmap acceptance story.

## Verdict

**Ready for owner merge.** This verdict applies to head `6d1db28c486b09000dbebb35b275863e6057682a`.
The earlier finding is fixed. The acceptance story holds, and the required checks pass.
