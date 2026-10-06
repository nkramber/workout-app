# Workout App - Phase 8 check

Status: the exit check of Phase 8 of `docs/roadmaps/high-level-roadmap.md`, and the check of the rotation of the muscle groups of policy version 9. It holds counts and rule ids alone (D-80). The owner decisions live in `docs/decisions.md`.

Date of the live reads: 2026-10-06, from 03:48Z to 04:06Z. The project is `nk-workout-app-prod`, and the code is the branch `feat/pr-42-rotation-and-phase-8-check` from base `efbafe7`.

## 1. Result

**Pass.** The policy refuses a plan that breaks the rotation, and Luna then plans again. The store, the logs, and the replay show no lost set, no refused valid sync, and no policy breach in the decision log.

| Measure | Result |
|---|---|
| Plans of the live model with the rotation | 100, and the policy refused 1 of them |
| Proposals of Luna that the policy accepted | 605 of 605 |
| Records of the active plan that the replay changed | 0 of 16 |
| Active plans that break the rotation of version 9 | 1 of 1, a plan of version 8 |
| Calls of the API with an HTTP error | 0 of 295 |
| Log entries of the API with the severity WARNING or more | 1, a request of a crawler for `/` |
| Workouts and sets in the store | 1 workout and 16 sets |
| Sets with a duplicate set id | 0 |
| Cost of the paid check | 0.163 USD, under the cap of 1 USD |

The owner dropped the four weeks of use (D-327). So this check reads the store, the logs, and the replay one time, and it does not cover four weeks.

## 2. The rotation of policy version 9

The owner asked for no legs, arms, or core in two sessions in a row, after a test workout. The decisions D-328 to D-331 give the rule:

- No primary group of an exercise is in two sessions in a row. The last session and the first session are in a row, because the week repeats.
- With 2 or 4 sessions, each selected group is in one of each two sessions in a row. So sessions 1 and 3 share their groups, and sessions 2 and 4 share theirs.
- With 3 sessions, each selected group is in one session.
- The rule does not need a selected group with no allowed exercise.

The body tables of D-210 give the primary groups. An exercise of two groups links them. For example, the chest press trains the chest and the triceps, and the shoulder press trains the shoulders and the triceps. So all three groups are in the same sessions.

When the allowed exercises can make no split, the rotation does not apply. A cardio exercise or an exercise with no group can fill a session. With the full gym of the owner, the rotation always applies.

The plan API refuses an output of Luna that breaks a rule, and it sends the cause to Luna with the retry (D-230). The cause holds rule ids, session numbers, and group ids alone.

The tests:

- `go/internal/policy/rotation_test.go` reads each rule for 2, 3, and 4 sessions.
- `go/internal/ai/rotation_test.go` proves that the layer refuses a plan with the same exercises in each session.
- `go/internal/plan/make_test.go` proves the retry and the layout of the saved plan.
- The fake provider gives a valid split, so the browser tests use plans of version 9.

## 3. The check with the live model

The owner approved 100 calls of the planner with a cap of 1 USD (D-25). Each call used `gpt-6-luna` at xhigh, the prompt `luna-prompt-v7` with the planner hash `be778498`, and policy version 9. The 20 synthetic profiles of `go/cmd/lunaeval` selected each group of D-210.

| Sessions | Calls | Valid plans | Refused by the rotation | Proposals accepted | Cost | Longest call |
|---|---|---|---|---|---|---|
| 2 | 60 | 60 | 0 | 366 of 366 | 0.0852 USD | 38.0 s |
| 3 | 20 | 20 | 0 | 122 of 122 | 0.0333 USD | 48.3 s |
| 4 | 20 | 19 | 1 | 117 of 117 | 0.0445 USD | 62.6 s |

The refused plan of 4 sessions had no exercise of the quadriceps and the glutes in any session. So the rule `rotation.cover` refused it. In the app, the plan API calls Luna again with that cause. 5 exercises of the valid plans of 4 sessions had no proposal of Luna, and the rules gave their targets (D-23).

## 4. The store and the logs

The owner approved three live reads (D-334). Each read was read-only.

The store holds 1 user, 1 workout that is not finished, 16 sets, and 17 applied sync entries. The 17 entries are the start of the workout and the 16 sets, so no set is missing from the server. No two sets have the same set id. The API logs show 2 calls of `DeleteHistory` after 2026-10-05, so the store holds no older workout.

The API logs from 2026-10-05T00:00Z to the read hold 295 calls of the contract, each with HTTP 200:

| Method | Calls |
|---|---|
| `SyncOutbox` | 42 |
| `GetMe` | 65 |
| `GetPlan` | 62 |
| `GetCatalog` and `GetInventory` | 57 each |
| `GetProfile` | 7 |
| `RequestPlan` | 3 |
| `DeleteHistory` | 2 |

No log line `plan attempt failed` exists in the period. The one entry of the severity WARNING is an HTTP 404 for `GET /` from a crawler, and no call of the app.

A refused sync entry leaves no trace on the server. The response gives the status `REFUSED` to the phone, the store keeps applied entries alone, and the API writes no log line for it. So the logs prove no refused call, and the store proves that each set of the workout arrived. The phone shows a refused entry as a failed sync. The owner reported that the test workout worked, and reported no failed sync.

## 5. The replay under policy version 9

The replay of `go/cmd/replay` read the active plan of the owner (`docs/operations.md` section 3).

| Measure | Live run |
|---|---|
| Records | 16, policy version 8, each of a new plan |
| Records with the stored input hash | 16 of 16 |
| Changed targets | 0 |
| Target copies | 0, because the store holds no finished workout |
| Plans where the rotation applies | 1 of 1 |
| Plans that break the rule `rotation.no-repeat` | 1 |
| Groups in two sessions in a row | chest, back, shoulders, biceps, triceps, quadriceps, hamstrings, glutes |

Version 9 changes no target, so the records give no breach of D-23. The active plan of version 8 holds the same exercises in each session, so it breaks the rotation. The owner deletes the history and makes a new plan after the deploy (D-333).

## 6. Limits

- The check covers one test workout, not four weeks of use (D-327).
- The server can not count a refused sync entry, as section 4 tells.
- The replay does not read the rule `rotation.cover`, because the store keeps no request of a plan.
