# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 1 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR-2 merged as GitHub #3. PR-3 of `docs/roadmaps/phase-1-risk-spikes.md` is in progress on the branch `feat/pr-3-recognition-test-set`, from the base `89046b4`, in the worktree `../workout-app-pr3`. No commit and no pull request exist yet.

The owner approved the milestone of PR-3 before the first edit (D-12). The owner answered Q-120 to Q-122 (D-102 to D-104). D-104 permits a photo with a face when few other photos exist for a catalog type.

Done in the worktree, not committed:

- `tools/spikes/recognition_set/`: `catalog.json` (26 types), `manifest_check.py`, `download.py`, `degrade.py`, `commons.py`, `test_recognition_set.py`, and `README.md`.
- Q-120 to Q-122 and D-102 to D-104 in the registers.

Open work:

1. Finish the image curation. The author reviews Commons thumbnails, and keeps each image with no face and a clear machine type. The session scratchpad holds the candidate pool and the picks. On 2026-09-28, the picks held 115 images. Seated row, triceps extension, back extension, and crunch have no image. Six other types have images from one source only.
2. Decide the catalog. Drop each type with images from fewer than two gyms, or accept the gap (D-103).
3. Write the selection, run `commons.py build`, and write `manifest.json` with the gyms, the splits, and the degraded photos.
4. Run `make verify`, the download script, and the degrade script. Then commit, push, and open the pull request.

Commons refuses a fast client with HTTP 429. One request each 2 seconds passes.

Q-91 to Q-93 and Q-98 to Q-107 stay open, and each names the phase that needs it.

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
