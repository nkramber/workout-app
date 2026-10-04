import "fake-indexeddb/auto";

import { clone, create, fromJson } from "@bufbuild/protobuf";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { PlanSchema, type Plan } from "../gen/workoutapp/v1/plan_service_pb";
import { CardioEntrySchema, SetEntrySchema, WorkoutHeaderSchema } from "../gen/workoutapp/v1/workout_service_pb";
import { OUTBOX_SCHEMA_VERSION, pendingOutbox, REST_KEY, WorkoutAppDB, type SetRecord } from "./db";
import {
  activeWorkout,
  adjustRest,
  calibrationLoad,
  clockText,
  currentExercise,
  dismissRest,
  doneSessions,
  editSet,
  finishWorkout,
  openExercises,
  restLeft,
  restTimer,
  rirChoices,
  rirText,
  skipExercise,
  startRest,
  logCardio,
  loggedText,
  logSet,
  nextSession,
  nextSet,
  noteLength,
  parseLevel,
  parseMiles,
  startWorkout,
  stepMinutes,
  stepReps,
  MAX_CARDIO_MINUTES,
  stepWeight,
  WorkoutClosedError,
  WorkoutInProgressError,
  workoutCardio,
  workoutExercises,
  workoutSets,
} from "./workout";

let store: WorkoutAppDB;
let n = 0;

beforeEach(async () => {
  store = new WorkoutAppDB(`workout-test-${n++}`);
  await store.open();
});

afterEach(async () => {
  vi.restoreAllMocks();
  await store.delete();
});

// A synthetic plan of two sessions. Session 1 has one exercise with a
// calibration set and two working sets, and 20 minutes of cardio.
const plan: Plan = create(PlanSchema, {
  createdAt: "2026-10-03T08:00:00Z",
  sessions: [
    {
      title: "Session 1",
      exercises: [
        {
          exerciseId: "leg_press",
          name: "Leg press",
          restSeconds: 180,
          calibrationSets: [{ reps: 8, loadTenthLb: 200 }],
          calibrationLoads: [
            { weightTenthLb: 100, downTenthLb: 100, keepTenthLb: 100, upOneTenthLb: 100, upTwoTenthLb: 200 },
            { weightTenthLb: 200, downTenthLb: 100, keepTenthLb: 200, upOneTenthLb: 200, upTwoTenthLb: 300 },
            { weightTenthLb: 300, downTenthLb: 200, keepTenthLb: 300, upOneTenthLb: 300, upTwoTenthLb: 300 },
          ],
          workingSets: [
            { reps: 8, loadTenthLb: 200, rirTarget: 3 },
            { reps: 8, loadTenthLb: 200, rirTarget: 3 },
          ],
        },
        { exerciseId: "chest_press", name: "Chest press", restSeconds: 120, workingSets: [{ reps: 10, loadTenthLb: 100, rirTarget: 3 }] },
      ],
      cardio: { exerciseId: "treadmill", name: "Treadmill", minutes: 20 },
    },
    { title: "Session 2", exercises: [{ exerciseId: "lat_pulldown", name: "Lat pulldown", restSeconds: 120 }] },
  ],
});

const weights = (id: string) => (id === "leg_press" ? [300, 100, 200, 200] : []);
const now = new Date("2026-10-03T15:00:00.000Z");

describe("startWorkout", () => {
  it("writes the workout and the outbox entry of its header in one step", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);

    expect(await activeWorkout(store)).toEqual(w);
    expect(w).toMatchObject({ planCreatedAt: plan.createdAt, sessionIndex: 0, title: "Session 1", finished: false, version: 0 });
    expect(w.cardio).toEqual({ exerciseId: "treadmill", name: "Treadmill", minutes: 20 });
    expect(w.exercises[0].weights).toEqual([100, 200, 300]);

    const [entry] = await pendingOutbox(store);
    expect(entry).toMatchObject({ entity: "workout", entityId: w.id, baseVersion: 0, attempts: 0, schemaVersion: OUTBOX_SCHEMA_VERSION });
    const header = fromJson(WorkoutHeaderSchema, entry.payload as never);
    expect(header.plan).toMatchObject({ planCreatedAt: plan.createdAt, sessionIndex: 0 });
    expect(header.finished).toBe(false);
    expect(header.date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    // The header holds the target of each exercise that the owner saw
    // (D-291).
    expect(header.targets.map((t) => t.exerciseId)).toEqual(w.exercises.map((e) => e.exerciseId));
    expect(header.targets[0]).toMatchObject({
      restSeconds: w.exercises[0].restSeconds,
      calibrationSets: w.exercises[0].calibrationSets.map((c) => expect.objectContaining({ reps: c.reps, loadTenthLb: c.loadTenthLb })),
      workingSets: w.exercises[0].workingSets.map((c) => expect.objectContaining({ reps: c.reps, loadTenthLb: c.loadTenthLb, rirTarget: c.rirTarget })),
    });
  });

  it("refuses a second open workout, and writes nothing", async () => {
    await startWorkout(store, plan, 0, weights, now);
    await expect(startWorkout(store, plan, 1, weights, now)).rejects.toBeInstanceOf(WorkoutInProgressError);
    expect(await store.workouts.count()).toBe(1);
    expect(await store.outbox.count()).toBe(1);
  });

  it("refuses a session that the plan does not have", async () => {
    await expect(startWorkout(store, plan, 2, weights, now)).rejects.toBeInstanceOf(RangeError);
    expect(await store.outbox.count()).toBe(0);
  });
});

describe("workoutExercises", () => {
  it("uses the loads of the targets when the machine has no weights", () => {
    const [, chest] = workoutExercises(plan, 0, () => []);
    expect(chest.weights).toEqual([100]);
    expect(chest.override).toBeUndefined();
  });
});

// A plan whose chest press has an override of the owner (D-69, D-293).
const overridden: Plan = create(PlanSchema, {
  createdAt: "2026-10-03T08:00:00Z",
  sessions: [
    {
      title: "Session 1",
      exercises: [
        {
          exerciseId: "chest_press",
          name: "Chest press",
          restSeconds: 60,
          workingSets: [
            { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
            { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
          ],
          override: {
            workingSets: [
              { reps: 8, loadTenthLb: 1100, rirTarget: 2 },
              { reps: 8, loadTenthLb: 1100, rirTarget: 2 },
            ],
            reason: "The last session felt easy.",
            recommendedWorkingSets: [
              { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
              { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
            ],
          },
        },
      ],
    },
  ],
});

describe("an override in the next workout (D-293)", () => {
  it("gives the sets of the override, and keeps the recommendation and the reason apart", () => {
    const [chest] = workoutExercises(overridden, 0, () => []);
    expect(chest.workingSets).toEqual([
      { reps: 8, loadTenthLb: 1100, rirTarget: 2 },
      { reps: 8, loadTenthLb: 1100, rirTarget: 2 },
    ]);
    expect(chest.override).toEqual({
      reason: "The last session felt easy.",
      recommendedWorkingSets: [
        { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
        { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
      ],
    });
  });

  it("sends the override, the recommendation, and the reason in the header", async () => {
    await startWorkout(store, overridden, 0, () => [], now);
    const [entry] = await pendingOutbox(store);
    const [t] = fromJson(WorkoutHeaderSchema, entry.payload as never).targets;
    expect(t.workingSets.map((s) => [s.reps, s.loadTenthLb, s.rirTarget])).toEqual([
      [8, 1100, 2],
      [8, 1100, 2],
    ]);
    expect(t.recommendedWorkingSets.map((s) => [s.reps, s.loadTenthLb, s.rirTarget])).toEqual([
      [10, 1000, 2],
      [10, 1000, 2],
    ]);
    expect(t.overrideReason).toBe("The last session felt easy.");
  });

  it("uses the recommendation for an expired override (D-294, D-295)", () => {
    const expired = clone(PlanSchema, overridden);
    expired.sessions[0].exercises[0].override!.expired = true;
    const [chest] = workoutExercises(expired, 0, () => []);
    expect(chest.workingSets).toEqual([
      { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
      { reps: 10, loadTenthLb: 1000, rirTarget: 2 },
    ]);
    expect(chest.override).toBeUndefined();
  });

  it("sends no override fields for an exercise with no override", async () => {
    await startWorkout(store, plan, 0, weights, now);
    const [entry] = await pendingOutbox(store);
    for (const t of fromJson(WorkoutHeaderSchema, entry.payload as never).targets) {
      expect(t.recommendedWorkingSets).toEqual([]);
      expect(t.overrideReason).toBe("");
    }
  });
});

describe("logSet", () => {
  it("writes the set and its outbox entry with the payload of SetEntry", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const s = await logSet(store, w.id, { exerciseId: "leg_press", kind: "calibration", reps: 8, weightTenthsLb: 200, rir: 4, note: "  seat 4  " }, now);

    expect(await workoutSets(store, w.id)).toEqual([s]);
    expect(s.note).toBe("seat 4");
    expect(s.pain).toBeUndefined();
    const entries = await pendingOutbox(store);
    expect(entries).toHaveLength(2);
    const last = entries.find((e) => e.entity === "set");
    expect(last).toMatchObject({ entityId: s.id, baseVersion: 0 });
    const payload = fromJson(SetEntrySchema, last?.payload as never);
    expect(payload).toMatchObject({ workoutId: w.id, exerciseId: "leg_press", kind: "calibration", reps: 8, weightTenthsLb: 200n, rir: 4, note: "seat 4" });
    expect(payload.pain).toBeUndefined();
  });

  it("keeps a pain rating of 0 as a value", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const s = await logSet(store, w.id, { exerciseId: "leg_press", kind: "working", reps: 8, weightTenthsLb: 200, rir: 2, pain: 0 }, now);
    expect(s.pain).toBe(0);
    const entry = (await pendingOutbox(store)).find((e) => e.entity === "set");
    expect(fromJson(SetEntrySchema, entry?.payload as never).pain).toBe(0);
  });

  it("refuses a set of a finished workout, and writes nothing", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await finishWorkout(store, w.id, now);
    const before = await store.outbox.count();
    await expect(
      logSet(store, w.id, { exerciseId: "leg_press", kind: "working", reps: 8, weightTenthsLb: 200, rir: 2 }, now),
    ).rejects.toBeInstanceOf(WorkoutClosedError);
    expect(await store.sets.count()).toBe(0);
    expect(await store.outbox.count()).toBe(before);
  });

  it("rolls back the set when the outbox write fails", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    vi.spyOn(store.outbox, "add").mockRejectedValueOnce(new Error("disk full"));
    await expect(
      logSet(store, w.id, { exerciseId: "leg_press", kind: "working", reps: 8, weightTenthsLb: 200, rir: 2 }, now),
    ).rejects.toThrow("disk full");
    expect(await store.sets.count()).toBe(0);
  });
});

describe("logCardio", () => {
  it("writes the log and its outbox entry with the payload of CardioEntry", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const c = await logCardio(store, w.id, { exerciseId: "treadmill", durationSeconds: 1200, effort: 6, distanceTenthsMi: 15 }, now);

    expect(await workoutCardio(store, w.id)).toEqual([c]);
    const entry = (await pendingOutbox(store)).find((e) => e.entity === "cardio");
    const payload = fromJson(CardioEntrySchema, entry?.payload as never);
    expect(payload).toMatchObject({ workoutId: w.id, exerciseId: "treadmill", durationSeconds: 1200, effort: 6, distanceTenthsMi: 15 });
    expect(payload.resistance).toBeUndefined();
    expect(payload.pain).toBeUndefined();
  });
});

describe("finishWorkout", () => {
  it("skips each exercise with no set, and ends the workout early", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await logSet(store, w.id, { exerciseId: "leg_press", kind: "calibration", reps: 8, weightTenthsLb: 200, rir: 4 }, now);
    const done = await finishWorkout(store, w.id, now);

    expect(done).toMatchObject({ finished: true, endedEarly: true, skippedExerciseIds: ["chest_press"] });
    expect(await activeWorkout(store)).toBeUndefined();
    const headers = (await pendingOutbox(store)).filter((e) => e.entity === "workout");
    expect(headers).toHaveLength(2);
    const last = fromJson(WorkoutHeaderSchema, headers[1].payload as never);
    expect(last).toMatchObject({ finished: true, endedEarly: true, skippedExerciseIds: ["chest_press"] });
  });

  it("ends early when an exercise with a set has a planned set with no log", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    for (const exerciseId of ["leg_press", "chest_press"]) {
      await logSet(store, w.id, { exerciseId, kind: "working", reps: 8, weightTenthsLb: 100, rir: 2 }, now);
    }
    expect(await finishWorkout(store, w.id, now)).toMatchObject({ endedEarly: true, skippedExerciseIds: [] });
  });

  it("does not end early when each planned set has a log, and stops the rest timer", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await logSet(store, w.id, { exerciseId: "leg_press", kind: "calibration", reps: 8, weightTenthsLb: 200, rir: 4 }, now);
    for (const exerciseId of ["leg_press", "leg_press", "chest_press"]) {
      await logSet(store, w.id, { exerciseId, kind: "working", reps: 8, weightTenthsLb: 200, rir: 3 }, now);
    }
    await startRest(store, w.id, 120, now);
    expect(await finishWorkout(store, w.id, now)).toMatchObject({ endedEarly: false, skippedExerciseIds: [] });
    expect(await store.meta.get(REST_KEY)).toBeUndefined();
  });

  it("does not end early after a skip, and the skipped exercise stays skipped (D-63)", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await skipExercise(store, w.id, "leg_press", now);
    await logSet(store, w.id, { exerciseId: "chest_press", kind: "working", reps: 10, weightTenthsLb: 100, rir: 3 }, now);
    expect(await finishWorkout(store, w.id, now)).toMatchObject({ endedEarly: false, skippedExerciseIds: ["leg_press"] });
  });

  it("counts the session as done for the next pick", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await finishWorkout(store, w.id, now);
    expect(await doneSessions(store, plan.createdAt)).toEqual([0]);
    expect(await doneSessions(store, "2026-01-01T00:00:00Z")).toEqual([]);
  });
});

describe("skipExercise", () => {
  it("writes the header with the skip and its outbox entry, and the next exercise comes next", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const after = await skipExercise(store, w.id, "leg_press", now);
    expect(after.skippedExerciseIds).toEqual(["leg_press"]);
    expect(currentExercise(after, [])?.exerciseId).toBe("chest_press");
    expect(openExercises(after, []).map((e) => e.exerciseId)).toEqual(["chest_press"]);
    const headers = (await pendingOutbox(store)).filter((e) => e.entity === "workout");
    expect(headers).toHaveLength(2);
    expect(fromJson(WorkoutHeaderSchema, headers[1].payload as never)).toMatchObject({ skippedExerciseIds: ["leg_press"], finished: false });

    // A second skip changes nothing.
    await skipExercise(store, w.id, "leg_press", now);
    expect((await pendingOutbox(store)).filter((e) => e.entity === "workout")).toHaveLength(2);
  });

  it("refuses an exercise that the workout does not have, and a finished workout", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await expect(skipExercise(store, w.id, "seated_row", now)).rejects.toThrow(RangeError);
    await finishWorkout(store, w.id, now);
    await expect(skipExercise(store, w.id, "leg_press", now)).rejects.toThrow(WorkoutClosedError);
  });
});

describe("editSet", () => {
  it("changes a logged set, keeps its id, kind, and time, and writes its outbox entry (D-63)", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const s = await logSet(store, w.id, { exerciseId: "leg_press", kind: "working", reps: 8, weightTenthsLb: 200, rir: 2, pain: 1, note: "a" }, now);
    const later = new Date(now.getTime() + 60_000);
    const e = await editSet(store, s.id, { reps: 7, weightTenthsLb: 100, rir: 4, note: " b " }, later);
    expect(e).toMatchObject({ id: s.id, kind: "working", at: s.at, reps: 7, weightTenthsLb: 100, rir: 4, note: "b" });
    expect(e.pain).toBeUndefined();
    expect(await store.sets.get(s.id)).toEqual(e);

    const entries = (await pendingOutbox(store)).filter((x) => x.entity === "set");
    expect(entries.map((x) => x.entityId)).toEqual([s.id, s.id]);
    expect(entries[1].at).toBe(later.toISOString());
    const payload = fromJson(SetEntrySchema, entries[1].payload as never);
    expect(payload).toMatchObject({ reps: 7, weightTenthsLb: 100n, rir: 4, note: "b" });
    expect(payload.pain).toBeUndefined();
  });

  it("refuses a set of a finished workout, and a set that does not exist", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const s = await logSet(store, w.id, { exerciseId: "leg_press", kind: "working", reps: 8, weightTenthsLb: 200, rir: 2 }, now);
    await finishWorkout(store, w.id, now);
    await expect(editSet(store, s.id, { reps: 1, weightTenthsLb: 200, rir: 1 }, now)).rejects.toThrow(WorkoutClosedError);
    await expect(editSet(store, "nope", { reps: 1, weightTenthsLb: 200, rir: 1 }, now)).rejects.toThrow(WorkoutClosedError);
    expect((await store.sets.get(s.id))?.reps).toBe(8);
  });
});

describe("the rest timer (D-59, D-270)", () => {
  it("starts from the rest of the target, moves by 15 s, never goes below 0, and stops", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await startRest(store, w.id, 120, now);
    const t = await restTimer(store, w.id);
    expect(t).toEqual({ workoutId: w.id, endsAt: now.getTime() + 120_000 });
    expect(restLeft(t!, now.getTime())).toBe(120);
    expect(restLeft(t!, now.getTime() + 500)).toBe(120);
    expect(restLeft(t!, now.getTime() + 119_001)).toBe(1);
    expect(restLeft(t!, now.getTime() + 600_000)).toBe(0);

    // A longer rest than the bound of D-172 is the choice of the owner.
    await adjustRest(store, w.id, 15, now);
    await adjustRest(store, w.id, 60, now);
    expect(restLeft((await restTimer(store, w.id))!, now.getTime())).toBe(195);

    // After the end, -15 s keeps 0, and +15 s counts from now.
    const late = new Date(now.getTime() + 300_000);
    await adjustRest(store, w.id, -15, late);
    expect(restLeft((await restTimer(store, w.id))!, late.getTime())).toBe(0);
    await adjustRest(store, w.id, 15, late);
    expect(restLeft((await restTimer(store, w.id))!, late.getTime())).toBe(15);
    await adjustRest(store, w.id, -60, late);
    expect(restLeft((await restTimer(store, w.id))!, late.getTime())).toBe(0);

    // The timer of another workout is not this one.
    expect(await restTimer(store, "other")).toBeNull();
    await adjustRest(store, "other", 15, late);
    expect(restLeft((await restTimer(store, w.id))!, late.getTime())).toBe(0);

    await dismissRest(store);
    expect(await restTimer(store, w.id)).toBeNull();
  });

  it("gives the time as m:ss", () => {
    expect(clockText(0)).toBe("0:00");
    expect(clockText(9)).toBe("0:09");
    expect(clockText(120)).toBe("2:00");
    expect(clockText(615)).toBe("10:15");
  });
});

describe("nextSession", () => {
  it("gives the first session with the fewest finished workouts", () => {
    expect(nextSession(3, [])).toBe(0);
    expect(nextSession(3, [0])).toBe(1);
    expect(nextSession(3, [0, 2])).toBe(1);
    expect(nextSession(3, [0, 1, 2])).toBe(0);
    expect(nextSession(3, [0, 1, 2, 0])).toBe(1);
  });
});

describe("nextSet and currentExercise", () => {
  const set = (exerciseId: string, kind: "working" | "calibration"): SetRecord => ({
    id: `${exerciseId}-${kind}-${Math.random()}`,
    workoutId: "w",
    exerciseId,
    kind,
    reps: 8,
    weightTenthsLb: 200,
    rir: 3,
    note: "",
    at: now.toISOString(),
    version: 0,
  });

  it("gives the calibration sets first, then the working sets", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const press = w.exercises[0];
    expect(nextSet(press, [])).toMatchObject({ kind: "calibration", number: 1, of: 1 });
    expect(nextSet(press, [set("leg_press", "calibration")])).toMatchObject({ kind: "working", number: 1, of: 2 });
    const all = [set("leg_press", "calibration"), set("leg_press", "working"), set("leg_press", "working")];
    expect(nextSet(press, all)).toBeNull();
    expect(currentExercise(w, all)?.exerciseId).toBe("chest_press");
    expect(currentExercise(w, [...all, set("chest_press", "working")])).toBeNull();
  });

  // The calibration table of D-150 and D-267, from the loads of the plan.
  it("gives each working set the load of the calibration table", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    const press = w.exercises[0];
    expect(press.calibrationLoads).toHaveLength(3);
    expect(press.calibrationLoads?.[1]).toEqual({ weight: 200, down: 100, keep: 200, upOne: 200, upTwo: 300 });
    expect(w.exercises[1].calibrationLoads).toBeUndefined();
    const cal = (rir: number, extra: Partial<SetRecord> = {}): SetRecord => ({ ...set("leg_press", "calibration"), rir, ...extra });
    const load = (s: SetRecord) => nextSet(press, [s]);

    expect(load(cal(6))).toMatchObject({ kind: "working", target: { loadTenthLb: 300, reps: 8, rirTarget: 3 }, fromCalibration: true });
    expect(load(cal(9))?.target.loadTenthLb).toBe(300);
    expect(load(cal(5))?.target.loadTenthLb).toBe(200);
    expect(load(cal(4))?.target.loadTenthLb).toBe(200);
    expect(load(cal(3))?.target.loadTenthLb).toBe(200);
    expect(load(cal(2))?.target.loadTenthLb).toBe(100);
    expect(load(cal(4, { pain: 1 }))?.target.loadTenthLb).toBe(100);
    expect(load(cal(4, { pain: 0 }))?.target.loadTenthLb).toBe(200);
    // The second working set gets the same load.
    expect(nextSet(press, [cal(6), set("leg_press", "working")])).toMatchObject({ number: 2, target: { loadTenthLb: 300 } });

    // The table applies to the weight that the owner logged (D-249): a
    // set at 300 goes down to 200 after a hard set, and a set at 100 goes
    // up to 200 after an easy set.
    expect(load(cal(1, { weightTenthsLb: 300 }))).toMatchObject({ target: { loadTenthLb: 200 }, fromCalibration: true });
    expect(load(cal(4, { weightTenthsLb: 300 }))).toMatchObject({ target: { loadTenthLb: 300 }, fromCalibration: true });
    expect(load(cal(6, { weightTenthsLb: 100 }))).toMatchObject({ target: { loadTenthLb: 200 }, fromCalibration: true });
    expect(load(cal(2, { weightTenthsLb: 100 }))).toMatchObject({ target: { loadTenthLb: 100 }, fromCalibration: true });
    // A weight with no row of the table gives the load of the plan.
    expect(load(cal(6, { weightTenthsLb: 250 }))).toMatchObject({ target: { loadTenthLb: 200 }, fromCalibration: false });
    expect(calibrationLoad(press, { weightTenthsLb: 250, rir: 6 })).toBeNull();
    // A plan of policy version 3 has no loads, so the plan load stays.
    expect(nextSet({ ...press, calibrationLoads: undefined }, [cal(6)])).toMatchObject({ target: { loadTenthLb: 200 }, fromCalibration: false });
  });

  it("does not give a skipped exercise as the current exercise", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    expect(currentExercise({ ...w, skippedExerciseIds: ["leg_press"] }, [])?.exerciseId).toBe("chest_press");
    expect(currentExercise({ ...w, skippedExerciseIds: ["leg_press", "chest_press"] }, [])).toBeNull();
  });
});

describe("the steps of the buttons", () => {
  it("moves the weight to the next weight of the list (D-264)", () => {
    const list = [100, 200, 300];
    expect(stepWeight(list, 200, 1)).toBe(300);
    expect(stepWeight(list, 200, -1)).toBe(100);
    expect(stepWeight(list, 300, 1)).toBe(300);
    expect(stepWeight(list, 100, -1)).toBe(100);
    expect(stepWeight(list, 250, 1)).toBe(300);
    expect(stepWeight(list, 250, -1)).toBe(200);
    expect(stepWeight([], 250, 1)).toBe(250);
  });

  it("keeps the reps at 0 or more (D-164)", () => {
    expect(stepReps(1, -1)).toBe(0);
    expect(stepReps(0, -1)).toBe(0);
    expect(stepReps(8, 1)).toBe(9);
  });
});

describe("the fields of a log", () => {
  it("counts the note after the trim (D-261)", () => {
    expect(noteLength("  ab  ")).toBe(2);
    expect(noteLength("é".repeat(3))).toBe(3);
  });

  it("reads the distance in tenths of a mile (D-165)", () => {
    expect(parseMiles("")).toBeUndefined();
    expect(parseMiles("2.5")).toBe(25);
    expect(parseMiles(" 3 ")).toBe(30);
    expect(parseMiles("0")).toBe(0);
    expect(parseMiles("2.55")).toBeNull();
    expect(parseMiles("-1")).toBeNull();
    expect(parseMiles("abc")).toBeNull();
  });

  it("refuses a distance that does not fit the int32 field", () => {
    expect(parseMiles("214748364.7")).toBe(2_147_483_647);
    expect(parseMiles("214748364.8")).toBeNull();
    expect(parseMiles("99999999999999999999")).toBeNull();
  });

  it("keeps the minutes of a cardio log from 1 to the int32 bound of its seconds", () => {
    expect(stepMinutes(1, -1)).toBe(1);
    expect(stepMinutes(20, 1)).toBe(21);
    expect(stepMinutes(MAX_CARDIO_MINUTES, 1)).toBe(MAX_CARDIO_MINUTES);
    expect(MAX_CARDIO_MINUTES * 60).toBeLessThanOrEqual(2_147_483_647);
  });

  it("reads the resistance level as a whole number", () => {
    expect(parseLevel("")).toBeUndefined();
    expect(parseLevel("8")).toBe(8);
    expect(parseLevel("8.5")).toBeNull();
    expect(parseLevel("2147483647")).toBe(2_147_483_647);
    expect(parseLevel("2147483648")).toBeNull();
    expect(parseLevel("99999999999")).toBeNull();
  });

  it("gives the text of a logged set", () => {
    expect(loggedText({ kind: "working", reps: 8, weightTenthsLb: 125, rir: 2 })).toBe("8 reps at 12.5 lb, 2 in reserve");
    expect(loggedText({ kind: "working", reps: 8, weightTenthsLb: 125, rir: 5 })).toBe("8 reps at 12.5 lb, 4+ in reserve");
    expect(loggedText({ kind: "calibration", reps: 1, weightTenthsLb: 200, rir: 5, pain: 3 })).toBe(
      "Calibration: 1 rep at 20 lb, 5 in reserve, pain 3",
    );
    expect(loggedText({ kind: "calibration", reps: 8, weightTenthsLb: 200, rir: 7 })).toBe("Calibration: 8 reps at 20 lb, 6+ in reserve");
  });

  it("offers 0 to 4+ for a working set and 0 to 6+ for a calibration set (D-249, D-268)", () => {
    expect(rirChoices("working").map((c) => c.label)).toEqual(["0", "1", "2", "3", "4+"]);
    expect(rirChoices("calibration").map((c) => c.label)).toEqual(["0", "1", "2", "3", "4", "5", "6+"]);
    expect(rirChoices("calibration").at(-1)?.value).toBe(6);
    expect(rirText("working", 4)).toBe("4+");
    expect(rirText("calibration", 4)).toBe("4");
  });
});
