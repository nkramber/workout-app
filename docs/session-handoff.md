# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 1 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR-3 of `docs/roadmaps/phase-1-risk-spikes.md` is GitHub PR #4, open on `feat/pr-3-recognition-test-set`. Its base and merge base are `89046b4219827aa3dbd5b505c228adcea48cf149`. Its reviewed effective head is `81f74d624feebc88bd20f602a707ff45f9735c69`.

The Codex review record is `docs/reviews/pr-4.md`. Verdict: `Changes required`, with open finding P2-1 on source URL and license URL verification. `make verify`, the manifest check, the source check, and the 130-image download and hash check passed. Two image tests skipped because Pillow is not installed in this environment. At the published metadata head, GitHub `review-gate` failed RG 4 because the verdict is `Changes required`. The other checks passed.

Next action: correct P2-1. Add URL mismatch regression tests. Push the correction and request a repeat review. Keep P2-1 and the earlier verdict in the same record. After the correction, refresh the checks and review-gate result.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. `make ruleset-check` passed. | 2026-09-28 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| `gpt-6-luna` accepts the strict JSON schema of the plan through the Responses API. 60 of 60 plans passed the schema. | 2026-09-28 | `docs/research/luna-plan-spike.md` |
| Commons answers HTTP 429 to a fast client. One request each 2 seconds passes, and it accepts only standard thumbnail widths. | 2026-09-28 | `tools/spikes/recognition_set/download.py` |
| The Commons API adds a tracking query to each file URL. The manifest holds the URL with no query. | 2026-09-28 | `tools/spikes/recognition_set/sources.py` |

## Next steps, in order

1. Close PR-3: CI, the Codex review, the owner confirmation, and the merge.
2. Start PR-4 of `docs/roadmaps/phase-1-risk-spikes.md` in a clean session. It needs PR-3 on `main`, the download script, and the degrade script.

## Session records

### Session 4 - 2026-09-28

Author provider: Claude Code

Branch: `feat/pr-3-recognition-test-set`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-120 to Q-123 (D-102 to D-105).
- Wrote the catalog shortlist, the manifest check, the download script, the degrade script, and the source script, with unit tests in `make test`.
- Curated 130 licensed images from 72 gyms, and 44 degraded photos. The manifest records 7 known gaps.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-3.

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
