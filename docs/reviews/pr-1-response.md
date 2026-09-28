# Pull request 1 - author response

Date: 2026-09-28. Review rounds 1, 2, and 3 recorded heads `d0be0f3`, `973703f`, and `918d8e6`. Each gave the verdict "Changes required".

## P1-1: The required review record has no verifiable reviewer identity

**Result: partial merit.**

The trigger reproduces. `evaluate()` in `docs/tools/review_gate.py` passes a record that the author writes, because it reads only the file, the verdict, and the head.

The correction that the finding asks for does not fit this repository. Every commit comes from one GitHub account, and no commit has a signature. So no check in this repository can prove which provider wrote a commit. A key or a token for the gate lives on the same machine as the author session. The owner chose the rule what-you-carry:D-198 for this case (D-87). The gate stops an accident and not an attack, and the owner reads the commit of the record before each merge.

**Correction.**

- `docs/tools/review_gate.py`: the new rule RG 6 names the commit that last changed `docs/reviews/pr-<n>.md`, with its subject. It gives the reason, and it cites D-87.
- `.claude/skills/one-pr-one-session/SKILL.md`: the "Codex review" part of the merge question names that commit (D-87).
- `AGENTS.md` rule 9 states the accepted risk (D-87).
- `docs/decisions.md` records D-87, and `docs/questions.md` records Q-109.

**Regression checks.**

- `RecordCommit.test_a_record_that_the_author_writes_passes_and_names_its_commit` gives the gate an author record. The gate passes, as D-87 accepts, and RG 6 names the commit of that record.
- `GitFacts.test_the_command_passes_an_approved_head_and_ignores_the_review_commit` runs the command on a real Git repository, and it reads the RG 6 line in the output.
- `make verify`: 287 tests pass, and every document check reports 0 findings.

## P1-2: The author loop can run a same-provider review

Round 2 recorded head `973703f`.

**Result: full merit.**

The author loop of `.claude/skills/one-pr-one-session/SKILL.md` sent every pull request to `make codex-review`. The answer guide told every author to write `Author provider: Claude Code`. A Codex author that followed both rules got a Codex review, which D-15 forbids.

**Correction.** The owner approved a port of the claude-review command of what-you-carry (D-88).

- `docs/tools/codex_review.py`: the new `--reviewer claude` path runs Claude Code in a detached worktree. It removes each API credential from the environment, needs a claude.ai account login, and runs a model probe first.
- `docs/tools/codex_review.py`: the new provider gate reads the author records of the branch in `docs/session-handoff.md` at the head. It refuses a review by the provider of the author, a branch with two author providers, and a branch with no author record.
- `Makefile`: the new paid target `make claude-review PR=<n>`.
- The `one-pr-one-session` skill, the `pr-review` skill and its references, `AGENTS.md`, and `README.md` name the target of each author provider. The answer guide tells the author to write its true provider.
- `docs/decisions.md` records D-88, and `docs/questions.md` records Q-110.

**Regression checks.**

- `ProviderGate.test_codex_refuses_a_codex_author` and `ProviderGate.test_claude_refuses_a_claude_code_author` prove that each target refuses the provider of the author.
- `ProviderGate.test_two_authors_of_one_branch_refuse_each_reviewer` and `ProviderGate.test_no_author_record_refuses` cover the conflict cases.
- `ClaudeReviewer` tests cover the environment, the login, the probe, and a usage error.
- `make verify`: 300 tests pass, and every document check reports 0 findings.

## P1-3: The provider gate trusts the author provider label

Round 3 recorded head `918d8e6`.

**Result: partial merit.** The trigger reproduces. The correction that the finding asks for has the same limit as P1-1. One GitHub account writes every commit, so no source that the author can not change names the provider. The owner extended the accepted risk of D-87 to the provider line (D-89).

**Correction.** The merge question of the `one-pr-one-session` skill names the author provider and the reviewer provider, and the owner reads both. `AGENTS.md` rule 9 and the docstring of `check_provider` state the risk.

**Regression check.** `ProviderGate.test_a_false_provider_line_passes_as_d89_accepts` records the accepted behavior.

## P1-4: The gate carries approval across new documentation changes

**Result: full merit.** D-4 needs a current review of each document change during the roadmap period.

**Correction.** `check_head` in `docs/tools/review_gate.py` carries an approval across later document commits only when `OVERRIDE_ENABLED` is `True`, after the D-4 period. During the period, only the metadata set keeps the effective head: the record, the response file, and the hand-off. The `review-record.md` reference of the `pr-review` skill states the rule.

**Regression checks.** `DocumentsDuringTheRoadmapPeriod` proves that a roadmap, decision, skill, or rules commit fails RG 5, and that a metadata commit keeps the approval. `GitFacts.test_the_command_fails_a_commit_of_documents_after_the_review_in_the_period` runs the command on a real Git repository.

## P1-5: Dependabot can skip the required review

**Result: full merit.** No owner decision of this repository made the exemption. It came with the port of the Decktome gate.

**Correction.** The owner removed the exemption (D-90). `docs/tools/review_gate.py` has no Dependabot rule, and the `pr-review` skill and its `review-record.md` reference say that no author is exempt.

**Regression checks.** `NoExemptAuthor` proves that a Dependabot pull request of code or of documents needs the record. `GitFacts.test_the_command_fails_a_dependabot_pull_request_with_no_record` runs the command. `make verify` passes 301 tests.

## Other changes in this round

The owner gave two rules during the round. Both changes are in this round, with their tests.

- D-85: a session continues until its pull request merges. The context checkpoint asks for a compaction, and it never ends the session. The hook `.claude/hooks/context_checkpoint.py`, its test, and section 4 of the `one-pr-one-session` skill changed.
- D-86: pull requests after #1 use the naming form of the-thing-below. `docs/tools/pr_check.py` checks the title and the branch. The branch of PR #1 keeps its name.
- `docs/tools/ref_check.py` skips the ids of the-thing-below and what-you-carry, as it skips the ids of Decktome.
