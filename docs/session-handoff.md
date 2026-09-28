# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 1 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR-1 merged as GitHub #2. PR-2 of `docs/roadmaps/phase-1-risk-spikes.md` is open as GitHub #3 on the branch `feat/pr-2-luna-plan-spike`, from the base `ddd0f49`.

PR-2 holds the Luna plan spike of work area 1.1:

- the harness, the 20 synthetic profiles, the plan schema and its validator, and the first draft of the policy rules table, in `tools/spikes/luna_plan/`,
- the results of the paid run, in `tools/spikes/luna_plan/results/`,
- the report `docs/research/luna-plan-spike.md`: go under the bar of D-101,
- the owner answer Q-119 and the decision D-101.

The owner approved the paid run at run time (D-25). The smoke call and the run of 60 plans cost 0.0564 USD in total, under the cap of 2 USD (D-98). `make verify` passed.

Codex round 1 reviewed effective head `365c2045ec80d68fe5f97914e05e74abe99b599e`. The verdict is `Changes required`, with finding P2-1 in `docs/reviews/pr-3.md`. The author found full merit, and corrected the budget guard. Now the harness reserves the worst-case cost of each attempt, retries included. `docs/reviews/pr-3-response.md` holds the answer and its regression checks. `make verify` passed after the correction.

Open work on PR-2:

1. Read the fresh CI result after the Codex review record and hand-off reach the branch. The review record says `Blocked` until `review-gate` passes.
2. If the gate passes, ask the owner about the merge with the four-part summary (D-13, D-87, D-89).

Q-91 to Q-93 and Q-98 to Q-107 stay open, and each names the phase that needs it. The draft rules use draft values for Q-92, Q-101, Q-102, Q-104, Q-105, and Q-106, and they answer none of them.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. `make ruleset-check` passed. | 2026-09-28 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| `gpt-6-luna` accepts the strict JSON schema of the plan through the Responses API. 60 of 60 plans passed the schema. | 2026-09-28 | `docs/research/luna-plan-spike.md` |

## Next steps, in order

1. Close PR-2: CI, the Codex review, the owner confirmation, and the merge.
2. Start the next Phase 1 pull request of `docs/roadmaps/phase-1-risk-spikes.md` in a clean session.

## Session records

### Session 3 - 2026-09-28

Author provider: Claude Code

Branch: `feat/pr-2-luna-plan-spike`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and set the go bar (D-101).
- Wrote the Luna plan harness, the fake provider, the fixtures, the schema, and the draft policy rules table, with 44 unit tests in `make test`.
- Ran the paid run after the owner approval. All 60 plans passed the schema, 4 plans broke a rule, and the cost was 0.0564 USD.
- Wrote the report `docs/research/luna-plan-spike.md`.
- Answered Codex finding P2-1 with full merit: each retry of a paid call now stays inside the cap.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-2.

### Session 2 - 2026-09-28

Author provider: Claude Code

Branch: `docs/pr-1-phase-1-roadmap`. Role: author.

Completed:

- Applied the ruleset and the merge settings of `main` after the owner approval, and `make ruleset-check` passed (D-91).
- Asked Q-94 to Q-97 and Q-113 to Q-118, and recorded D-91 to D-100.
- Wrote `docs/roadmaps/phase-1-risk-spikes.md` with PR-1 to PR-6, and moved the start of the development project into work area 1.3 (D-99).
- Changed the branch form of the `one-pr-one-session` skill to D-86.

Open work:

- None. PR #2 merged.

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
