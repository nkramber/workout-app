# Pull request 33 review response

Date: 2026-10-03

Round 1 of the Codex review gave "Changes required" at `e2721cd532d7747708118802eff71aff51133e30`, with finding P2-1.

## P2-1: A failed sync does not stop a plan request

Result: full merit.

The trigger reproduces in the code. `SyncEngine.syncNow` keeps the failure in its status and does not throw. So `requestPlan` and `excludeExercise` in `web/src/lib/plan-api.ts` started the plan call with an inventory entry in the outbox. A plan reads the confirmed machines of the server (D-193), so the plan read the old inventory.

The planner reads no workout history today, so a workout entry that waits does not change a plan. An inventory entry that the server refused left the outbox, and the plan then reads the inventory of the server, which wins (D-258).

Correction:

- `syncBeforePlan` in `web/src/lib/sync.ts` runs the sync, then throws `InventoryNotSyncedError` when an inventory entry still waits in the outbox (D-272).
- `web/src/lib/plan-api.ts` runs `syncBeforePlan` before `RequestPlan` and before `ExcludeExercise`.
- `planErrorText` in `web/src/lib/plan.ts` gives "Your equipment changes did not reach the server. Your plan did not change. Try again when the line above says Synced." An exclusion adds "The exercise is not excluded." first.

Regression checks:

- `web/e2e/plan.spec.ts`, "a plan request waits for the inventory changes that did not sync": a removal of a machine waits after a failed `SyncOutbox` call. "Make a plan" shows the text, and the page sends no `RequestPlan` call. After the sync passes, the request runs, and the plan has no exercise of the removed machine. It passed 2 of 2 runs in WebKit and in Chromium.
- `web/src/lib/sync.test.ts`, "syncBeforePlan": an inventory entry after a failed sync throws `InventoryNotSyncedError`. A refused inventory entry and a workout entry that waits let the request start.
- `web/src/lib/plan.test.ts`: the text of the error, and the text of an exclusion.

## The milestone after round 1

After round 1, the owner widened the milestone (D-281). PR-32 now also holds the rest of 60 seconds with policy version 5 (D-279), and the "Screen lock test" screen (D-280, D-282). The repeat review reads these changes as new code.

## P2-2: A video probe can not stop while it prepares its source

Round 2 gave "Changes required" at `9c645d6a61c93d16324840e8aff57e9dd6d02305`, with finding P2-2. Round 2 found P2-1 fixed.

Result: full merit.

The trigger reproduces. A Stop during the 1 s that `silentSource` records left `stopRef` empty. After the recording, `start` started the probe of the video, and the log got a line after the Stop.

Correction: `web/src/pages/lock-test.tsx` gives each start a run number, and Stop and the exit of the screen add 1 to it. When the source is ready, a start with an old number stops its source and starts no probe. A failure of such a start shows no error (D-282).

Regression check: `web/e2e/lock-test.spec.ts`, "a stop while the video file prepares starts no probe", taps Stop at once, and reads the log and the video after 2 s. On the old code in Chromium, it failed with 2 lines of the log for 1. With the correction, it passed 2 of 2 runs in WebKit and in Chromium. In CI, WebKit on Linux refuses the start of the video file at once. No preparation runs there, so the test skips with that reason.
