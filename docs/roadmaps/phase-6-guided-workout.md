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

No open question blocks PR-29. Section 4 names the questions that each session asks. Q-194 and Q-202 stay open for Phase 7. Q-103 stays open for the deferred photo work.

## 3. Rules for each pull request of this phase

- The phone keeps each log first. Each change and its outbox entry go into one Dexie transaction (D-62, D-77, D-132).
- The API alone reads and writes Firestore (D-77). The rules of `firestore.rules` refuse each client read and write.
- A sync is idempotent. The server applies each client op id one time alone, and a replay changes nothing.
- The server checks each log with the bounds of D-164, as the domain model of `go/internal/domain` does.
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
| PR-30 | 6.1 | `feat: the workout screen and set log (PR-30)` | none |
| PR-31 | 6.2 | `feat: the rest timer and automatic advance (PR-31)` | none |
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

Before the work, the session reads the deploys of the merge of PR-28. It states the expected cost of one live plan at xhigh, and asks the owner (D-212). After the approval, the owner requests one plan on the iPhone and reads the bottom edge of the plan screen.

Concerns:

- a workout service in `proto/workoutapp/v1`. One unary call takes a batch of outbox entries and gives the result of each entry. Another call reads the logged sessions,
- the Firestore store of the logged sessions, with the form of the domain model of `go/internal/domain/log.go`,
- the idempotency of each client op id: the server applies an entry one time alone, and a replay gives the same result,
- the check of each log with the bounds of D-164, and the refusal of an unknown entity or schema version,
- the link of each logged session to the session of the plan that it started from (D-248).

Acceptance story: the emulator tests send a batch of entries for one session through the API, and the server holds each set one time. The same batch again changes nothing. The server refuses a set outside the bounds of D-164, and no other entry of the batch changes.

Checks: `make contract`, `make go-test`, `make emulator-test`, and `make verify`, free. Codex reviews PR-29.

Questions for the session: the Firestore paths of the logged sessions and of the applied op ids. Also the time that the server keeps an op id. Also the rule of a conflict with `baseVersion` (D-132), and the size limit of one batch.

### PR-30 - The workout screen and set log

Branch: `feat/pr-30-workout-screen`. Work area 6.1. It needs PR-29 on `main`.

Concerns:

- the start of a workout from the next session of the plan, or from another session that the owner picks (D-248),
- the set log of D-249: reps and weight from the target, and one tap on the reps in reserve. Pain and a note are optional (D-57, D-162),
- the cardio log of D-123: the duration and the effort rating, with the optional fields,
- the button "Report a symptom" on each workout screen, with the warning of D-153 and the confirmation of D-40 (D-251),
- the refusal of a new plan and of an exclusion while a workout is in progress (D-252),
- the wake lock, so the screen stays on during a workout,
- each log and its outbox entry in one Dexie transaction (D-132). The sync comes in PR-32.

Acceptance story: the browser tests start the next session, and log a set in three taps or fewer. A symptom report shows the warning, and the owner continues after the confirmation. After a stop and an open of the app, the workout and each logged set stay on the phone.

Checks: `make web`, `make verify`, and the Go checks of PR-29 when `go/` changes, free. Codex reviews PR-30.

Questions for the session: the fixed list of symptoms and the text of each warning (D-153). Also the support of the wake lock in the Home Screen app on iOS 27 with Chrome 154 (research, with the date of each fact). Also the step of the plus and minus buttons of the weight.

### PR-31 - The rest timer and automatic advance

Branch: `feat/pr-31-rest-timer`. Work area 6.2. It needs PR-30 on `main`.

Concerns:

- the rest timer from a stored end time. It starts when the owner logs a set, and the owner can change or dismiss it (D-59). The rest comes from the target (D-172),
- the preview of the next machine after the last set of an exercise, then the automatic advance (D-60),
- the edit of a logged set, the skip of an exercise, and "finish now", which ends the session early (D-63, D-170),
- visual cues alone, and no notification (D-58, D-61).

Acceptance story: the UI tests prove that the timer shows the correct time after a screen lock and a return. The advance comes after the last set. A skip and "finish now" give the correct session log.

Checks: `make web` and `make verify`, free. Codex reviews PR-31.

Questions for the session: the time of the preview before the advance, and the controls of the timer.

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

Questions for the session: the confirmation of a machine with no connection (D-201). Also the order of the inventory entries and the workout entries in one sync.

## 5. Exit of the phase

Phase 6 ends when PR-32 merges and the device check passes. These items give the exit evidence of `docs/roadmaps/high-level-roadmap.md`:

- The browser tests of PR-30 log a set in three taps or fewer.
- The UI tests of PR-31 prove that the timer shows the correct time after a screen lock. The advance comes after the last set.
- The offline tests of PR-32 replay a full workout with a dropped connection and a stop of the app. The server holds each set one time.
- After the deploy of PR-32, the owner completes a full workout on the iPhone with no connection. The logs reach Firestore when the app is open and online again.
