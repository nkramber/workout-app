# Pull request 1 review

Date: 2026-09-28

## Identity

- PR: 1
- Work area: 0.1 Foundation documents and process tooling
- Milestone: Foundation registers, research and design, roadmap, and process gates
- Target: `main`
- Base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Merge base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Head: `918d8e66779ae0ff411ce4363d318739fdb04048`
- Branch: `docs/foundation-roadmap`

## Provider gate

The hand-off names Claude Code (Anthropic) as the author. This session uses Codex (OpenAI). The provider line names the source. No owner statement conflicts with it, so the declared providers differ. Finding P1-3 describes why the branch cannot prove that declaration.

## Intended behavior and scope

The milestone adds the Phase 0 foundation and process tooling. Its acceptance story says a fresh session can find decisions, questions, and phases, and the ported gates pass. D-4 requires Codex review of each documentation pull request during the roadmap period. D-8 requires the author session to start the review. D-13 reserves merge confirmation to the owner. D-15 requires cross-provider review for code or safety changes.

The full pull request diff has 54 paths and 8,812 insertions. The last review covered effective head `973703fea92df0ef5cdb480038bbe845f11ae2e4`. This review covers the 14 paths changed since that head. I read the changed review tool, tests, make target, decision and question rows, README, hand-off, response, review record, and PR body.

I read each requested bootstrap rule file as code: `.claude/skills/one-pr-one-session/SKILL.md`, `.claude/skills/pr-review/SKILL.md`, the five reference files, `.claude/skills/ste-writing/SKILL.md`, `AGENTS.md`, and `CLAUDE.md`. I also read the project decisions, questions, roadmap, gate, workflows, ruleset, and the prior review history. I did not inspect every line of the design and research documents.

The provider-specific author commands correct P1-2. The gate still accepts branch-supplied provider claims, inherits review approval across later documentation changes, and exempts Dependabot without an owner decision.

## Findings

### P1-1: The required review record has no verifiable reviewer identity

Status: accepted risk, D-87.

Open at: `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

File: `docs/tools/review_gate.py:190-202`.

Trigger: A pull request author adds `docs/reviews/pr-<n>.md` with a current effective-head hash and the text `**Ready for owner merge.**`.

Expected: D-4 and D-8 require a Codex review of each pull request during the roadmap period. The required check must reject a record that the author can create without that review.

Actual: `evaluate()` checks that the file exists, the verdict says approved, and the recorded hash matches. The owner accepted this risk in D-87. RG 6 names the commit that last changed the record.

Consequence: The gate cannot prove that Codex wrote the record. The owner must inspect the named commit before each merge.

Correction: None. D-87 accepts this risk.

Regression check: The prior review reproduced the gate pass for an author-created record. The prior round's `make verify` passed, but its gate tests do not reject that record.

### P1-2: The author loop can run a same-provider review

Status: fixed in `918d8e66779ae0ff411ce4363d318739fdb04048`.

Open at: `973703fea92df0ef5cdb480038bbe845f11ae2e4`.

File: `.claude/skills/one-pr-one-session/SKILL.md:71-81`, `.claude/skills/pr-review/references/answer-review.md:7-14`.

Trigger: A Codex session authors a pull request, then follows the author loop after CI passes.

Expected: D-15 requires Claude Code to review Codex-authored work. The author guidance must select the other provider and record the true author provider.

Actual: The prior round sent every pull request to Codex and told each author to record Claude Code. This round adds `make claude-review` for Codex authors and tells each author to record its true provider.

Consequence: The correction restores the required provider-specific author path.

Correction: Add the Claude Code review target, a reviewer-provider gate, and provider-specific instructions.

Regression check: `make verify` passed 300 tests. Two tests reject a Claude review for a Claude Code author and a Codex review for a Codex author. The tests trust the provider values. P1-3 covers that source.

### P1-3: The provider gate trusts the PR author's provider label

Status: open.

Open at: `918d8e66779ae0ff411ce4363d318739fdb04048`.

File: `docs/tools/codex_review.py:290-310`.

Trigger: A Codex author writes `Author provider: Claude Code` in the hand-off for the PR branch.

Expected: D-15 and D-88 require a review by the provider that did not write the pull request. The provider gate must use evidence that does not come only from the PR author.

Actual: `check_provider()` reads the hand-off from the PR head, extracts its provider line, and accepts that claim as fact. I passed a fabricated hand-off that names Claude Code to `check_provider()` for a Codex review. The gate returned success.

Consequence: A Codex author can label the PR as Claude Code work and make the Codex target accept a same-provider review. The skill also accepts the line when no separate owner statement exists.

Correction: Verify the author provider with a source that the PR author cannot change. Otherwise, treat the provider as unknown and block the review until the owner confirms it.

Regression check: The inline Python reproduction at `918d8e6` returned success for the fabricated provider line. `make verify` passed 300 tests, but the provider tests supply trusted hand-off text and do not test a false provider claim.

### P1-4: The gate carries approval across new documentation changes

Status: open.

Open at: `918d8e66779ae0ff411ce4363d318739fdb04048`.

File: `docs/tools/review_gate.py:150-163`, `docs/tools/review_gate.py:82-83`, `.claude/skills/pr-review/references/review-record.md:21-31`.

Trigger: A pull request gets an approved record, then adds a later commit that changes only a document, skill, roadmap, or other eligible path.

Expected: D-4 requires Codex review of each documentation pull request until the owner ends the roadmap period. A changed review rule or roadmap must receive current review.

Actual: `documents_since()` lets the old head pass RG 5 when every later commit changes eligible documentation alone. The test `DocumentsAfterTheApproval.test_a_commit_of_each_kind_of_document_keeps_the_gate` asserts that an edit to `.claude/skills/pr-review/SKILL.md` keeps the old approval.

Consequence: The required check can pass for changed documentation that the recorded review never examined, while D-4 still applies.

Correction: While D-4 holds, require a review record for the newest documentation change. Keep metadata-only commits outside the effective head.

Regression check: The inline Python reproduction at `918d8e6` returned `PASS` for an old approved head followed by a roadmap change. `make verify` passed 300 tests, including tests that preserve this behavior.

### P1-5: Dependabot can skip the required review

Status: open.

Open at: `918d8e66779ae0ff411ce4363d318739fdb04048`.

File: `docs/tools/review_gate.py:191-193`, `.claude/skills/pr-review/SKILL.md:20-25`, `.claude/skills/pr-review/references/review-record.md:150-154`.

Trigger: Dependabot opens a documentation-only pull request and writes every commit.

Expected: D-4 requires Codex review of every documentation-only pull request until the owner ends the roadmap period. D-4 defines no Dependabot exception.

Actual: RG 2 returns `PASS` for any pull request that Dependabot opens and whose commits all use Dependabot's author email. It does not check the changed paths or the D-4 period. The inline reproduction for a Dependabot documentation pull request returned the RG 2 pass without a review record.

Consequence: A documentation-only pull request can merge without the review that D-4 requires.

Correction: Remove the exception for documentation-only pull requests while D-4 holds, or get an owner decision that amends D-4.

Regression check: The inline Python reproduction at `918d8e6` returned `RG 2 PASS` for `docs/decisions.md` with no review record. `make verify` passed 300 tests, including the broad Dependabot exemption.

## Out of scope

None.

## PR comments

No general comments or review threads exist. `gh pr view --comments` returned no comments. The GraphQL review-thread query returned zero threads.

## Description edits

None.

## Verification

- `make where`: clean detached worktree at `918d8e6`, based on `origin/main` at `613c3cb`.
- `git diff --stat origin/main...HEAD`: 54 paths, 8,812 insertions.
- `git diff --stat 973703f...HEAD`: 14 paths, 312 insertions and 50 deletions.
- `python3 docs/tools/review_gate.py --effective-head 1`: `918d8e66779ae0ff411ce4363d318739fdb04048`.
- `make verify` at `918d8e6`: passed. STE, reference, lifecycle, and context checks passed. All 300 tests passed.
- `gh pr checks 1 --repo nkramber/workout-app`: `pr-contract`, `verify:lint`, and `verify:test` passed at `918d8e6`. `review-gate` did not run because `origin/main` does not contain the workflow.
- `gh pr view 1 --repo nkramber/workout-app`: PR is open at `918d8e66779ae0ff411ce4363d318739fdb04048`, based on `613c3cb08733338b9ac73b33e19d2380220b41da`.
- `gh pr view --comments` and the GraphQL review-thread query: no comments and zero review threads.
- Inline reproduction: RG 5 passed an old approval after a roadmap change.
- Inline reproduction: RG 2 passed a Dependabot documentation pull request without a record.
- Inline reproduction: `check_provider()` accepted a fabricated Claude Code provider line.
- `make hooks`: passed. Hooks are installed from `.githooks`.
- Live ruleset check: not run. Q-94 remains open, and the hand-off says no live ruleset exists.
- Paid review targets: not run.
- Push: pending.

## Open questions and accepted risks

- D-87 accepts the risk that the gate cannot prove which provider wrote the review record. RG 6 names the commit, and the owner must inspect it.
- Q-94 remains open. The live `main` ruleset does not exist, so GitHub does not enforce the committed required checks.
- P1-3 needs a trusted provider source or an owner decision.
- P1-4 and P1-5 conflict with D-4 and need correction or an owner decision.

## Earlier verdicts

**Changes required.** This verdict applied to head `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

The review gate accepted an author-created approval record. The owner accepted this risk in D-87.

**Changes required.** This verdict applied to head `973703fea92df0ef5cdb480038bbe845f11ae2e4`.

The author loop sent Codex-authored work to Codex. P1-2 records that finding.

## Verdict

**Changes required.** This verdict applies to head `918d8e66779ae0ff411ce4363d318739fdb04048`.

P1-3 through P1-5 leave the provider gate or the required documentation review open to bypass.
