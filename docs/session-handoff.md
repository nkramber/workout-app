# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-09-30. Roadmap PR-15 of `docs/roadmaps/phase-3-workout-domain.md`, work area 3.2, on branch `feat/pr-15-policy-fallback`, from base `0b80066`. It holds the re-entry, the calibration, and the rules fallback.

The pull request holds these concerns, all in `go/internal/policy`, at policy version 2:

- the long-break table and the first sessions after a break (D-151, D-179), with 3 reps in reserve and rep progression only (D-37),
- the start of a new exercise and the calibration of its first 3 sessions (D-150, D-177, D-178, D-180),
- the rules fallback when Luna gives no proposal or the policy refuses it (D-23),
- the decision record of D-176, with a SHA-256 hash of the input,
- the golden tests of scenarios E and F, and the property tests of the fallback, the first sessions, the start, and the calibration.

The owner approved the milestone before the first edit, and answered Q-188 to Q-193 (D-175 to D-180). Q-194 holds REC-7 for Phase 7. The deploy of `0b80066` passed: the build `deploy-api` `bef45ff1` gave SUCCESS, and the live `/version` names `0b80066`.

State: GitHub PR 16 is open, pending CI, the Codex review, and the owner merge. `make go-test` and `make verify` pass.

Next action: wait for green CI, then run `make codex-review` (D-8).

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
| The project `nk-workout-app-prod` holds billing, the 10 USD budget, Firestore with PITR, daily backups, and delete protection, the service `api` of build `0b80066`, three triggers, and the allowlist entry. | 2026-09-30 | `docs/setup-gcp.md`, `/version` |
| The bucket `nk-workout-app-prod-deploy-lock` exists, and each deployer account holds `roles/storage.objectUser` on it alone. | 2026-09-29 | `gcloud storage buckets describe`, `get-iam-policy` |
| `workout-app-prod` is in use by another Google Cloud project. | 2026-09-29 | `gcloud projects create` |
| The old project `gym-route-dev` is `DELETE_REQUESTED` since 2026-09-30T03:19:47Z. `gcloud projects undelete` can restore it for 30 days. Its site still gave HTTP 200 at 04:38:56Z. | 2026-09-30 | `gcloud projects describe`, `curl` |
| The live `/version` names `0b8006641dd58996d45bc23eeebd3ccdcaa65520`, from the build `deploy-api` `bef45ff1`. The live `/version.json` names `73b1964340ebc2a14f93b75f1be19aba024918f5`, from the build `deploy-web` `c5ac0f62`. The rules release of `df79c16` has the update time 02:33:00Z. | 2026-09-30 | `curl`, `gcloud builds list` |
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

1. Close PR-15: CI, the Codex review, the owner confirmation, and the merge.
2. Start PR-16 of `docs/roadmaps/phase-3-workout-domain.md` in a clean session.

## Session records

### Session 16 - 2026-09-30

Author provider: Claude Code

Branch: `feat/pr-15-policy-fallback`. Role: author.

Completed:

- Read the deploy of `0b80066`. The owner approved the milestone before the first edit (D-12), and answered Q-188 to Q-193 (D-175 to D-180).
- Added to `go/internal/policy` the start, the calibration, the long-break table, the first sessions after a break, the rules fallback, and the decision record.
- Wrote the golden tests of scenarios E and F, and the property tests of the fallback half of the Luna property of section 6.3.
- Changed the design, both roadmaps, the registers, and `go/README.md`.

Open work:

- CI, the Codex review, the owner confirmation, and the merge of PR-15.

### Session 15 - 2026-09-30

Author provider: Claude Code

Branch: `feat/pr-14-policy-progression`. Role: author.

Completed:

- Read the deploy of `c3ab54c`. The owner approved the milestone before the first edit (D-12), and answered Q-180 to Q-187 (D-167 to D-174).
- Wrote `go/internal/policy`: the rules, the bounds, the rounding, the next target, the check of a proposal, and the pain warning.
- Wrote the golden tests of scenarios A to D and the property tests of section 6.3 of the high-level roadmap.
- Changed the design, both roadmaps, the registers, and `go/README.md`.

Open work:

- None. GitHub PR 15 merged as `0b80066`.

### Session 14 - 2026-09-30

Author provider: Claude Code

Branch: `feat/pr-13-domain-model`. Role: author.

Completed:

- The owner approved the milestone before the first edit (D-12), and answered Q-172 to Q-179 (D-159 to D-166).
- Wrote `go/internal/domain`: the types, the catalog of D-155, the lookups, and the checks, with table tests.
- Changed the design, the Phase 3 roadmap, and `go/README.md` for the new decisions.
- Answered Codex findings P2-1 and P2-2 with full merit: fixed sets of kinds and regions, and the 100 lb dumbbell cap.

Open work:

- None. GitHub PR 14 merged as `c3ab54c`.
