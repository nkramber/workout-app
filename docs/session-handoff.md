# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-29. GitHub pull request 8 is open on branch `docs/pr-7-phase-2-roadmap`, from base `71b5220`. Its id is PR-7 in `docs/roadmaps/phase-2-platform-skeleton.md`.

The pull request holds the focused roadmap of Phase 2, with PR-8 to PR-11 (D-128). It records the owner answers D-122 to D-128. It also ends the D-4 period (D-125): `OVERRIDE_ENABLED` in `docs/tools/review_gate.py` is `True`, and the rule text changes with it.

State: Codex round 1 reviewed effective head `f1a9321`, with the verdict `Changes required` and finding P2-1 (owner billing details in public files). The correction removes the details, and `docs/reviews/pr-8-response.md` holds the answer. `make verify` passed.

Next action: wait for green CI, and run round 2 of `make codex-review PR=8`.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| The owner ended the D-4 period. `OVERRIDE_ENABLED` is `True` on the branch of PR-7, and `False` on `main` until PR-7 merges. | 2026-09-29 | D-125, `docs/tools/review_gate.py` |
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

1. Close PR-7: CI, the Codex review, the owner confirmation, and the merge.
2. After the merge, start PR-8 (work area 2.1) in a clean session, with `docs/roadmaps/phase-2-platform-skeleton.md` section 4.

## Session records

### Session 8 - 2026-09-29

Author provider: Claude Code

Branch: `docs/pr-7-phase-2-roadmap`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-91, Q-93, Q-99, Q-100, and Q-140 to Q-144 (D-122 to D-128).
- Wrote the focused roadmap `docs/roadmaps/phase-2-platform-skeleton.md`.
- Set `OVERRIDE_ENABLED` to `True`, with the tests of the gate, and changed the rule text of the D-4 period (D-125).
- Changed the high-level roadmap: the exit evidence of work area 2.1 (D-126), the backups of work area 2.3 (D-124), and the cited ids.

Open work:

- Q-142, the billing account of work area 2.3. The session of PR-10 reads the billing state, then asks.
- The Codex review, the owner confirmation, and the merge of PR-7.

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

- None. PR #7 merged.

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
