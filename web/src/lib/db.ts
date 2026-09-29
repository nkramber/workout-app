import { Dexie, type EntityTable } from "dexie";

import { uuidv7 } from "./uuidv7";

// The offline store of the phone (D-62, D-77). Dexie on IndexedDB holds
// the local state and the outbox (REC-1, D-132). Each change and its
// outbox entry go into one transaction, so the phone never keeps a change
// that it can not sync, and never syncs a change that it did not keep.
// The sync call of the outbox comes in Phase 6.

// The version of the outbox entry form. The server of Phase 6 reads it.
export const OUTBOX_SCHEMA_VERSION = 1;

// An outbox entry, in the form of the platform research, section 5.2.
export type OutboxEntry = {
  opId: string; // a UUIDv7 that the phone makes
  entity: string;
  entityId: string;
  baseVersion: number; // the server version that the change starts from, 0 before the first sync
  payload: unknown;
  at: string; // the time of the change, ISO 8601
  attempts: number;
  schemaVersion: number;
};

// A setting is the one entity of the skeleton. Phase 3 adds the entities
// of the domain model. `version` is the last server version that the
// phone knows, and the sync of Phase 6 sets it.
export type Setting = {
  id: string;
  value: unknown;
  version: number;
  updatedAt: string;
};

// A local fact of the phone, such as the result of the persistent
// storage request (REC-5, D-134). It never syncs.
export type Meta = { key: string; value: unknown };

export class GymRouteDB extends Dexie {
  settings!: EntityTable<Setting, "id">;
  outbox!: EntityTable<OutboxEntry, "opId">;
  meta!: EntityTable<Meta, "key">;

  constructor(name = "gym-route") {
    super(name);
    this.version(1).stores({
      settings: "id",
      outbox: "opId, at",
      meta: "key",
    });
  }
}

export const db = new GymRouteDB();

// Safari on iOS can close the database when the app goes to the
// background. withReopen opens it again and tries the work one more time
// (platform research, section 5.1).
const reopenErrors = new Set(["DatabaseClosedError", "InvalidStateError"]);

export async function withReopen<T>(store: GymRouteDB, work: () => Promise<T>): Promise<T> {
  try {
    return await work();
  } catch (err) {
    if (!(err instanceof Error) || !reopenErrors.has(err.name)) throw err;
    store.close({ disableAutoOpen: false });
    await store.open();
    return work();
  }
}

// saveSetting writes a setting and its outbox entry in one transaction.
// When one write fails, Dexie rolls back both.
export function saveSetting(store: GymRouteDB, id: string, value: unknown, now: Date = new Date()): Promise<OutboxEntry> {
  return withReopen(store, () =>
    store.transaction("rw", store.settings, store.outbox, async () => {
      const at = now.toISOString();
      const baseVersion = (await store.settings.get(id))?.version ?? 0;
      await store.settings.put({ id, value, version: baseVersion, updatedAt: at });
      const entry: OutboxEntry = {
        opId: uuidv7(now.getTime()),
        entity: "setting",
        entityId: id,
        baseVersion,
        payload: value,
        at,
        attempts: 0,
        schemaVersion: OUTBOX_SCHEMA_VERSION,
      };
      await store.outbox.add(entry);
      return entry;
    }),
  );
}

export function pendingOutbox(store: GymRouteDB): Promise<OutboxEntry[]> {
  return withReopen(store, () => store.outbox.orderBy("at").toArray());
}
