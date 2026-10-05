import "fake-indexeddb/auto";

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import { OUTBOX_SCHEMA_VERSION, REST_KEY, WorkoutAppDB, type OutboxEntry } from "./db";
import { canDelete, DELETE_CONFIRMATION, deleteAllData, deleteLocalHistory } from "./history";

let store: WorkoutAppDB;
let n = 0;

beforeEach(() => {
  store = new WorkoutAppDB(`history-test-${n++}`);
});

afterEach(async () => {
  store.close();
  await store.delete();
});

const entry = (opId: string, entity: string): OutboxEntry => ({
  opId,
  entity,
  entityId: `${entity}-1`,
  baseVersion: 0,
  payload: {},
  at: "2026-10-05T08:00:00Z",
  attempts: 0,
  schemaVersion: OUTBOX_SCHEMA_VERSION,
});

// seed gives the phone a workout with a set and a cardio log, their
// outbox entries, a refused set, an inventory entry, the copies, and the
// rest timer.
async function seed() {
  await store.workouts.put({
    id: "w1",
    date: "2026-10-05",
    planCreatedAt: "2026-10-04T08:00:00Z",
    sessionIndex: 0,
    title: "Session 1",
    exercises: [],
    cardio: null,
    skippedExerciseIds: [],
    endedEarly: false,
    finished: false,
    startedAt: "2026-10-05T08:00:00Z",
    version: 0,
  });
  await store.sets.put({ id: "s1", workoutId: "w1", exerciseId: "chest_press", kind: "working", reps: 10, weightTenthsLb: 500, rir: 3, note: "", at: "2026-10-05T08:01:00Z", version: 0 });
  await store.cardio.put({ id: "c1", workoutId: "w1", exerciseId: "treadmill", durationSeconds: 600, effort: 5, note: "", at: "2026-10-05T08:30:00Z", version: 0 });
  await store.outbox.bulkPut([entry("op1", "workout"), entry("op2", "set"), entry("op3", "cardio"), entry("op4", "machine")]);
  await store.refused.bulkPut([
    { ...entry("op5", "set"), code: "invalid_argument", message: "m", refusedAt: "2026-10-05T08:02:00Z" },
    { ...entry("op6", "note"), code: "invalid_argument", message: "m", refusedAt: "2026-10-05T08:02:00Z" },
  ]);
  await store.copies.bulkPut([
    { key: "plan", json: {}, savedAt: "2026-10-05T08:00:00Z" },
    { key: "inventory", json: {}, savedAt: "2026-10-05T08:00:00Z" },
    { key: "profile", json: {}, savedAt: "2026-10-05T08:00:00Z" },
  ]);
  await store.meta.bulkPut([
    { key: REST_KEY, value: {} },
    { key: "persistence-requested", value: true },
  ]);
}

describe("the deletion of all data (D-314, D-315)", () => {
  it("needs the switch at yes and the exact text", () => {
    expect(DELETE_CONFIRMATION).toBe("Delete all data");
    expect(canDelete(true, "Delete all data")).toBe(true);
    expect(canDelete(false, "Delete all data")).toBe(false);
    expect(canDelete(true, "")).toBe(false);
    expect(canDelete(true, "delete all data")).toBe(false);
    expect(canDelete(true, "Delete all data ")).toBe(false);
    expect(canDelete(true, "Delete all")).toBe(false);
  });

  it("deletes the history on the phone, and keeps the profile, the inventory, and their entries", async () => {
    await seed();
    await deleteLocalHistory(store);
    expect(await store.workouts.count()).toBe(0);
    expect(await store.sets.count()).toBe(0);
    expect(await store.cardio.count()).toBe(0);
    expect((await store.outbox.toArray()).map((e) => e.entity)).toEqual(["machine"]);
    expect((await store.refused.toArray()).map((e) => e.entity)).toEqual(["note"]);
    expect((await store.copies.toArray()).map((c) => c.key).sort()).toEqual(["inventory", "profile"]);
    expect((await store.meta.toArray()).map((m) => m.key)).toEqual(["persistence-requested"]);
  });

  it("deletes on the phone first, then waits for a sync, then deletes on the server, then syncs again", async () => {
    await seed();
    const steps: string[] = [];
    const deleted = await deleteAllData(store, {
      sync: async () => {
        steps.push(`sync ${await store.outbox.count()} ${await store.workouts.count()}`);
      },
      deleteOnServer: async () => {
        steps.push("server");
        return 4;
      },
    });
    expect(deleted).toBe(4);
    // At the first sync, the outbox holds the inventory entry alone, and
    // the phone holds no workout.
    expect(steps).toEqual(["sync 1 0", "server", "sync 1 0"]);
  });

  it("throws a failed call of the server after the phone deleted its copy, and a second call passes", async () => {
    await seed();
    const deps = {
      sync: async () => {},
      deleteOnServer: async (): Promise<number> => {
        throw new Error("unavailable");
      },
    };
    await expect(deleteAllData(store, deps)).rejects.toThrow("unavailable");
    expect(await store.workouts.count()).toBe(0);
    expect(await store.outbox.count()).toBe(1);
    await expect(deleteAllData(store, { ...deps, deleteOnServer: async () => 1 })).resolves.toBe(1);
  });
});
