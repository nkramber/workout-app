import "fake-indexeddb/auto";

import { create, fromJson } from "@bufbuild/protobuf";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { PlanSchema, type Plan } from "../gen/workoutapp/v1/plan_service_pb";
import { CardioEntrySchema, SetEntrySchema, WorkoutHeaderSchema } from "../gen/workoutapp/v1/workout_service_pb";
import { OUTBOX_SCHEMA_VERSION, pendingOutbox, WorkoutAppDB, type SetRecord } from "./db";
import {
  activeWorkout,
  currentExercise,
  doneSessions,
  finishWorkout,
  logCardio,
  loggedText,
  logSet,
  nextSession,
  nextSet,
  noteLength,
  parseLevel,
  parseMiles,
  startWorkout,
  stepReps,
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

  it("does not end early when each exercise has a set", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    for (const exerciseId of ["leg_press", "chest_press"]) {
      await logSet(store, w.id, { exerciseId, kind: "working", reps: 8, weightTenthsLb: 100, rir: 2 }, now);
    }
    expect(await finishWorkout(store, w.id, now)).toMatchObject({ endedEarly: false, skippedExerciseIds: [] });
  });

  it("counts the session as done for the next pick", async () => {
    const w = await startWorkout(store, plan, 0, weights, now);
    await finishWorkout(store, w.id, now);
    expect(await doneSessions(store, plan.createdAt)).toEqual([0]);
    expect(await doneSessions(store, "2026-01-01T00:00:00Z")).toEqual([]);
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

  it("reads the resistance level as a whole number", () => {
    expect(parseLevel("")).toBeUndefined();
    expect(parseLevel("8")).toBe(8);
    expect(parseLevel("8.5")).toBeNull();
    expect(parseLevel("99999999999")).toBeNull();
  });

  it("gives the text of a logged set", () => {
    expect(loggedText({ kind: "working", reps: 8, weightTenthsLb: 125, rir: 2 })).toBe("8 reps at 12.5 lb, 2 in reserve");
    expect(loggedText({ kind: "calibration", reps: 1, weightTenthsLb: 200, rir: 5, pain: 3 })).toBe(
      "Calibration: 1 rep at 20 lb, 4+ in reserve, pain 3",
    );
  });
});
