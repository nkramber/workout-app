# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-29. GitHub pull request 9 is open on branch `feat/pr-8-api-skeleton`, from base `7aa7bd4`. Its id is PR-8 in `docs/roadmaps/phase-2-platform-skeleton.md`, work area 2.1.

The pull request holds the contract and the API skeleton:

- the buf v2 contract `proto/gymroute/v1/user_service.proto` with one call `GetMe`, and the generated Go code in `go/gen`,
- the Go API in `go/`: the token check, the allowlist of uids (D-131), and CORS for one origin (D-82),
- the route `/version` of the build commit, and the start guard of D-129,
- the Auth and Firestore emulators in `firebase.json`, with `firebase-tools` pinned in `emulators/`,
- the targets `make proto`, `make contract`, `make go-test`, and `make emulator-test`, and the CI jobs `verify:contract`, `verify:go`, and `verify:emulator` (D-126),
- the three jobs as required checks in `.github/rulesets/review-gate.json` (D-127),
- the owner answers Q-145 to Q-147 (D-129 to D-131).

State: `make verify`, `make contract`, `make go-test`, and `make emulator-test` passed on the machine of the owner. Each CI job of the first push passed, except `review-gate`, which waits for the Codex record. The owner approved the change of the live ruleset at run time. The session applied it, and `make ruleset-check` passed.

Next action: run the Codex review, then answer each finding.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| The owner ended the D-4 period. `OVERRIDE_ENABLED` is `True` on `main`. | 2026-09-29 | D-125, `docs/tools/review_gate.py` |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. The ruleset requires `verify:contract`, `verify:go`, and `verify:emulator` too. `make ruleset-check` passed. | 2026-09-29 | `make ruleset-check` |
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
| Go 1.27.1, buf v1.73.0, connect v1.21.0, and firebase-admin-go v4.22.0 are the newest releases. | 2026-09-29 | go.dev and proxy.golang.org, `go/go.mod` |
| `firebase-tools` 15.32.0 refuses a Java version before 21. The Firestore emulator is v1.22.0. | 2026-09-29 | `make emulator-test` |
| Decktome uses the emulator ports 8281, 9199, 9150, 4490, and 4590, and the probe uses 9099. Gym Route uses other ports. | 2026-09-29 | `go/README.md` |
| `firebase deploy --only auth` turns on the email and password provider from the `auth` block of `firebase.json` (firebase-tools 15.32.0). | 2026-09-28 | `tools/spikes/iphone_probe/firebase.json` |
| Commons answers HTTP 429 to a fast client. One request each 2 seconds passes, and it accepts only standard thumbnail widths. | 2026-09-28 | `tools/spikes/recognition_set/download.py` |
| The Commons API adds a tracking query to each file URL. The manifest holds the URL with no query. | 2026-09-28 | `tools/spikes/recognition_set/sources.py` |

## Next steps, in order

1. Close PR-8: the Codex review, the owner confirmation, and the merge.
2. After the merge, start PR-9 (work area 2.2) in a clean session, with `docs/roadmaps/phase-2-platform-skeleton.md` section 4.

## Session records

### Session 9 - 2026-09-29

Author provider: Claude Code

Branch: `feat/pr-8-api-skeleton`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-145 to Q-147 (D-129 to D-131).
- Wrote the contract, the Go API skeleton, and its unit tests, with the patterns of `decktome:go/internal/auth/auth.go` and `decktome:go/cmd/api/main.go`.
- Wrote the emulator test of the acceptance story: no token, a bad token, and a token of another project give `unauthenticated`. A uid outside the allowlist gives `permission_denied`. An allowed uid gets its uid back.
- Proved that `make contract` refuses a changed field number, with a copy of the tree as the base.
- Installed the Homebrew keg `openjdk@21` on the machine of the owner, because the Firestore emulator needs Java 21.
- Applied the three new required checks to the live ruleset after the owner approval (D-127). `make ruleset-check` passed.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-8.

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
- None for PR-7. PR #8 merged.

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
