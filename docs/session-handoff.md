# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-03. Roadmap PR-28 of `docs/roadmaps/phase-6-guided-workout.md`, on branch `docs/pr-28-phase-6-roadmap`, from base `9d6f8eb`. GitHub PR 29.

Before the work, the session read the deploy of `9d6f8eb`. The build `deploy-api` `cb49ad85` and the build `deploy-web` `7345e0d5` gave `SUCCESS`, and the live `/version` and `/version.json` both name `9d6f8eb` (D-137).

The owner approved the live check (D-212). Request 2 gave a plan of 9 exercises from Luna, with no violation, in 1 call of 0.00104 USD. Section 1.1 of the roadmap records it as the exit evidence of Phase 5.

The owner approved the milestone, then widened it to "the Phase 5 live check and its follow-ups" (D-12, D-244). The pull request holds:

- the Phase 6 roadmap, with PR-29 to PR-32 (D-247 to D-252),
- the flag `-effort` of `go/cmd/lunaeval`, two paid runs of 0.1166 USD in total, the report `docs/research/luna-effort-check.md`, and the effort xhigh of both roles (D-242, D-243, D-253),
- the weight list at the save (D-245), and the confirmation of a cardio machine at the save (D-246),
- the shell to the bottom edge of the Home Screen app (D-120). The live shell of `100dvh` ended 62 pt above the bottom edge.

`make verify`, `make go-test`, `make emulator-test`, and `make web` passed.

State: the review record gives `Ready for owner merge` at `3317cde`, and `review-gate` passes. Pending the owner merge. Next action: ask the owner to confirm the merge (D-13).

The merge changes `go/` and `web/`, so it deploys both. The session of PR-29 reads both deploys, then asks the owner for one live plan at xhigh (D-212). The owner also reads the bottom edge of the plan screen on the iPhone.

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
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `9d6f8eb`, three triggers, and the allowlist entry. | 2026-10-03 | `docs/setup-gcp.md`, `/version` |
| The service `api` revision `api-00013-w8w` holds `OPENAI_API_KEY` from `openai-api-key:latest`, `LUNA_CAP_USER_USD=1`, `LUNA_CAP_PROJECT_USD=2`, and a request timeout of 420 s. The TTL policy of `aiErrors.expire_at` is `ACTIVE`. | 2026-10-03 | `gcloud run services describe`, `gcloud firestore fields ttls list` |
| A Cloud Run service with a secret needs the accessor role for its service identity alone. The page lists `roles/run.admin` for the deployer. | 2026-10-02 | Cloud Run docs, "Configure secrets for services" |
| The secret `openai-api-key` has version 1, enabled. A free call to the OpenAI model list with it gave HTTP 200 and lists `gpt-6-luna`. | 2026-10-01 | `gcloud secrets versions list`, `curl` |
| With the same 50 calls of `go/cmd/lunaeval`, xhigh cost 0.0740 USD and medium 0.0426 USD. The longest xhigh call took 63.6 s, and each effort passed 50 of 50. | 2026-10-03 | `docs/research/luna-effort-check.md` |
| In the Home Screen app on iOS 27.0 with Chrome 154, `100dvh` leaves out the band of the status bar. A shell of `100dvh` ended 62 pt above the bottom edge of an iPhone 16 Pro. | 2026-10-03 | The screenshot of the owner, `web/src/lib/app-height.ts` |
| Chromium of Playwright 1.63.0 does not apply the display mode `standalone` of `Emulation.setEmulatedMedia`. | 2026-10-03 | `web/e2e/shell.spec.ts` |
| `gpt-6-luna` at medium effort passed the schema `luna_plan_v2` in 50 of 50 calls. A planner call cost 0.0017 USD on average, and the longest call took 33.0 s. | 2026-10-02 | `docs/research/phase-3-check.md` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` and the live `/version.json` name `9d6f8ebdb3dc3532a99ad4858d65056c1851eba0`, from the builds `deploy-api` `cb49ad85` and `deploy-web` `7345e0d5`. The rules release of `df79c16` has the update time 02:33:00Z, read 2026-09-30. | 2026-10-03 | `curl`, `gcloud builds list` |
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

1. Close PR-28: the owner confirmation and the merge.
2. Start PR-29 in a clean session (D-12). Read the deploys of the merge of PR-28 first (D-137).
3. State the expected cost of one live plan at xhigh, and ask the owner (D-212). After the approval, the owner requests a plan and reads the bottom edge.

## Session records

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

- The owner confirmation and the merge of PR-28.

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

### Session 27 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-26-plan-api`. Role: author.

Completed:

- Read the deploy of `42fb3da`.
- The owner approved the milestone before the first edit (D-12), answered Q-240 to Q-251 (D-226 to D-238), and approved the revised milestone.
- Wrote the plan service, `go/internal/plan`, `go/internal/plansvc`, prompt v3, and the API wiring, with unit tests and emulator tests.
- With the approval of the owner, changed the service `api` and the TTL policy. Changed the design, both roadmaps, the registers, `go/README.md`, `docs/setup-gcp.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 27 merged as `0a7b742`.
