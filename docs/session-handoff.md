# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-29. GitHub PR 11 is roadmap PR-10 of `docs/roadmaps/phase-2-platform-skeleton.md`, work area 2.3, on branch `feat/pr-10-deploy`, from base `6117952`.

The pull request holds the project and the deploy:

- the new app name "Workout App", the project `nk-workout-app-prod`, and the contract package `workoutapp.v1` (D-136 to D-138),
- the billing link and an alerts-only budget of 10 USD (D-139),
- Firestore Standard in `us-central1`, deny-all rules, PITR, daily backups of 10 days, and delete protection (D-140, D-124),
- the service `api` on Cloud Run with `api-runtime`, and the secret `openai-api-key` with no value (D-141),
- the triggers `deploy-api`, `deploy-web`, and `deploy-rules` on `main`, each with its own account (D-142),
- the allowlist entry of the owner in the live project alone,
- `docs/setup-gcp.md` and `docs/deploy-and-rollback.md`, and the tests `docs/tools/test_deploy_config.py` and `go/cmd/api/rules_emulator_test.go`.

State: the live project is complete, and each part reads back. Codex required changes at effective head `e5f724b` for P2-1, an older build that replaces a newer deploy. Round 2 adds the guard `docs/tools/deploy_order.py` before each deploy step, and `docs/reviews/pr-11-response.md` answers the finding.

Next action: repeat the Codex review. Then the owner confirms the merge. The merge starts the three builds.

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
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` with the `hello` placeholder, three triggers, and the allowlist entry. | 2026-09-29 | `docs/setup-gcp.md`, the read back of each step |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` has no billing, and its site serves the probe of build `d6e3c54`. | 2026-09-29 | `gcloud billing projects describe`, `firebase hosting:sites:list` |
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

1. Close PR-10: the Codex review, the owner confirmation, and the merge.
2. After the merge, read the three builds with `gcloud builds list --region=us-central1 --project nk-workout-app-prod`.
3. Start PR-11 (work areas 2.2 and 2.3) in a clean session. It reads `/version` and `/version.json`, runs the device check, shuts down `gym-route-dev`, and adds the package rules of D-138.

## Session records

### Session 11 - 2026-09-29

Author provider: Claude Code

Branch: `feat/pr-10-deploy`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12). Then the owner added the new name and the new project, and approved the changed milestone.
- Read the billing state, then asked Q-142 and Q-152 to Q-159 (D-136 to D-142).
- Made each part of the project with the approval of the owner at run time, and read each part back.
- Renamed each current file. The past records keep the old name (D-136).
- Wrote the deploy files, the setup and rollback documents, and the tests. The rules test fails with open rules.
- Answered P2-1 of the Codex review with a guard before each deploy step.

Open work:

- The Codex review, the owner confirmation, and the merge of PR-10.

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
