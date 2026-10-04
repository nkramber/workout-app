# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-04. Author of pull request PR-37, work area 7.3 of `docs/roadmaps/phase-7-adaptation-loop.md`, on `feat/pr-37-calibration-evaluation` from base `e18781f`.

The deploys of `e18781f` passed: `deploy-api` `87489cf0` and `deploy-web` `83635766`. The live `/version` and `/version.json` name `e18781f`, and the revision `api-00020-cfh` has all traffic.

The owner approved one live revision with policy version 6 (D-212), and approved the milestone (D-12). The owner answered Q-310, Q-308, and Q-311 (D-299 to D-303). The live revision waits for the owner, who finishes a workout on the iPhone. The session then reads the reviser call and the plan with ids and counts alone.

The pull request holds:

- the policy version 7: the first set of a new exercise is the calibration, and a new exercise starts at the estimate,
- the history of a new plan through `Maker.History`, and a deload week with no load step (D-303),
- the flag `first_set_calibration` in the contract, the stores, and the phone,
- the collapse of a finished exercise list (D-298),
- the scenarios G to J in `go/cmd/lunaeval`, the numbers of each scenario, and the paid run of D-302,
- the report `docs/research/reviser-evaluation.md`.

`make verify`, `make go-test`, `make contract`, `make emulator-test`, and `make web` passed. Pending the CI, the Codex review, and the owner merge. Next action: push the first round, open the pull request, and run the Codex review after CI is green.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The paid evaluation of D-302: 50 reviser calls of the 10 scenarios with policy version 7 and xhigh, all `ok`. The check accepted 110 reasons and refused 5. Cost 0.0179 USD, the longest call 10.8 s. | 2026-10-04 | `docs/research/reviser-evaluation.md` |
| The live `/version.json` and `/version` name `e18781f6cf201cf0a7028eee44c95b906c4813d2`, from the builds `deploy-web` `83635766` and `deploy-api` `87489cf0`. The revision `api-00020-cfh` has all traffic. | 2026-10-04 | `curl`, `gcloud builds list`, `gcloud run services describe` |
| The live revision of `d8f766e`: 1 reviser call of 0.0017 USD with the status `ok`, in a `SyncOutbox` call of 19.8 s. The plan got 8 new targets, each with a reason of Luna. Each of the 16 exercises of the plan has a calibration set. | 2026-10-04 | `gcloud logging read`, `users/{uid}/plan/active` |
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

1. Close PR-37: CI, the Codex review, the live revision of the owner, the owner confirmation, and the merge.
2. After the merge, read the deploys of the merge.
3. Ask the owner to check the first-set calibration and the collapse on the iPhone. Phase 7 then ends.

## Session records

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

Open work:

- The live revision of the owner, CI, the Codex review, the owner confirmation, and the merge of PR-37.

### Session 37 - 2026-10-04

Author provider: Claude Code

Branch: `feat/pr-36-overrides-disruptions`. Role: author.

Completed:

- Read the deploys of `d8f766e`. The owner approved one live revision (D-212), and the session read it.
- The owner approved the milestone (D-12), and answered Q-305 to Q-307 and Q-309 (D-293 to D-296).
- Recorded the requests of the owner for PR-37 (D-297, D-298, Q-310).
- Answered P2-1 and P2-2 of Codex round 1.
- Added the policy version 6, the deload dates, the targets on a date, and the overrides.
- Added `TestDisruptionAcceptanceStory`, unit tests, and a browser test.
- Changed the registers, both roadmaps, the design, both READMEs, and `AGENTS.md`.

Open work:

- None. GitHub PR 37 merged as `e18781f`.

### Session 36 - 2026-10-04

Author provider: Claude Code

Branch: `feat/pr-35-revision`. Role: author.

Completed:

- The owner approved the milestone (D-12), and answered Q-302 to Q-304 (D-290 to D-292).
- Added the target copy to the workout header, and the revision of the plan in the sync of a finished workout.
- Added the reviser of the reason alone, with its check (D-288).
- Added the screen "Workout done" with the next targets and their reasons, and the target copy in the header of the phone.
- Added the emulator tests of scenarios A to F, unit tests, and a browser test.
- Changed the registers, both roadmaps, the design, both READMEs, and `AGENTS.md`.

Open work:

- None. GitHub PR 36 merged as `d8f766e`.
