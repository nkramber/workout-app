# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-05. Author of pull request PR-39 of `docs/roadmaps/phase-8-personal-use-operations.md`, on the branch `docs/pr-39-restore-drill` from base `e09b7ff`.

The deploys of `e09b7ff` passed: `deploy-web` `ddf4e3c4` and `deploy-api` `12c94777`. The live `/version` and `/version.json` name `e09b7ff`, and the revision `api-00022-tjh` has all traffic.

The owner approved the milestone (D-12), the live check, and the restore (D-212). The three iPhone checks of Phase 7 passed, so Phase 7 ended (D-310). The check made 1 planner call and 3 reviser calls, each `ok`, for 0.0044 USD. The owner approved a read of the plan document, and each reason of the rules came from a skipped exercise.

The pull request holds:

- the report `docs/research/restore-drill.md`, with times and counts alone,
- the restore of the daily backup of 07:28Z into `restore-20261005`, with the count of each collection, and its delete,
- the rollback drill of `api` and of Hosting,
- the changes of `docs/deploy-and-rollback.md` and `docs/setup-gcp.md` that the drill found,
- the result of the check in both roadmaps of Phase 7 and Phase 8, the high-level roadmap, and `AGENTS.md`.

The restore took 8 min 48 s. Each difference of a count comes from the "Delete all data" of the owner at 18:52Z. The restored database had delete protection, so section 6 of the runbook now turns it off first. The Hosting rollback now has REST steps.

`make verify` passed. Codex reviewed GitHub PR 40 at effective head `a019d5caa7b7711fe327be94afaa6e62235a458b`. P2-1 is fixed in `docs/research/restore-drill.md`. P2-2 names the missing operation id in step 5 of `docs/deploy-and-rollback.md`. All checks passed except `review-gate`, which rejects this verdict.

Next action: correct step 5, then run Codex round 3.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The paid evaluation of D-302: 50 reviser calls of the 10 scenarios with policy version 7 and xhigh, all `ok`. The check accepted 110 reasons and refused 5. Cost 0.0179 USD, the longest call 10.8 s. | 2026-10-04 | `docs/research/reviser-evaluation.md` |
| The live `/version.json` and `/version` name `e09b7ff0a4b788ce83d9c64e8d21607d861b8714`, from the builds `deploy-web` `ddf4e3c4` and `deploy-api` `12c94777`. The revision `api-00022-tjh` has all traffic. | 2026-10-05 | `curl`, `gcloud builds list`, `gcloud run services describe` |
| The Phase 7 check of `e09b7ff`: 1 `DeleteHistory` of 11 workouts, 1 planner call, and 3 reviser calls, each `ok`, for 0.004379230 USD. No log entry had the severity WARNING or more. | 2026-10-05 | `docs/research/restore-drill.md` |
| The restore of a daily backup took 8 min 48 s, and the restored database had delete protection. The drill deleted it. `(default)` is the one database. | 2026-10-05 | `docs/research/restore-drill.md` |
| `nk-workout-app-prod` has 0 alert policies and 0 notification channels of Cloud Monitoring. | 2026-10-04 | Monitoring API |
| 6 daily backups have the state READY, from 2026-09-30 to 2026-10-05. | 2026-10-05 | `gcloud firestore backups list` |
| The live check of `4bbf6c8`: 1 planner call of 0.0045 USD and 1 reviser call of 0.0014 USD, each `ok`. The revision of 8 exercises gave 1 reason of Luna and 7 reasons of the rules. No log entry had the severity WARNING or more. | 2026-10-04 | `gcloud logging read` |
| The repository is public. | 2026-09-27 | GitHub repository settings |
| The owner ended the D-4 period. `OVERRIDE_ENABLED` is `True` on `main`. | 2026-09-29 | D-125, `docs/tools/review_gate.py` |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. The ruleset requires `verify:contract`, `verify:go`, `verify:emulator`, and `verify:web` too. `make ruleset-check` passed. | 2026-09-29 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| In the short context of 272K input tokens or fewer, `gpt-6-luna` costs 0.10 USD per million input tokens, 0.01 USD cached, 0.125 USD for a cache write, and 0.50 USD per million output tokens. | 2026-10-01 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| `gpt-6-luna` accepts the strict JSON schema of the plan through the Responses API. 60 of 60 plans passed the schema. | 2026-09-28 | `docs/research/luna-plan-spike.md` |
| `gpt-6-luna` accepts an image with a strict JSON schema. One photo at 1536 px costs 0.00037 USD. | 2026-09-28 | `docs/research/recognition-spike.md` |
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `e18781f`, three triggers, and the allowlist entry. | 2026-10-04 | `docs/setup-gcp.md`, `/version` |
| The service `api` revision `api-00013-w8w` holds `OPENAI_API_KEY` from `openai-api-key:latest`, `LUNA_CAP_USER_USD=1`, `LUNA_CAP_PROJECT_USD=2`, and a request timeout of 420 s. The TTL policy of `aiErrors.expire_at` is `ACTIVE`. | 2026-10-03 | `gcloud run services describe`, `gcloud firestore fields ttls list` |
| A Cloud Run service with a secret needs the accessor role for its service identity alone. The page lists `roles/run.admin` for the deployer. | 2026-10-02 | Cloud Run docs, "Configure secrets for services" |
| The secret `openai-api-key` has version 1, enabled. A free call to the OpenAI model list with it gave HTTP 200 and lists `gpt-6-luna`. | 2026-10-01 | `gcloud secrets versions list`, `curl` |
| With the same 50 calls of `go/cmd/lunaeval`, xhigh cost 0.0740 USD and medium 0.0426 USD. The longest xhigh call took 63.6 s, and each effort passed 50 of 50. | 2026-10-03 | `docs/research/luna-effort-check.md` |
| A live plan at xhigh of `luna-prompt-v5` and policy version 5 took 1 call of 0.0037 USD, with 2 sessions of 8 exercises each. Each exercise has a rest of 60 seconds, and a calibration row for each weight of its machine. | 2026-10-04 | `gcloud logging read`, `users/{uid}/plan/active` |
| WebKit fixed the Screen Wake Lock for Home Screen apps in iOS 18.4. On iOS 27.0, the lock worked in the Home Screen app of the probe, from Chrome 154. | 2026-10-03 | PC-2, PC-8, `docs/research/iphone-platform-spike.md` |
| On iOS 27.0, the workout of `6274c19` kept the screen on after a return with no tap, and showed no notice. | 2026-10-04 | The owner, on the iPhone |
| In the Home Screen app on iOS 27.0 with Chrome 154, `100dvh` leaves out the band of the status bar. A shell of `100dvh` ended 62 pt above the bottom edge of an iPhone 16 Pro. | 2026-10-03 | The screenshot of the owner, `web/src/lib/app-height.ts` |
| Chromium of Playwright 1.63.0 does not apply the display mode `standalone` of `Emulation.setEmulatedMedia`. | 2026-10-03 | `web/e2e/shell.spec.ts` |
| `gpt-6-luna` at medium effort passed the schema `luna_plan_v2` in 50 of 50 calls. A planner call cost 0.0017 USD on average, and the longest call took 33.0 s. | 2026-10-02 | `docs/research/phase-3-check.md` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| WebKit of Playwright 1.63.0 refuses a navigation of an offline context with "WebKit encountered an internal error", also when the service worker can serve the page. Chromium serves it from the service worker. | 2026-10-03 | `web/e2e/support.ts` |
| In WebKit of Playwright 1.63.0, a route of the context did not apply to the API calls of a page that opened again after a stop. The service worker controls such a page (unverified cause). | 2026-10-03 | `web/e2e/support.ts` |
| With 10 workers of Playwright on a Mac of 10 cores, 5 Chromium tests of `make web` waited more than 10 s for a plan of the fake in 1 of 2 runs. The plan requests share the cap documents. More reads of the emulator in each `GetPlan` made such waits occur in 4 of 5 runs. | 2026-10-04 | `make web` |
| In the Firestore emulator v1.22.0, 10 transactions at the same time on one document abort with "Transaction lock timeout" after about 3 s, and the Go client tries again. With 5 attempts, a test of 10 transactions took up to 19 s. | 2026-10-02 | `go/internal/capstore/firestore_emulator_test.go` |
| `api-runtime` holds `roles/datastore.user`, and no other project role. | 2026-10-02 | `gcloud projects get-iam-policy` |
| With `context.setOffline(true)` of Playwright 1.63.0, a call of the API fails in WebKit and in Chromium. Connect gives the code `unknown` with a `TypeError` as its cause. | 2026-10-02 | `web/e2e/inventory.spec.ts` |
| `rules-deployer` holds `roles/firebaserules.admin`, `roles/serviceusage.serviceUsageViewer`, and `roles/logging.logWriter`. | 2026-09-30 | `gcloud projects get-iam-policy` |
| Since iOS 26, a Home Screen app blurs a band below the status bar, and a page can not turn it off. On iOS 27.0 with Chrome 154, 16 px above the title keeps it sharp. The owner saw the fix in the live app of `73b1964`. | 2026-09-30 | `docs/research/phase-2-check.md`, the owner |
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

1. Close PR-39: CI, the Codex review, the owner confirmation, and the merge.
2. Start PR-40 of `docs/roadmaps/phase-8-personal-use-operations.md`, the alerts of work area 8.2, in a clean session.

## Session records

### Session 40 - 2026-10-05

Author provider: Claude Code

Branch: `docs/pr-39-restore-drill`. Role: author.

Completed:

- Read the deploys of `e09b7ff`. The owner approved the milestone (D-12), the live check, and the restore (D-212).
- The owner passed the three iPhone checks of Phase 7 (D-310). Read the logs and, with the approval of the owner, the plan document.
- Restored a daily backup into a new database, counted each collection, and deleted it.
- Moved the traffic of `api` and the release of Hosting to the last version and back.
- Wrote `docs/research/restore-drill.md`, and changed both runbooks, three roadmaps, and `AGENTS.md`.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-39.

### Session 39 - 2026-10-05

Author provider: Claude Code

Branch: `docs/pr-38-phase-8-roadmap`. Role: author.

Completed:

- Read the deploys of `4bbf6c8` and the logs of the live check of the owner.
- The owner approved the milestone (D-12), added the fix (D-310), and answered Q-313 to Q-319 (D-306 to D-313).
- Added the policy version 8, the limit in the contract, the stores, and the phone, and the read of the followed load.
- Added the unit tests, the golden files, and a browser test of the live case.
- Wrote the focused roadmap of Phase 8, and changed the registers, both roadmaps, the design, both READMEs, and `AGENTS.md`.
- Answered P2-1 of Codex round 1. The owner then added "Delete all data" (D-314 to D-316), with its tests.

Open work:

- None. GitHub PR 39 merged as `e09b7ff`.

### Session 38 - 2026-10-04

Author provider: Claude Code

Branch: `feat/pr-37-calibration-evaluation`. Role: author.

Completed:

- Read the deploys of `e18781f`. The owner approved one live revision (D-212) and the milestone (D-12).
- The owner answered Q-310, Q-308, and Q-311 (D-299 to D-303).
- Added the policy version 7, the history of a new plan, and the flag in the contract and the phone.
- Added the collapse of a finished workout.
- Added the scenarios G to J and the numbers of each scenario to `go/cmd/lunaeval`, and made the paid run of D-302.
- Wrote `docs/research/reviser-evaluation.md`, and changed the registers, both roadmaps, the design, both READMEs, and `AGENTS.md`.
- Answered P2-1 to P2-3 of Codex, and read the live check of the owner. Added the claim of D-304.

Open work:

- None. GitHub PR 38 merged as `4bbf6c8`.
