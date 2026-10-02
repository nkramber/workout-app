# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-02. Roadmap PR-21 of `docs/roadmaps/phase-4-equipment-inventory.md`, on branch `fix/pr-21-inventory-writes`, from base `dc89e30`. Work area 4.1.

The session read the deploy of `dc89e30` first. The build `deploy-web` `c78c7c6e` gave SUCCESS, and the live `/version.json` names `dc89e30`. The live `/version` names `b3484b6`, because no API deploy followed.

The live check of D-203 failed. The app showed "The API did not answer. The change is not saved." The request log gave HTTP 500 for each `SaveMachine` call, and `api-runtime` held `roles/datastore.viewer` alone. The owner chose a Phase 4 correction in this session (D-204), and approved the milestone before the first edit (D-12).

The pull request holds:

- the role change of `api-runtime` to `roles/datastore.user`, applied at 16:22:48Z and read back, in `docs/setup-gcp.md` (D-206),
- the text "The server failed." for the code `internal`, in `web/src/lib/errors.ts` (D-206),
- the A to Z order of the inventory list and of each kind group of the catalog list, in `web/src/lib/inventory.ts` (D-205),
- the unit tests and the browser tests of both, and the owner answers D-204 to D-206 to Q-218 to Q-220.

State: CI passed except `review-gate`, which needs the review record. The Codex review says `Ready for owner merge` for effective head `5ba1eb38f2af08675364a89c788062297a90d6cc`. Finding ids: none. The result is pending the owner merge.

Next action: after the merge, read the `deploy-web` build of the merge and the live `/version.json`. Then the owner adds and confirms one machine in the live app on the iPhone (D-203). The Phase 5 roadmap session records the result as PR-22 (D-204).

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
| The live `/version` names `b3484b631e10b1fe18c75a5c4bb139c732b8dc51`, from the build `deploy-api` `19e8e84a`. The live `/version.json` names `dc89e30d2fd996510b68e442f9b95394c1744338`, from the build `deploy-web` `c78c7c6e`. The rules release of `df79c16` has the update time 02:33:00Z, read 2026-09-30. | 2026-10-02 | `curl`, `gcloud builds list` |
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

1. Close PR-21: CI, the Codex review, the owner confirmation, and the merge.
2. Read the `deploy-web` build of the merge, and the live `/version.json`.
3. The owner adds and confirms one machine in the live app on the iPhone (D-203).
4. Start the Phase 5 focused roadmap as PR-22 in a clean session. It records the Phase 4 exit evidence (D-204).

## Session records

### Session 22 - 2026-10-02

Author provider: Claude Code

Branch: `fix/pr-21-inventory-writes`. Role: author.

Completed:

- Read the deploy of `dc89e30`. The live check of D-203 failed, and the request log and the IAM policy gave the cause.
- The owner answered Q-218 to Q-220 (D-204 to D-206), and approved the milestone before the first edit (D-12).
- Changed the role of `api-runtime`. Wrote the A to Z order and the text of a server fault, with their tests.

Open work:

- CI, the Codex review, the owner confirmation, and the merge of PR-21.
- The owner check on the iPhone after the deploy (D-203).

### Session 21 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-20-inventory-screens`. Role: author.

Completed:

- Read the deploys of `b3484b6`. The owner approved the milestone before the first edit (D-12), and answered Q-216 and Q-217 (D-202, D-203).
- Wrote the inventory screens of `web/src/pages/inventory`, with 25 unit tests and 7 browser tests in each engine.
- Changed the design, both roadmaps, the registers, `web/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 21 merged as `dc89e30`. The live check of D-203 failed, and PR-21 corrects the fault (D-204).

### Session 20 - 2026-10-02

Author provider: Claude Code

Branch: `feat/pr-19-inventory-api`. Role: author.

Completed:

- Read the merge of PR-18. The owner approved the milestone before the first edit (D-12), and answered Q-211 to Q-215 (D-197 to D-201).
- Wrote `InventoryService`, `go/internal/inventory`, and `go/internal/inventorysvc`, with the unit tests and the emulator tests.
- Changed the design, both roadmaps, the registers, `go/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 20 merged as `b3484b6`.
