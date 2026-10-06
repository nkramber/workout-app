# Workout App - Phase 8 focused roadmap: personal-use operations

This roadmap splits Phase 8 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-10-05.

## 1. Scope and start state

Phase 8 keeps the one copy of the history of the owner safe, the cost bounded, and the model current. At the end of the phase, the system runs for weeks with no data loss and no surprise bill. A known path leads to a new model or a new policy version.

### 1.1 The exit evidence of Phase 7

Section 5 of `docs/roadmaps/phase-7-adaptation-loop.md` gives the tests and the paid evaluation of each pull request of Phase 7. This session read these facts on 2026-10-04:

- PR-37 merged to `main` as `4bbf6c8`. Its commit time is 22:17:56Z.
- PR-37 changed `go/` and `web/`, so both deploys ran (D-137). The build `deploy-web` `21decb0b` gave SUCCESS at 22:20:40Z, and the build `deploy-api` `2dadce0f` gave SUCCESS at 22:22:45Z.
- The live `/version` and `/version.json` named `4bbf6c8`. The revision `api-00021-xxp` had all traffic.

The owner then made a plan and finished a workout on the iPhone. The session read the logs of the revision `api-00021-xxp` with ids and counts alone:

| Time (UTC) | Event | Result |
|---|---|---|
| 22:24:44 | `RequestPlan`, 47.3 s | 1 planner call, `ok`, 0.0045 USD |
| 22:26:11 to 22:26:47 | 5 `SyncOutbox` calls | each HTTP 200 |
| 22:26:54 | 1 reviser call | `ok`, 0.0014 USD |
| 22:26:54 | the revision of 8 exercises | 1 reason of Luna, 7 reasons of the rules |

No log entry had the severity WARNING or more. The two calls cost 0.0059 USD, inside the caps of D-188. The session did not state this cost before the calls, as D-212 tells. The plan document holds the cause of each reason of the rules, and the session did not read it.

The check found a defect. On the first exercise, the owner logged 30 lb at 3 reps in reserve for the first set, and the next set showed 20 lb. The exercise had history, so its plan had no first-set calibration (D-301). So the code did what D-301 tells, but not what the owner wants. The owner chose the rule of D-306 to D-309, and PR-38 holds the fix (D-310).

Phase 7 ends with the iPhone check at the start of PR-39, after the deploy of the merge of PR-38 (D-310). No work of Phase 8 starts before that check passes. This roadmap and the fix of PR-38 come before the check, because the owner put them there (D-310).

The check passed on 2026-10-05, on the deploy of `e09b7ff`. The three checks of PR-39 passed, so Phase 7 ended. Section 2 of `docs/research/restore-drill.md` gives the logs, the cost, and the cause of each reason of the rules.

### 1.2 The start state of the operations

This session read these facts on 2026-10-04:

- The database `(default)` is the one Firestore database of `nk-workout-app-prod`. Point-in-time recovery and delete protection are on, and the earliest version time is 2026-09-30T00:09:00Z.
- The daily backup schedule keeps each backup for 10 days (D-124). 5 backups have the state READY, from 2026-09-30 to 2026-10-04.
- Cloud Monitoring has 0 alert policies and 0 notification channels.

`docs/setup-gcp.md` gives the other facts. The budget of 10 USD sends email at 50%, 90%, and 100%, and it does not stop spend (D-139). The service `api` has no spend cap (D-141). The caps of the AI calls are 1 USD for the user and 2 USD for the project for each month (D-188, D-190). `docs/deploy-and-rollback.md` section 4 gives the rollback of each part.

This session did not read a retirement date of `gpt-6-luna`. That fact is unverified.

### 1.3 Inputs from the earlier phases

The session read the code of `4bbf6c8` for this list:

- Each plan decision has a record with the policy version and a hash of its input (D-176). `go/internal/plan` stores the records in the plan.
- Each workout holds the target that the owner saw at its start (D-291). `go/internal/revise` reads it as the history of each exercise.
- `policy.Next` gives the same target from the same history and the same policy version. So a replay under a new version gives a diff with no model call.
- `go/internal/ai/role.go` gives the model of each role, and no call site names a model (D-24).
- `go/cmd/lunaeval` runs the planner and the reviser on synthetic scenarios, with the fake provider or with Luna.
- A log line holds ids and counts alone (D-80). The line `ai call` holds the role, the status, and the cost of each call.

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-313, the later sets after a first set at another weight | They use the weight of the first set, inside a limit of the policy. Policy version 8. | D-306 |
| Q-314, the limit of a heavier first set | One weight of the list of the machine above the target. | D-307 |
| Q-315, a lighter first set | The later sets follow it, with no limit. | D-308 |
| Q-316, the load that the rules read | The load that the later sets followed. | D-309 |
| The place of the fix | PR-38, with this roadmap. The iPhone check of Phase 7 moves to the start of PR-39. | D-310 |
| Q-317, the split | Four pull requests: PR-39 to PR-42, as section 4 gives them. | D-311 |
| Q-318, the spend guard | No new cap. The alerts of this answer are dropped (D-317). | D-312 |
| Q-319, the alert channel | No channel, because the owner dropped the alerts (D-317). | D-313 |
| Q-320, a control that deletes the history | Yes, one control, "Delete all data" in the diagnostics. This amends D-78. | D-314 |
| Q-321, the data that it deletes | The workouts and the plan, on the phone and on the server. The profile and the inventory stay. | D-315 |
| Q-322, the backups | They keep a copy for up to 10 days. | D-316 |
| Q-323, the alerts | Dropped. Work area 8.2 holds the steady workout screen. | D-317 |
| Q-324 to Q-329, the parts of the workout screen that come and go | The rest card keeps its space, and the preview has the height of the set logger. The notes show below the buttons, a band at the bottom holds the notices, and the update banner waits for the end of the workout. | D-318 to D-323 |

No question of this phase stays open. Q-103 stays open for the deferred photo work.

## 3. Rules for each pull request of this phase

- No work of a work area of Phase 8 starts before the iPhone check of Phase 7 passes (AGENTS.md hard rule 1, D-310).
- No file, log, or report holds an email address, a workout log, or a value of the owner. A report holds ids and counts alone (D-80).
- A paid AI run needs the approval of the owner (D-25). The session states the expected cost first (D-212).
- Recommendation: the session also states the expected cost of each paid step of Google Cloud, such as a restore, and asks the owner first.
- A drill never writes to the database `(default)`, and delete protection stays on.
- A merge that changes `go/` deploys the API, and a merge that changes `web/` deploys the web app (D-137). The next session reads each deploy first.
- Each pull request with code gets the review of the other provider (D-15). A pull request of documents alone can use the `review-override` label (D-125).
- After a change in `proto/`, run `make proto`, and commit the generated code. `make contract` runs `buf breaking` against `main`.

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-38 | all | `docs: the Phase 8 focused roadmap (PR-38)` | none by the session |
| PR-39 | 8.1 | `docs: the restore drill and the Phase 7 check (PR-39)` | the live revision of the Phase 7 check (D-212), and the restore |
| PR-40 | 8.2 | `fix: the steady workout screen (PR-40)` | none |
| PR-41 | 8.3 | `feat: the policy replay and the incident runbook (PR-41)` | none, because the replay calls no model |
| PR-42 | exit | `docs: the four-week check of Phase 8 (PR-42)` | none |

### PR-38 - The Phase 8 focused roadmap, the first-set limit, and the deletion of all data

Branch: `docs/pr-38-phase-8-roadmap`. On 2026-10-05, the owner added the fix of the live check to this pull request (D-310). After Codex round 2, the owner added the deletion of all data (D-314).

Concerns:

- policy version 8: the limit `follow_max_tenth_lb` of each target with no calibration, and the read of the followed load (D-306 to D-309),
- the phone: the later sets follow the first set inside the limit, with a note, and the plan screen tells the limit,
- this file, with the exit evidence of Phase 7 in section 1.1 and the start state in section 1.2,
- the split of the work areas into pull requests, each with its concerns, acceptance story, checks, and paid step (D-311),
- the owner answers of section 2, with their change of `docs/roadmaps/high-level-roadmap.md` and `docs/design.md` (D-312, D-313),
- "Delete all data" in the diagnostics, with the call `DeleteHistory` of `UserService` (D-314 to D-316).

Acceptance story: a target of 20 lb with a first set of 30 lb gives 25 lb to the later sets, the limit. The rules then read 25 lb. A browser test logs that case with no network, and the next target starts from the followed load.

A second browser test opens "Delete all data". Its button stays off until the switch is at "Yes" and the text is "Delete all data". After it, the server and the phone hold no workout and no plan, and the profile and the inventory stay. This roadmap names each pull request of Phase 8, and `make verify` passes.

Checks: `make contract`, `make go-test`, `make emulator-test`, `make web`, `make verify`, and `make pr-check`, free. Codex reviews PR-38, because it changes code (D-15, D-310).

### PR-39 - The restore drill and the Phase 7 check

Branch: `docs/pr-39-restore-drill`. Work area 8.1. It needs PR-38 on `main`.

Recommendation: the type is `docs`, because the drill gives a report and a runbook. When the milestone needs a script, the type is `feat`, and the branch takes that type.

Before the work, the session reads both deploys of the merge of PR-38. It states the expected cost of one live revision with policy version 8, and asks the owner (D-212). The owner then checks on the iPhone:

1. On a new exercise, the first set is the calibration, and the other sets use its weight (D-297, D-299).
2. On an exercise with history, a first set two weights above the target gives one weight above the target (D-306, D-307).
3. When each exercise is done, the exercise list collapses (D-298).

Gate: Phase 7 ends after the three checks pass. No work of work area 8.1 starts before that. When a check fails, the session stops, and asks the owner for the fix. The work of PR-39 starts only after a fix passes the three checks. The session also reads the cause of each reason of the rules of that revision, with ids alone. In PR-38, the permission check of the session refused a read of the plan document, so the session asks the owner first.

Concerns:

- the restore of a daily backup into a new database, and its delete after the count (D-124),
- the count of the documents of each collection in the new database and in `(default)`, with the cause of each difference,
- a rollback drill of the service `api` and of Hosting, with the steps of `docs/deploy-and-rollback.md` section 4,
- a report in `docs/research/`, with the times and the counts alone,
- the changes of `docs/setup-gcp.md` and `docs/deploy-and-rollback.md` that the drill finds.

Acceptance story: the report shows that a daily backup restores into a new database, and it gives the count of each collection. The rollback drill moves the traffic of `api` to the last revision and back. `make verify` passes.

Checks: `make verify` and `make pr-check`, free. The restore costs money, so the session states the cost first. Without code, PR-39 can use the `review-override` label (D-125).

Result: `docs/research/restore-drill.md` gives the drill of 2026-10-05. The restore took 8 min 48 s, and each difference of a count has a cause. The rollback of `api` and of Hosting passed.

### PR-40 - The steady workout screen

Branch: `fix/pr-40-steady-workout-screen`. Work area 8.2. It needs PR-39 on `main`.

The owner does the first real workout with the app on 2026-10-06. On 2026-10-05, the owner dropped the alerts of D-312 and D-313, and put the quality of the workout screen first (D-317). The session read these facts on 2026-10-05:

- Cloud Monitoring starts to charge for alerts on 2027-09-01 at the earliest. Then each metric reference of an alert policy costs 0.35 USD each month, and a filter condition is one reference.
- The Firestore documents name no metric, log entry, or audit entry for a failed scheduled backup. The one backup metric gives the storage size.
- The log of the failed build `18b68239` ends with the line `ERROR`. No Cloud Build document states this line, and the documents give the build status `FAILURE`.

Concerns:

- the rest card always keeps its space, with the stored end, "Rest done", and no sound and no vibration (D-58, D-59, D-270, D-318),
- the preview of the next machine keeps the height of the set logger (D-319),
- the notes of a set show below the buttons of the reps in reserve (D-320),
- a band at the bottom holds the wake notice and the error line (D-321, D-322),
- the update banner waits for the end of the workout (D-323),
- the record of the drop of the alerts in the registers and the roadmaps (D-317).

Acceptance story: a browser test of `web/e2e/workout.spec.ts`, in Chromium and WebKit, logs a set, waits through the rest, and dismisses it. The top of the set logger keeps the same position before the rest, during it, at "Rest done", and after the dismiss. `make web` and `make verify` pass.

Checks: `make web`, `make verify`, and `make pr-check`, free. The merge changes `web/`, so it deploys the web app (D-137). Codex reviews PR-40.

### PR-41 - The policy replay and the incident runbook

Branch: `feat/pr-41-version-migration`. Work area 8.3. It needs PR-40, the steady workout screen, on `main`.

Concerns:

- a replay in `go/cmd`. It reads the stored workouts and their target copies, and gives the targets of the current policy version,
- a diff report against the stored decision records, with the counts and the rule ids alone (D-80, D-176),
- the steps of a model change when `gpt-6-luna` retires: the change of the role layer, the prompt version, and the paid evaluations (D-24, D-25),
- an incident runbook for a failed deploy, a failed sync, a refusal of the caps, a data loss, and an outage of the model.

Acceptance story: an emulator test replays stored sessions of policy version 7 under policy version 8. The diff report gives the count of each changed target and its rule id. A read of the live store needs the approval of the owner. `make verify` passes.

Checks: `make go-test`, `make emulator-test`, `make verify`, and `make pr-check`, free. The replay calls no model. Codex reviews PR-41.

### PR-42 - The four-week check of Phase 8

Branch: `docs/pr-42-phase-8-check`. The exit of the roadmap. It needs PR-41 on `main`, and four weeks of the use of the owner.

Recommendation: the four weeks start at the end of the Phase 7 check of PR-39. The session asks the owner when another start applies.

Concerns:

- the count of the workouts and the sets of the four weeks, from the store, with counts alone,
- the count of the refused syncs and of the errors, from the logs,
- the replay of PR-41 over the four weeks, and the count of the targets outside a bound,
- a report in `docs/research/`.

Acceptance story: the report shows no lost set, no refused valid sync, and no policy breach in the decision log for four weeks. `make verify` passes.

Checks: `make verify` and `make pr-check`, free. PR-42 changes documents alone, so it can use the `review-override` label (D-125).

## 5. Exit of the phase

Phase 8 ends when PR-42 merges. These items give the exit evidence of `docs/roadmaps/high-level-roadmap.md`:

- The restore drill report of PR-39 (work area 8.1).
- The browser test of the steady workout screen of PR-40 (work area 8.2).
- The replay diff report of PR-41 (work area 8.3).
- The four weeks of use of PR-42, the exit of the roadmap.
