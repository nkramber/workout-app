# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-28. Phase 1 of `docs/roadmaps/high-level-roadmap.md` is in progress. PR-5 of `docs/roadmaps/phase-1-risk-spikes.md` is open on `feat/pr-5-iphone-platform-probe`, from the base `abf6a09`.

PR-5 holds the iPhone web platform probe of work area 1.3:

- the probe web app, its browser tests, and its configuration tests, in `tools/spikes/iphone_probe/`,
- the target `make probe` and the CI job `verify:probe` (D-113, D-114),
- the development project `gym-route-dev` with Firebase Hosting and Firebase Authentication only (D-99, D-116),
- the owner answers Q-131 to Q-134 and the decisions D-113 to D-116.

State: the probe builds, and its 12 browser tests pass in WebKit and Chromium on the owner machine. The Firebase CLI login of the owner machine gave HTTP 401, so the project does not exist yet. The owner must sign in to the Firebase CLI again.

Next action: after the owner signs in again, create the project with the four approved steps of D-116. Then commit the web configuration and `.firebaserc`, and record the evidence.

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
| WebKit refuses a page on port 4190. The probe uses port 4173. | 2026-09-28 | `tools/spikes/iphone_probe/README.md` |
| `firebase deploy --only auth` turns on the email and password provider from the `auth` block of `firebase.json` (firebase-tools 15.32.0). | 2026-09-28 | `tools/spikes/iphone_probe/firebase.json` |
| Commons answers HTTP 429 to a fast client. One request each 2 seconds passes, and it accepts only standard thumbnail widths. | 2026-09-28 | `tools/spikes/recognition_set/download.py` |
| The Commons API adds a tracking query to each file URL. The manifest holds the URL with no query. | 2026-09-28 | `tools/spikes/recognition_set/sources.py` |

## Next steps, in order

1. Create the project `gym-route-dev` after the owner signs in to the Firebase CLI again (D-116).
2. Close PR-5: CI, the Codex review, the owner confirmation, and the merge.
3. Start PR-6 of `docs/roadmaps/phase-1-risk-spikes.md` in a clean session. Deploy the probe by hand from `main` first (D-14).

## Session records

### Session 6 - 2026-09-28

Author provider: Claude Code

Branch: `feat/pr-5-iphone-platform-probe`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-131 to Q-134 (D-113 to D-116).
- Wrote the probe web app with five pages and no camera page (D-112), 12 browser tests, and 11 configuration tests.
- Added `make probe` and the CI job `verify:probe`.

Open work:

- The project `gym-route-dev`, the web configuration, the Codex review, the owner confirmation, and the merge of PR-5.

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
