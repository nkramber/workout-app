# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-02. Roadmap PR-25 of `docs/roadmaps/phase-5-onboarding-and-plan.md`, on branch `feat/pr-25-cap-store`, from base `a1b88d4`. Work area 5.2.

Before the work, the session read the deploy of `a1b88d4`. The build `deploy-web` `b57b45ae` gave `SUCCESS`, and the live `/version.json` names `a1b88d4`. The merge changed no file in `go/`, so the live `/version` stays at `2e22c27` (D-137).

The owner approved the milestone before the first edit (D-12), and answered Q-238 and Q-239 (D-224, D-225).

The pull request holds:

- the lasting cap hook `go/internal/capstore`, with the spend of each calendar month in UTC in Firestore (D-189, D-190, D-224),
- a reservation of the worst-case cost before each call and a charge after it, each in one transaction for the user and the project,
- the worst-case charge of a failed call (D-225), and the flag `Unsettled` of a cost record when the charge does not reach the store,
- a `context.Context` in `ai.CapHook`, and an error from its settle function,
- the cap text of `go/internal/ai/cost.go`, `go/internal/ai/doc.go`, and `go/README.md`.

The emulator tests of `go/internal/capstore/firestore_emulator_test.go` hold the acceptance story. After a restart, a planner call over the cap gets `capped`, and no call reaches the fake provider. `make go-test` and `make emulator-test` passed.

Codex review: `docs/reviews/pr-26.md` says Ready for owner merge for effective head `96c8d663c7ed7f121edc4243d99762fea34c6ed7`. No finding is open.

State: pending the owner merge. Next action: get the owner confirmation, then turn on the auto-merge. After the merge, read the deploy of the API first. Then start PR-26, the plan API, in a clean session.

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
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `2e22c27`, three triggers, and the allowlist entry. | 2026-10-02 | `docs/setup-gcp.md`, `/version` |
| The secret `openai-api-key` has version 1, enabled. A free call to the OpenAI model list with it gave HTTP 200 and lists `gpt-6-luna`. | 2026-10-01 | `gcloud secrets versions list`, `curl` |
| `gpt-6-luna` at medium effort passed the schema `luna_plan_v2` in 50 of 50 calls. A planner call cost 0.0017 USD on average, and the longest call took 33.0 s. | 2026-10-02 | `docs/research/phase-3-check.md` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` names `2e22c2774dff177363d6d049b070429300a61054`, from the build `deploy-api` `d2933354`. The live `/version.json` names `a1b88d40c8c6c65d8fc958abff94453cbbf6687d`, from the build `deploy-web` `b57b45ae`. The rules release of `df79c16` has the update time 02:33:00Z, read 2026-09-30. | 2026-10-02 | `curl`, `gcloud builds list` |
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

1. Close PR-25: CI, the Codex review, the owner confirmation, and the merge.
2. Read the deploy of the API of the merge of PR-25 (D-137).
3. Start PR-26, the plan API, in a clean session (D-207). The owner approves its milestone first (D-12).

## Session records

### Session 26 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-25-cap-store`. Role: author.

Completed:

- Read the deploy of `a1b88d4`.
- The owner approved the milestone before the first edit (D-12), and answered Q-238 and Q-239 (D-224, D-225).
- Wrote the lasting cap hook, with unit tests and emulator tests. Changed the cap hook interface of `go/internal/ai`, the design, both roadmaps, the registers, `go/README.md`, and `AGENTS.md`.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-25.

### Session 25 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-24-onboarding-screens`. Role: author.

Completed:

- Read the deploys of `2e22c27`.
- The owner approved the milestone before the first edit (D-12), and answered Q-236 and Q-237 (D-222, D-223).
- Wrote the onboarding screen, the profile gate, and the "Profile" button, with unit tests and browser tests. Changed the design, both roadmaps, the registers, `web/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 25 merged as `a1b88d4`.

### Session 24 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-23-profile-api`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-227 to Q-235 (D-213 to D-221).
- Answered Codex findings P1-1 and P2-1 of round 1 with full merit.
- Wrote the profile service, the store, the checks, the two tables, and `profile.ForPlan`, with unit tests and emulator tests.
- Added the research of both tables to `docs/research/exercise-safety.md`. Changed the design, both roadmaps, the registers, `go/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 24 merged as `2e22c27`.
