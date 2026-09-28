# Pull request 1 review

Date: 2026-09-28

## Identity

- PR: 1
- Work area: 0.1 Foundation documents and process tooling
- Milestone: Foundation registers, research and design, roadmap, and process gates
- Target: `main`
- Base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Merge base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Head: `d0be0f3d98ce5b6d286847f224c59e73755ead96`
- Branch: `docs/foundation-roadmap`

## Provider gate

The hand-off names Claude Code (Anthropic) as author. The PR has one commit and names no other substantive provider. This session uses Codex (OpenAI). The providers differ, so this gate passes. The hand-off gives the source. No owner statement conflicts with it.

## Intended behavior and scope

The milestone adds the Phase 0 foundation and process tooling. Its acceptance story says a fresh session can find decisions, questions, and phases, and the ported gates pass. D-4 requires Codex review during this roadmap period. D-8 requires the author session to start that review. D-13 reserves merge confirmation to the owner. D-15 requires cross-provider review for code or safety changes, while D-4 also applies to document changes.

I inspected these bootstrap files as code: `.claude/skills/one-pr-one-session/SKILL.md`, `.claude/skills/pr-review/SKILL.md`, its five named references, `.claude/skills/ste-writing/SKILL.md`, `AGENTS.md`, and `CLAUDE.md`. No instruction in these files weakens the provider gate or applies their rules to this review. I traced the gate implementation and ruleset. The diff statistics list the other foundation documents and tools. I did not read `docs/design.md`, the research documents, or the roadmap sections after Phase 2 line by line.

## Findings

### P1-1: The required review record has no verifiable reviewer identity

Status: open.

Open at: `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

File: `docs/tools/review_gate.py:190-202`.

Trigger: A PR author adds `docs/reviews/pr-<n>.md` with a current effective-head hash and the text `**Ready for owner merge.**`.

Expected: D-4 and D-8 require an actual Codex review of each PR during the roadmap period. The required check must reject a record that the PR author can create without that review.

Actual: `evaluate()` checks that the file exists, the verdict text says approved, and the recorded hash matches. I called `evaluate()` with a fabricated record and an author-owned commit. RG 3, RG 4, and RG 5 all passed. `.github/rulesets/review-gate.json` sets `required_approving_review_count` to 0. No independent GitHub approval closes this gap.

Consequence: A PR author can satisfy the required review gate without running `make codex-review` or receiving a cross-provider review. This defeats the review gate and provider gate for future PRs.

Correction: Make the gate verify an approval that the PR author cannot forge. Use an authenticated GitHub review by the required provider, or a trusted signed review result. Keep the effective-head check. Add a regression test for an author-created approval record.

Regression check: The inline Python reproduction against `docs/tools/review_gate.py` at `d0be0f3` returned `RG 3 PASS`, `RG 4 PASS`, and `RG 5 PASS` for the fabricated record. `make verify` passed, but its gate tests do not reject this record.

## Out of scope

None.

## PR comments

No general comments or review threads exist. `gh pr view --comments` returned an empty file, and the GraphQL review-thread query returned zero nodes.

## Description edits

None.

## Verification

- `git fetch origin`: passed. `origin/main` is `613c3cb08733338b9ac73b33e19d2380220b41da`.
- `git status --short --branch`: tree was clean at review start. The tool detached HEAD as instructed.
- `git diff --stat 613c3cb08733338b9ac73b33e19d2380220b41da..HEAD`: 52 paths, 8,313 insertions.
- `python3 docs/tools/review_gate.py --effective-head 1`: `d0be0f3d98ce5b6d286847f224c59e73755ead96`.
- Inline `review_gate.evaluate()` forged-record reproduction at `d0be0f3`: RG 3, RG 4, and RG 5 passed.
- `make verify` at `d0be0f3`: passed. STE and reference checks reported 0 findings. Lifecycle check reported 0 errors. The test command passed 276 unit tests.
- `gh pr checks 1 --repo nkramber/workout-app`: `pr-contract`, `verify:lint`, and `verify:test` passed. An earlier `pr-contract` run failed on a body text mismatch. A later run passed. `review-gate` did not run because `main` does not contain the workflow.
- Live ruleset check: not run. Q-94 is open, and the hand-off says the live ruleset does not exist.
- Paid target: not run.
- Push: <pending>.

## Open questions and accepted risks

- Q-94 remains open. The live `main` ruleset does not exist, so GitHub does not yet enforce the committed required checks.

## Verdict

**Changes required.** This verdict applies to head `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

The review gate accepts an author-created approval record, which defeats the cross-provider requirement. Correct the gate and add a regression test before merge.
