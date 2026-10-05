import { createClient, type Transport } from "@connectrpc/connect";
import { liveQuery } from "dexie";

import { create, fromJson, toJson, type JsonValue } from "@bufbuild/protobuf";
import { InventoryService, GetCatalogResponseSchema, GetInventoryResponseSchema } from "../gen/workoutapp/v1/inventory_service_pb";
import { GetPlanResponseSchema, PlanService } from "../gen/workoutapp/v1/plan_service_pb";
import { UserService } from "../gen/workoutapp/v1/user_service_pb";
import {
  EntryResult_Status,
  OutboxEntrySchema,
  WorkoutService,
  type OutboxEntry as OutboxEntryMessage,
  type SyncOutboxResponse,
} from "../gen/workoutapp/v1/workout_service_pb";
import { HISTORY_GENERATION_KEY, INVENTORY_ENTITIES, withReopen, type CopyKey, type OutboxEntry, type WorkoutAppDB } from "./db";
import { isNoConnection } from "./errors";
import { localInventory } from "./inventory-api";
import { planRequest } from "./today";

// The sync of the outbox (work area 6.3). The phone sends its outbox
// through SyncOutbox of `proto/workoutapp/v1/workout_service.proto`, in
// batches of 100 entries or fewer, in the order of the op ids (D-259,
// D-275). The server applies each entry one time by its op id, so a
// batch that the phone sends again after a dropped answer changes
// nothing (D-257). The phone removes each applied entry, and moves each
// refused entry to the list of refused entries (D-274).
//
// The sync runs while the app is open alone: at the open, at each focus
// and return, at each reconnect, after each new entry, and on a timer
// after a failure (D-21, D-277). A reconnect starts the delays of the
// timer again. iOS has no background sync for a web
// app, so the service worker does not sync.

// MAX_BATCH is the largest count of entries in one call (D-259).
export const MAX_BATCH = 100;

// RETRY_DELAYS_MS gives the time before each try after a failure. After
// the last delay, the phone tries again each 5 minutes (D-277).
export const RETRY_DELAYS_MS: readonly number[] = [5_000, 15_000, 60_000, 300_000];

// The codes of a refusal of the server, and the code of an entry that
// the phone can not read.
export const CODE_INVALID = "invalid_argument";

// SyncClient holds the calls of the sync. A test gives a fake.
export type SyncClient = {
  syncOutbox: (entries: OutboxEntryMessage[]) => Promise<SyncOutboxResponse>;
  // copies reads the answers of the server that the sync keeps a copy of
  // (D-250, D-278), as JSON. The profile screen keeps the profile copy.
  copies: () => Promise<Partial<Record<CopyKey, JsonValue>>>;
  // generation reads the generation of the history with GetMe (D-315).
  // A fake of a test can leave it out.
  generation?: () => Promise<number>;
};

// connectClient gives the calls of the API over the transport.
export function connectClient(transport: Transport): SyncClient {
  const workouts = createClient(WorkoutService, transport);
  const inventories = createClient(InventoryService, transport);
  const plans = createClient(PlanService, transport);
  const users = createClient(UserService, transport);
  return {
    syncOutbox: (entries) => workouts.syncOutbox({ entries }),
    generation: async () => (await users.getMe({})).historyGeneration,
    copies: async () => {
      const [catalog, inventory, plan] = await Promise.all([inventories.getCatalog({}), inventories.getInventory({}), plans.getPlan(planRequest())]);
      return {
        catalog: toJson(GetCatalogResponseSchema, catalog),
        inventory: toJson(GetInventoryResponseSchema, inventory),
        plan: toJson(GetPlanResponseSchema, plan),
      };
    },
  };
}

// toMessage gives the message of the contract for an outbox entry. The
// payload of a workout, set, or cardio entry is the JSON form of its
// message, and the payload of an inventory entry is the JSON form of the
// payload field of OutboxEntry, such as {"saveMachine": {...}}. It throws
// for a payload that the contract does not accept.
export function toMessage(e: OutboxEntry): OutboxEntryMessage {
  const payload = INVENTORY_ENTITIES.has(e.entity) ? (e.payload as Record<string, JsonValue>) : { [e.entity]: e.payload as JsonValue };
  return fromJson(OutboxEntrySchema, {
    opId: e.opId,
    entity: e.entity,
    entityId: e.entityId,
    baseVersion: String(e.baseVersion),
    at: e.at,
    schemaVersion: e.schemaVersion,
    ...payload,
  });
}

// DrainResult counts the entries of one drain of the outbox.
export type DrainResult = { applied: number; refused: number };

// drainOutbox sends the outbox in batches until it is empty. An entry
// that the phone can not read goes to the refused entries with the code
// "invalid_argument", so it does not block the outbox. A failed call
// adds 1 to the attempts of each entry of its batch, and throws. Each
// applied entry and each refused entry left the outbox before the
// failure.
export async function drainOutbox(store: WorkoutAppDB, client: SyncClient, now: () => Date = () => new Date()): Promise<DrainResult> {
  const out: DrainResult = { applied: 0, refused: 0 };
  for (;;) {
    const batch = await withReopen(store, () => store.outbox.limit(MAX_BATCH).toArray());
    if (batch.length === 0) return out;

    const sent: OutboxEntry[] = [];
    const messages: OutboxEntryMessage[] = [];
    const unreadable: OutboxEntry[] = [];
    for (const e of batch) {
      try {
        messages.push(toMessage(e));
        sent.push(e);
      } catch {
        unreadable.push(e);
      }
    }
    if (unreadable.length > 0) {
      await refuse(store, unreadable, CODE_INVALID, "the phone can not read the entry", now());
      out.refused += unreadable.length;
    }
    if (sent.length === 0) continue;

    let res: SyncOutboxResponse;
    try {
      res = await client.syncOutbox(messages);
    } catch (err) {
      await withReopen(store, () =>
        store.outbox.bulkUpdate(sent.map((e) => ({ key: e.opId, changes: { attempts: e.attempts + 1 } }))),
      );
      throw err;
    }
    if (res.results.length !== sent.length || res.results.some((r, i) => r.opId !== sent[i].opId)) {
      throw new Error("the results of the sync do not agree with the entries");
    }
    const at = now().toISOString();
    await withReopen(store, () =>
      store.transaction("rw", [store.outbox, store.refused, store.workouts, store.sets, store.cardio, store.copies], async () => {
        const inventory: OutboxEntry[] = [];
        for (const [i, e] of sent.entries()) {
          const r = res.results[i];
          if (r.status === EntryResult_Status.APPLIED) {
            await store.outbox.delete(e.opId);
            await keepVersion(store, e, Number(r.version));
            if (INVENTORY_ENTITIES.has(e.entity)) inventory.push(e);
            out.applied++;
          } else if (r.status === EntryResult_Status.REFUSED) {
            await store.outbox.delete(e.opId);
            await store.refused.put({ ...e, code: r.code, message: r.message, refusedAt: at });
            out.refused++;
          }
        }
        await keepApplied(store, inventory);
      }),
    );
    // An entry with no status stays in the outbox. So a server that gives
    // no status stops the drain here, and the timer tries again.
    if (res.results.some((r) => r.status !== EntryResult_Status.APPLIED && r.status !== EntryResult_Status.REFUSED)) {
      throw new Error("an entry of the sync has no status");
    }
  }
}

function refuse(store: WorkoutAppDB, entries: OutboxEntry[], code: string, message: string, now: Date): Promise<void> {
  const refusedAt = now.toISOString();
  return withReopen(store, () =>
    store.transaction("rw", store.outbox, store.refused, async () => {
      await store.outbox.bulkDelete(entries.map((e) => e.opId));
      await store.refused.bulkPut(entries.map((e) => ({ ...e, code, message, refusedAt })));
    }),
  );
}

// keepApplied puts the applied inventory entries on the inventory copy,
// in the transaction that removes them from the outbox. So the screens
// show each applied change until the sync reads the inventory again, also
// when that read fails.
async function keepApplied(store: WorkoutAppDB, entries: OutboxEntry[]): Promise<void> {
  if (entries.length === 0) return;
  const [catalogCopy, inventoryCopy] = await Promise.all([store.copies.get("catalog"), store.copies.get("inventory")]);
  if (!catalogCopy || !inventoryCopy) return;
  const catalog = fromJson(GetCatalogResponseSchema, catalogCopy.json as JsonValue, { ignoreUnknownFields: true });
  const server = fromJson(GetInventoryResponseSchema, inventoryCopy.json as JsonValue, { ignoreUnknownFields: true }).inventory;
  const { inventory } = localInventory(server, entries, catalog);
  await store.copies.put({ ...inventoryCopy, json: toJson(GetInventoryResponseSchema, create(GetInventoryResponseSchema, { inventory })) });
}

// keepVersion keeps the server version of an applied workout, set, or
// cardio entry on the entity of the phone, so its next change starts from
// it. A larger stored version stays. The inventory has no version.
async function keepVersion(store: WorkoutAppDB, e: OutboxEntry, version: number): Promise<void> {
  const keep = (row: { version: number }) => {
    if (version > row.version) row.version = version;
  };
  if (e.entity === "workout") await store.workouts.where("id").equals(e.entityId).modify(keep);
  else if (e.entity === "set") await store.sets.where("id").equals(e.entityId).modify(keep);
  else if (e.entity === "cardio") await store.cardio.where("id").equals(e.entityId).modify(keep);
}

// refreshCopies reads the catalog, the inventory, and the plan, and keeps
// each copy (D-250, D-278). The server wins: the new inventory copy
// replaces the old one, and the screens put the inventory entries of the
// outbox on it again (D-258).
//
// It also keeps the generation of the history (D-315). So a phone learns
// of a deletion on another device before its next workout starts.
export async function refreshCopies(store: WorkoutAppDB, client: SyncClient, now: () => Date = () => new Date()): Promise<void> {
  const [copies, generation] = await Promise.all([client.copies(), client.generation?.()]);
  const savedAt = now().toISOString();
  await withReopen(store, () =>
    store.copies.bulkPut((Object.keys(copies) as CopyKey[]).map((key) => ({ key, json: copies[key] ?? null, savedAt }))),
  );
  if (generation !== undefined) await setHistoryGeneration(store, generation);
}

// setHistoryGeneration keeps the generation of the history that the
// server gave (D-315).
export function setHistoryGeneration(store: WorkoutAppDB, generation: number): Promise<unknown> {
  return withReopen(store, () => store.meta.put({ key: HISTORY_GENERATION_KEY, value: generation }));
}

// keepCopy keeps one copy, for example the plan that a plan request gave.
export function keepCopy(store: WorkoutAppDB, key: CopyKey, json: JsonValue, now: Date = new Date()): Promise<unknown> {
  return withReopen(store, () => store.copies.put({ key, json, savedAt: now.toISOString() }));
}

// SyncStatus is the state of the sync for the screen (D-276). error is
// "offline" when the phone has no connection, and "failed" for another
// failure, until the next sync passes.
export type SyncStatus = { running: boolean; error: null | "offline" | "failed"; lastSyncAt: string | null };

// SyncEnv holds the events and the timers of the browser. A test gives
// its own.
export type SyncEnv = {
  window: EventTarget;
  document: EventTarget & { visibilityState?: string };
  setTimeout: (fn: () => void, ms: number) => unknown;
  clearTimeout: (id: unknown) => void;
  now: () => Date;
};

function browserEnv(): SyncEnv {
  return {
    window,
    document,
    setTimeout: (fn, ms) => window.setTimeout(fn, ms),
    clearTimeout: (id) => window.clearTimeout(id as number),
    now: () => new Date(),
  };
}

// SyncEngine runs one sync at a time. A request during a sync runs one
// more sync after it, so each entry that existed at the request goes in
// it. A sync drains the outbox, then reads the copies again. The sync
// passes when both steps pass.
export class SyncEngine {
  private status: SyncStatus = { running: false, error: null, lastSyncAt: null };
  private readonly listeners = new Set<() => void>();
  private inflight: Promise<void> | null = null;
  private again = false;
  private failures = 0;
  private timer: unknown = null;
  private env: SyncEnv | null = null;

  constructor(
    private readonly store: WorkoutAppDB,
    private readonly client: SyncClient,
    private readonly makeEnv: () => SyncEnv = browserEnv,
  ) {}

  subscribe = (fn: () => void): (() => void) => {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  };

  getStatus = (): SyncStatus => this.status;

  // start runs a sync now, and at each focus, return, reconnect, and new
  // outbox entry. It gives the function that stops it.
  start(): () => void {
    const env = this.makeEnv();
    this.env = env;
    const run = () => void this.syncNow();
    const onVisible = () => {
      if (env.document.visibilityState !== "hidden") run();
    };
    // A reconnect starts the delays again, so a failure just after it
    // gets a new try after 5 s.
    const onOnline = () => {
      this.failures = 0;
      run();
    };
    env.window.addEventListener("online", onOnline);
    env.window.addEventListener("focus", run);
    env.window.addEventListener("pageshow", run);
    env.document.addEventListener("visibilitychange", onVisible);
    let last = -1;
    const sub = liveQuery(() => this.store.outbox.count()).subscribe({
      next: (count) => {
        if (last >= 0 && count > last) run();
        last = count;
      },
      error: () => {},
    });
    run();
    return () => {
      env.window.removeEventListener("online", onOnline);
      env.window.removeEventListener("focus", run);
      env.window.removeEventListener("pageshow", run);
      env.document.removeEventListener("visibilitychange", onVisible);
      sub.unsubscribe();
      this.clearTimer();
      this.env = null;
    };
  }

  // syncNow runs a sync, or one more sync after the sync that runs. It
  // never throws: the status holds the result.
  syncNow(): Promise<void> {
    if (this.inflight) {
      this.again = true;
      return this.inflight;
    }
    this.inflight = (async () => {
      try {
        do {
          this.again = false;
          await this.once();
        } while (this.again);
      } finally {
        this.inflight = null;
      }
    })();
    return this.inflight;
  }

  private async once(): Promise<void> {
    const now = this.env?.now ?? (() => new Date());
    this.clearTimer();
    this.set({ running: true });
    // The copies come from the server alone, so the phone reads them after
    // a failed drain too. The screens put the entries that wait on them.
    let failure: unknown = null;
    try {
      await drainOutbox(this.store, this.client, now);
    } catch (err) {
      failure = err;
    }
    try {
      await refreshCopies(this.store, this.client, now);
    } catch (err) {
      failure ??= err;
    }
    try {
      if (failure !== null) throw failure;
      this.failures = 0;
      this.set({ running: false, error: null, lastSyncAt: now().toISOString() });
    } catch (err) {
      const delay = RETRY_DELAYS_MS[Math.min(this.failures, RETRY_DELAYS_MS.length - 1)];
      this.failures++;
      this.set({ running: false, error: isNoConnection(err) ? "offline" : "failed" });
      const env = this.env;
      if (env) this.timer = env.setTimeout(() => void this.syncNow(), delay);
    }
  }

  private clearTimer() {
    if (this.timer !== null && this.env) this.env.clearTimeout(this.timer);
    this.timer = null;
  }

  private set(change: Partial<SyncStatus>) {
    this.status = { ...this.status, ...change };
    for (const fn of this.listeners) fn();
  }
}


// syncLineText gives the line of the shell (D-276): "Syncing…",
// "Offline", or "Sync failed", the count of the entries that wait or
// "Synced", and the count of the refused entries.
export function syncLineText(s: Pick<SyncStatus, "running" | "error"> & { waiting: number; refused: number }): string {
  const parts: string[] = [];
  if (s.running) parts.push("Syncing…");
  else if (s.error === "offline") parts.push("Offline");
  else if (s.error === "failed") parts.push("Sync failed");
  if (s.waiting > 0) parts.push(`${s.waiting} waiting`);
  else if (!s.running && !s.error) parts.push("Synced");
  if (s.refused > 0) parts.push(`${s.refused} refused`);
  return parts.join(" · ");
}

// Names gives the catalog names of the machines and of the exercises.
export type Names = { machine: (id: string) => string; exercise: (id: string) => string };

// refusedLabel names the item of a refused entry for the owner.
export function refusedLabel(e: OutboxEntry, names: Names): string {
  const p = (e.payload ?? {}) as Record<string, unknown>;
  switch (e.entity) {
    case "workout":
      return "Workout";
    case "set":
      return `Set of ${names.exercise(String(p.exerciseId ?? ""))}`;
    case "cardio":
      return `Cardio of ${names.exercise(String(p.exerciseId ?? ""))}`;
    case "machine": {
      const name = names.machine(e.entityId);
      if ("confirmMachine" in p) return `Confirmation of ${name}`;
      if ("removeMachine" in p) return `Removal of ${name}`;
      return `Save of ${name}`;
    }
    case "note":
      return "removeNote" in p ? "Removal of a note" : "Note";
    default:
      return e.entity;
  }
}

// refusedReason gives the cause of a refusal for the owner. The message
// of the server, with ids and numbers alone, follows it on the screen.
export function refusedReason(e: OutboxEntry, code: string): string {
  if (code === "failed_precondition") {
    if (e.entity === "machine") return "The server holds other weights, or no such machine. Read the machine again.";
    return "The server holds no such workout.";
  }
  return "The server did not accept the values.";
}

// noCopyText gives the message of a screen whose offline copy the phone
// does not have yet, after a failed sync.
export function noCopyText(error: "offline" | "failed"): string {
  return error === "offline"
    ? "No connection. The phone has no copy yet. Try again when the phone is online."
    : "The sync failed, and the phone has no copy yet. Try again.";
}

// InventoryNotSyncedError tells that an inventory entry still waits in
// the outbox after a sync. A plan reads the inventory of the server, so a
// plan request stops before its call (D-193, D-272).
export class InventoryNotSyncedError extends Error {
  constructor() {
    super("an inventory change waits in the outbox");
    this.name = "InventoryNotSyncedError";
  }
}

// syncBeforePlan runs a sync, then throws InventoryNotSyncedError when an
// inventory entry still waits in the outbox. An entry that the server
// refused left the outbox, and the plan then reads the inventory of the
// server, which wins (D-258).
export async function syncBeforePlan(store: WorkoutAppDB, sync: () => Promise<void>): Promise<void> {
  await sync();
  const waiting = await withReopen(store, () => store.outbox.filter((e) => INVENTORY_ENTITIES.has(e.entity)).count());
  if (waiting > 0) throw new InventoryNotSyncedError();
}
