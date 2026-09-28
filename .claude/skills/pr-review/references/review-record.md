# PR review: the review record

Part of the `pr-review` skill. Load this file before you write a review record or correct the body of a pull request. Load it too before you read the head and the verdict for `review-gate`.

## The record

Use one file for each pull request, `docs/reviews/pr-<number>.md`. The number is the GitHub number of the pull request. Keep the name and the finding ids on a repeat review.

The record and the `Author provider` line of the hand-off can name a provider (D-15). No commit, body, branch, or comment names one (D-14).

The `review-gate` check reads three parts of the record. Keep their form exact:

| Part | Exact form | Rule |
|---|---|---|
| The file name | `docs/reviews/pr-<number>.md` | The GitHub number of the pull request. |
| The head field | `- Head: ` and the hash in backticks, in the `## Identity` list | The effective head, with 7 characters or more. The check reads that list alone. |
| The verdict | The verdict line of the `## Verdict` section, such as `**Ready for owner merge.**` | The section holds one bold span, and that span is one of the three verdict names. |

## The effective head

The effective head is the newest commit that changes a path outside the metadata set. The metadata set holds three paths of the pull request:

- `docs/reviews/pr-<number>.md`
- `docs/reviews/pr-<number>-response.md`
- `docs/session-handoff.md`

A commit that changes those paths alone is a metadata commit. It does not move the effective head. So the commit of the record does not make the record stale, and a later hand-off commit of the author does not either. A commit that changes the record of another pull request moves the effective head.

The check also passes a record of an earlier commit, when each later commit changes documents alone. The documents are the documentation set of `docs/tools/review_gate.py`. So a later commit of the roadmap, a decision, or a skill keeps the approval. Still record the effective head, because `make codex-review` reads a new record against it.

A merge commit always moves the effective head, because it brings new code into the branch.

Record the effective head, and not the tip of the branch. Read it with the rule of the check itself, so the two never disagree:

```bash
git fetch origin main
python3 docs/tools/review_gate.py --effective-head <number>
```

## The skeleton

Keep the text of each heading, and keep the order.

```markdown
# Pull request <number> review

Date: <YYYY-MM-DD>

## Identity

- PR: <number>
- Work area: <the work area of the high-level roadmap>
- Milestone: <the milestone of the pull request body>
- Target: `main`
- Base: `<sha>`
- Merge base: `<sha>`
- Head: `<effective head sha>`
- Branch: `<branch>`

## Provider gate

The author provider, the source of that fact, the reviewer provider, and the result.

## Intended behavior and scope

The milestone, its acceptance story, and each contract that the change touches.
Each path of `git diff --stat`, as inspected or as not inspected.

## Findings

One subsection for each finding, in severity order. Write "No finding." when the review found none.

## Out of scope

One line for each concern that a later work area holds, with the name of that work area. Write "None." when there is none.

## PR comments

One line for each thread: the claim, the answer of the author, and what the review verified.

## Description edits

One line for each correction of the body: the old value and the new value. Write "None." when there is none.

## Verification

One line for each command or check, with its result. Name each check that did not run, and the reason.
End with the push line: `- Push: <sha> is the head of origin/<branch>, verified with gh pr view.`

## Open questions and accepted risks

Each open question and each accepted risk, with its id.

## Verdict

**<Blocked | Changes required | Ready for owner merge>.** This verdict applies to head `<sha>`.
The reason in one or two sentences, with no bold text.
```

A repeat review replaces the verdict of the `## Verdict` section. Put each earlier verdict under `## Earlier verdicts`, above `## Verdict`. A second bold span in the `## Verdict` section fails the check.

## The finding form

Give each finding a stable id: the letter `P`, the severity, a hyphen, and an index. `P1-1` is the first P1 finding. Never give a finding a new number on a repeat review.

```markdown
### P<severity>-<n>: <a short title that states the defect>

Status: <open | fixed in `<sha>` | accepted risk, D-<id> | withdrawn>.

Open at: `<effective head sha>`, `<effective head sha>`.

File: `<path>:<line range>`, or Commit: `<sha>`.

Trigger: the input or the state that causes the defect.

Expected: the correct behavior, with the contract or the D- id.

Actual: the observed behavior.

Consequence: the effect on the user, the data, the build, or the maintainer.

Correction: the smallest change that restores the contract.

Regression check: the command or the test that proves the fix, and its result.
```

A withdrawn finding stays in the file with the evidence that refuted it. Never delete a finding.

## The open rounds

The `Open at:` line lists each effective head at which a review found the finding open, oldest first. Keep each earlier head. When this review finds the finding open, add the effective head of this review. When the finding is fixed at this head, add no head. A second review of one head adds that head one time only.

`make codex-review` counts the distinct heads of an open P0, P1, or P2 finding. At three heads it stops the fix loop, and the owner decides. A finding that closes and then opens again keeps its earlier heads. A record without the line counts the current head alone. The `review-gate` check does not read the line.

## Correct the body

A body that names a stale head, an old count, or a replaced correction gives the owner wrong facts at the merge. The reviewer corrects such a fact directly, with no finding. The reviewer changes only a fact that the review verified:

- The effective head, the base, or the merge base.
- A count that the review read, such as the number of tests or findings.
- The result of a check that the review read.

The reviewer never changes what the author says the pull request does, or why. It never changes a decision or a recommendation. Name no provider in the body (D-14). Write one line for each edit under `## Description edits`.

## The review-gate check

The check has five rules:

1. RG 1: when the `review-override` label is on and `OVERRIDE_ENABLED` is `True`, each changed path is in the documentation set. While D-4 holds, the constant is `False`, and the label satisfies no rule.
2. RG 2: Dependabot opened the pull request and wrote every commit, and GitHub committed each one. This rule passes the check alone.
3. RG 3: `docs/reviews/pr-<number>.md` exists on the head, and the head is a branch of this repository, not a fork.
4. RG 4: the verdict is `Ready for owner merge`.
5. RG 5: the head field names the effective head, or an earlier commit that documents alone follow.

`docs/tools/review_gate.py` holds each rule, and `docs/tools/test_review_gate.py` holds its tests. A push of code after the approval fails RG 5. That result is correct: review the new diff, then change the head field and the verdict together.

The check can not run on a pull request that changes `.github/workflows/review-gate.yml`. GitHub starts `pull_request_target` from `main` alone, so the check reads the rules of `main`. The first pull request that adds the workflow gets no run of the check. So the ruleset of `main` goes live after that merge.
