# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 0 is complete (D-91, D-92). Phase 1 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR-1 of `docs/roadmaps/phase-1-risk-spikes.md` is open as GitHub #2 on the branch `docs/pr-1-phase-1-roadmap`.

PR-1 holds the Phase 1 focused roadmap, the owner answers D-91 to D-100, and the D-86 branch form in the `one-pr-one-session` skill. Codex reviewed effective head `dd6bd403acf9c1c3123e3bf8a61958616f99de50` and recorded `Ready for owner merge` with no findings. `make verify` passed. The review record is in `docs/reviews/pr-2.md`. Its first publication passed `review-gate` and all other CI checks. The owner confirmation remains.

Open work on PR-1:

1. Confirm the review commit and green CI on GitHub.
2. Ask the owner about the merge with the four-part summary (D-13, D-87, D-89).

Every Phase 1 question has an answer. Q-91 to Q-93 and Q-98 to Q-107 stay open, and each names the phase that needs it.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. `make ruleset-check` passed. | 2026-09-28 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |

## Next steps, in order

1. Close PR-1: the Codex review, the owner confirmation, and the merge.
2. Start PR-2, the Luna plan spike of `docs/roadmaps/phase-1-risk-spikes.md`, in a clean session.

## Session records

### Session 2 - 2026-09-28

Author provider: Claude Code

Branch: `docs/pr-1-phase-1-roadmap`. Role: author.

Completed:

- Applied the ruleset and the merge settings of `main` after the owner approval, and `make ruleset-check` passed (D-91).
- Asked Q-94 to Q-97 and Q-113 to Q-118, and recorded D-91 to D-100.
- Wrote `docs/roadmaps/phase-1-risk-spikes.md` with PR-1 to PR-6, and moved the start of the development project into work area 1.3 (D-99).
- Changed the branch form of the `one-pr-one-session` skill to D-86.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-1.

### Session 1 - 2026-09-27 to 2026-09-28

Author provider: Claude Code (Anthropic)

Branch: `docs/foundation-roadmap`. Role: author.

Completed:

- Answered Codex findings P1-1 to P1-5 (D-87 to D-90). Recorded the session rule D-85 and the naming rule D-86, each with a check.
- Asked the 73 launch questions and 16 follow-up questions. Recorded D-1 to D-84 and Q-1 to Q-107.
- The owner changed the product form to an installable web app for one user (D-17, D-67), and made Luna central (D-22).
- Wrote the design, the three research documents, the high-level roadmap, and the roadmap rules.
- Ported the Decktome process tooling without Gitar (D-3, D-6), with tests.

Open work:

- None. PR #1 merged.
