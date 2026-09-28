# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 0 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR #1 is open on the branch `docs/foundation-roadmap`.

The first session ended at a context checkpoint. CI is green at `d0be0f3`. The first Codex round gave "Changes required" with one open finding, P1-1, in `docs/reviews/pr-1.md`. The review gate accepts a review record that the author can write without a Codex review.

Open work on PR #1:

1. Answer P1-1. Every commit comes from one GitHub account, so a commit author does not identify a reviewer. Ask the owner how to prove the review. Options: a GitHub review from a second account, or a signed review result. Then fix `docs/tools/review_gate.py` and add a regression test for a record that the author writes.
2. Record the naming decision of 2026-09-28 with the next free D- and Q- ids. The owner chose the form of the-thing-below for every pull request after #1. The title is `<type>: <summary> (PR-<n>)`, and the branch is `<type>/pr-<n>-<slug>`. Focused roadmaps assign the PR-<n> ids, from PR-1. Mark D-9 as extended by the new row. Update the rule in `docs/roadmaps/README.md` and rule 7 of `AGENTS.md`. Add a title and branch check to `docs/tools/pr_check.py` in the same change.
3. Push, wait for green CI, and run `make codex-review PR=1` again (D-8).
4. After the approval, ask the owner about the merge with the four-part summary (D-13).

Every launch question has an answer or a "Not applicable" status. Q-91 to Q-107 in `docs/questions.md` stay open, and each names the phase that needs it.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| `main` has no live ruleset. The committed ruleset waits for Q-94. | 2026-09-28 | `.github/rulesets/review-gate.json` |
| The review-gate workflow runs from `main`. It can not run on the first pull request, because `main` does not hold it yet. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| On the first pull request, `make codex-review` takes the rule files from the pull request head, because `main` does not hold them yet. | 2026-09-28 | `docs/tools/codex_review.py` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |

## Next steps, in order

1. Close PR #1: the open work of the resume section, then the owner confirmation and the merge.
2. Ask the owner about Q-94, then apply the ruleset and run `make ruleset-check`.
3. Write the Phase 1 focused roadmap: the Luna plan spike, the recognition spike, and the iPhone web platform spike.

## Session records

### Session 1 - 2026-09-27 to 2026-09-28

Author provider: Claude Code (Anthropic)

Branch: `docs/foundation-roadmap`. Role: author.

Completed:

- Asked the 73 launch questions and 16 follow-up questions. Recorded D-1 to D-84 and Q-1 to Q-107.
- The owner changed the product form to an installable web app for one user (D-17, D-67), and made Luna central (D-22).
- Wrote the design, the three research documents, the high-level roadmap, and the roadmap rules.
- Ported the Decktome process tooling without Gitar (D-3, D-6), with tests.

Open work:

- P1-1 of the first Codex round, the naming decision, a second Codex round, and the merge of PR #1.
- Q-91 to Q-107.
