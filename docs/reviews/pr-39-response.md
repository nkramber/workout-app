# Pull request 39 response

Date: 2026-10-05

This file answers the findings of round 1 of `docs/reviews/pr-39.md`, at `dcd23494a8212291ddc4788ceca8d85cb0f1de29`.

## P2-1: Stage 8 starts before the Phase 7 exit check

Result: partial merit.

Evidence: the trigger reproduced. `AGENTS.md:9` named Phase 8 as the stage, and the three iPhone checks of Phase 7 did not run. The roadmap put the checks before the work of PR-39, but no line said that the work of work area 8.1 waits for them. No line said what a failed check does.

The order itself stays. The owner put the fix and this roadmap before the check, and the check at the start of PR-39 (D-310). The owner approved that milestone with those terms before the first edit (D-12). So the finding has no merit where it asks to move the check out of PR-39.

Correction:

- `AGENTS.md` names the exit of Phase 7 as the stage, and says that no Phase 8 work starts before the iPhone check passes (D-310).
- `docs/roadmaps/phase-8-personal-use-operations.md` adds the rule to section 3, and a gate to PR-39. When a check fails, the session stops and asks the owner for the fix. The work of PR-39 starts only after a fix passes the three checks.
- `docs/roadmaps/high-level-roadmap.md` adds the same gate to the exit note of Phase 7.

Regression check: `make ste-check`, `make ref-check`, and `make verify` passed. The stage line, section 1.1, section 3, and the PR-39 gate of the roadmap now agree. Each one keeps the work of work area 8.1 after the three checks.

## Round 3

This part answers the findings of round 3, at `6beaab0f643529e75d2de435d97f53c91f824c36`.

## P2-2: Another tab can restore history after deletion

Result: full merit.

Evidence: the trigger holds. The tabs of one origin share the local database, so the deletion removes their rows. But a `SyncOutbox` call that a second tab sent before the deletion still writes on the server. A second device keeps its own outbox, and sends its entries later. `DeleteAll` had no fence, so each such entry wrote a workout again.

Correction:

- `DeleteAll` of `go/internal/workout/store.go` first writes the time of the deletion at `users/{uid}/history/deleted`, on the clock of the server. Then it deletes the workouts and the op ids.
- The transaction of `Apply` reads that document. An entry that the phone made at that time or before gets `ErrBeforeDeletion`. A deletion that writes the time before the commit makes Firestore run the transaction again, so a call in flight sees the fence.
- The memory store does the same, and `SyncOutbox` refuses such an entry with `failed_precondition`. The phone moves it to the refused entries, and never sends it again (D-274).

Accepted risk: the fence compares the time of the phone with the time of the server. A phone clock that is behind the server can refuse an entry that the owner made in the seconds after a deletion. The refused entry then shows on the phone.

Regression check: `TestMemoryDeleteAll`, `TestFirestoreDeleteAll` on the emulator, and `TestSyncAfterDeletion` send an entry of a time before the deletion after `DeleteAll`. Each one gets the refusal, and no workout comes back. An entry after the deletion applies. Without the fence, the three tests fail.

## Round 4

This part answers the findings of round 4, at `e0a01c9b560980b6c6a6cf8dea18432302406fb7`.

## P2-3: A phone clock ahead can bypass the deletion fence

Result: full merit.

Evidence: the fence of round 3 compared the time of the server with `Entry.At`, a time of the phone. A phone clock ahead of the server gives an old entry a later time, and that entry passed.

Correction: the fence uses no clock of a phone.

- The document `users/{uid}/history/deleted` holds a generation of the history. `DeleteAll` adds 1 to it in a transaction, first.
- `DeleteAll` then deletes each workout of an older generation, and each op id that the server applied before the change. A time of the server alone decides which op ids go.
- Each workout header carries the generation that the phone knew at the start of the workout. The phone reads it with `GetMe` at each read of the copies, and from the answer of `DeleteHistory`.
- The transaction of `Apply` refuses a header of an older generation, and each entry of a stored workout of an older generation, with `ErrBeforeDeletion`.

Accepted risk: a device that starts a workout after a deletion, before it reads the new generation, sends a workout of the old generation. The server refuses it, and the phone shows the refused entries. A device reads the generation at each sync of its copies, so this needs a workout with no connection.

Regression check: `TestMemoryDeleteAll` and `TestFirestoreDeleteAll` give a header of the old generation a phone time far after the deletion, and each one gets the refusal. A header of the new generation applies. `TestFenced` proves each case of the fence. `TestSyncAfterDeletion` proves the refusal and the apply through `SyncOutbox`. The browser test of the deletion starts a new workout after the deletion, and the server applies it.

## Round 5

This part answers the findings of round 5, at `4cd1bd51bb36d59b07a2f03e31782abcc5e379dc`.

## P2-4: A plan request can save after history deletion

Result: full merit.

Evidence: the trigger holds. A plan request waits up to about 47 s in its call of Luna, and the owner can leave the plan screen during it. `Save` read the exclusions alone, so a deletion during the request did not stop its save, and a plan came back.

Correction:

- The new package `go/internal/history` holds the fence document of the generation. The workout store and the plan store read it.
- `Maker.Make` reads the generation at the start of the request, and the plan keeps it.
- The transaction of `Save` reads the fence, and refuses a plan of an older generation with `ErrHistoryDeleted`. A deletion that changes the fence before the commit makes Firestore run the transaction again.
- `plansvc` gives `ABORTED` for it, and the phone tells the owner that the exclusions or the history changed during the request.

Regression check: `TestDeleteHistoryDuringPlanRequest` in `go/cmd/api` holds a plan request in its call of Luna, runs `DeleteHistory`, then releases the call. The request gets `ABORTED`, `GetPlan` gives no plan, and a later request saves its plan. `TestFirestoreSaveFence` proves the refusal of the store on the emulator.

## Round 6

Codex round 6 gave `Ready for owner merge` for effective head `0e094c5ab28b614f20f24234a9ea860b0f8f6190`.

## P3-1: Remove trailing blank lines from two changed files

Result: partial merit, an optional improvement with no broken contract.

Correction: this file ends with no blank line now. `web/src/lib/sync.test.ts` keeps its blank line, because a change of code after the approval needs a new review (D-90). No check reads it.
