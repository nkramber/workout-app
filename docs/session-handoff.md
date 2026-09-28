# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 0 of `docs/roadmaps/high-level-roadmap.md` is in progress.

The first session wrote the foundation documents, the research, the high-level roadmap, and the process tooling on the branch `docs/foundation-roadmap`. The Codex review of PR #1 found P1-1 at `d0be0f3`: the review gate accepts an author-created approval record. The verdict is Changes required. Every launch question has an answer or a "Not applicable" status. Q-91 to Q-107 in `docs/questions.md` stay open, and each names the phase that needs it.

How to resume:

1. Run `make where`. It prints the branch, the tree, and the state of the pull request.
2. If the Phase 0 pull request is still open, finish its review loop with the `one-pr-one-session` and `pr-review` skills.
3. If it merged, ask the owner about Q-94, the live ruleset of `main`.
4. Then start the Phase 1 focused roadmap in a new clean session, after the owner approves the work (D-12).

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| `main` has no live ruleset. The committed ruleset waits for Q-94. | 2026-09-28 | `.github/rulesets/review-gate.json` |
| The review-gate workflow runs from `main`. It can not run on the first pull request, because `main` does not hold it yet. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| On the first pull request, `make codex-review` takes the rule files from the pull request head, because `main` does not hold them yet. | 2026-09-28 | `docs/tools/codex_review.py` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |

## Next steps, in order

1. Close the Phase 0 pull request: CI, Codex review, owner confirmation, merge.
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

- The review loop and the merge of the Phase 0 pull request.
- Q-91 to Q-107.
