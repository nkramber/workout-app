# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 0 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR #1 is open on the branch `docs/foundation-roadmap`. The first session continues until PR #1 merges (D-85).

Codex review round 1 gave "Changes required" with P1-1. The owner resolved it as an accepted risk (D-87). Round 2 found P1-2, and the owner approved `make claude-review` for Codex authors (D-88). The fix at `918d8e6` passed the provider-path tests. This review marks P1-2 fixed and finds P1-3, P1-4, and P1-5 open. The record gives the evidence at `docs/reviews/pr-1.md`.

Open work on PR #1:

1. Resolve P1-3, P1-4, and P1-5, or ask the owner about each risk.
2. Push the corrections and hand-off, then wait for green CI.
3. Run the required cross-provider review again.
4. After approval, ask the owner about the merge with the four-part summary. Name the commit that last changed the review record (D-13, D-87).

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
2. Name the first focused roadmap pull request with the form of D-86.
3. Ask the owner about Q-94, then apply the ruleset and run `make ruleset-check`.
4. Write the Phase 1 focused roadmap: the Luna plan spike, the recognition spike, and the iPhone web platform spike.

## Session records

### Session 1 - 2026-09-27 to 2026-09-28

Author provider: Claude Code (Anthropic)

Branch: `docs/foundation-roadmap`. Role: author.

Completed:

- Answered Codex findings P1-1 (D-87) and P1-2 (D-88). Recorded the session rule D-85 and the naming rule D-86, each with a check.

- Asked the 73 launch questions and 16 follow-up questions. Recorded D-1 to D-84 and Q-1 to Q-107.
- The owner changed the product form to an installable web app for one user (D-17, D-67), and made Luna central (D-22).
- Wrote the design, the three research documents, the high-level roadmap, and the roadmap rules.
- Ported the Decktome process tooling without Gitar (D-3, D-6), with tests.

Open work:

- Round 2 of the Codex review, and the merge of PR #1.
- Q-91 to Q-107.
