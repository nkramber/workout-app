import "fake-indexeddb/auto";

import { Dexie } from "dexie";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import { WorkoutAppDB, OUTBOX_SCHEMA_VERSION, pendingOutbox, withReopen, type OutboxEntry } from "./db";

let store: WorkoutAppDB;
let n = 0;

beforeEach(async () => {
  store = new WorkoutAppDB(`test-${n++}`);
  await store.open();
});

afterEach(async () => {
  await store.delete();
});

const entry = (opId: string, entity = "set"): OutboxEntry => ({
  opId,
  entity,
  entityId: "01920000-0000-7000-9000-000000000001",
  baseVersion: 0,
  payload: {},
  at: "2026-10-03T12:00:00.000Z",
  attempts: 0,
  schemaVersion: OUTBOX_SCHEMA_VERSION,
});

describe("pendingOutbox", () => {
  it("gives the entries in the order of the op ids, the order of the sync (D-275)", async () => {
    await store.outbox.bulkAdd([entry("0192000b-0000-7000-8000-000000000000"), entry("0192000a-0000-7000-8000-000000000000")]);
    expect((await pendingOutbox(store)).map((e) => e.opId)).toEqual([
      "0192000a-0000-7000-8000-000000000000",
      "0192000b-0000-7000-8000-000000000000",
    ]);
  });

  it("opens the store again after the browser closed it", async () => {
    await store.outbox.add(entry("0192000a-0000-7000-8000-000000000000"));
    store.close();
    expect(await withReopen(store, () => pendingOutbox(store))).toHaveLength(1);
  });
});

describe("version 3", () => {
  it("removes the settings of the skeleton and their outbox entries, and keeps each other entry", async () => {
    const name = `upgrade-${n++}`;
    const old = new Dexie(name);
    old.version(2).stores({ settings: "id", outbox: "opId, at", meta: "key", workouts: "id, startedAt", sets: "id, workoutId, at", cardio: "id, workoutId, at" });
    await old.open();
    await old.table("settings").put({ id: "rest-seconds", value: 90, version: 0, updatedAt: "2026-09-29T12:00:00.000Z" });
    await old.table("outbox").bulkAdd([entry("0192000a-0000-7000-8000-000000000000", "setting"), entry("0192000b-0000-7000-8000-000000000000")]);
    old.close();

    const upgraded = new WorkoutAppDB(name);
    await upgraded.open();
    try {
      expect(upgraded.tables.map((t) => t.name)).not.toContain("settings");
      expect((await pendingOutbox(upgraded)).map((e) => e.entity)).toEqual(["set"]);
      expect(await upgraded.refused.count()).toBe(0);
      expect(await upgraded.copies.count()).toBe(0);
    } finally {
      await upgraded.delete();
    }
  });
});
