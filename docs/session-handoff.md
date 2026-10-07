# Workout App - session hand-off

This file is the resume point of the next session. Read `AGENTS.md` first, then this file.

## Resume here

Date: 2026-10-06. Author of pull request PR-43, on the branch `ci/pr-43-gitar-review` from base `d9e3590`. PR-43 is a change of the process, outside the focused roadmaps. Phase 8 ended with the merge of PR-42, and the owner gave this work by a new decision (D-335).

The owner approved the milestone "the mandatory Gitar review" (D-12), and answered Q-341 to Q-343 (D-335 to D-338). The pull request ports the Gitar review of Decktome and what-you-carry:

- the rules of `AGENTS.md` and the decisions that amend D-3 and D-6 (D-335),
- the skill `.claude/skills/gitar-review/SKILL.md` and its reference `references/traps.md`,
- the Gitar gate of `docs/tools/codex_review.py`, for `make codex-review` and `make claude-review`, with the flag of D-337 (D-336),
- the wait `docs/tools/gitar_wait.py` and `make gitar-wait` (D-338),
- the wiring check of `docs/tools/pr_check.py` in place of the ban, and the Gitar steps of `one-pr-one-session`, `pr-review`, and the pull request template.

GitHub pull request 44 holds PR-43. The gate applies to PR-43 itself, because `make codex-review` runs from the checkout. The Gitar app is installed, as the owner said (Q-342). Its review of `679655b` came 125 s after the push, and `make gitar-wait` read it.

Gitar round 1 gave one finding: `SKIP_GITAR=0` also skipped the Gitar pass. The author found full merit. Now only the exact value 1 skips, and a test of `make -n` proves each value.

Codex reviewed PR-44 at effective head `88b647d`. The verdict is `Ready for owner merge`, with no finding. The record is `docs/reviews/pr-44.md`.

Next action: continue the owner merge steps for PR-43. The review record is on the branch, and `review-gate` passes.

## Facts that expire

| Fact | Date read | Source |
|---|---|---|
| The Gitar app `gitar-bot` gave no check run and no comment on the GitHub pull requests 41 to 43 of this repository. On Decktome and what-you-carry, it gives a check run `Gitar` and one dashboard comment on each pull request. | 2026-10-06 | `gh api` check runs and issue comments |
| The live `/version` names `efbafe7`, from the build `deploy-api` `7a5d44c3`. The revision `api-00023-7x9` has all traffic. | 2026-10-06 | `curl`, `gcloud builds list`, `gcloud run services describe` |
| The live store holds 1 user, 1 workout that is not finished, 16 sets, and 17 applied sync entries. The active plan has 16 records of policy version 8, and it breaks the rotation of version 9. | 2026-10-06 | `docs/research/phase-8-check.md` |
| The paid check of the rotation: 100 planner calls of `luna-prompt-v7` at xhigh, 99 valid plans, 1 refused by `rotation.cover`. Cost 0.163 USD, the longest call 62.6 s. | 2026-10-06 | `docs/research/phase-8-check.md` |
| From 2026-10-05T00:00Z, the API gave HTTP 200 to each of 295 calls. The one WARNING entry is a `GET /` of a crawler. | 2026-10-06 | `gcloud logging read` |
| The deprecation page of OpenAI lists no retirement of `gpt-6-luna`, and names it as the replacement of `gpt-5.4-nano`. | 2026-10-06 | `docs/operations.md` section 4 |
| The paid evaluation of D-302: 50 reviser calls of the 10 scenarios with policy version 7 and xhigh, all `ok`. The check accepted 110 reasons and refused 5. Cost 0.0179 USD, the longest call 10.8 s. | 2026-10-04 | `docs/research/reviser-evaluation.md` |
| The Phase 7 check of `e09b7ff`: 1 `DeleteHistory` of 11 workouts, 1 planner call, and 3 reviser calls, each `ok`, for 0.004379230 USD. No log entry had the severity WARNING or more. | 2026-10-05 | `docs/research/restore-drill.md` |
| The restore of a daily backup took 8 min 48 s, and the restored database had delete protection. The drill deleted it. `(default)` is the one database. | 2026-10-05 | `docs/research/restore-drill.md` |
| `nk-workout-app-prod` has 0 alert policies and 0 notification channels of Cloud Monitoring. | 2026-10-04 | Monitoring API |
| Cloud Monitoring starts to charge for alerts on 2027-09-01 at the earliest. Then each metric reference of an alert policy costs 0.35 USD each month. | 2026-10-05 | "Pricing", Google Cloud Observability |
| No Firestore document names a metric, a log entry, or an audit entry for a failed scheduled backup. | 2026-10-05 | Firestore backups, metrics, and audit logging pages |
| The log of the failed build `18b68239` ends with the line `ERROR`. No Cloud Build document states this line. | 2026-10-05 | `gcloud builds log` |
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

1. Close PR-43: the Gitar pass, CI, the Codex review, the owner confirmation, and the merge.
2. Read the API deploy of the merge of PR-42, and check that `/version` names the merge commit.
3. The owner deletes the history and makes a new plan of policy version 9 (D-333).
4. Phase 8 ends with the merge of PR-42. No phase comes after it. A new owner decision gives the next work.

## Session records

### Session 44 - 2026-10-06

Author provider: Claude Code

Branch: `ci/pr-43-gitar-review`. Role: author.

Completed:

- Read the Gitar parts of Decktome and what-you-carry. The owner approved the milestone and answered Q-341 to Q-343 (D-335 to D-338).
- Ported the Gitar gate of the review targets, the wait, the skill, and the rules, with their tests.

Open work:

- The Gitar pass, the Codex review, the owner confirmation, and the merge of PR-43.

### Session 43 - 2026-10-06

Author provider: Claude Code

Branch: `feat/pr-42-rotation-and-phase-8-check`. Role: author.

Completed:

- Read the deploy of `efbafe7`. The owner answered Q-333 to Q-340 (D-327 to D-334) and approved the milestone (D-12).
- Wrote policy version 9, the rotation of the role layer, the fake split, the layout replay, and their tests.
- Ran the paid check of the planner, the live replay, and the counts of the store and the logs.
- Wrote `docs/research/phase-8-check.md`, and changed the registers, both roadmaps, the design, the runbook, `go/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 43 merged as `d9e3590`.

### Session 42 - 2026-10-06

Author provider: Claude Code

Branch: `feat/pr-41-version-migration`. Role: author.

Completed:

- The owner approved the milestone (D-12), and answered Q-330 to Q-332 (D-324 to D-326).
- Wrote the replay, the command, its tests, and the emulator test of the acceptance story.
- Ran the replay on the live store two times with reads alone. Found and fixed a fault of the replay.
- Read the deprecation page of OpenAI.
- Wrote `docs/operations.md` and `docs/research/policy-replay.md`, and changed the registers, both roadmaps, `go/README.md`, and `AGENTS.md`.

Open work:

- None. GitHub PR 42 merged as `efbafe7`.
