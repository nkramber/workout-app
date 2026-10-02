# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-02. Roadmap PR-20 of `docs/roadmaps/phase-4-equipment-inventory.md`, open as GitHub PR 21 on branch `feat/pr-20-inventory-screens`, from base `b3484b6`. Work area 4.1.

The session read the deploys of `b3484b6` first. The builds `deploy-api` `19e8e84a` and `deploy-web` `98ec4b70` gave SUCCESS. The live `/version` and `/version.json` name `b3484b6`.

The pull request holds the inventory screens:

- `web/src/pages/inventory`: the list, the catalog list by kind and the text entry, the weights and the estimates, and the review screen,
- `web/src/lib/inventory.ts`, `web/src/lib/inventory-api.ts`, and `web/src/lib/errors.ts`, with the unit tests of `web/src/lib/inventory.test.ts`,
- the browser tests of `web/e2e/inventory.spec.ts`, and the shared steps of `web/e2e/support.ts`,
- the owner answers D-202 and D-203 to Q-216 and Q-217, with the design, both roadmaps, `web/README.md`, and `AGENTS.md`.

The owner approved the milestone before the first edit (D-12).

State: `make verify` passed. GitHub `verify:*` and `pr-contract` passed. The Codex review says `Ready for owner merge` for effective head `b46c0a3`. Finding ids: none.

The author ran `make web` under Node 22, and it passed. The Codex checkout has Node 20, so `make web` did not run there, and GitHub `verify:web` passed. The result is pending the owner confirmation and merge.

Next action: the owner reads the review record and the author provider, confirms the merge, then the author session turns on auto-merge under D-13.

The merge changes `web/` alone, so `deploy-web` deploys the web app, and no API deploy follows. After that deploy, the owner adds and confirms one machine in the live app on the iPhone (D-203). The Phase 5 roadmap session records the result as the exit evidence of Phase 4.

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
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `b3484b6`, three triggers, and the allowlist entry. | 2026-10-02 | `docs/setup-gcp.md`, `/version` |
| The secret `openai-api-key` has version 1, enabled. A free call to the OpenAI model list with it gave HTTP 200 and lists `gpt-6-luna`. | 2026-10-01 | `gcloud secrets versions list`, `curl` |
| `gpt-6-luna` at medium effort passed the schema `luna_plan_v2` in 50 of 50 calls. A planner call cost 0.0017 USD on average, and the longest call took 33.0 s. | 2026-10-02 | `docs/research/phase-3-check.md` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` names `b3484b631e10b1fe18c75a5c4bb139c732b8dc51`, from the build `deploy-api` `19e8e84a`. The live `/version.json` names the same commit, from the build `deploy-web` `98ec4b70`. The rules release of `df79c16` has the update time 02:33:00Z, read 2026-09-30. | 2026-10-02 | `curl`, `gcloud builds list` |
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

1. Close PR-20: CI, the Codex review, the owner confirmation, and the merge.
2. Read the `deploy-web` build of the merge, and the live `/version.json`.
3. The owner adds and confirms one machine in the live app on the iPhone (D-203).
4. Start the Phase 5 focused roadmap in a clean session. It records the Phase 4 exit evidence.

## Session records

### Session 21 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-20-inventory-screens`. Role: author.

Completed:

- Read the deploys of `b3484b6`. The owner approved the milestone before the first edit (D-12), and answered Q-216 and Q-217 (D-202, D-203).
- Wrote the inventory screens of `web/src/pages/inventory`, with 25 unit tests and 7 browser tests in each engine.
- Changed the design, both roadmaps, the registers, `web/README.md`, and `AGENTS.md`.

Open work:

- CI, the Codex review, the owner confirmation, and the merge of PR-20.
- The owner check on the iPhone after the deploy (D-203).

### Session 20 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-19-inventory-api`. Role: author.

Completed:

- Read the merge of PR-18. The owner approved the milestone before the first edit (D-12), and answered Q-211 to Q-215 (D-197 to D-201).
- Wrote `InventoryService`, `go/internal/inventory`, and `go/internal/inventorysvc`, with the unit tests and the emulator tests.
- Changed the design, both roadmaps, the registers, `go/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 20 merged as `b3484b6`.

### Session 19 - 2026-10-02

Author provider: Claude Code

Branch: `docs/pr-18-phase-4-roadmap`. Role: author.

Completed:

- Read the deploy of `9d0e5d0`. The owner approved the milestone before the first edit (D-12).
- Asked Q-98 and Q-203 to Q-210, and recorded D-188 to D-196.
- Wrote the Phase 4 focused roadmap, and changed the high-level roadmap, the design, the roadmap index, and `AGENTS.md`.

Open work:

- None. GitHub PR 19 merged as `a95ce8c`.
