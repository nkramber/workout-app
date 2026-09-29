# Pull request 8 review

Date: 2026-09-29

## Identity

- PR: 8
- Work area: Phase 2 platform skeleton
- Milestone: the Phase 2 focused roadmap, owner answers, and the end of the D-4 period
- Target: `main`
- Base: `71b5220c6ca2f1c159023ceefc6675135144ebfe`
- Merge base: `71b5220c6ca2f1c159023ceefc6675135144ebfe`
- Head: `f1a9321a02fbfa15017b82da69b9c93195260e56`
- Branch: `docs/pr-7-phase-2-roadmap`

## Provider gate

The author provider is Claude Code, as stated in the Session 8 record of `docs/session-handoff.md`. The changed files are the work of that session, as the pull request body and hand-off state. The reviewer provider is Codex, from the active environment. The providers differ, so the provider gate passes.

## Intended behavior and scope

The PR adds a focused Phase 2 roadmap and records owner answers. It enables the document review override after the owner ended D-4. The roadmap gives each Phase 2 PR an id, branch, title, concerns, acceptance story, and checks. `make verify` must pass.

I inspected each changed path: `.claude/skills/one-pr-one-session/SKILL.md`, `.claude/skills/pr-review/SKILL.md`, `.claude/skills/pr-review/references/answer-review.md`, `.claude/skills/pr-review/references/review-record.md`, `AGENTS.md`, `README.md`, `docs/decisions.md`, `docs/questions.md`, `docs/research/decktome-patterns.md`, `docs/reviews/README.md`, `docs/roadmaps/README.md`, `docs/roadmaps/high-level-roadmap.md`, `docs/roadmaps/phase-2-platform-skeleton.md`, `docs/session-handoff.md`, `docs/tools/review_gate.py`, and `docs/tools/test_review_gate.py`.

## Findings

### P2-1: The public records disclose owner billing details

Status: open.

Open at: `f1a9321a02fbfa15017b82da69b9c93195260e56`.

File: `docs/questions.md:235`.

Trigger: A reader of the public repository reads the new Q-142 answer or the matching entry in `docs/session-handoff.md:21`.

Expected: The repository must not contain personal data of the owner. The exception in D-106 applies only to required test-image author credit.

Actual: The Q-142 row and hand-off disclose an owner-specific billing trial status and remaining credit and time. The details are unverified. The context identifies the owner and the information gives private financial data.

Consequence: The pull request publishes private owner billing information in a public repository.

Correction: Remove the billing status, credit, and remaining-time details from both public records. Keep Q-142 open and state only that the account check remains pending before someone asks the owner.

Regression check: After correction, search both files for the account status and amounts. The private details must be absent. Run `make ref-check` and `make ste-check`.

## Out of scope

None.

## PR comments

No comments or review threads exist on PR #8.

## Description edits

None.

## Verification

- `make verify` at `f1a9321a02fbfa15017b82da69b9c93195260e56`: passed. STE, references, lifecycle, context budget, and all 302 tests passed.
- GitHub `pr-contract` at `f1a9321a02fbfa15017b82da69b9c93195260e56`: passed.
- GitHub `verify:lint` at `f1a9321a02fbfa15017b82da69b9c93195260e56`: passed.
- GitHub `verify:probe` at `f1a9321a02fbfa15017b82da69b9c93195260e56`: passed.
- GitHub `verify:test` at `f1a9321a02fbfa15017b82da69b9c93195260e56`: passed.
- GitHub `review-gate` at published head `853376fe2200f7ed8ac0da214e6fabc61d85c8b6`: failed at RG 4. It read this record and rejected the `Changes required` verdict. RG 5 passed for effective head `f1a9321a02fbfa15017b82da69b9c93195260e56`. The gate must remain red until the author fixes the finding and gets a new review.
- `git diff --stat $(git merge-base origin/main HEAD)..HEAD`: inspected. The intended behavior and scope section lists all 16 changed paths.
- `make ste-check`: passed with 0 findings for this record and handoff update.
- `make ref-check`: passed with 0 findings for this record and handoff update.
- Push: `853376fe2200f7ed8ac0da214e6fabc61d85c8b6` was the head of `origin/docs/pr-7-phase-2-roadmap`, verified with `gh pr view` after the first publication.

## Open questions and accepted risks

Q-142 remains open. Later work area 2.3 must verify the billing account and budget alert. This does not block the Phase 2 roadmap acceptance story.

## Verdict

**Changes required.** This verdict applies to head `f1a9321a02fbfa15017b82da69b9c93195260e56`.
The public records disclose private owner billing information. Remove those details before merge. The owner can answer Q-142 after the account check.
