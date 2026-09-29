# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-29. PR-6 of `docs/roadmaps/phase-1-risk-spikes.md` is open on the branch `docs/pr-6-iphone-platform-report`, from the base `d6e3c54`. It is the last pull request of Phase 1.

PR-6 holds the iPhone web platform spike report of work area 1.3:

- the report `docs/research/iphone-platform-spike.md`: go under the bar of D-118,
- the owner answers Q-136 to Q-139, and the decisions D-118 to D-121,
- the result line of PR-6 in the focused roadmap, and the dated results in `docs/research/platform-cloud-and-ai.md`,
- the stage lines of `AGENTS.md` and `README.md`: Phase 1 gave its three reports.

The session deployed the probe by hand from a clean worktree of `main` at `d6e3c54` (D-14). The live page showed `build d6e3c54`. The owner made the probe account in the Firebase console (D-117). The owner ran the checklist in the Home Screen app (D-119), on an iPhone 16 Pro with iOS 27.0 and Chrome 154. Items 1 to 5 passed, and the median first contentful paint of 5 cold starts was 33 ms.

The owner saw two display faults of the probe: the layout sat too high, and a pinch zoomed the page. D-120 makes the app of the roadmap fix both, in work area 2.2.

State: Codex round 1 reviewed effective head `61d9309a63001f3bdf7c6ec82ad0d0a4db5d4659`, with the verdict `Changes required` and finding P2-1. The owner defined a cold start for the bar (D-121). `docs/reviews/pr-7-response.md` holds the answer. `make verify` passed.

Next action: wait for CI on the answer, then run the Codex round 2 with `make codex-review`.

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
| The project `gym-route-dev` exists. Billing is off, and the project has no Storage bucket, no Firestore database, and no Realtime Database. | 2026-09-28 | `gcloud billing projects describe`, `gcloud storage buckets list`, `firebase database:instances:list` |
| The Hosting site `gym-route-dev` at `https://gym-route-dev.web.app` serves the probe of build `d6e3c54`, from a hand deploy of `main`. | 2026-09-29 | `firebase deploy --only hosting`, the header of the live page |
| Chrome 154 on iOS 27.0 adds the probe to the Home Screen, and the app opens in the `standalone` display mode. | 2026-09-29 | `docs/research/iphone-platform-spike.md` |
| GitHub secret scanning flags the Firebase browser key (alert 1). The key allows only the Auth APIs and the probe sites. | 2026-09-29 | `gcloud services api-keys describe`, secret scanning of the repository |
| WebKit refuses a page on port 4190. The probe uses port 4173. | 2026-09-28 | `tools/spikes/iphone_probe/README.md` |
| `firebase deploy --only auth` turns on the email and password provider from the `auth` block of `firebase.json` (firebase-tools 15.32.0). | 2026-09-28 | `tools/spikes/iphone_probe/firebase.json` |
| Commons answers HTTP 429 to a fast client. One request each 2 seconds passes, and it accepts only standard thumbnail widths. | 2026-09-28 | `tools/spikes/recognition_set/download.py` |
| The Commons API adds a tracking query to each file URL. The manifest holds the URL with no query. | 2026-09-28 | `tools/spikes/recognition_set/sources.py` |

## Next steps, in order

1. Close PR-6: CI, the Codex review, the owner confirmation, and the merge.
2. After the merge, Phase 1 ends (section 5 of `docs/roadmaps/phase-1-risk-spikes.md`). Start the focused roadmap of Phase 2 in a clean session. Its ids start at PR-7 (D-96).

## Session records

### Session 7 - 2026-09-29

Author provider: Claude Code

Branch: `docs/pr-6-iphone-platform-report`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-136 and Q-137 (D-118, D-119).
- Deployed the probe by hand from a clean worktree of `main` at `d6e3c54`, after a check of the branch and the commit (D-14).
- Wrote the report `docs/research/iphone-platform-spike.md` from the device results of the owner: a go under D-118.
- Recorded the display faults of the probe (D-120), and answered Codex finding P2-1 with the owner definition of a cold start (D-121).
- Changed the focused roadmap, the platform research, and the stage lines of `AGENTS.md` and `README.md`.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-6.

### Session 6 - 2026-09-28

Author provider: Claude Code

Branch: `feat/pr-5-iphone-platform-probe`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-131 to Q-135 (D-113 to D-117).
- Wrote the probe web app with five pages and no camera page (D-112), 12 browser tests, and 12 configuration tests.
- Added `make probe` and the CI job `verify:probe`.
- The owner signed in to the Firebase CLI again after an HTTP 401. Created the project `gym-route-dev` with the approved steps (D-116). The default Hosting site came with the project.
- Committed the public web configuration and `.firebaserc`. A sign-in from the production build reached the real project in WebKit and Chromium.
- After the alert of GitHub secret scanning, limited the browser key and turned off self sign-up (D-117).
- Answered Codex finding P2-1 with full merit: the stage lines of `AGENTS.md` and `README.md`.

Open work:

- None. PR #6 merged.

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

- None. PR #5 merged.
