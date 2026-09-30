# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-30. Roadmap PR-12 of `docs/roadmaps/phase-3-workout-domain.md`, all work areas of Phase 3, on branch `docs/pr-12-phase-3-roadmap`, from base `73b1964`. It holds the Phase 3 focused roadmap.

The pull request holds these concerns:

- the focused roadmap `docs/roadmaps/phase-3-workout-domain.md`, with PR-12 to PR-17 (D-158),
- the owner answers to Q-92, Q-101, Q-102, Q-104 to Q-107, and Q-165 to Q-171 (D-147 to D-158),
- the change of scope of D-154: dumbbells and two cable exercises, in `docs/design.md` and `docs/roadmaps/high-level-roadmap.md`,
- the free-weight research in `docs/research/exercise-safety.md` (D-156),
- the row of the roadmap in `docs/roadmaps/README.md`, and the Phase 3 stage in `AGENTS.md`.

Checks of the base before the work, on 2026-09-30:

- The build `deploy-web` `c5ac0f62-7d38-46b5-8f68-6f04da88f5a4` of `73b1964` passed at 04:30:24Z. `/version.json` names `73b1964340ebc2a14f93b75f1be19aba024918f5`.
- `deploy-api` and `deploy-rules` did not run for `73b1964`. `/version` still names `df79c169ae38326c0a876ac678a5465f80103d13`.
- The owner opened the Home Screen app again. The title is sharp, with the 16 px space above it (D-145).
- `gym-route-dev` is still `DELETE_REQUESTED`. Its site still gave HTTP 200 at 04:38:56Z.

State: Codex reviewed effective head `4bf0f925225f2d806e17f7c5a9cdc2bfadbbb567` and found no defect. The record is ready for publication. GitHub CI passed except `review-gate`, which needs this record.

Next action: publish the review record, verify the review-gate result, then wait for the owner confirmation and merge of PR-12.

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
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` names `df79c169ae38326c0a876ac678a5465f80103d13`. The live `/version.json` names `73b1964340ebc2a14f93b75f1be19aba024918f5`, from the build `deploy-web` `c5ac0f62`. The rules release of `df79c16` has the update time 02:33:00Z. | 2026-09-30 | `curl`, `gcloud builds list` |
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

1. Close PR-12: CI, the Codex review, the owner confirmation, and the merge.
2. Start PR-13 of `docs/roadmaps/phase-3-workout-domain.md` in a clean session.

## Session records

### Session 13 - 2026-09-30

Author provider: Claude Code

Branch: `docs/pr-12-phase-3-roadmap`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12). Then the owner approved the change of scope and the research as more concerns.
- Read the deploy of `73b1964` and the state of `gym-route-dev`. The owner confirmed the title fix in the live app.
- Asked the seven open questions of Phase 3 and Q-165 to Q-171 (D-147 to D-158).
- Wrote the focused roadmap, and changed the high-level roadmap, the design, `AGENTS.md`, and the roadmap list.
- Added the free-weight research to `docs/research/exercise-safety.md` (D-156).

Open work:

- CI, the Codex review, the owner confirmation, and the merge of PR-12.

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

- None. GitHub PR 12 merged as `73b1964`.

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
