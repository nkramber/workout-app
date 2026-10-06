# Workout App - policy replay report

Status: the replay diff report of work area 8.3 of `docs/roadmaps/high-level-roadmap.md`. It holds counts and rule ids alone (D-80). The owner decisions live in `docs/decisions.md`, and `docs/operations.md` section 3 gives the steps of a replay.

Date of the live runs: 2026-10-06, 02:22:06Z and 02:26:42Z. The project is `nk-workout-app-prod`, and the code is the branch `feat/pr-41-version-migration` from base `c712f56`.

## 1. Result

**Pass.** The emulator test replays the records of policy version 7 under policy version 8. The diff report gives the count of each changed target with its rule id. The live store holds no record of an older version, and each live record replays to the same target.

| Measure | Emulator test | Live run 2 |
|---|---|---|
| Users | 1 | 1 |
| Records | 4, policy version 7 | 9, policy version 8 |
| Records of a new plan, and of a revision | 1 and 3 | 4 and 5 |
| Records with the stored input hash | 4 of 4 | 9 of 9 |
| Changed targets | 3 | 0 |
| Changed targets by rule id | `follow.first-set`: 3 | none |
| Changed fields | `follow_max`: 3 | none |
| Target copies | 2 | 15 |
| Copies equal to the target of the rules | 2 | 10 |
| Copies with a stale history | 0 | 1 |
| Copies outside the bounds | 0 | 0 |

The replay called no model, and it wrote nothing. Each live run read the list of the users, and the workouts, the inventory, and the plan of one user. That count of reads is far inside the free daily quota of Firestore.

## 2. The emulator test

The test `go/cmd/replay/replay_emulator_test.go` uses its own database of the emulator. It stores a plan of the rules for the chest press, the seated row, and the leg press. One workout logs the chest press and the seated row, and the revision of policy version 8 gives three records. The test then removes the limit of D-306 from each record and each target, and gives each record the version 7. So the plan holds the records that policy version 7 gave.

The replay under policy version 8 gives the limit to the three records of the revision. The report counts them under `follow.first-set`, the rule of D-306. The record of the leg press holds the first-set calibration, which has no limit (D-299), so it stays the same.

## 3. The live runs

The owner approved the live runs (D-324, D-326). The live store holds the data of the checks of PR-39 on 2026-10-05, after the deletion of the history of that day.

Run 1 gave one copy outside the bounds, with the rule `load.ceiling`. With the approval of the owner, the session read that copy and its history on its own terminal alone (D-326). The cause:

- Two workouts of the leg press started from the same plan target, a first-set calibration.
- The second workout started 4 s before the server saved the revision of the first workout. The logs `plan revised` and the start times of the workout ids give both times.
- So the phone showed the target before that revision, and the copy equals the plan target of that time.
- The replay of run 1 read the first workout as part of the history of the second copy. The rules of that history give a lower ceiling.

So the copy was not a breach of D-23. The store keeps no time of a revision, so the replay now checks a copy with each history of its plan up to its start. A copy that passes only with a shorter history counts as a stale history. The test `TestReplayStaleHistory` holds this case, and a copy above each ceiling still counts as outside the bounds.

Run 2 used the fixed code. The 5 copies that differ from the target of the rules are inside the bounds. A Luna proposal, an estimate, or a plan of an earlier date can give such a target.
