# Gym Route - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-29. PR-9 of `docs/roadmaps/phase-2-platform-skeleton.md`, work area 2.2, is on branch `feat/pr-9-web-shell`, from base `d9b192e`.

The pull request holds the installable web shell in `web/`:

- a React and Vite client on the stack of D-84, with the phone layout alone (D-20), as one npm package (D-135),
- a shell of the dynamic viewport height with the safe areas, and a viewport meta that blocks the pinch zoom (D-120),
- sign-in with email and password on Firebase Authentication, with no form that makes an account (D-75, D-117),
- an empty home screen that calls `GetMe` through the API and shows the uid,
- the Dexie store and the outbox, with one transaction for a change and its outbox entry (D-132),
- the service worker in the `prompt` mode (D-133), and the persistent storage request (D-134),
- the generated TypeScript code in `web/src/gen`, the target `make web`, and the CI job `verify:web` as a required check (D-126, D-127),
- the owner answers Q-148 to Q-151 (D-132 to D-135).

State: `make web` passed: 21 unit tests, and 17 browser tests in WebKit and Chromium. The WebKit copy of the pinch test skips, because Playwright can pinch in Chromium alone. `make verify` passed. The pull request, CI, the live ruleset, and the Codex review are pending.

Next action: open the pull request, read CI, apply the ruleset after the owner approval, and run `make codex-review`.

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
| The newest npm releases: React 19.3.0, Vite 8.3.1, `vite-plugin-pwa` 1.3.0, Dexie 4.4.6, Connect Query 2.3.1, `@bufbuild/protobuf` 2.16.0, and Playwright 1.63.0. TypeScript 7.0.2 exists, and `web/` keeps 5.9.3, as in Decktome. | 2026-09-29 | `npm view`, `web/package.json` |
| In Chromium, a new service worker controls a navigation about 300 ms after it shows the `activated` state. | 2026-09-29 | `web/e2e/shell.spec.ts` |
| Chromium reads the viewport meta at load. A change of the meta after the load does not allow the pinch zoom. | 2026-09-29 | `web/e2e/shell.spec.ts` |
| In Chromium on Linux, `Input.synthesizePinchGesture` zooms no page, and a touch of two fingers through `Input.dispatchTouchEvent` zooms it. The first CI run of PR-9 showed the fault, and the Playwright 1.63.0 image gave the same result. | 2026-09-29 | `web/e2e/shell.spec.ts` |
| Decktome uses the emulator ports 8281, 9199, 9150, 4490, and 4590, and the probe uses 9099. Gym Route uses other ports. | 2026-09-29 | `go/README.md` |
| `firebase deploy --only auth` turns on the email and password provider from the `auth` block of `firebase.json` (firebase-tools 15.32.0). | 2026-09-28 | `tools/spikes/iphone_probe/firebase.json` |
| Commons answers HTTP 429 to a fast client. One request each 2 seconds passes, and it accepts only standard thumbnail widths. | 2026-09-28 | `tools/spikes/recognition_set/download.py` |
| The Commons API adds a tracking query to each file URL. The manifest holds the URL with no query. | 2026-09-28 | `tools/spikes/recognition_set/sources.py` |

## Next steps, in order

1. Close PR-9: CI, the live ruleset, the Codex review, the owner confirmation, and the merge.
2. After the merge, start PR-10 (work area 2.3) in a clean session. Read the billing state first, then ask Q-142.
3. In PR-10, check that the browser key allows the Hosting site of the web app (D-117).

## Session records

### Session 10 - 2026-09-29

Author provider: Claude Code

Branch: `feat/pr-9-web-shell`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-148 to Q-151 (D-132 to D-135).
- Wrote the web shell in `web/` with the patterns of `decktome:web/apps/web/src/lib/api.ts` and `decktome:web/apps/web/e2e/phone.spec.ts`.
- Wrote the browser tests of the acceptance story, with a control that proves the pinch block of the viewport meta.
- Moved the Java 21 lookup of `make emulator-test` into `scripts/java21.sh`, so `make web` uses it too.

Open work:

- CI, the live ruleset, the Codex review, the owner confirmation, and the merge of PR-9.

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

- None. PR #9 merged.

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
