# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 1 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR-4 of `docs/roadmaps/phase-1-risk-spikes.md` is GitHub PR #5, open on `feat/pr-4-recognition-spike`, from the base `8a034c9`.

PR-4 holds the recognition spike of work area 1.2:

- the harness, the fake provider, and the tests, in `tools/spikes/recognition/`,
- the results of the paid run, in `tools/spikes/recognition/results/`,
- the report `docs/research/recognition-spike.md`: no-go under the bar of D-107,
- the owner answers Q-125 to Q-130 and the decisions D-107 to D-112.

The owner approved the paid run at run time (D-25). The smoke call and the run of 174 photos cost 0.0651 USD, under the cap of 2 USD (D-94). One photo costs 0.00037 USD (Q-96).

After the no-go, the owner changed the roadmap. Phase 4 gives manual selection and text entry only (D-110). No phase holds photo recognition (D-111). The iPhone probe of PR-5 has no camera page (D-112).

Review: Codex approves effective head `63fe65df968140dead8aae5f497766fd41d8cb47` as Ready for owner merge. Open findings: none.

Next action: the owner confirms the merge, then the author session turns on auto-merge.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. `make ruleset-check` passed. | 2026-09-28 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| `gpt-6-luna` accepts the strict JSON schema of the plan through the Responses API. 60 of 60 plans passed the schema. | 2026-09-28 | `docs/research/luna-plan-spike.md` |
| `gpt-6-luna` accepts an image with a strict JSON schema. One photo at 1536 px costs 0.00037 USD. | 2026-09-28 | `docs/research/recognition-spike.md` |
| Commons answers HTTP 429 to a fast client. One request each 2 seconds passes, and it accepts only standard thumbnail widths. | 2026-09-28 | `tools/spikes/recognition_set/download.py` |
| The Commons API adds a tracking query to each file URL. The manifest holds the URL with no query. | 2026-09-28 | `tools/spikes/recognition_set/sources.py` |

## Next steps, in order

1. Close PR-4: CI, the Codex review, the owner confirmation, and the merge.
2. Start PR-5 of `docs/roadmaps/phase-1-risk-spikes.md` in a clean session. The probe has no camera page (D-112).

## Session records

### Session 5 - 2026-09-28

Author provider: Claude Code

Branch: `feat/pr-4-recognition-spike`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and set the go bar and the photo set (D-107, D-108).
- Wrote the recognition harness, the fake provider, and 19 unit tests in `make test`. The harness uses the call loop and the cap gate of the Luna plan spike.
- Ran the smoke call and the paid run after the owner approval (D-25, D-109). The result is a no-go under D-107, at a cost of 0.0651 USD.
- Wrote the report `docs/research/recognition-spike.md`.
- The owner changed the roadmap after the no-go (D-110 to D-112). Changed the high-level roadmap, the focused roadmap, the design, and the research.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-4.

### Session 4 - 2026-09-28

Author provider: Claude Code

Branch: `feat/pr-3-recognition-test-set`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-120 to Q-123 (D-102 to D-105).
- Wrote the catalog shortlist, the manifest check, the download script, the degrade script, and the source script, with unit tests in `make test`.
- Curated 130 licensed images from 72 gyms, and 44 degraded photos. The manifest records 7 known gaps.
- Answered Codex findings P2-1 to P2-4. The owner answered Q-124 (D-106).

Open work:

- None. PR #4 merged.

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

- None. PR #3 merged.
