# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-03. Roadmap PR-29 of `docs/roadmaps/phase-6-guided-workout.md`, on branch `feat/pr-29-workout-api`, from base `fe7fe20`. GitHub PR 30.

Before the work, the session read the deploy of `fe7fe20`. The build `deploy-api` `1bfa4330` and the build `deploy-web` `9e3f4aff` gave `SUCCESS`, and the live `/version` and `/version.json` both name `fe7fe20` (D-137).

The owner approved the live check (D-212). One plan at xhigh took 1 call of 0.0022 USD and 29.7 s, with no violation. The owner saw the plan to the bottom edge (D-120). Section 1.4 of the roadmap records it. The owner then asked for 20 to 30 minutes of cardio in each session (D-254, D-255).

The owner approved the milestone, then widened it with the cardio rule (D-12, D-254). The pull request holds:

- the workout service of `proto/workoutapp/v1/workout_service.proto`, with `SyncOutbox` and `ListWorkouts`,
- `go/internal/workout` and `go/internal/workoutsvc`: the check of each entry, and one apply of each op id,
- the Firestore store at the paths of D-256, with no end date for an op id (D-257),
- the phone rule of D-258, the batch limit of D-259, the cardio log of D-260, and the note limit of D-261,
- the cardio rule of D-255 in prompt v4, the output check, the fake provider, and the plan screen.

`make contract`, `make go-test`, `make emulator-test`, `make web`, and `make verify` passed.

Codex round 2 reviewed effective head `a20f04b`. The review withdrew P2-1 after the focused test confirmed the nil plan refusal. It found P2-2, a version collision across entity types. `docs/reviews/pr-30.md` holds the record. State: changes required. Next action: fix P2-2 and ask Codex to review the new effective head (D-8).

The merge changes `go/` and `web/`, so it deploys both. The session of PR-30 reads both deploys first. The next live plan gives 20 to 30 minutes of cardio in each session. The stored plan of 10 minutes stays until the owner requests a new one.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The repository is public. | 2026-09-27 | GitHub repository settings |
| The owner ended the D-4 period. `OVERRIDE_ENABLED` is `True` on `main`. | 2026-09-29 | D-125, `docs/tools/review_gate.py` |
| `main` has the live ruleset `review-gate` and the merge settings of `.github/rulesets`. The ruleset requires `verify:contract`, `verify:go`, `verify:emulator`, and `verify:web` too. `make ruleset-check` passed. | 2026-09-29 | `make ruleset-check` |
| The review-gate workflow runs from `main`, so it runs on each pull request. | 2026-09-28 | `.github/workflows/review-gate.yml` |
| In the short context of 272K input tokens or fewer, `gpt-6-luna` costs 0.10 USD per million input tokens, 0.01 USD cached, 0.125 USD for a cache write, and 0.50 USD per million output tokens. | 2026-10-01 | `docs/research/platform-cloud-and-ai.md` |
| The OpenAI usage policies page returns HTTP 403. The owner accepted a copy printed on 2025-11-07 (D-93). | 2026-09-28 | `docs/research/platform-cloud-and-ai.md` |
| `gpt-6-luna` accepts the strict JSON schema of the plan through the Responses API. 60 of 60 plans passed the schema. | 2026-09-28 | `docs/research/luna-plan-spike.md` |
| `gpt-6-luna` accepts an image with a strict JSON schema. One photo at 1536 px costs 0.00037 USD. | 2026-09-28 | `docs/research/recognition-spike.md` |
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `fe7fe20`, three triggers, and the allowlist entry. | 2026-10-03 | `docs/setup-gcp.md`, `/version` |
| The service `api` revision `api-00013-w8w` holds `OPENAI_API_KEY` from `openai-api-key:latest`, `LUNA_CAP_USER_USD=1`, `LUNA_CAP_PROJECT_USD=2`, and a request timeout of 420 s. The TTL policy of `aiErrors.expire_at` is `ACTIVE`. | 2026-10-03 | `gcloud run services describe`, `gcloud firestore fields ttls list` |
| A Cloud Run service with a secret needs the accessor role for its service identity alone. The page lists `roles/run.admin` for the deployer. | 2026-10-02 | Cloud Run docs, "Configure secrets for services" |
| The secret `openai-api-key` has version 1, enabled. A free call to the OpenAI model list with it gave HTTP 200 and lists `gpt-6-luna`. | 2026-10-01 | `gcloud secrets versions list`, `curl` |
| With the same 50 calls of `go/cmd/lunaeval`, xhigh cost 0.0740 USD and medium 0.0426 USD. The longest xhigh call took 63.6 s, and each effort passed 50 of 50. | 2026-10-03 | `docs/research/luna-effort-check.md` |
| A live plan at xhigh of `luna-prompt-v3` took 1 call of 0.0022 USD and 29.7 s, with 2 sessions and no violation. | 2026-10-03 | `gcloud logging read`, `users/{uid}/plan/active` |
| In the Home Screen app on iOS 27.0 with Chrome 154, `100dvh` leaves out the band of the status bar. A shell of `100dvh` ended 62 pt above the bottom edge of an iPhone 16 Pro. | 2026-10-03 | The screenshot of the owner, `web/src/lib/app-height.ts` |
| Chromium of Playwright 1.63.0 does not apply the display mode `standalone` of `Emulation.setEmulatedMedia`. | 2026-10-03 | `web/e2e/shell.spec.ts` |
| `gpt-6-luna` at medium effort passed the schema `luna_plan_v2` in 50 of 50 calls. A planner call cost 0.0017 USD on average, and the longest call took 33.0 s. | 2026-10-02 | `docs/research/phase-3-check.md` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` and the live `/version.json` name `fe7fe20261f299bb426af3a5ef4ac29fcc0343b2`, from the builds `deploy-api` `1bfa4330` and `deploy-web` `9e3f4aff`. The revision `api-00014-j4h` serves it. | 2026-10-03 | `curl`, `gcloud builds list`, `gcloud logging read` |
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

1. Close PR-29: CI, the Codex review, the owner confirmation, and the merge.
2. Start PR-30 in a clean session (D-12). Read the deploys of the merge of PR-29 first (D-137).

## Session records

### Session 30 - 2026-10-03

Author provider: Claude Code

Branch: `feat/pr-29-workout-api`. Role: author.

Completed:

- Read the deploys of `fe7fe20`, and read the live check of D-212 in the logs and Firestore, with ids alone.
- The owner approved the milestone and widened it (D-12, D-254), and answered Q-268 to Q-275 (D-254 to D-261).
- Wrote the workout service, `go/internal/workout`, and `go/internal/workoutsvc`, with unit tests and emulator tests.
- Changed the cardio rule in the prompt, the output check, the fake provider, and the plan screen, with tests.
- Changed the registers, both roadmaps, the design, `go/README.md`, and `AGENTS.md`.

Open work:

- CI, the Codex review, the owner confirmation, and the merge of PR-29.

### Session 29 - 2026-10-03

Author provider: Claude Code

Branch: `docs/pr-28-phase-6-roadmap`. Role: author.

Completed:

- Read the deploys of `9d6f8eb`, and read the live check of D-212 in Firestore and the logs, with ids alone.
- The owner approved the milestone and widened it (D-12, D-244), and answered Q-256 to Q-267 (D-242 to D-253).
- Wrote the effort flag with tests, ran the two paid runs after the approval, and wrote the report. Set the effort to xhigh.
- Changed the weight list, the cardio confirmation, and the shell, with unit, emulator, and browser tests.
- Wrote the Phase 6 roadmap, and changed the registers, the design, the high-level roadmap, both READMEs, and `AGENTS.md`.

Open work:

- None. GitHub PR 29 merged as `fe7fe20`.

### Session 28 - 2026-10-03

Author provider: Claude Code

Branch: `feat/pr-27-plan-screens`. Role: author.

Completed:

- Read the deploys of `0a7b742`.
- The owner approved the milestone before the first edit (D-12), and answered Q-252 to Q-255 (D-239 to D-241).
- Wrote the plan screen, the stream hook, and the texts, with unit tests and browser tests. Changed the fake provider and the start guard, with Go tests. Changed the design, both roadmaps, the registers, both READMEs, and `AGENTS.md`.
- Answered Codex findings P2-1 to P2-3 with full merit, and refuted P2-4.

Open work:

- None. GitHub PR 28 merged as `9d6f8eb`.
