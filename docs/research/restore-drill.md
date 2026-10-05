# Workout App - restore drill and Phase 7 check

Status: the drill report of work area 8.1 of `docs/roadmaps/high-level-roadmap.md`, with the iPhone check of Phase 7 (D-310). It holds times and counts alone. The owner decisions live in `docs/decisions.md`.

Date of the drill: 2026-10-05, from 18:52:55Z to 19:30:40Z. The project is `nk-workout-app-prod`, and the tool is Google Cloud SDK 533.0.0.

## 1. Result

**Pass.** The three iPhone checks of Phase 7 passed, so Phase 7 ends. A daily backup restored into a new database, and the count of each collection agrees with the logs. The rollback drill moved the traffic of `api` to the last revision and back, and it moved Hosting to the last release and back.

| Measure | Result |
|---|---|
| iPhone checks of Phase 7 | 3 of 3 passed, the owner on the iPhone |
| Paid calls of the check | 1 planner call and 3 reviser calls, each `ok` |
| Known cost of the calls | 0.004379230 USD, under the estimate of about 0.006 USD |
| Restore of the backup | 8 min 48 s, state `SUCCESSFUL` |
| Collection paths | 10 in `(default)` and 9 in the restore, each with a count. `users/{uid}/history` is in `(default)` alone. |
| Differences from `(default)` | 3 collection paths, each with a known cause |
| Rollback of `api` | 6 s to the last revision, 5 s back |
| Rollback of Hosting | under 1 s to the last release, under 1 s back |
| Log entries with the severity WARNING or more | 0 |

## 2. The Phase 7 check

### 2.1 The deploys

PR-38 merged to `main` as `e09b7ff`. It changed `go/` and `web/`, so both deploys ran (D-137):

- The build `deploy-web` `ddf4e3c4` gave SUCCESS at 18:15:42Z.
- The build `deploy-api` `12c94777` gave SUCCESS at 18:17:06Z.
- The live `/version` and `/version.json` named `e09b7ff`. The revision `api-00022-tjh` had all traffic.

### 2.2 The checks of the owner

The session stated the cost of one live plan and one revision with policy version 8 first, and the owner approved it (D-212). The owner applied "Update ready", made a new plan, and did the three checks of D-310:

1. On a new exercise, the first set is the calibration, and the other sets use its weight (D-297, D-299). Pass.
2. On an exercise with history, a first set two weights above the target gives one weight above the target (D-306, D-307). Pass.
3. When each exercise is done, the exercise list collapses (D-298). Pass.

### 2.3 The logs

The session read the logs of the revision `api-00022-tjh` with ids and counts alone:

| Time (UTC) | Event | Result |
|---|---|---|
| 18:52:55 | `DeleteHistory`, 0.6 s | HTTP 200, 11 workouts and 0 error records deleted |
| 18:53:05 | `RequestPlan`, 29.8 s | 1 planner call, `ok`, 0.002253925 USD |
| 19:16:35 | `SyncOutbox`, 7.4 s | 1 reviser call, `ok`, 0.000708925 USD |
| 19:16:42 | revision 1 of 5 exercises | 1 reason of Luna, 4 reasons of the rules |
| 19:16:56 | `SyncOutbox`, 10.1 s | 1 reviser call, `ok`, 0.000682380 USD |
| 19:17:06 | revision 2 of 5 exercises | 1 reason of Luna, 4 reasons of the rules |
| 19:17:59 | `SyncOutbox`, 7.9 s | 1 reviser call, `ok`, 0.000734380 USD |
| 19:18:07 | revision 3 of 5 exercises | 5 reasons of Luna, 0 reasons of the rules |

The owner used "Delete all data" before the new plan (D-314). The check made 3 revisions, and the estimate was for one revision. The known cost of the 4 calls was 0.004379230 USD, inside the caps of D-188. Each call had a known cost, and the reserve of each call was 0.0184 USD or less.

### 2.4 The cause of each reason of the rules

The owner approved a read of the plan document, with ids alone. Revision 3 wrote the plan after revisions 1 and 2. So the session read the plan of each earlier revision with point-in-time recovery, at a `readTime` of 19:17:00Z and 19:18:00Z.

| Revision | Exercises with a reason of the rules | Rule | Cause |
|---|---|---|---|
| 1 | `seated_leg_curl`, `chest_press`, `lat_pulldown`, `abdominal_crunch` | `progress.skipped` | `no-reason` |
| 2 | the same 4 exercises | `progress.skipped` | `no-reason` |
| 3 | none | none | none |

Each reason of the rules comes from an exercise that the workout skipped. After revision 3, each of the 9 exercises of the plan has policy version 8. The 5 exercises of the first session have the limit `follow_max_tenth_lb`. The 4 exercises of the second session have no history, so they keep the first-set calibration.

## 3. The restore

### 3.1 Method

The session did the steps of `docs/deploy-and-rollback.md` section 6:

1. The newest READY backup was `e50e86ed`, with the snapshot time 07:28:41Z. 6 backups had the state READY.
2. The restore into the new database `restore-20261005` started at 19:20:09Z.
3. The operation gave `SUCCESSFUL` at 19:28:57Z, after 8 min 48 s.
4. The session counted each collection of both databases with the REST API of Firestore.
5. The delete step of section 6 failed, because the new database had delete protection.
6. The session turned off the delete protection of `restore-20261005` alone, and deleted it at 19:30:40Z.

The count walks the tree with `listCollectionIds`. For each collection, it runs a `count` aggregation. The paths below put `{uid}` in place of the id of the user. The count of `(default)` ran from 19:20:41Z to 19:20:55Z, and the count of the restore ran from 19:29:22Z to 19:29:53Z.

### 3.2 Counts

| Collection | `(default)` at 19:20Z | Restore of 07:28Z | Cause of the difference |
|---|---|---|---|
| `aiSpend` | 1 | 1 | none |
| `allowlist` | 1 | 1 | none |
| `users` | 0 | 0 | none. The document of the user has no fields, and it holds the collections below. |
| `users/{uid}/aiSpend` | 1 | 1 | none |
| `users/{uid}/history` | 1 | 0 | `DeleteHistory` wrote the fence document `deleted` at 18:52:56Z. |
| `users/{uid}/inventory` | 1 | 1 | none |
| `users/{uid}/ops` | 25 | 115 | `DeleteHistory` deleted the 115 ops: 22 of workouts, 91 of sets, and 2 of cardio. Each of the 25 ops of `(default)` came after it. |
| `users/{uid}/plan` | 1 | 1 | none. `(default)` holds the new plan. |
| `users/{uid}/profile` | 1 | 1 | none |
| `users/{uid}/workouts` | 3 | 11 | `DeleteHistory` deleted 11 workouts, and the owner then did 3 workouts. |

The restore holds each workout that "Delete all data" deleted. So a restore can bring back the deleted data, as D-316 tells. The drill deleted the new database after the count, and it copied no document into `(default)`.

### 3.3 Findings

- A restored database gets the delete protection of its source. So the delete step of section 6 failed with `FAILED_PRECONDITION`, until a step turned off that protection.
- The restore of the database of one user took 8 min 48 s. A plan for a repair must allow for that time.
- A read with a `readTime` of point-in-time recovery gives the old plan document. The drill used it to read the overwritten revisions of section 2.4.
- The backup record of `gcloud firestore backups describe` holds no size. So the session can not state the billed size of a restore before the restore.
- The pricing pages of Firestore did not give the unit price of a restore on 2026-10-05 through the fetch of the session. The cost of the restore is unverified, and the billing data of the drill was not read.

## 4. The rollback drill

### 4.1 The service `api`

The session did the steps of `docs/deploy-and-rollback.md` section 4.1. The last revision before `api-00022-tjh` was `api-00021-xxp`, of `4bbf6c8`.

| Time (UTC) | Step | Result |
|---|---|---|
| 19:21:13 to 19:21:19 | `update-traffic --to-revisions=api-00021-xxp=100` | done |
| 19:21:19 | `/version` | `4bbf6c8` |
| 19:21:23 to 19:21:28 | `update-traffic --to-latest` | `api-00022-tjh` has 100% |
| 19:21:29 | `/version` | `e09b7ff` |

During the pin, the session called `/version` alone. The old revision has policy version 7, and it read no plan of policy version 8. The drill did not test a longer rollback over a change of the policy version.

### 4.2 Hosting

Section 4.2 gives the console alone. The session used the Firebase Hosting REST API in its place. A `POST` to `channels/live/releases` with the `versionName` of an earlier release makes a release of the type `ROLLBACK`.

| Time (UTC) | Step | Result |
|---|---|---|
| 19:22:01.9 | release of version `6aed659c85ee649c` | type `ROLLBACK` |
| 19:22:02 | `/version.json` | `4bbf6c8` |
| 19:22:02.8 | release of version `b12859f4771f0074` | type `ROLLBACK` |
| 19:22:04 | `/version.json` | `e09b7ff` |

The list of the releases names the commit of each release in its message. The service worker of the phone did not get a check during the 3 s of the drill.

## 5. Changes from the drill

- `docs/deploy-and-rollback.md` section 4.2 gives the REST steps of a Hosting rollback.
- `docs/deploy-and-rollback.md` section 6 turns off the delete protection of the restored database before its delete, and gives the time and the count method.
- `docs/setup-gcp.md` names this report for the backups, and gives the cost of a restore.

## 6. Sources and dates

| Fact | Date read | Source |
|---|---|---|
| A restore writes a new database, and Firestore charges a restore on the size of the backup. | 2026-10-05 | Firebase docs, "Back up and restore data" and "Understand Cloud Firestore billing" |
| The unit price of a restore did not show on the pricing page through the fetch of the session (unverified). | 2026-10-05 | `cloud.google.com/firestore/pricing` |
| A restored database had the state `DELETE_PROTECTION_ENABLED`, and its delete gave `FAILED_PRECONDITION`. | 2026-10-05 | `gcloud firestore databases describe` and `delete` |
| A release of an earlier version through the Hosting REST API has the type `ROLLBACK`. | 2026-10-05 | `firebasehosting.googleapis.com/v1beta1` |
