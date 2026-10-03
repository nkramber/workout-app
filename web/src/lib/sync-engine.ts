import { useLiveQuery } from "dexie-react-hooks";
import { useEffect, useSyncExternalStore } from "react";

import { transport } from "./api";
import { db, withReopen, type RefusedEntry } from "./db";
import { connectClient, SyncEngine, type SyncStatus } from "./sync";

// The one sync engine of the app (work area 6.3). The signed-in part of
// the app starts it, and a sign-out stops it, so a sync always has a
// signed-in owner.
export const engine = new SyncEngine(db, connectClient(transport));

// useSync starts the engine while the component shows.
export function useSync(): void {
  useEffect(() => engine.start(), []);
}

// useSyncStatus gives the state of the sync, the count of the entries
// that wait, and the refused entries, oldest first (D-274, D-276). The
// counts are null while the store loads.
export function useSyncStatus(): SyncStatus & { waiting: number | null; refused: RefusedEntry[] | null } {
  const status = useSyncExternalStore(engine.subscribe, engine.getStatus);
  const waiting = useLiveQuery(() => withReopen(db, () => db.outbox.count()), [], null);
  const refused = useLiveQuery(() => withReopen(db, () => db.refused.orderBy("at").toArray()), [], null);
  return { ...status, waiting, refused };
}

// dismissRefused removes a refused entry from the list (D-274).
export function dismissRefused(opId: string): Promise<void> {
  return withReopen(db, () => db.refused.delete(opId));
}

// useSyncWhenMissing asks for a sync when the phone has no copy that a
// screen needs, and gives the state of the sync.
export function useSyncWhenMissing(missing: boolean): SyncStatus {
  const status = useSyncExternalStore(engine.subscribe, engine.getStatus);
  useEffect(() => {
    if (missing) void engine.syncNow();
  }, [missing]);
  return status;
}
