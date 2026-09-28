# Pull request 1 review

Date: 2026-09-28

## Identity

- PR: 1
- Work area: 0.1 Foundation documents and process tooling
- Milestone: Foundation registers, research and design, roadmap, and process gates
- Target: `main`
- Base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Merge base: `613c3cb08733338b9ac73b33e19d2380220b41da`
- Head: `7931be813f2262c559bfa89878cdf2e79b14e6a6`
- Branch: `docs/foundation-roadmap`

## Provider gate

The hand-off names Claude Code (Anthropic) as the author. The prior review names this session as Codex (OpenAI). The hand-off names the source, and the owner has no conflicting statement in the reviewed record. The declared providers differ. P1-3 records the accepted risk that the author controls the provider line.

## Intended behavior and scope

The milestone adds the Phase 0 foundation and process tooling. Its acceptance story says a fresh session can find decisions, questions, and phases, and the ported gates pass. D-4 requires Codex review of each documentation pull request during the roadmap period. D-8 requires the author session to start the review. D-13 reserves merge confirmation to the owner. D-15 requires cross-provider review for code or safety changes.

The full pull request has 54 paths and 8,938 insertions. The prior review covered head `918d8e66779ae0ff411ce4363d318739fdb04048`. This review covers 13 paths since that head. I read each changed path and each bootstrap file named in the request. I read the hand-off, milestone, pull request body, decisions, questions, gate code, tests, review history, workflows, roadmap, and contract references.

I did not inspect every line of the design and research documents. This round changes review process documents and tools.

The changes remove approval inheritance across document commits during D-4 and remove the Dependabot exemption. They record D-89 as an accepted risk for the author-controlled provider line.

## Findings

### P1-1: The review gate cannot prove who wrote the review record

Status: accepted risk, D-87.

Open at: `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

File: `docs/tools/review_gate.py:190-202`.

Trigger: A pull request author adds a record with a current head hash and the text `**Ready for owner merge.**`.

Expected: D-4 and D-8 require a Codex review of each pull request during the roadmap period. The gate cannot identify the writer of the record.

Actual: The gate checks the file, verdict, and head. D-87 accepts this limit. RG 6 reports the last commit that changed the record.

Consequence: A same-provider or fabricated review record can pass the automated check. The owner must inspect the record commit before merge.

Correction: None. D-87 accepts this risk.

Regression check: The prior review reproduced an author-created record that passes the gate. `make verify` passed 301 tests at the current head.

### P1-2: The author loop can run a same-provider review

Status: fixed in `918d8e66779ae0ff411ce4363d318739fdb04048`.

Open at: `973703fea92df0ef5cdb480038bbe845f11ae2e4`.

File: `.claude/skills/one-pr-one-session/SKILL.md:71-81`, `.claude/skills/pr-review/references/answer-review.md:7-14`.

Trigger: A Codex session authors a pull request and follows the review instructions after CI passes.

Expected: D-15 requires Claude Code to review Codex-authored work.

Actual: The prior round sent every pull request to Codex. The correction adds `make claude-review` for Codex authors and records the true declared author provider.

Consequence: The author instructions select the other provider.

Correction: Add the Claude Code review target and provider-specific author instructions.

Regression check: The current `make verify` passed 301 tests. Provider tests reject the configured same-provider cases. P1-3 records the separate limit in the source of the provider value.

### P1-3: The provider gate trusts the author-controlled provider line

Status: accepted risk, D-89.

Open at: `918d8e66779ae0ff411ce4363d318739fdb04048`.

File: `docs/tools/codex_review.py:298-310`, `AGENTS.md:34`.

Trigger: A Codex author writes `Author provider: Claude Code` in the hand-off.

Expected: D-15 and D-88 require review by the provider that did not write the pull request. The automated provider gate must not treat an author-controlled claim as independent evidence.

Actual: `check_provider()` reads the hand-off from the pull request head and accepts its provider line. The regression test explicitly confirms that a false line passes. D-89 accepts this risk and tells the owner to read the author provider before merge.

Consequence: A same-provider review can pass the automated provider gate if the author misstates the provider. The bootstrap rules weaken the provider gate at this point, despite the manual owner check.

Correction: None. D-89 accepts this risk. The owner must verify the provider outside the pull request before merge.

Regression check: `ProviderGate.test_a_false_provider_line_passes_as_d89_accepts` passed. It proves the accepted bypass remains possible. `make verify` passed 301 tests.

### P1-4: The gate carries approval across new document changes

Status: fixed in `7931be813f2262c559bfa89878cdf2e79b14e6a6`.

Open at: `918d8e66779ae0ff411ce4363d318739fdb04048`.

File: `docs/tools/review_gate.py:159-165`, `.claude/skills/pr-review/references/review-record.md:21-31`.

Trigger: A pull request receives an approval, then a later commit changes an eligible document during the D-4 period.

Expected: D-4 requires a current review of each document change during that period.

Actual: `check_head()` now allows document-only commits to preserve approval only when `OVERRIDE_ENABLED` is true. The constant is false during D-4. Metadata-only commits still preserve the effective head.

Consequence: A later document change fails the review gate until a new review updates the record.

Correction: Require current review for document changes during D-4.

Regression check: The new during-period tests and the Git repository reproduction fail RG 5 for a later roadmap change. `make verify` passed 301 tests.

### P1-5: Dependabot can skip the required review

Status: fixed in `7931be813f2262c559bfa89878cdf2e79b14e6a6`.

Open at: `918d8e66779ae0ff411ce4363d318739fdb04048`.

File: `docs/tools/review_gate.py:191-202`, `.claude/skills/pr-review/SKILL.md:20-22`.

Trigger: Dependabot opens a pull request and writes each commit.

Expected: D-4 requires a review of every pull request during the roadmap period. D-90 removes the Dependabot exemption.

Actual: The gate no longer has a Dependabot bypass. The new tests require a review record for Dependabot changes to code and documents.

Consequence: Dependabot changes must satisfy the review gate.

Correction: Remove the exemption.

Regression check: Dependabot gate tests and the Git repository reproduction require RG 3 to find a review record. `make verify` passed 301 tests.

## Out of scope

None.

## PR comments

No issue comments, submitted reviews, or review threads exist. The GitHub API returned zero for each.

## Description edits

None.

## Verification

- `make where`: clean detached worktree at `7931be8`, based on `origin/main` at `613c3cb`.
- `git diff --stat 918d8e6..HEAD`: 13 paths, 228 insertions, and 102 deletions.
- `python3 docs/tools/review_gate.py --effective-head 1`: `7931be813f2262c559bfa89878cdf2e79b14e6a6`.
- `make verify` at `7931be8`: passed. STE, reference, lifecycle, and context checks passed. All 301 tests passed.
- `make ste-check`: 0 findings.
- `make ref-check`: 0 findings.
- `git diff --check 918d8e6..HEAD`: passed.
- `git diff --check 613c3cb..HEAD`: failed because `docs/tools/test_pr_check.py:368` has an extra blank line at EOF. That line is in the original PR commit, not this review round.
- Pre-commit `python3 docs/tools/review_gate.py --event /tmp/pr1-event.json --head HEAD`: failed on the old committed record. The gate reads committed files, so it did not read this working copy.
- `gh pr checks 1`: `pr-contract`, `verify:lint`, and `verify:test` passed at head `7931be8`. The review-gate workflow did not run because `main` does not contain it.
- GitHub issue comments, submitted reviews, and review threads: zero.
- Live ruleset check: not run. Q-94 remains open, and the hand-off says no live ruleset exists.
- Paid review targets: not run.
- Push: pending.

## Open questions and accepted risks

- D-87 accepts that the gate cannot prove who wrote the review record. The owner must inspect the record commit.
- D-89 accepts that the provider gate trusts the provider line in the author-controlled hand-off. A false line passes the provider gate. The owner must independently verify both providers before merge.
- Q-94 remains open. The live `main` ruleset does not exist, so GitHub does not enforce the committed required checks.

## Earlier verdicts

**Changes required.** This verdict applied to head `d0be0f3d98ce5b6d286847f224c59e73755ead96`.

The gate accepted an author-created approval record. D-87 accepted that risk.

**Changes required.** This verdict applied to head `973703fea92df0ef5cdb480038bbe845f11ae2e4`.

The author loop sent Codex-authored work to Codex. P1-2 records that finding.

**Changes required.** This verdict applied to head `918d8e66779ae0ff411ce4363d318739fdb04048`.

P1-3 through P1-5 left the provider gate and documentation review open to bypass.

## Verdict

**Ready for owner merge.** This verdict applies to head `7931be813f2262c559bfa89878cdf2e79b14e6a6`.

The earlier findings are fixed or accepted by D-87 and D-89. The owner must verify the provider identities and review commit, and Q-94 still leaves the live ruleset unenforced.
