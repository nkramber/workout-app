# Workout App - Phase 7 focused roadmap: adaptation loop

This roadmap splits Phase 7 of `docs/roadmaps/high-level-roadmap.md` into pull requests. `docs/roadmaps/README.md` gives the rules. The high-level roadmap keeps the objective, the order, and the exit evidence of the phase.

The date of this version is 2026-10-04.

## 1. Scope and start state

Phase 7 adapts the next session from the logged history under the policy (D-43, D-64). At the end of the phase, the owner sees validated targets for the next session after each session, each with a concise reason (D-68). An override keeps the recommendation and the reason (D-69). Missed sessions and long breaks change the targets (D-66).

The rules of the policy give each target of a revision. Luna writes the reason alone (D-288).

### 1.1 The exit evidence of Phase 6

The exit of Phase 6 holds. Section 5 of `docs/roadmaps/phase-6-guided-workout.md` gives the tests of each pull request, and its section 1.8 gives the offline workout on the iPhone. This session read these facts on 2026-10-04:

- PR-33 merged to `main` as `6274c19` at 03:10:55Z.
- PR-33 changed only `web/` and `docs/`, so only the build `deploy-web` ran (D-137). The build `deploy-web` `6fe91aa9` gave SUCCESS at 03:13:03Z.
- The live `/version.json` named `6274c19`. No build `deploy-api` ran, and the live `/version` still named `ee3b89b`.

The owner applied "Update ready" on the iPhone, and started a workout. The owner left the app and came back with no tap. The screen stayed on, and the notice "The screen can turn off." did not show (D-285).

### 1.2 The live plan of policy version 5

Before this check, the active plan had policy version 4. A new plan must have policy version 5, with a rest of 60 seconds for each exercise (D-279).

The session stated the expected cost of one live plan at xhigh with `luna-prompt-v5`. It was about 0.002 to 0.004 USD for each call, with 4 calls or fewer. The owner approved it (D-212). The owner then requested one plan on the iPhone. The session read the logs and Firestore with ids and counts alone:

- At 03:17:54Z, the revision `api-00018-vhd` made 1 planner call, with the status `ok` and a cost of 0.0037 USD.
- The plan has 1 attempt, policy version 5, `luna-prompt-v5`, the effort xhigh, and the schema `luna_plan_v2`.
- The plan has 2 sessions with 8 exercises each, and 20 minutes of cardio in each session (D-255).
- Each of the 16 exercises has a rest of 60 seconds (D-279), and the reason "This exercise is new." (D-262).
- Each exercise has a calibration row for each weight of its machine. A machine has 29 rows, and the cable station has 15 rows.

### 1.3 Inputs from the earlier phases

The session read the code of `6274c19` for this list:

- `policy.Next` in `go/internal/policy/progress.go` gives the next target of one exercise from its history. It holds the rules of D-147, D-151, D-168 to D-174, and D-179.
- `policy.Decide` in `go/internal/policy/record.go` checks a proposal and gives the decision record of D-176. The policy version is 5.
- `RoleReviser` in `go/internal/ai/role.go` and `Client.Revise` in `go/internal/ai/client.go` exist. Only `go/cmd/lunaeval` calls the reviser. No RPC calls it.
- `go/cmd/lunaeval` runs scenarios A to F of the high-level roadmap through the reviser, with the fake provider or with Luna. It has no `make` target.
- The workout store holds each workout at `users/{uid}/workouts/{workoutId}`, with a link to the session of the plan (D-248, D-256). It holds no copy of the target that the owner saw.
- A new plan replaces the old plan and its decision records, with no history (D-227). Each new exercise of a plan starts as a return after a break of 91 days or more (D-238).
- No code reads the logged history for the next targets. `WorkoutService` has `SyncOutbox` and `ListWorkouts`, and the web app does not call `ListWorkouts`.
- The phone repeats the sessions of the one week of the plan, with the same targets (D-211).
- A symptom report shows the warning of D-263, and the app keeps no record of it.
- No override exists in the code or in the contract.

In the evaluation of Phase 3, Luna proposed the target of the rules in each of 197 decisions. Some reasons of Luna were not accurate (`docs/research/phase-3-check.md`). This result gave D-288.

## 2. Owner answers for this phase

| Question | Answer | Decision |
|---|---|---|
| Q-300, a third work area | Yes: work area 7.3, "Reviser evaluation". | D-286 |
| Q-301, the split | Three pull requests: PR-35 to PR-37, as section 4 gives them. | D-287 |
| Q-202, the role of Luna in the targets | For a revision, Luna writes the reason alone, and the rules give each target. | D-288 |
| Q-194, the reactive deload of REC-7 | Yes, in work area 7.2, with a new policy version. | D-289 |

These questions stay open, and section 4 names the session that asks each one:

| Question | Session |
|---|---|
| Q-302, the targets that a revision changes | PR-35 |
| Q-303, the place of the target that the owner saw | PR-35 |
| Q-304, the time of a revision with no connection | PR-35 |
| Q-305, the fields of an override | PR-36 |
| Q-306, a missed session under 14 days | PR-36 |
| Q-307, the numbers of the deload | PR-36 |
| Q-308, the size and the cap of the paid evaluation | PR-37 |

Q-103 stays open for the deferred photo work.

## 3. Rules for each pull request of this phase

- The rules of the policy give each target of a revision. Luna writes the reason alone (D-288).
- The policy checks each target before the owner sees it (D-23).
- A reason names the logged evidence (D-68). A check refuses a reason of Luna that names no logged set, and the reason of the rules then shows (D-288).
- The same history and the same policy version give the same targets (section 6.3 of the high-level roadmap).
- A change of a rule changes the policy version. The golden files of `go/internal/policy/testdata` change in the same pull request.
- No call site names a model. A call names a role (D-24).
- No log, metric, or error report holds a reason, a load, a pain rating, or a note. It holds ids alone (D-80).
- A paid run needs the approval of the owner (D-25). The session states the expected cost first (D-212). The monthly caps of D-188 apply.
- The tests use synthetic data alone. The repository is public.
- Each pull request with code gets the review of the other provider (D-15).
- A merge that changes `go/` deploys the API, and a merge that changes `web/` deploys the web app (D-137). The next session reads each deploy first.
- After a change in `proto/`, run `make proto`, and commit the generated code. `make contract` runs `buf breaking` against `main`.

## 4. Pull requests

The PR-<n> order is the order of work. Each pull request needs the one before it on `main`.

| Id | Work area | Title | Paid step |
|---|---|---|---|
| PR-34 | all | `docs: the Phase 7 focused roadmap (PR-34)` | the live plan of policy version 5 before the work (D-212) |
| PR-35 | 7.1 | `feat: the revision after a session (PR-35)` | none, because the tests use the fake provider |
| PR-36 | 7.2 | `feat: the overrides and the disruptions (PR-36)` | the live revision before the work (D-212) |
| PR-37 | 7.3 | `test: the reviser evaluation (PR-37)` | the live revision before the work (D-212), and the paid evaluation (D-25, D-286) |

### PR-34 - The Phase 7 focused roadmap

Branch: `docs/pr-34-phase-7-roadmap`.

Concerns:

- this file, with the exit evidence of Phase 6 in section 1.1, and the live plan of policy version 5 in section 1.2 (D-212, D-285),
- the split of the work areas into pull requests, each with its concerns, acceptance story, checks, and questions (D-287),
- the paid step of each pull request (D-25, D-212),
- the owner answers of section 2, with their change of `docs/roadmaps/high-level-roadmap.md` and `docs/design.md` (D-286, D-288, D-289).

Acceptance story: this roadmap names each pull request of Phase 7 with its concerns, its acceptance story, its checks, and its questions. The high-level roadmap holds work area 7.3, and `make verify` passes.

Checks: `make verify` and `make pr-check`, free. Codex reviews PR-34, because it sets the role of Luna in the targets (D-288) and the reactive deload (D-289).

### PR-35 - The revision after a session

Branch: `feat/pr-35-revision`. Work area 7.1. It needs PR-34 on `main`.

PR-34 changes only `docs/`, so its merge deploys nothing. The session reads no deploy first.

Concerns:

- the record of the target that the owner saw, for the history of each exercise (Q-303),
- a read of the logged history of each exercise into `policy.Input`, and the revision at the end of a workout (Q-302, Q-304),
- the targets of `policy.Next`, the policy check, and the write of the revised targets with their decision records (D-23, D-176),
- the reviser call for the reason alone, with a new prompt version and a schema for a reason (D-288),
- the check of each reason against the logged sets, and the reason of the rules as the fallback (D-68, D-288),
- the screen of the next targets with each reason after a workout, and the offline copy of the revised plan (D-278),
- the scenarios A to F of the high-level roadmap end to end with the fake provider.

Recommendation: a failed call or a call that the cap refuses gives the reason of the rules too, and the targets stay the same.

Acceptance story: the emulator tests send the logged sets of each scenario of section 5 of the high-level roadmap through `SyncOutbox`, and finish the workout. The plan then holds the targets of the rules, each with a reason that names a logged set. A reason of the fake provider with no logged set gives the reason of the rules. A browser test shows the next targets and their reasons after the end of a workout.

Checks: `make contract`, `make go-test`, `make emulator-test`, `make web`, and `make verify`, free. Codex reviews PR-35.

The session asks the owner Q-302 to Q-304 before the work.

### PR-36 - The overrides and the disruptions

Branch: `feat/pr-36-overrides-disruptions`. Work area 7.2. It needs PR-35 on `main`.

Before the work, the session reads the deploys of the merge of PR-35. It states the expected cost of one live revision, and asks the owner (D-212). After the approval, the owner finishes a workout on the iPhone. The owner then reads the next targets and their reasons. The session reads the reviser call and the revised plan with ids and counts alone.

Concerns:

- the override of a target, with the recommendation, the override, and the reason as separate records (D-69, Q-305),
- the rule of a missed session (D-66, Q-306),
- the reactive deload of REC-7, with a new policy version and the numbers of Q-307 (D-289),
- the scenario tests of a long break of D-151 and D-179, with the history of the workout store.

Acceptance story: the scenario tests prove the targets after a missed week and after a break of 14 days or more. They also prove the deload after a decline. An override shows in the next workout, and the recommendation and the reason stay as separate records. `make verify` passes.

Checks: `make contract`, `make go-test`, `make emulator-test`, `make web`, and `make verify`, free. Codex reviews PR-36.

The session asks the owner Q-305 to Q-307 before the work.

### PR-37 - The reviser evaluation

Branch: `test/pr-37-reviser-evaluation`. Work area 7.3. It needs PR-36 on `main`.

Before the work, the session reads the deploys of the merge of PR-36. It states the expected cost of one live revision with the new policy version, and asks the owner (D-212).

Concerns:

- the evaluation of the reviser in `go/cmd/lunaeval`, with the prompt of D-288, on the scenarios A to F and on the scenarios of PR-36,
- the check of each reason of Luna against the logged sets (D-68, D-288),
- the paid run, with the size and the cap of Q-308 (D-25, D-286),
- a report in `docs/research/`, with the measured numbers of each scenario.

Acceptance story: the report gives, for each scenario, the count of reasons that the check accepted and refused, the cost, and the longest call. The properties of section 6.3 of the high-level roadmap hold for each output. `make verify` passes.

Checks: `make go-test` and `make verify`, free. The evaluation costs money, and runs only after the owner approves the cap. Codex reviews PR-37.

The session asks the owner Q-308 before the paid run.

## 5. Exit of the phase

Phase 7 ends when PR-37 merges and the device checks pass. These items give the exit evidence of `docs/roadmaps/high-level-roadmap.md`:

- The emulator tests of PR-35 pass the scenarios A to F end to end with the fake provider (work area 7.1).
- After the deploy of PR-35, the owner sees the next targets with their reasons on the iPhone. The session of PR-36 records the result.
- The scenario tests of PR-36 prove a missed week, a break of the Q-102 length, and a deload (work area 7.2).
- The report of PR-37 gives the measured numbers of the paid evaluation (work area 7.3).
