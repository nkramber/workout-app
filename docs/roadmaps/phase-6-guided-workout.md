# Workout App - Phase 6 focused roadmap: guided workout and offline logging

This roadmap splits Phase 6 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-10-03.

## 1. Scope and start state

Phase 6 delivers the workout flow for one-handed use in a gym with poor signal (D-62, D-71). At the end of the phase, the owner completes a full workout on the iPhone with no connection. The logs reach Firestore when the app is open and online again (D-77).

The phase makes no new call to Luna. The reviser call after a logged session comes in Phase 7 (D-64).

### 1.1 The exit evidence of Phase 5

The exit of Phase 5 holds. This session read these facts on 2026-10-03:

- PR-27 merged to `main` as `9d6f8eb`.
- The build `deploy-api` `cb49ad85` of `9d6f8eb` gave SUCCESS at 03:46:13Z. The build `deploy-web` `7345e0d5` gave SUCCESS at 03:44:46Z.
- At 03:49:25Z, the live `/version` and the live `/version.json` both named `9d6f8eb` (D-137).
- The revision `api-00013-w8w` kept the key of `openai-api-key:latest`, the caps of D-188, and the request timeout of 420 s.

The session stated the expected cost of the live check, and the owner approved it (D-25, D-212). The owner then requested plans in the live app on the iPhone:

| Request | Machines | Calls | Time of `RequestPlan` | Cost | Plan |
|---|---|---|---|---|---|
| 1, at 04:01Z | 1 confirmed machine | 1 | 13.8 s | 0.00083 USD | 1 exercise |
| 2, at 04:07Z | 10 machines | 1 | 19.6 s | 0.00104 USD | 2 sessions with 4 and 5 exercises |

The first request had one confirmed machine alone, so its plan had one exercise. A plan reads the confirmed machines alone, so this is the correct result (D-49, D-193). The owner then added machines, and the second request is the exit evidence.

The plan of request 2 is in `users/{uid}/plan/active`. Each of its 9 exercises has the source Luna, no violation, and no fallback. The plan names `gpt-6-luna` at medium effort, policy version 3, `luna-prompt-v3`, and the schema `luna_plan_v2`. No request wrote an error record of D-236. The spend of 2026-10 was 0.00187 USD, with no reservation open. The live app on the iPhone showed the plan, and the policy accepted it (D-23).

### 1.2 The follow-ups of the live check

The live check found three faults, and the owner asked for an effort check. PR-28 holds each of them (D-244):

- The owner had to tap "Make the list" before a save. Now the save makes the weight list (D-245).
- A cardio machine needed a confirmation of weights that it does not have. Now its save confirms it (D-246).
- In the Home Screen app, the shell of `100dvh` ended 62 pt above the bottom edge, and the plan text stopped there. Now the shell takes the screen height in the standalone mode, and the main region holds the bottom safe area (D-120).
- `docs/research/luna-effort-check.md` compares medium and xhigh with the same 50 calls. The owner chose xhigh for both roles (D-253).

The deploy of the merge of PR-28 changes the live API and the live web app. The session of PR-29 reads both deploys first. It then asks the owner for one live plan at xhigh (D-212), and the owner reads the bottom edge on the iPhone.

### 1.3 Inputs from the earlier phases

- `go/internal/domain/log.go` holds the set log, the cardio log, and the session log, with the bounds of D-164 and the pain rating of D-162.
- `go/internal/policy` reads the logs for the next targets (D-168, D-170). Phase 7 calls it after a session.
- `users/{uid}/plan/active` holds the current plan, with the sessions, the targets, and the decision records (D-226).
- `web/src/lib/db.ts` holds the offline store of Dexie and the outbox entry form of D-132. No sync call exists yet.
- The inventory and the profile use direct calls to the API, with an error when the phone has no connection (D-196).
- The service worker applies an update only after the owner taps "Update ready", and never during a workout (D-133).
- No workout service exists in `proto/workoutapp/v1`.

### 1.4 The live check at xhigh

The session of PR-29 read these facts on 2026-10-03:

- PR-28 merged to `main` as `fe7fe20`.
- The build `deploy-web` `9e3f4aff` gave SUCCESS at 05:27:39Z. The build `deploy-api` `1bfa4330` gave SUCCESS at 05:28:54Z.
- The live `/version` and the live `/version.json` both named `fe7fe20` (D-137).

The session stated the expected cost of one live plan at xhigh, and the owner approved it (D-212). At 06:10Z the owner requested one plan on the iPhone. The revision `api-00014-j4h` made 1 planner call at xhigh, with the status `ok`, a cost of 0.0022 USD, and a `RequestPlan` time of 29.7 s. The plan has 2 sessions with 5 and 4 exercises, no violation, and 10 minutes of cardio in each session. The owner saw the plan to the bottom edge of the screen (D-120).

The owner then asked for 20 to 30 minutes of cardio in each session, and PR-29 holds the change (D-254, D-255).

### 1.5 The live check of the cardio rule

The session of PR-30 read these facts on 2026-10-03:

- PR-29 merged to `main` as `ac739c1`.
- The build `deploy-web` `5ba685fe` gave SUCCESS at 07:56:55Z. The build `deploy-api` `65ad6292` gave SUCCESS at 07:58:06Z.
- The live `/version` and the live `/version.json` both named `ac739c1`, and the revision `api-00015-5tp` served all traffic (D-137).

The session stated the expected cost of one live plan at xhigh with `luna-prompt-v4`, and the owner approved it (D-212). At 08:02Z the owner requested one plan on the iPhone. The API made 1 planner call at xhigh, with the status `ok` and a cost of 0.0039 USD. The plan has 2 sessions with 6 exercises each, 1 attempt, and no violation. Each session has 20 minutes of cardio, so the rule of D-255 holds.

Each of the 12 exercises had the reason "The exercise log has a gap". The owner has no log yet, so each exercise had its start target (D-150). The evidence list of prompt v4 had no item for a new exercise, so Luna named a gap. Prompt v5 of PR-30 corrects the reason (D-262).

### 1.6 The live check of prompt v5

The session of PR-31 read these facts on 2026-10-03:

- PR-30 merged to `main` as `eaa18b6`.
- The build `deploy-web` `247ef56d` gave SUCCESS at 16:03:37Z. The build `deploy-api` `715c817d` gave SUCCESS at 16:05:29Z.
- The live `/version` and the live `/version.json` both named `eaa18b6`, and the revision `api-00016-c2h` served all traffic (D-137).

The session stated the expected cost of one live plan at xhigh with `luna-prompt-v5`, and the owner approved it (D-212). At 16:48Z the owner requested one plan on the iPhone. The API made 1 planner call at xhigh, with the status `ok` and a cost of 0.0019 USD. The plan has 2 sessions with 5 and 4 exercises, 1 attempt, and 25 minutes of cardio in each session (D-255).

The reason of each of the 9 exercises is "This exercise is new.", with no gap, so prompt v5 corrects the reason (D-262). The owner saw the wake lock hold, and logged a set on the iPhone.

The owner then found two faults on the workout screen of PR-30. First, a calibration set did not change the load of the working sets, so it was only one more set. Second, the screen showed "The screen can turn off." after the owner left the app and came back. PR-31 holds both corrections (D-266 to D-268, D-271).

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-256, the effort comparison | In PR-28, with a paid run at each effort. | D-242 |
| Q-257, a change of the role | In PR-28. | D-243 |
| Q-258, the scope of PR-28 | The exit evidence, the effort check, the role change, the three fixes, and this roadmap. | D-244 |
| Q-259, the weight list | The save makes it. | D-245 |
| Q-260, a cardio machine | The save confirms it. | D-246 |
| Q-261, the split | PR-29 to PR-32, as section 4 gives them. | D-247 |
| Q-262, the session of a workout | The next session that the owner did not do. The owner can pick another. | D-248 |
| Q-263, the set log | Reps and weight from the target. One tap on the reps in reserve logs the set. | D-249 |
| Q-264, the inventory offline | PR-32 puts the changes of the inventory in the outbox. | D-250 |
| Q-265, a warning symptom | A button "Report a symptom" on each workout screen. | D-251 |
| Q-266, a plan request during a workout | The plan screen refuses it until the workout ends. | D-252 |
| Q-267, the effort of Luna | xhigh, for both roles. | D-253 |
| Q-268, the place of the cardio rule | PR-29, with a wider milestone. | D-254 |
| Q-269, the cardio of a plan | 20 to 30 minutes in each session, when the owner likes a cardio exercise. | D-255 |
| Q-270, the Firestore paths | `users/{uid}/workouts/{workoutId}` and `users/{uid}/ops/{opId}`. | D-256 |
| Q-271, the time of an op id | With no end date. | D-257 |
| Q-272, a conflict with `baseVersion` | The phone wins for a workout entry. | D-258 |
| Q-273, the size of a batch | 100 entries. | D-259 |
| Q-274, the least time of a cardio log | None. The log holds the true duration. | D-260 |
| Q-275, the length of a note | 280 characters. | D-261 |
| Q-276, the reason of a new exercise | Prompt v5, in PR-30. | D-262 |
| Q-277, the symptoms and the warnings | Seven symptoms. "You reported <symptom>. Stop this exercise." | D-263 |
| Q-278, the step of the weight buttons | The next weight of the list of the machine. | D-264 |
| Q-279, the wake lock | The Screen Wake Lock API, with a notice when the phone refuses it. | D-265 |
| Q-280, the place of the calibration step | PR-31. | D-266 |
| Q-281, the count of calibration sets | One. The table gives the load of the working sets one time. | D-267 |
| Q-282, the reps in reserve of a calibration set | 0 to 6+. A working set keeps 0 to 4+. | D-268 |
| Q-283, the time of the preview | 10 seconds, with "Go now". | D-269 |
| Q-284, the controls of the timer | "-15 s", "+15 s", and "Dismiss". | D-270 |
| Q-285, the wake lock after a return | A request at each tap, focus, `pageshow` event, and return. | D-271 |

No open question blocks PR-32. Section 4 names the questions that each session asks. Q-194 and Q-202 stay open for Phase 7. Q-103 stays open for the deferred photo work.

## 3. Rules for each pull request of this phase

- The phone keeps each log first. Each change and its outbox entry go into one Dexie transaction (D-62, D-77, D-132).
- The API alone reads and writes Firestore (D-77). The rules of `firestore.rules` refuse each client read and write.
- A sync is idempotent. The server applies each client op id one time alone, and a replay changes nothing.
- The server checks each log with the bounds of D-164, as the domain model of `go/internal/domain` does. A note has 280 characters or fewer (D-261).
- For a workout entry, the phone wins. The server records the base version, and does not refuse a different one (D-258).
- The workout screens use large targets, few taps, and little typing (D-71). Each cue is visual alone, with no audio and no vibration (D-58). The app gives no notification (D-61).
- A warning symptom gives the warning of D-153. The owner can continue after a confirmation (D-40).
- No log, metric, or error report holds the pain rating, a note, a symptom, or a load. It holds ids alone (D-80).
- The tests use synthetic data alone. The repository is public.
- Each pull request with code gets the review of the other provider (D-15).
- A merge that changes `go/` deploys the API, and a merge that changes `web/` deploys the web app (D-137). The next session reads each deploy first.
- After a change in `proto/`, run `make proto`, and commit the generated code. `make contract` runs `buf breaking` against `main`.

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-28 | all | `docs: the Phase 6 focused roadmap (PR-28)` | the two runs of the effort check (D-242) |
| PR-29 | 6.1, 6.3 | `feat: the workout log API and store (PR-29)` | the live plan at xhigh before the work (D-212) |
| PR-30 | 6.1 | `feat: the workout screen and set log (PR-30)` | the live plan at xhigh before the work (D-212) |
| PR-31 | 6.2 | `feat: the rest timer and automatic advance (PR-31)` | the live plan of prompt v5 before the work (D-212) |
| PR-32 | 6.3 | `feat: the outbox sync (PR-32)` | none |

### PR-28 - The Phase 6 focused roadmap

Branch: `docs/pr-28-phase-6-roadmap`.

Concerns:

- this file, with the exit evidence of Phase 5 in section 1.1 (D-212),
- the flag `-effort` of `go/cmd/lunaeval`, the two paid runs, the report `docs/research/luna-effort-check.md`, and the effort xhigh of both roles (D-242, D-243, D-253),
- the weight list at the save (D-245), and the confirmation of a cardio machine at the save (D-246),
- the shell to the bottom edge of the Home Screen app (D-120),
- the owner answers of section 2, with their change of `docs/roadmaps/high-level-roadmap.md` and `docs/design.md`.

Acceptance story: the live app on the iPhone showed a plan that the policy accepted. Section 1.1 records it as the exit evidence of Phase 5. The report compares medium and xhigh with measured numbers. The browser tests prove the weight list, the cardio confirmation, and the shell to the bottom edge. This roadmap maps each work area of Phase 6 to one or more pull requests, and `make verify` passes.

Checks: `make verify`, `make go-test`, `make emulator-test`, `make web`, and `make pr-check`, free. The effort check costs money, and the owner approved it with a cap of 2 USD for each run. Codex reviews PR-28, because it changes code (D-15).

### PR-29 - The workout log API and store

Branch: `feat/pr-29-workout-api`. Work areas 6.1 and 6.3. It needs PR-28 on `main`.

Before the work, the session reads the deploys of the merge of PR-28. It states the expected cost of one live plan at xhigh, and asks the owner (D-212). After the approval, the owner requests one plan on the iPhone and reads the bottom edge of the plan screen. Section 1.4 gives the result.

Concerns:

- a workout service in `proto/workoutapp/v1`. One unary call takes a batch of 100 outbox entries or fewer, and gives the result of each entry (D-259). Another call reads the logged sessions,
- the Firestore store of the logged sessions at `users/{uid}/workouts/{workoutId}`, with the form of the domain model of `go/internal/domain/log.go` (D-256),
- the idempotency of each client op id: the server applies an entry one time alone, and a replay gives the same result. Each applied op id stays at `users/{uid}/ops/{opId}` with no end date (D-257),
- the check of each log with the bounds of D-164, D-260, and D-261, and the refusal of an unknown entity or schema version. For a workout entry, the phone wins (D-258),
- the link of each logged session to the session of the plan that it started from (D-248),
- the cardio rule of D-255 in the prompt, the output check, the fake provider, and the plan screen (D-254).

Acceptance story: the emulator tests send a batch of entries for one session through the API, and the server holds each set one time. The same batch again changes nothing. The server refuses a set outside the bounds of D-164, and no other entry of the batch changes. The unit tests show that an output breaks the cardio rule when a session has fewer than 20 minutes of cardio. A session with no cardio breaks it too, when the owner likes a cardio exercise. Such an output uses one retry (D-230).

Checks: `make contract`, `make go-test`, `make emulator-test`, `make web`, and `make verify`, free. Codex reviews PR-29.

The owner answered the questions of the session: Q-270 to Q-275 (D-256 to D-261).

### PR-30 - The workout screen and set log

Branch: `feat/pr-30-workout-screen`. Work area 6.1. It needs PR-29 on `main`.

Concerns:

- the start of a workout from the next session of the plan, or from another session that the owner picks (D-248),
- the set log of D-249: reps and weight from the target, and one tap on the reps in reserve. Pain and a note are optional (D-57, D-162),
- the cardio log of D-123: the duration and the effort rating, with the optional fields,
- the button "Report a symptom" on each workout screen, with the warning of D-153 and the confirmation of D-40 (D-251),
- the refusal of a new plan and of an exclusion while a workout is in progress (D-252),
- the wake lock, so the screen stays on during a workout,
- each log and its outbox entry in one Dexie transaction (D-132), with the entities and payloads of `proto/workoutapp/v1/workout_service.proto`. The sync comes in PR-32,
- the reason of a new exercise in prompt v5, after the live check of section 1.5 (D-262).

An open workout also holds the update of the app, as D-133 says.

Acceptance story: the browser tests start the next session, and log a set in three taps or fewer. A symptom report shows the warning, and the owner continues after the confirmation. After a stop and an open of the app, the workout and each logged set stay on the phone.

Checks: `make web`, `make verify`, and the Go checks of PR-29 when `go/` changes, free. Codex reviews PR-30.

Answers of the session: the symptoms and the warnings of D-263, the step of the weight buttons of D-264, and the wake lock of D-265. On 2026-09-29, the device checklist showed that the wake lock works in the Home Screen app on iOS 27.0 (`docs/research/iphone-platform-spike.md`). Chrome 154 added that app, and a request after the background worked. WebKit fixed the lock for Home Screen apps in iOS 18.4, on 2025-03-31 (PC-2, PC-8).

### PR-31 - The rest timer and automatic advance

Branch: `feat/pr-31-rest-timer`. Work area 6.2. It needs PR-30 on `main`.

Concerns:

- the rest timer from a stored end time. It starts when the owner logs a set, and the owner can change or dismiss it with "-15 s", "+15 s", and "Dismiss" (D-59, D-270). The rest comes from the target (D-172),
- the preview of the next machine for 10 seconds after the last set of an exercise, with "Go now", then the automatic advance (D-60, D-269),
- the edit of a logged set and the skip of an exercise, each with its outbox entry (D-63, D-170, D-132). "Finish now" ends the session early when an exercise that the owner did not skip has a set with no log,
- visual cues alone, and no notification (D-58, D-61),
- the calibration step on the workout screen: one calibration set, and the table of D-150 gives the load of the working sets one time. A calibration set offers 0 to 6+ reps in reserve. The plan holds the 4 loads for each weight of the machine. The phone needs no network, and the policy goes to version 4 (D-266 to D-268),
- the wake lock: a request at each tap, focus, `pageshow` event, and return, with the error name in the notice (D-271).

Acceptance story: the UI tests prove that the timer shows the correct time after a screen lock and a return. The advance comes after the last set. A skip and "finish now" give the correct session log. After a calibration set, the working sets show the load of the calibration table with no network. A refused wake lock comes back at the next tap.

Before the work, the session read the deploys of PR-30. Then one live plan of `luna-prompt-v5` checked the reason of a new exercise (D-212, D-262). Section 1.6 records the check.

Checks: `make contract`, `make go-test`, `make emulator-test`, `make web`, and `make verify`, free. Codex reviews PR-31.

After the deploy of the merge, the owner checks the wake lock after a return to the app. The owner also logs one calibration set on the iPhone.

The owner answered the questions of the session: Q-280 to Q-285 (D-266 to D-271).

### PR-32 - The outbox sync

Branch: `feat/pr-32-outbox-sync`. Work area 6.3. It needs PR-31 on `main`.

Concerns:

- the sync of the outbox through the call of PR-29: on open, on focus, and on reconnect,
- no sync while the app is closed, because iOS has no background sync (D-21),
- a retry after a failed sync, and the removal of each entry that the server applied,
- the offline copy of the inventory, with its changes in the outbox (D-196, D-250),
- the state of the sync on the screen, so the owner knows when each log reached the server.

Acceptance story: the offline tests replay a full workout with a dropped connection and a stop of the app. After the reconnect, the server holds each set one time. An inventory change with no connection reaches the server after the reconnect.

Checks: `make web`, `make go-test`, `make emulator-test`, and `make verify`, free. Codex reviews PR-32.

After the deploy of the merge, the owner completes a full workout on the iPhone with no connection, and then opens the app online. The Phase 7 roadmap session records the result (recommendation, as this roadmap does for Phase 5).

Questions for the session: the confirmation of a machine with no connection (D-201). Also the place on the phone of an entry that the server refused. Also the order of the inventory entries and the workout entries in one sync.

## 5. Exit of the phase

Phase 6 ends when PR-32 merges and the device check passes. These items give the exit evidence of `docs/roadmaps/high-level-roadmap.md`:

- The browser tests of PR-30 log a set in three taps or fewer.
- The UI tests of PR-31 prove that the timer shows the correct time after a screen lock. The advance comes after the last set. A calibration set gives the load of the working sets with no network.
- The offline tests of PR-32 replay a full workout with a dropped connection and a stop of the app. The server holds each set one time.
- After the deploy of PR-32, the owner completes a full workout on the iPhone with no connection. The logs reach Firestore when the app is open and online again.
