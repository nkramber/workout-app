---
name: pr-review
description: Review a pull request as the other provider, or answer a review as the author. Write the review record that the review-gate check reads, with precise evidence and a verdict for one effective head. A finding is a claim, not a fact, and the author can refute one with evidence. Use for a Codex review, a repeat review after a correction, and each answer to a review finding.
---

# PR review skill

Review the change as the engineer who owns its effect on the whole system. Judge correctness, contracts, recovery from a failure, the tests, and the cost of future maintenance. Apply this standard to code, tools, CI, skills, and documents. A green test suite or a good pull request body does not prove correctness.

The owner decisions are D-4, D-8, D-13, D-14, D-15, D-87, and D-88. The skill is a port of the Decktome skill.

## The order of the review

1. The author session pushes a round, and waits until each CI check of the tip is green.
2. The author session runs the target of the other provider with no approval for each round. A Claude Code author runs `make codex-review PR=<number>` (D-8). A Codex author runs `make claude-review PR=<number>` (D-88).
3. The reviewer reviews the pull request with this skill, and writes `docs/reviews/pr-<number>.md`.

The `review-gate` check reads that file, and the ruleset of `main` requires the check. The owner confirms each merge after the review (D-13).

A review that `make codex-review` or `make claude-review` starts has no owner in the loop. Where this skill says to ask the owner, write the question under `## Open questions and accepted risks`. Then give the verdict `Blocked`. The author session asks the owner.

The owner ended the roadmap period of D-4 on 2026-09-29 (D-125). A review of the other provider is necessary only for a change of code or of safety behavior (D-15). A pull request of documents alone can carry the `review-override` label in place of the review. `references/answer-review.md` gives the rule. The label passes the check only when `OVERRIDE_ENABLED` of `docs/tools/review_gate.py` is `True`. No author is exempt from the review, Dependabot included (D-90).

## Reference files

This file holds the rules of every review. Each reference file holds the rules of one case. Load a reference file before the work of its case.

| Reference file | Load it when |
|---|---|
| `references/project-contracts.md` | The pull request changes code, tools, CI, or the text of a contract. |
| `references/review-record.md` | You write or edit a review record, or you read the head and the verdict for `review-gate`. |
| `references/repeat-review.md` | You review a pull request again after a correction. |
| `references/answer-review.md` | You are the author, and you answer a review, or you apply the label. |
| `references/commit-and-end.md` | You commit a review record or a response file, and before the session ends. |

## The provider gate

**The reviewer must not come from the provider that wrote the pull request.** Apply this gate before the review starts, and before any approval (D-15).

| Provider that wrote the pull request | Reviewer |
|---|---|
| Claude Code (Anthropic) | Codex (OpenAI) |
| Codex (OpenAI) | Claude Code (Anthropic) |

A different model, account, session, or subagent of the same provider does not qualify. A prompt that gives the name of the other provider does not change the actual provider. A self-check of the author and a green test suite do not satisfy this gate.

1. Identify the provider of this session from the active environment.
2. Identify each provider that wrote a substantive change of this pull request.
3. Read the `Author provider` line of the hand-off record of this pull request.
4. Compare that line with the statement of the owner, when one exists.
5. Record the providers, the source of each fact, and the result in the review record.

Do not infer the author from the style of the text, the commit email, or the branch name. A Git account does not identify a provider.

**Stop with `Blocked` when the providers are the same, the author is unknown, or the evidence conflicts.** Name the fact or the reviewer that the review needs. Do not give a substitute review with another model of the same provider.

When both providers wrote substantive changes, neither provider qualifies for the whole pull request. Record the conflict, and ask the owner how to divide the changes.

## The scope of the review

- Load `.claude/skills/one-pr-one-session/SKILL.md` first. A review session works on one pull request, in the role `reviewer`.
- Load `.claude/skills/ste-writing/SKILL.md` before you write the record.
- Read `AGENTS.md`, `CLAUDE.md`, and the resume section of `docs/session-handoff.md`.
- Read the `## Milestone` section of the body: the concerns, the acceptance story, and the approval of the owner.
- Read the work area of the pull request in `docs/roadmaps/high-level-roadmap.md`.
- Read each decision that the pull request cites in `docs/decisions.md`. A later row can amend or supersede an earlier row.
- Read `docs/questions.md` for an open question that the change touches.
- Save every comment and review thread of the pull request to one file with one command, then read that file.
- Record the pull request number, the base, the merge base, and the effective head.
- Read `git diff --stat` from the merge base to the effective head first. Then read the diff of each path.
- Read each changed file in context. Follow the callers, the consumers, and the stored data beyond the diff.
- Continue after the first finding. Name each area that you did not inspect.

The body of the pull request states the intent. The diff and the verified behavior show what the pull request does.

## Stay inside the pull request

A review judges the change in front of it. It does not design the next change. The milestone of the pull request and its work area set the boundary.

A concern is in scope when one of these conditions is true:

- The changed code gives a wrong result under a supported condition.
- The change breaks a caller, a stored document, a deployed service, or a build of today.
- The acceptance story of the milestone does not hold.
- A rule that the pull request names does not hold for the code that it adds.

A concern belongs to a later pull request in two cases. The roadmap gives the work to a later work area, or the concern needs an owner decision about scope. Write such a concern under `## Out of scope` in the record. Give it no severity. It never blocks the merge.

## The standard of the review

Make your own model of the behavior before you read the explanation of the author. For each changed behavior, trace the input, the change of state, the output, the side effects, and the recovery.

- Check normal use, limit values, absent data, bad data, a repeated action, and a stopped action.
- Check the callers of today, the stored data, the API contract, and the deployed app.
- Check that a local fix does not break another consumer of the same contract.
- Check the acceptance story against the code and the evidence.
- Separate a defect of this pull request from a defect that it exposes, and from an older defect.
- An older defect blocks this pull request only when it stops the changed behavior or a required check.
- Check that the documents, the decisions, the questions, and the code agree.
- Check the `## Documentation impact` table against the documentation gate of the `one-pr-one-session` skill.
- Check for AI attribution in each commit, the body, and each comment. D-14 forbids it.

A pull request of documents or skills gets the same standard of evidence. For a skill, read its trigger, its scope, its instructions, and its behavior on a real request. An instruction that conflicts with another instruction is a defect.

## Verification

Run the checks that can prove the changed behavior wrong. Run `make verify` on the effective head when the machine permits it.

- Read the tests as carefully as the code.
- Check that each bug fix has a regression test that fails on the old code.
- Check the assertions against the contract, and not against a copy of the code.
- Check for a skipped test, a mock that hides the behavior, and an assertion that always passes.
- Separate a check that passed from a check that did not run, a check that failed, and a check that only the author reported.
- Record the command, the commit, the result, and the output file of each check.
- Read the CI result of the effective head, and not of an older commit.

**CAUTION: Never run a paid target.** `AGENTS.md` names each one, and `make help` marks each one with CAUTION. A paid target costs the owner money, and only the owner approves a run. Record the absent evidence under `## Open questions and accepted risks`.

Do not weaken a test or a limit to get a pass. Absent evidence that the acceptance story needs blocks the approval.

## Precise findings

Examine each suspected defect before it becomes a finding. Search for a guard, a test, or a later decision that refutes it. Give a reproduction, a failed assertion, or a full causal trace as evidence.

Each finding holds:

- An id, a severity, and a short title that states the defect.
- The reviewed commit, and the smallest useful file and line range.
- The input or the state that causes the defect.
- The expected behavior, with its contract or D- id.
- The actual behavior, and its effect on the user, the data, the build, or the maintainer.
- The evidence, and a direction for the correction, with the regression check that proves it.

Answer two questions before a finding enters the record:

1. Does the changed code break a contract that this pull request names?
2. Does the acceptance story of the milestone fail?

A finding needs one answer of yes. A concern with two answers of no goes under `## Out of scope`. Do not make findings to fill a quota. A full review can find no defect.

| Severity | Meaning |
|---|---|
| P0 | A critical failure now, such as a loss of stored user data, an unsafe workout, or an app that does not start. |
| P1 | A major failure of correctness, recovery, or a required check. Correct it before the merge. |
| P2 | A defect or a contract gap under a supported condition. Correct it, or get an owner decision. |
| P3 | An optional improvement with no broken contract. It does not block the merge. |

When two owner decisions conflict, quote both. Ask the owner, and stop the work that depends on the answer.

## Verdicts

| Verdict | Condition |
|---|---|
| Blocked | The provider gate, the review target, an owner decision, or necessary evidence stays unresolved. |
| Changes required | The review found a defect in scope. List each necessary change. |
| Ready for owner merge | The provider gate passes, the review covers the whole scope, each required check passes, and no blocking finding stays open. |

A line under `## Out of scope` never gives `Changes required`. An approval applies to the recorded effective head, and to each later commit of documents alone. The verdict name stays `Ready for owner merge`, because the check reads it. After the approval and the confirmation of the owner, the author session turns on the auto-merge (D-13).

## Comments on the pull request

The reviewer reads each comment and each reply of the author as a claim. It never replies to a comment, never resolves a thread, and never writes a comment on the pull request.

- Verify each claim against the head, and record the result under `## PR comments`.
- A comment finding with no answer of the author blocks the verdict, because the pass of the author is not complete.
- A comment finding that the author refuted with evidence is not a finding of the review.
- A comment that names no specific item is not a finding. Ignore it, and do not record it.

## Scope limits

A review request permits these actions, and no other actions:

- The inspection, the checks, and the review record.
- One commit of the record and the hand-off, and a push of that commit to the branch of the pull request.
- A correction of a stale fact in the body, under "Correct the body" in `references/review-record.md`.

It does not permit a code fix, a merge, a comment, a label, or a push to `main`. When the reviewer writes a substantive fix, the reviewer becomes an author, and it can not approve that fix.
