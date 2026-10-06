# Operations: incidents, the policy replay, and a model change

This document gives the steps of five incidents, the policy replay, and a change of the model. `docs/deploy-and-rollback.md` gives the deploy, the rollback of each part, and the restore of the data. `docs/setup-gcp.md` gives the project `nk-workout-app-prod` and its accounts.

The date of this version is 2026-10-05.

## 1. Find an incident

The project has no alert (D-317). Cloud Monitoring has 0 alert policies and 0 notification channels. So a person finds each incident with one of these signals:

| Signal | Command or place | It shows |
|---|---|---|
| The logs of `api` | `gcloud logging read` with the filter of section 1.1 | each error, each AI call, and each revision |
| The build status | `gcloud builds list --project nk-workout-app-prod --region us-central1 --limit 10` | each deploy and its status |
| The live commit | `curl` of `/version` of the API and `/version.json` of the web app | the commit of each part |
| The backup list | `gcloud firestore backups list --project nk-workout-app-prod` | each daily backup and its state |
| The phone | "Diagnostics" on the home screen | the build, and the count of the changes that wait to sync |
| The budget | the email of the budget at 50%, 90%, and 100% of 10 USD | the spend of the month |

Read the signals after each merge, and once each week. A log line holds ids and counts alone (D-80), so a read of the logs shows no workout data.

### 1.1 The log filters

Each filter reads the service `api`. Add `AND timestamp>="<time>"` to read a short period.

```
P=nk-workout-app-prod
BASE='resource.type="cloud_run_revision" AND resource.labels.service_name="api"'
# Each warning and each error.
gcloud logging read "$BASE AND severity>=WARNING" --project $P --limit 50
# Each AI call, with its role, status, and cost.
gcloud logging read "$BASE AND jsonPayload.msg=\"ai call\"" --project $P --limit 50
# Each revision, and each failed revision.
gcloud logging read "$BASE AND jsonPayload.msg=(\"plan revised\" OR \"revision failed\")" --project $P --limit 50
# Each sync call, with its HTTP status.
gcloud logging read "$BASE AND httpRequest.requestUrl:\"SyncOutbox\"" --project $P --limit 50
```

## 2. The incidents

Each incident gives its signal, its check, and its steps. Stop and ask the owner before a step that changes the project or costs money.

### 2.1 A failed deploy

Signal: a build of `deploy-api`, `deploy-web`, or `deploy-rules` has the status `FAILURE` or `TIMEOUT`. The log of a failed build ends with the line `ERROR`. Or `/version` and `/version.json` do not name the commit of the last merge.

1. Read the build list, and find the failed build and its trigger.
2. Read the log of the build with `gcloud builds log <id> --region us-central1`.
3. Find the step that failed. `docs/deploy-and-rollback.md` section 1 gives the steps of each build.
4. When the deploy step did not start, the live part stays at the last good commit. Repair the cause in a new pull request.
5. When the live part is bad, roll it back with `docs/deploy-and-rollback.md` section 4.
6. When the cause was outside the repository, start the build again with the Cloud Build API call `builds/{id}:retry`.
7. Read `/version` and `/version.json` again.

A lock that stays after a failed build becomes stale after 20 minutes, and the next build removes it.

### 2.2 A failed sync

Signal: on the phone, "Changes waiting to sync" stays above 0 with a network. Or the logs show `SyncOutbox` calls with an HTTP status of 400 or more, or the line `revision failed`.

The phone keeps each entry in its outbox until the server applies it. A failed call adds 1 to the attempts of each entry of its batch, and the phone sends the batch again later. The server refuses an entry that it can not read with the code `invalid_argument`. The phone then moves that entry to its refused entries, so the entry does not stop the other entries.

1. Read the sync calls and the warnings of the period.
2. For the code `unauthenticated`, sign in again on the phone.
3. For the code `unavailable` with the line `revision failed`, read the other errors of that period.
4. Let the phone send the batch again. The server does not revise a plan two times.
5. For the code `internal`, read the error line of the workout store. Firestore and its quota are the first suspects.
6. For a refused entry, find the code and the entity in the refused entries of the phone. A refused valid entry is a defect. Report it to the owner, with ids alone.
7. Read "Changes waiting to sync" again after the repair.

CAUTION: Do not delete the data of the app on the phone while "Changes waiting to sync" is above 0. The outbox holds the only copy of those sets.

### 2.3 A refusal of the caps

Signal: the line `ai call` has the status `capped`. A plan request then gives the code `resource_exhausted`, and the phone shows an error. A revision with a refused reviser call shows the reasons of the rules.

The caps are 1 USD for the user and 2 USD for the project in each month (D-188, D-190). The service `api` reads them from `LUNA_CAP_USER_USD` and `LUNA_CAP_PROJECT_USD`. The cap documents of Firestore hold the spend of the month.

1. Read the AI calls of the month, and add the cost of each call.
2. Compare the sum with the caps. A sum near a cap tells that the caps work.
3. Look for a loop: many calls of one role in a short period, or many retries of one request. A plan request makes 4 calls at most.
4. When the spend is correct, wait for the next month, or ask the owner for a higher cap.
5. A change of a cap is a change of the service. Make it only with the approval of the owner (D-25).

The workout screen and the rules need no AI call, so the owner can train during this incident.

### 2.4 A data loss

Signal: the phone or the plan screen shows fewer workouts or no plan, and the owner did not use "Delete all data". Or the backup list shows no backup of the last day.

1. Stop. Do not write to the database `(default)`.
2. Read the backup list. Each daily backup stays for 10 days (D-124).
3. Find the time of the loss from the logs. The line `history deleted` marks "Delete all data".
4. Ask the owner for the approval of a restore. A restore costs money.
5. Restore with `docs/deploy-and-rollback.md` section 6. Restore into a new database, and count each collection first.
6. When the loss is less than 7 days old, point-in-time recovery can read the database before the loss.

The phone keeps its own copy of each workout until a sync applies it. So a loss on the server alone does not lose the sets of the outbox.

### 2.5 An outage of the model

Signal: the line `ai call` has the status `timeout`, `error`, `incomplete`, or `malformed` for each call of a period. A plan request then gives the code `unavailable` after 4 calls. The collection `aiErrors` holds a record of each failed attempt for a short time.

The policy needs no model. A revision with a failed reviser call gives the targets of the rules and their reasons (D-288). So the owner can train during an outage with the current plan.

1. Read the AI calls of the period, and count each status.
2. Read the status page of OpenAI.
3. When the status is `refusal` or `malformed` for each call, look for a change of the model. Section 4 gives the steps.
4. Tell the owner not to request a new plan until the calls give `ok` again.
5. After the outage, read one `ok` call in the logs.

## 3. The policy replay

The command `go/cmd/replay` replays the stored data under the current policy version (D-176). It calls no model, so it costs nothing. It reads the store and writes nothing to it.

It reads two items for each user:

- Each decision record of the active plan. A record of a new plan replays through `policy.Decide` with its stored proposal. A record of a revision replays through `policy.Revise`.
- Each target copy of a finished workout (D-291). The copy replays through `policy.Next` and `policy.Check`. An override copy replays through `policy.CheckOverride`, against the recommendation that its workout keeps (D-293).
- The sessions of the active plan. The replay reads them against the rule `rotation.no-repeat` of policy version 9 (D-328). The store keeps no request of a plan, so the replay does not read the rule `rotation.cover`.

### 3.1 Run the replay

The emulator test `go/cmd/replay/replay_emulator_test.go` runs with `make emulator-test`. A run on the live store reads the workouts of the owner, so it needs the approval of the owner for each run (D-324).

1. Ask the owner for the approval of the run.
2. Sign in with `gcloud auth application-default login`.
3. Unset `FIRESTORE_EMULATOR_HOST`.
4. In `go/`, run `go run ./cmd/replay -project nk-workout-app-prod -live -out <path>`.
5. Keep the report outside the repository. Copy only its counts into a document.

The command refuses a run with no `-project`. It refuses a live run with no `-live` flag, and `-live` with the emulator.

### 3.2 Read the report

The report holds counts and rule ids alone (D-80).

| Field | It counts |
|---|---|
| `records.by_version` | the records of each stored policy version |
| `records.by_origin` | the records of a new plan and of a revision |
| `records.not_rebuilt` | the records whose input the replay could not make again |
| `records.same_input_hash` | the records whose new input has the stored hash |
| `records.changed` | the records whose target changed under the current version |
| `records.changed_by_rule` | each changed target under each rule id that explains the change |
| `records.changed_fields` | each changed field of a target, such as `load` or `follow_max` |
| `copies.same_as_rules` | the copies that equal the target of the rules on their date |
| `copies.stale_history` | the copies that pass the bounds only without the newest workouts of their plan |
| `copies.outside_bounds` | the copies that break a bound of the current version |
| `copies.violations_by_rule` | each broken rule of those copies |
| `layouts.applies` | the active plans whose exercises can make a split of the muscle groups |
| `layouts.breaks` | the active plans with a group in two sessions in a row |
| `layouts.groups_in_a_row` | each group in two sessions in a row, one time for each plan |

A different input hash tells that the replay made another input, or that a new version changed the form of the input. A copy of a Luna proposal can differ from the target of the rules, and it is still inside the bounds.

A copy in `stale_history` is not a breach. The store keeps no time of a revision. A phone can start a workout before the revision of the last workout, and then it shows the target before that revision. The live run of 2026-10-06 found one such copy.

A plan in `layouts.breaks` is a plan of an older policy version, or a breach of D-329. A new plan of version 9 or later never breaks the rotation, because the plan API refuses such an output. So the owner makes a new plan after a deploy of version 9.

A copy in `outside_bounds` can be a breach of D-23. Ask the owner for the approval to read that copy, and find its cause.

### 3.3 Before a new policy version merges

1. Change `policy.Version` and the rules in the pull request.
2. Run the replay on the emulator test, and update its expected report.
3. Ask the owner for the approval of a live run.
4. Run the replay on the live store with the code of the pull request.
5. Record the counts of `records.changed_by_rule` in the pull request.
6. Explain each changed rule id. An unexplained change stops the merge.

## 4. A change of the model

On 2026-10-06, the deprecation page of OpenAI listed no retirement of `gpt-6-luna`. It named `gpt-6-luna` as the replacement of `gpt-5.4-nano`, with a shutdown on 2027-04-01. Read the page again each month.

The role layer holds the one model id of the module (D-24). The file `go/internal/ai/role.go` gives the model, the effort, the prices, and the limits of each role. A test fails when another Go file names a model. The policy checks each target of each model, so a new model can not give a target outside the bounds (D-23).

Do these steps in one pull request:

1. Read the page of the new model. Record its prices, its context limit, and its efforts in `docs/research/platform-cloud-and-ai.md`.
2. Change the model, the effort, the prices, and the limits in `go/internal/ai/role.go`.
3. Change `PromptVersion` in `go/internal/ai/prompt.go` when the prompt changes.
4. Change `SchemaName` in `go/internal/ai/schema.go` when the schema changes.
5. Run `make go-test` and `make emulator-test` with the fake provider.
6. Run `go/cmd/lunaeval` with the fake provider. It costs nothing.
7. State the expected cost of a live evaluation, and ask the owner for the approval (D-25).
8. Run `go/cmd/lunaeval -live` with the cap that the owner approved.
9. Compare the schema pass rate, the policy refusals, the cost, and the longest call with the last evaluation.
10. Record the result in `docs/research/`, with the date.
11. After the merge, ask the owner for the approval of one live plan (D-212).

Each decision record names its model, its effort, and its prompt hash. So the records of the old model and of the new model stay apart in the replay.

## 5. Sources and dates

| Fact | Date read | Source |
|---|---|---|
| `nk-workout-app-prod` has 0 alert policies and 0 notification channels. | 2026-10-04 | Monitoring API |
| The log of a failed build ends with the line `ERROR`. | 2026-10-05 | `gcloud builds log` |
| The deprecation page lists no retirement of `gpt-6-luna`. | 2026-10-06 | `https://developers.openai.com/api/docs/deprecations` |
| The live replay: 9 records unchanged, 15 copies, 1 stale history, 0 outside the bounds. | 2026-10-06 | `docs/research/policy-replay.md` |
