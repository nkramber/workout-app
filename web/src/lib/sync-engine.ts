import { createConnectQueryKey } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useLiveQuery } from "dexie-react-hooks";
import { useEffect, useRef, useSyncExternalStore } from "react";

import { PlanService } from "../gen/workoutapp/v1/plan_service_pb";

import { transport } from "./api";
import { db, withReopen, type RefusedEntry } from "./db";
import { planRequest } from "./today";
import { revisedAt } from "./revision";
import { connectClient, SyncEngine, type SyncStatus } from "./sync";

// The one sync engine of the app (work area 6.3). The signed-in part of
// the app starts it, and a sign-out stops it, so a sync always has a
// signed-in owner.
export const engine = new SyncEngine(db, connectClient(transport));

// useSync starts the engine while the component shows. A sync of a
// finished workout revises the plan on the server (D-292), and the sync
// then keeps the new plan copy. So a new revision in the plan copy marks
// the cached plan as stale, and the plan screen reads the plan again. A
// sync with no revision leaves the cache, so it never races a plan
// request.
export function useSync(): void {
  const queryClient = useQueryClient();
  useEffect(() => engine.start(), []);
  const revised = useLiveQuery(() => withReopen(db, async () => revisedAt((await db.copies.get("plan"))?.json)), [], null);
  const seen = useRef<string | null>(null);
  useEffect(() => {
    if (revised === null) return;
    if (seen.current !== null && seen.current !== revised) {
      const key = createConnectQueryKey({ schema: PlanService.method.getPlan, transport, input: planRequest(), cardinality: "finite" });
      void queryClient.invalidateQueries({ queryKey: key });
    }
    seen.current = revised;
  }, [revised, queryClient]);
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
