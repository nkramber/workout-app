# Pull request 31 - author response

Date: 2026-10-03. Review round 1 recorded effective head `1ef826c`, with the verdict "Changes required".

## P2-1: Distance input exceeds the `int32` contract

**Result: full merit.** The trigger reproduces. At `1ef826c`, `parseMiles("214748364.8")` gives 2147483648. The field `distance_tenths_mi` of `CardioEntry` in `proto/workoutapp/v1/workout_service.proto` is an `int32`, so the outbox entry can not sync. The minutes button had the same gap: the field `duration_seconds` is an `int32` too, and the button had no upper limit.

**Correction.** In `web/src/lib/workout.ts`, `parseMiles` counts the tenths as a `bigint` and refuses a value above 2147483647. The new `stepMinutes` keeps the minutes from 1 to `MAX_CARDIO_MINUTES`, so the duration in seconds fits its field. `parseLevel` uses the same bound `MAX_INT32`. The cardio form marks a refused distance as invalid, and turns off "Log cardio" (D-123, D-165).

**Regression checks.**

- The unit test "refuses a distance that does not fit the int32 field" of `web/src/lib/workout.test.ts`: `214748364.7` gives 2147483647, and `214748364.8` gives no value. It fails at `1ef826c`.
- The unit test of `stepMinutes` checks each limit, and the test of `parseLevel` checks 2147483647 and 2147483648.
- `make web` passes.

## P2-2: A released wake lock leaves the screen state stale

**Result: full merit.** At `1ef826c`, `holdWakeLock` of `web/src/lib/wake-lock.ts` did not listen for the `release` event of the sentinel. A release while the app shows, as in a power-save mode, left the state "on", and the app showed no notice (D-265).

**Correction.** The code listens for the `release` event of each sentinel. A release while the app goes to the back needs no step, because the return to the front requests the lock again. A release while the app shows requests the lock one more time. A second release in the same visit gives "off", and then the workout screen shows its notice. A limit of one request stops a loop of requests on a phone that releases each lock at once. A return to the front resets the limit.

**Regression checks.**

- The unit test "requests the lock again after a release while the app shows, and gives off after a second release" of `web/src/lib/wake-lock.test.ts`. At `1ef826c` it fails, because the code makes no second request.
- `make web` passes.
