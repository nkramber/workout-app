# Pull request 1 - author response

Date: 2026-09-28. Review round 1 recorded head `d0be0f3` with the verdict "Changes required".

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

## Other changes in this round

The owner gave two rules during the round. Both changes are in this round, with their tests.

- D-85: a session continues until its pull request merges. The context checkpoint asks for a compaction, and it never ends the session. The hook `.claude/hooks/context_checkpoint.py`, its test, and section 4 of the `one-pr-one-session` skill changed.
- D-86: pull requests after #1 use the naming form of the-thing-below. `docs/tools/pr_check.py` checks the title and the branch. The branch of PR #1 keeps its name.
- `docs/tools/ref_check.py` skips the ids of the-thing-below and what-you-carry, as it skips the ids of Decktome.
