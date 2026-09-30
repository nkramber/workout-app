# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-30. GitHub PR 12 is roadmap PR-11 of `docs/roadmaps/phase-2-platform-skeleton.md`, work areas 2.2 and 2.3, on branch `docs/pr-11-phase-2-check`, from base `df79c16`. It holds the exit evidence of Phase 2.

The pull request holds these concerns:

- the deploy check: `/version` and `/version.json` name `df79c16`, the merge commit of PR-10,
- the repair of `deploy-rules`: `rules-deployer` got the viewer role of Service Usage, and the build of `df79c16` passed on a retry (D-146),
- the device check on the iPhone: items 1 to 5 pass,
- the header space of 16 px that keeps the title out of the iOS blur band, with a Chromium test (D-145),
- the shutdown of `gym-route-dev` (D-137),
- `PACKAGE_NO_DELETE` and `PACKAGE_SERVICE_NO_DELETE` in `buf.yaml`, with `scripts/package_move_probe.sh` in `make contract` (D-138),
- the report `docs/research/phase-2-check.md`.

State: Codex reviewed effective head `4d643e3b81061535cc69591f568a4a295e04bfe8` and set the verdict to `Ready for owner merge`. No finding remains open. GitHub CI passed all checks except `review-gate`, which failed because this record was absent. The record and this hand-off update await publication.

Next action: publish the review record and hand-off, verify the `review-gate` result, and wait for the owner confirmation and merge of PR-11. The live web app gets the header space after the merge, through `deploy-web`.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| The owner ended the D-4 period. `OVERRIDE_ENABLED` is `True` on `main`. | 2026-09-29 | D-125, `docs/tools/review_gate.py` |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. The ruleset requires `verify:contract`, `verify:go`, `verify:emulator`, and `verify:web` too. `make ruleset-check` passed. | 2026-09-29 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| `gpt-6-luna` costs 0.10 USD per million input tokens and 0.50 USD per million output tokens. | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| `gpt-6-luna` accepts the strict JSON schema of the plan through the Responses API. 60 of 60 plans passed the schema. | 2026-09-28 | `docs/research/luna-plan-spike.md` |
| `gpt-6-luna` accepts an image with a strict JSON schema. One photo at 1536 px costs 0.00037 USD. | 2026-09-28 | `docs/research/recognition-spike.md` |
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `df79c16`, three triggers, and the allowlist entry. | 2026-09-30 | `docs/setup-gcp.md`, `/version` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 03:54:47Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` and `/version.json` name `df79c169ae38326c0a876ac678a5465f80103d13`. The rules release of `df79c16` has the update time 02:33:00Z. | 2026-09-30 | `docs/research/phase-2-check.md` |
| `rules-deployer` holds `roles/firebaserules.admin`, `roles/serviceusage.serviceUsageViewer`, and `roles/logging.logWriter`. | 2026-09-30 | `gcloud projects get-iam-policy` |
| Since iOS 26, a Home Screen app blurs a band below the status bar, and a page can not turn it off. On iOS 27.0 with Chrome 154, 16 px above the title keeps it sharp. | 2026-09-30 | `docs/research/phase-2-check.md` |
| Google Cloud SDK 533.0.0 has no `gcloud builds retry`. The Cloud Build API call `builds/{id}:retry` works. | 2026-09-30 | `docs/deploy-and-rollback.md` |
| Chrome 154 on iOS 27.0 adds the probe to the Home Screen, and the app opens in the `standalone` display mode. | 2026-09-29 | `docs/research/iphone-platform-spike.md` |
| The browser key of `nk-workout-app-prod` allows only the Auth APIs and the sites of the project and the local ports. GitHub secret scanning flagged the old key (alert 1). | 2026-09-29 | `gcloud services api-keys describe` |
| WebKit refuses a page on port 4190. The probe uses port 4173. | 2026-09-28 | `tools/spikes/iphone_probe/README.md` |
| Go 1.27.1, buf v1.73.0, connect v1.21.0, and firebase-admin-go v4.22.0 are the newest releases. | 2026-09-29 | go.dev and proxy.golang.org, `go/go.mod` |
| The spend cap of Cloud Run is Preview, and the console alone sets it. | 2026-09-29 | Google Cloud, "Spend cap budgets" |
| `buf breaking` with the rules of `buf.yaml` passes a move of the package `gymroute.v1` to `workoutapp.v1`. | 2026-09-29 | `make contract` |
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

1. Close PR-11: CI, the Codex review, the owner confirmation, and the merge.
2. After the merge, read `deploy-web` for the merge commit, and read `/version.json` again.
3. Start the first work area of Phase 3 of `docs/roadmaps/high-level-roadmap.md` in a clean session.

## Session records

### Session 12 - 2026-09-30

Author provider: Claude Code

Branch: `docs/pr-11-phase-2-check`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-161 to Q-164 (D-144 to D-146).
- Read the live endpoints and the builds. Repaired `deploy-rules` after the owner approval, and read the live rules release back.
- The owner ran the device check on the iPhone. Items 1 to 5 pass.
- Tested four fixes of the iOS blur band on the iPhone, from the Mac on the local network. The 16 px header space works (D-145).
- Shut down `gym-route-dev` after the owner approval (D-137).
- Added the package rules and their probe to `make contract` (D-138).

Open work:

- CI, the Codex review, the owner confirmation, and the merge of PR-11.

### Session 11 - 2026-09-29

Author provider: Claude Code

Branch: `feat/pr-10-deploy`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12). Then the owner added the new name and the new project, and approved the changed milestone.
- Read the billing state, then asked Q-142 and Q-152 to Q-159 (D-136 to D-142).
- Made each part of the project with the approval of the owner at run time, and read each part back.
- Renamed each current file. The past records keep the old name (D-136).
- Wrote the deploy files, the setup and rollback documents, and the tests. The rules test fails with open rules.
- Answered P2-1 of the Codex review with a guard, then with a lock for each part (D-143). Answered P2-2 with a time limit on the read of `main`.

Open work:

- None. PR #11 merged.

### Session 10 - 2026-09-29

Author provider: Claude Code

Branch: `feat/pr-9-web-shell`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-148 to Q-151 (D-132 to D-135).
- Wrote the web shell in `web/` with the patterns of `decktome:web/apps/web/src/lib/api.ts` and `decktome:web/apps/web/e2e/phone.spec.ts`.
- Wrote the browser tests of the acceptance story, with a control that proves the pinch block of the viewport meta.
- Moved the Java 21 lookup of `make emulator-test` into `scripts/java21.sh`, so `make web` uses it too.
- Changed the pinch test to a touch of two fingers after the first CI run, because Chromium on Linux ignores `Input.synthesizePinchGesture`.
- Applied `verify:web` as a required check to the live ruleset after the owner approval (D-127). `make ruleset-check` passed.

Open work:

- None. PR #10 merged.
