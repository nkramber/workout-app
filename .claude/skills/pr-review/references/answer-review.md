# PR review: the author answer

Part of the `pr-review` skill. Load this file when you are the author, and you answer a review or apply the `review-override` label.

## Start the review of the other provider

Start the review yourself after the Gitar pass and after CI is green (D-8, D-88, D-335). The owner approved each round of the loop. The target spends the Codex plan or the Claude plan of the owner.

| Author provider | Target | Reviewer |
|---|---|---|
| Claude Code | `make codex-review PR=<number>` | Codex |
| Codex | `make claude-review PR=<number>` | Claude Code |

1. Write the line `Author provider: <your provider>` in the hand-off record, with the true provider of this session, and push it. The target refuses a record that names the provider of the reviewer.
2. Complete the Gitar pass with the `gitar-review` skill. The target refuses an incomplete pass (D-336).
3. Wait until each check of the tip completes and passes.
4. Run the target of the table in the background, and wait for the notice of its end.
5. Read the last line of the output: `outcome: <name> (exit <n>)`.
6. Run `git pull --ff-only`, because the reviewer pushed the record.

| Outcome | Next step |
|---|---|
| approve | Go to the merge of the `one-pr-one-session` skill. The owner confirms the merge after the summary of four sections (D-13). |
| changes | Answer each finding with the procedure below, push, and start the next round. |
| three-strike stop | Do the procedure of "The three-strike stop" below. |
| refusal | Correct the condition that the output names, then run the target again. |
| fault | Read the transcript that the output names. Ask the owner when the cause is not clear. |

A refusal spends nothing. The target refuses a dirty tree, a checkout that differs from origin, an incomplete Gitar pass, and an open review thread. Only in a pause of Gitar that the owner states, add `SKIP_GITAR=1` to the target (D-337).

A reply to a comment names no provider, harness, or model (D-14).

## The three-strike stop

The target exits 4 when a blocking finding is open at its third effective head. Do these steps:

1. Turn off the auto-merge with `gh pr merge <number> --disable-auto`.
2. Stop the fix loop. Change no file for that finding.
3. Ask the owner with `AskUserQuestion`.
4. Record the answer in `docs/reviews/pr-<number>-response.md`.
5. Record the answer as a new D- row too, when it sets a rule.

The question gives the finding, the evidence of the reviewer, each answer of the author so far, and the options with their pros and cons.

## The label

D-4 kept the Codex review for every pull request of documents alone until the owner ended the roadmap period. The owner ended it on 2026-09-29 (D-125), and `OVERRIDE_ENABLED` in `docs/tools/review_gate.py` is `True`.

Now a pull request of documents alone can merge with no Codex review (D-15). A change of the constant is a change of code, so it needs a Codex review. Apply the `review-override` label yourself when all of these conditions are true:

- Each changed path is in the documentation set of `docs/tools/review_gate.py`.
- The change does not change safety behavior (D-15).
- Each check of the pull request is green, except `review-gate`.
- The Gitar pass of the head is complete (D-335). No target reads it for a label.
- The completion gate of the `one-pr-one-session` skill holds.

The documentation set holds `docs/`, `.claude/`, `CLAUDE.md`, `AGENTS.md`, `README.md`, and `.github/pull_request_template.md`. These paths are not in it: `docs/tools/`, `.claude/hooks/`, `.claude/settings.json`, the local settings file of the harness, and each workflow.

Apply the label with this command:

```bash
gh pr edit <number> --add-label review-override
```

The label event runs `review-gate` again. Read its result. When you push again after the label, remove the label first. Apply it again when each condition above holds again.

## Answer the findings of a review

**A finding is a claim, not a fact.** A review can be wrong. Examine each finding against the evidence before you change anything.

1. Run `git fetch` and `git status --short --branch`.
2. Read the finding, and read the file and the lines that it names.
3. Reproduce the trigger. A finding that does not reproduce has no merit.
4. Read the contract that the finding cites. Look for a later decision that amends it.
5. Decide the result: full merit, partial merit, or no merit.
6. Correct each part with merit. Make the smallest change that restores the contract.
7. Record each result in `docs/reviews/pr-<number>-response.md`.
8. Commit the response, the corrections, and the hand-off, then push.
9. When CI is green, start the repeat review with the target of the table above.

Refute a finding when the evidence supports it:

| Reason | What to show |
|---|---|
| The finding reads a rule too broadly. | Quote the rule. Name the other files that the broad reading also condemns. |
| The finding cites an amended decision. | Quote the later decision. |
| The trigger does not reproduce. | Give the command, the commit, and the result. |
| The correction breaks another contract. | Name the contract and the caller. |
| The finding asks for work of a later work area. | Quote the roadmap work area and the milestone. |

Never accept a finding only to close the review faster. Ask the owner when a finding and an owner decision conflict, and quote both.

## The response file

Write `docs/reviews/pr-<number>-response.md` when the verdict is `Changes required` or `Blocked`. The `review-gate` check does not read it. For each finding, the file states:

- The result: full merit, partial merit, or no merit.
- The evidence, for partial merit or no merit.
- The correction, with the file and the decision id.
- The regression check that ran, and its result.
