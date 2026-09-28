# Pull request 1 review

Date: 2026-09-28

## Identity

- PR: 1
- Work area: 0.1 Foundation documents and process tooling
- Milestone: Foundation registers, research and design, roadmap, and process gates
- Target: `main`
- Base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Merge base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Head: `973703fea92df0ef5cdb480038bbe845f11ae2e4`
- Branch: `docs/foundation-roadmap`

## Provider gate

The hand-off names Claude Code (Anthropic) as author. The commit history shows the correction round on top of the initial Claude-authored change. This session uses Codex (OpenAI). The providers differ, so this gate passes. The hand-off gives the source. No owner statement conflicts with it.

## Intended behavior and scope

The milestone adds the Phase 0 foundation and process tooling. Its acceptance story says a fresh session can find decisions, questions, and phases, and the ported gates pass. D-4 requires Codex review during this roadmap period. D-8 requires the author session to start that review. D-13 reserves merge confirmation to the owner. D-15 requires cross-provider review for code or safety changes, while D-4 also applies to document changes.

I inspected these bootstrap files as code: `.claude/skills/one-pr-one-session/SKILL.md`, `.claude/skills/pr-review/SKILL.md`, its five named references, `.claude/skills/ste-writing/SKILL.md`, `AGENTS.md`, and `CLAUDE.md`. The author loop and author answer file conflict with the cross-provider rule for a Codex-authored pull request. I traced `make codex-review`, `docs/tools/codex_review.py`, the review gate, its tests, the workflows, and the ruleset. I read the hand-off, the prior review and response, the cited decisions and questions, and the PR body. I did not inspect every line of the design and research documents.

## Findings

### P1-1: The required review record has no verifiable reviewer identity

Status: accepted risk, D-87.

Open at: `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

File: `docs/tools/review_gate.py:190-202`.

Trigger: A PR author adds `docs/reviews/pr-<n>.md` with a current effective-head hash and the text `**Ready for owner merge.**`.

Expected: D-4 and D-8 require an actual Codex review of each PR during the roadmap period. The required check must reject a record that the PR author can create without that review.

Actual: `evaluate()` checks that the file exists, the verdict text says approved, and the recorded hash matches. I called `evaluate()` with a fabricated record and an author-owned commit. RG 3, RG 4, and RG 5 all passed. `.github/rulesets/review-gate.json` sets `required_approving_review_count` to 0. No independent GitHub approval closes this gap.

Consequence: A PR author can satisfy the required review gate without running `make codex-review` or receiving a cross-provider review. This defeats the review gate and provider gate for future PRs.

Correction: Make the gate verify an approval that the PR author cannot forge. Use an authenticated GitHub review by the required provider, or a trusted signed review result. Keep the effective-head check. Add a regression test for an author-created approval record.

Regression check: The inline Python reproduction against `docs/tools/review_gate.py` at `d0be0f3` returned `RG 3 PASS`, `RG 4 PASS`, and `RG 5 PASS` for the fabricated record. `make verify` passed, but its gate tests do not reject this record.

### P1-2: The author loop can run a same-provider review

Status: open.

Open at: `973703fea92df0ef5cdb480038bbe845f11ae2e4`.

File: `.claude/skills/one-pr-one-session/SKILL.md:71-81`, `.claude/skills/pr-review/references/answer-review.md:7-14`.

Trigger: A Codex session authors a pull request, then follows the author loop after CI passes.

Expected: D-15 and the provider gate require Claude Code to review Codex-authored work. The author guidance must select a reviewer from the other provider and record the real author provider.

Actual: The author loop always directs the author to `make codex-review`, whose implementation invokes Codex. The answer guide also instructs every author to write `Author provider: Claude Code`.

Consequence: For a Codex-authored pull request, following these rules runs a same-provider review and writes a false author-provider claim. The required cross-provider gate fails.

Correction: Add the Claude Code review path for Codex-authored pull requests. Or limit the Codex target to Claude-authored pull requests, and define the matching author flow. Set the hand-off line to the real provider.

Regression check: Not run. The cited files prove the flow conflict. Add a test that blocks Codex review of Codex-authored work.

## Out of scope

None.

## PR comments

No general comments or review threads exist. `gh pr view --comments` returned an empty file, and the GraphQL review-thread query returned zero nodes.

## Description edits

None.

## Verification

- `git fetch origin`: passed. `origin/main` is `613c3cb08733338b9ac73b33e19d2380220b41da`.
- `git status --short --branch`: tree was clean at review start. The tool detached HEAD as instructed.
- `make where`: clean detached review worktree at `973703f`. Base `origin/main` is `613c3cb`.
- `git fetch origin`: passed. `origin/main` is `613c3cb08733338b9ac73b33e19d2380220b41da`.
- `git merge-base origin/main HEAD`: `613c3cb08733338b9ac73b33e19d2380220b41da`.
- `git diff --stat origin/main...HEAD`: 54 paths, 8,550 insertions.
- `python3 docs/tools/review_gate.py --effective-head 1`: `973703fea92df0ef5cdb480038bbe845f11ae2e4`.
- `make verify` at `973703f`: passed. STE, reference, and lifecycle checks reported 0 findings/errors. The test command passed 287 tests.
- `gh pr checks 1 --repo nkramber/workout-app`: `pr-contract`, `verify:lint`, and `verify:test` passed at `973703f`. `review-gate` did not run because `main` does not contain the workflow.
- `gh api graphql` review-thread query: zero threads. `gh pr view` showed no general comments or reviews.
- `make hooks`: passed. Hooks installed from `.githooks`.
- Live ruleset check: not run. Q-94 is open, and the hand-off says the live ruleset does not exist.
- Paid target: not run.
- Push: `73cc7ba699e14aedaa9c2f485e49f0651d356ede` was the head of `origin/docs/foundation-roadmap` when verified with `gh pr view`.

## Open questions and accepted risks

- D-87 accepts the risk that the gate cannot prove the provider of a review-record commit. RG 6 names the commit, and the owner must inspect it.
- Q-94 remains open. The live `main` ruleset does not exist, so GitHub does not yet enforce the committed required checks.

## Earlier verdicts

**Changes required.** This verdict applied to head `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

The review gate accepted an author-created approval record. The owner accepted this risk in D-87.

## Verdict

**Changes required.** This verdict applies to head `973703fea92df0ef5cdb480038bbe845f11ae2e4`.

The review instructions direct Codex-authored work to a Codex review and require a false author-provider line. Correct the author and reviewer paths before merge.
