import "fake-indexeddb/auto";

import { create, fromJson, type JsonValue } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { GetCatalogResponseSchema, MachineState } from "../gen/workoutapp/v1/inventory_service_pb";
import { PlanSchema } from "../gen/workoutapp/v1/plan_service_pb";
import {
  EntryResult_Status,
  EntryResultSchema,
  SyncOutboxResponseSchema,
  type OutboxEntry as OutboxEntryMessage,
} from "../gen/workoutapp/v1/workout_service_pb";
import { OUTBOX_SCHEMA_VERSION, pendingOutbox, WorkoutAppDB, type CopyKey, type OutboxEntry } from "./db";
import { confirmMachine, readOfflineInventory, removeNote, saveMachine, saveNote } from "./inventory-api";
import {
  drainOutbox,
  MAX_BATCH,
  refreshCopies,
  refusedLabel,
  refusedReason,
  RETRY_DELAYS_MS,
  SyncEngine,
  syncLineText,
  toMessage,
  type SyncClient,
  type SyncEnv,
} from "./sync";
import { finishWorkout, logSet, startWorkout } from "./workout";

let store: WorkoutAppDB;
let n = 0;

beforeEach(async () => {
  store = new WorkoutAppDB(`sync-test-${n++}`);
  await store.open();
});

afterEach(async () => {
  vi.useRealTimers();
  await store.delete();
});

const plan = create(PlanSchema, {
  createdAt: "2026-10-03T08:00:00Z",
  sessions: [
    {
      title: "Session 1",
      exercises: [
        {
          exerciseId: "chest_press",
          name: "Chest press",
          restSeconds: 120,
          workingSets: [
            { reps: 10, loadTenthLb: 500, rirTarget: 2 },
            { reps: 10, loadTenthLb: 500, rirTarget: 2 },
          ],
        },
      ],
    },
  ],
});

const catalogJson = {
  version: 1,
  machines: [
    { id: "leg_press", name: "Leg press", kind: "machine" },
    { id: "chest_press", name: "Chest press", kind: "machine" },
    { id: "treadmill", name: "Treadmill", kind: "cardio" },
  ],
  exercises: [{ id: "chest_press", name: "Chest press", machineId: "chest_press", region: "upper_push" }],
};

// FakeServer applies each op id one time, as the server of D-257. It
// refuses an entry when refuse gives a code. dropNext applies the next
// batch, then fails as a dropped connection does.
class FakeServer implements SyncClient {
  applied = new Map<string, bigint>();
  versions = new Map<string, bigint>();
  batches: OutboxEntryMessage[][] = [];
  refuse: (e: OutboxEntryMessage) => string = () => "";
  dropNext = false;
  down = false;
  copiesJson: Partial<Record<CopyKey, JsonValue>> = { catalog: catalogJson, inventory: { inventory: {} }, plan: {} };

  async syncOutbox(entries: OutboxEntryMessage[]) {
    if (this.down) throw new ConnectError("fetch failed", Code.Unknown, undefined, undefined, new TypeError("Failed to fetch"));
    this.batches.push(entries);
    const results = entries.map((e) => {
      const code = this.refuse(e);
      if (code) return create(EntryResultSchema, { opId: e.opId, status: EntryResult_Status.REFUSED, code, message: `entity ${e.entityId}` });
      let version = this.applied.get(e.opId);
      if (version === undefined) {
        const key = e.entity === "set" || e.entity === "cardio" ? e.entityId : e.entity === "workout" ? e.entityId : "";
        version = key ? (this.versions.get(key) ?? 0n) + 1n : 0n;
        if (key) this.versions.set(key, version);
        this.applied.set(e.opId, version);
      }
      return create(EntryResultSchema, { opId: e.opId, status: EntryResult_Status.APPLIED, version });
    });
    if (this.dropNext) {
      this.dropNext = false;
      throw new ConnectError("fetch failed", Code.Unknown, undefined, undefined, new TypeError("Failed to fetch"));
    }
    return create(SyncOutboxResponseSchema, { results });
  }

  async copies() {
    if (this.down) throw new ConnectError("fetch failed", Code.Unknown, undefined, undefined, new TypeError("Failed to fetch"));
    return this.copiesJson;
  }
}

const now = new Date("2026-10-03T15:00:00.000Z");
const weights = () => [100, 200, 300, 400, 500];

async function workoutWithSets(count: number) {
  const w = await startWorkout(store, plan, 0, weights, now);
  for (let i = 0; i < count; i++) {
    await logSet(store, w.id, { exerciseId: "chest_press", kind: "working", reps: 10, weightTenthsLb: 500, rir: 2 }, now);
  }
  return w;
}

describe("toMessage", () => {
  it("puts a workout payload under its entity, and an inventory payload as it is", async () => {
    const w = await workoutWithSets(1);
    await saveMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200] }, now);
    const [header, set, machine] = (await pendingOutbox(store)).map(toMessage);
    expect(header.payload.case).toBe("workout");
    expect(header.entityId).toBe(w.id);
    expect(set.payload.case).toBe("set");
    expect(set.schemaVersion).toBe(OUTBOX_SCHEMA_VERSION);
    expect(machine.payload).toMatchObject({ case: "saveMachine", value: { weightsTenthLb: [100, 200] } });
    expect(machine.baseVersion).toBe(0n);
  });
});

describe("drainOutbox", () => {
  it("removes each applied entry, and keeps the server version on the entity", async () => {
    const w = await workoutWithSets(2);
    await finishWorkout(store, w.id, now);
    const server = new FakeServer();

    expect(await drainOutbox(store, server)).toEqual({ applied: 4, refused: 0 });
    expect(await store.outbox.count()).toBe(0);
    expect((await store.workouts.get(w.id))?.version).toBe(2);
    expect((await store.sets.toArray()).map((s) => s.version)).toEqual([1, 1]);
  });

  it("sends the entries in the order of the op ids, in batches of 100 or fewer (D-259, D-275)", async () => {
    await workoutWithSets(MAX_BATCH + 20);
    await saveNote(store, "rope handle", "", now);
    const order = (await pendingOutbox(store)).map((e) => e.opId);
    const server = new FakeServer();

    await drainOutbox(store, server);
    expect(server.batches.map((b) => b.length)).toEqual([MAX_BATCH, 22]);
    expect(server.batches.flat().map((e) => e.opId)).toEqual(order);
    expect(server.batches[1].at(-1)?.entity).toBe("note");
  });

  it("moves each refused entry to the refused list, and applies the next entry (D-274)", async () => {
    await workoutWithSets(2);
    const [, firstSet] = await pendingOutbox(store);
    const server = new FakeServer();
    server.refuse = (e) => (e.opId === firstSet.opId ? "invalid_argument" : "");

    expect(await drainOutbox(store, server, () => now)).toEqual({ applied: 2, refused: 1 });
    expect(await store.outbox.count()).toBe(0);
    const [refused] = await store.refused.toArray();
    expect(refused).toMatchObject({ opId: firstSet.opId, code: "invalid_argument", refusedAt: now.toISOString() });
    // The set stays on the phone.
    expect(await store.sets.count()).toBe(2);
  });

  it("refuses an entry that the phone can not read, so it does not block the outbox", async () => {
    await store.outbox.add({
      opId: "01920000-0000-7000-8000-000000000001",
      entity: "set",
      entityId: "01920000-0000-7000-9000-000000000001",
      baseVersion: 0,
      payload: { reps: "many" },
      at: now.toISOString(),
      attempts: 0,
      schemaVersion: OUTBOX_SCHEMA_VERSION,
    });
    await saveNote(store, "rope handle", "", now);
    const server = new FakeServer();

    expect(await drainOutbox(store, server)).toEqual({ applied: 1, refused: 1 });
    expect((await store.refused.toArray())[0]).toMatchObject({ code: "invalid_argument", message: "the phone can not read the entry" });
  });

  it("keeps each entry after a failed call, and adds 1 to its attempts", async () => {
    await workoutWithSets(1);
    const server = new FakeServer();
    server.down = true;

    await expect(drainOutbox(store, server)).rejects.toThrow();
    expect((await pendingOutbox(store)).map((e) => e.attempts)).toEqual([1, 1]);
  });

  it("sends the batch again after a dropped answer, and the server applies each entry one time", async () => {
    await workoutWithSets(3);
    const server = new FakeServer();
    server.dropNext = true;

    await expect(drainOutbox(store, server)).rejects.toThrow();
    expect(await store.outbox.count()).toBe(4);
    await drainOutbox(store, server);
    expect(await store.outbox.count()).toBe(0);
    expect(server.batches).toHaveLength(2);
    expect(server.applied.size).toBe(4);
    expect([...server.versions.values()]).toEqual([1n, 1n, 1n, 1n]);
  });
});

describe("the inventory of the phone", () => {
  it("shows the changes that wait on the copy, and the server copy after the sync (D-258, D-273)", async () => {
    const server = new FakeServer();
    await refreshCopies(store, server);
    await saveMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200, 300] }, now);
    await confirmMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200, 300] }, now);
    await saveMachine(store, { machineId: "treadmill" }, now);
    await saveNote(store, "  rope handle ", "", now);

    let local = await readOfflineInventory(store);
    expect(local?.inventory.machines.map((m) => [m.machineId, m.state])).toEqual([
      ["leg_press", MachineState.CONFIRMED],
      ["treadmill", MachineState.CONFIRMED],
    ]);
    expect(local?.inventory.notes.map((n) => n.text)).toEqual(["rope handle"]);
    expect([...(local?.pending.machines ?? [])]).toEqual(["leg_press", "treadmill"]);
    const noteId = local!.inventory.notes[0].id;

    // The server refuses the confirmation: its weights are other weights.
    server.refuse = (e) => (e.payload.case === "confirmMachine" ? "failed_precondition" : "");
    server.copiesJson.inventory = {
      inventory: {
        machines: [{ machineId: "leg_press", weightsTenthLb: [100, 200, 300], state: "MACHINE_STATE_DRAFT" }],
        notes: [{ id: noteId, text: "rope handle" }],
      },
    };
    await drainOutbox(store, server);
    await refreshCopies(store, server);
    local = await readOfflineInventory(store);
    expect(local?.inventory.machines.map((m) => [m.machineId, m.state])).toEqual([["leg_press", MachineState.DRAFT]]);
    expect(local?.pending.machines.size).toBe(0);

    // A removal that waits goes on the new copy.
    await removeNote(store, noteId, now);
    local = await readOfflineInventory(store);
    expect(local?.inventory.notes).toEqual([]);
    expect([...(local?.pending.notes ?? [])]).toEqual([noteId]);
  });

  it("keeps each applied change on the copy when the read of the inventory fails after the drain", async () => {
    const server = new FakeServer();
    await refreshCopies(store, server);
    await saveMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200] }, now);
    await confirmMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200] }, now);
    await drainOutbox(store, server);
    const local = await readOfflineInventory(store);
    expect(local?.inventory.machines.map((m) => [m.machineId, m.state])).toEqual([["leg_press", MachineState.CONFIRMED]]);
    expect(local?.pending.machines.size).toBe(0);
  });

  it("keeps a confirmed machine confirmed when a save keeps its weights", async () => {
    const server = new FakeServer();
    server.copiesJson.inventory = {
      inventory: { machines: [{ machineId: "leg_press", weightsTenthLb: [100, 200], state: "MACHINE_STATE_CONFIRMED" }] },
    };
    await refreshCopies(store, server);
    await saveMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200], estimates: [] }, now);
    expect((await readOfflineInventory(store))?.inventory.machines[0].state).toBe(MachineState.CONFIRMED);
    await saveMachine(store, { machineId: "leg_press", weightsTenthLb: [100, 200, 300] }, now);
    expect((await readOfflineInventory(store))?.inventory.machines[0].state).toBe(MachineState.DRAFT);
  });

  it("gives null with no copy of the server", async () => {
    expect(await readOfflineInventory(store)).toBeNull();
  });
});

// A test environment: two event targets and the fake timers of vitest.
function testEnv(): SyncEnv & { document: EventTarget & { visibilityState: string } } {
  const doc = Object.assign(new EventTarget(), { visibilityState: "visible" });
  return {
    window: new EventTarget(),
    document: doc,
    setTimeout: (fn, ms) => setTimeout(fn, ms),
    clearTimeout: (id) => clearTimeout(id as ReturnType<typeof setTimeout>),
    now: () => new Date(),
  };
}

async function settle(engine: SyncEngine) {
  for (let i = 0; i < 20 && engine.getStatus().running; i++) await vi.advanceTimersByTimeAsync(1);
  await vi.advanceTimersByTimeAsync(1);
}

describe("SyncEngine", () => {
  it("syncs at the start, then tries again after 5 s, 15 s, 60 s, and each 5 minutes (D-277)", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
    await workoutWithSets(1);
    const server = new FakeServer();
    server.down = true;
    const calls = vi.spyOn(server, "syncOutbox");
    const env = testEnv();
    const engine = new SyncEngine(store, server, () => env);
    const stop = engine.start();
    await settle(engine);
    expect(calls).toHaveBeenCalledTimes(1);
    expect(engine.getStatus().error).toBe("offline");

    // settle moves the clock a few milliseconds, so each check comes 50 ms
    // before the delay ends.
    let count = 1;
    for (const delay of [...RETRY_DELAYS_MS, 300_000]) {
      await vi.advanceTimersByTimeAsync(delay - 50);
      expect(calls).toHaveBeenCalledTimes(count);
      await vi.advanceTimersByTimeAsync(50);
      await settle(engine);
      expect(calls).toHaveBeenCalledTimes(++count);
    }

    server.down = false;
    await vi.advanceTimersByTimeAsync(300_000);
    await settle(engine);
    expect(engine.getStatus()).toMatchObject({ running: false, error: null });
    expect(await store.outbox.count()).toBe(0);
    // After a pass, no timer stays.
    await vi.advanceTimersByTimeAsync(600_000);
    expect(calls).toHaveBeenCalledTimes(count + 1);
    stop();
  });

  it("starts the delays again at a reconnect", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
    await workoutWithSets(1);
    const server = new FakeServer();
    server.down = true;
    const calls = vi.spyOn(server, "syncOutbox");
    const env = testEnv();
    const engine = new SyncEngine(store, server, () => env);
    const stop = engine.start();
    await settle(engine);
    for (const delay of RETRY_DELAYS_MS.slice(0, 3)) {
      await vi.advanceTimersByTimeAsync(delay);
      await settle(engine);
    }
    expect(calls).toHaveBeenCalledTimes(4);

    // The next try would come after 5 minutes. A reconnect that fails
    // gives a try after 5 s again.
    env.window.dispatchEvent(new Event("online"));
    await settle(engine);
    expect(calls).toHaveBeenCalledTimes(5);
    server.down = false;
    await vi.advanceTimersByTimeAsync(RETRY_DELAYS_MS[0]);
    await settle(engine);
    expect(calls).toHaveBeenCalledTimes(6);
    expect(engine.getStatus().error).toBeNull();
    stop();
  });

  it("syncs at a reconnect, a focus, and a return to the front, and at a new entry", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
    const server = new FakeServer();
    const copies = vi.spyOn(server, "copies");
    const env = testEnv();
    const engine = new SyncEngine(store, server, () => env);
    const stop = engine.start();
    await settle(engine);
    expect(copies).toHaveBeenCalledTimes(1);

    env.window.dispatchEvent(new Event("online"));
    await settle(engine);
    env.window.dispatchEvent(new Event("focus"));
    await settle(engine);
    env.document.visibilityState = "hidden";
    env.document.dispatchEvent(new Event("visibilitychange"));
    await settle(engine);
    expect(copies).toHaveBeenCalledTimes(3);
    env.document.visibilityState = "visible";
    env.document.dispatchEvent(new Event("visibilitychange"));
    await settle(engine);
    expect(copies).toHaveBeenCalledTimes(4);

    await saveNote(store, "rope handle", "", now);
    await vi.waitFor(() => expect(server.batches).toHaveLength(1));
    await settle(engine);
    expect(await store.outbox.count()).toBe(0);

    stop();
    env.window.dispatchEvent(new Event("online"));
    await settle(engine);
    expect(copies).toHaveBeenCalledTimes(5);
  });

  it("runs one more sync after a request during a sync, so a new entry goes in it", async () => {
    const server = new FakeServer();
    let release!: () => void;
    const gate = new Promise<void>((r) => (release = r));
    const first = server.copies.bind(server);
    let calls = 0;
    server.copies = async () => {
      if (calls++ === 0) await gate;
      return first();
    };
    const engine = new SyncEngine(store, server, testEnv);
    const a = engine.syncNow();
    await saveNote(store, "rope handle", "", now);
    const b = engine.syncNow();
    release();
    await Promise.all([a, b]);
    expect(calls).toBe(2);
    expect(await store.outbox.count()).toBe(0);
    expect(engine.getStatus().lastSyncAt).not.toBeNull();
  });
});

describe("syncLineText", () => {
  it("gives the line of the shell (D-276)", () => {
    const base = { running: false, error: null, waiting: 0, refused: 0 } as const;
    expect(syncLineText(base)).toBe("Synced");
    expect(syncLineText({ ...base, waiting: 3 })).toBe("3 waiting");
    expect(syncLineText({ ...base, error: "offline", waiting: 3 })).toBe("Offline · 3 waiting");
    expect(syncLineText({ ...base, error: "offline" })).toBe("Offline");
    expect(syncLineText({ ...base, error: "failed", waiting: 1, refused: 2 })).toBe("Sync failed · 1 waiting · 2 refused");
    expect(syncLineText({ ...base, running: true, waiting: 2 })).toBe("Syncing… · 2 waiting");
    expect(syncLineText({ ...base, refused: 1 })).toBe("Synced · 1 refused");
  });
});

describe("refusedLabel and refusedReason", () => {
  const catalog = fromJson(GetCatalogResponseSchema, catalogJson);
  const names = {
    machine: (id: string) => catalog.machines.find((m) => m.id === id)?.name ?? id,
    exercise: (id: string) => catalog.exercises.find((e) => e.id === id)?.name ?? id,
  };
  const entry = (entity: string, payload: unknown, entityId = "leg_press"): OutboxEntry => ({
    opId: "01920000-0000-7000-8000-000000000001",
    entity,
    entityId,
    baseVersion: 0,
    payload,
    at: now.toISOString(),
    attempts: 0,
    schemaVersion: 1,
  });

  it("names the item of each entry", () => {
    expect(refusedLabel(entry("machine", { confirmMachine: {} }), names)).toBe("Confirmation of Leg press");
    expect(refusedLabel(entry("machine", { saveMachine: {} }), names)).toBe("Save of Leg press");
    expect(refusedLabel(entry("machine", { removeMachine: {} }), names)).toBe("Removal of Leg press");
    expect(refusedLabel(entry("set", { exerciseId: "chest_press" }), names)).toBe("Set of Chest press");
    expect(refusedLabel(entry("workout", {}), names)).toBe("Workout");
    expect(refusedLabel(entry("note", { saveNote: { text: "x" } }), names)).toBe("Note");
  });

  it("gives the cause of the code", () => {
    expect(refusedReason(entry("machine", { confirmMachine: {} }), "failed_precondition")).toMatch(/other weights/);
    expect(refusedReason(entry("set", {}), "failed_precondition")).toBe("The server holds no such workout.");
    expect(refusedReason(entry("set", {}), "invalid_argument")).toBe("The server did not accept the values.");
  });
});
