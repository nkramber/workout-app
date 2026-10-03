import { fromJson, type JsonValue } from "@bufbuild/protobuf";
import { useLiveQuery } from "dexie-react-hooks";
import { useState } from "react";

import { GetCatalogResponseSchema } from "./gen/workoutapp/v1/inventory_service_pb";
import { db } from "./lib/db";
import { refusedLabel, refusedReason, syncLineText, type Names } from "./lib/sync";
import { dismissRefused, engine, useSyncStatus } from "./lib/sync-engine";
import { secondary } from "./pages/inventory/ui";

// SyncLine is the line of the state of the sync, below the header of the
// shell, on each screen of a signed-in owner (D-276). A tap opens the
// detail: the counts, the time of the last sync, "Sync now", and each
// refused entry with "Dismiss" (D-274).
export function SyncLine() {
  const s = useSyncStatus();
  const [open, setOpen] = useState(false);
  const names = useCatalogNames();
  if (s.waiting === null || s.refused === null) return null;
  const text = syncLineText({ running: s.running, error: s.error, waiting: s.waiting, refused: s.refused.length });
  const warn = !!s.error || s.refused.length > 0;

  return (
    <div className="border-b border-slate-800 px-4 text-sm">
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen(!open)}
        className={`flex min-h-11 w-full items-center justify-between text-left ${warn ? "text-amber-300" : "text-slate-400"}`}
      >
        <span data-testid="sync-line">{text}</span>
        <span aria-hidden="true">{open ? "▴" : "▾"}</span>
      </button>
      {open && (
        <div className="space-y-3 pb-3" data-testid="sync-detail">
          <p className="text-slate-300">
            {s.lastSyncAt ? `Last sync: ${new Date(s.lastSyncAt).toLocaleTimeString()}` : "No sync yet in this visit."}
          </p>
          <button type="button" className={`${secondary} w-full`} disabled={s.running} onClick={() => void engine.syncNow()}>
            Sync now
          </button>
          {s.refused.length > 0 && (
            <ul className="space-y-2" aria-label="Refused changes">
              {s.refused.map((e) => (
                <li key={e.opId} data-testid="refused-entry" className="space-y-1 rounded-lg border border-amber-900 p-2">
                  <p className="font-medium text-slate-100">{refusedLabel(e, names)}</p>
                  <p className="text-slate-300">{refusedReason(e, e.code)}</p>
                  <p className="text-xs break-words text-slate-500">{`${e.code}: ${e.message}`}</p>
                  <button type="button" className={`${secondary} w-full`} onClick={() => void dismissRefused(e.opId)}>
                    Dismiss
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}

// useCatalogNames reads the names of the catalog copy. An id with no
// copy shows as it is.
function useCatalogNames(): Names {
  const copy = useLiveQuery(() => db.copies.get("catalog"), [], undefined);
  let machines = new Map<string, string>();
  let exercises = new Map<string, string>();
  if (copy) {
    const c = fromJson(GetCatalogResponseSchema, copy.json as JsonValue, { ignoreUnknownFields: true });
    machines = new Map(c.machines.map((m) => [m.id, m.name]));
    exercises = new Map(c.exercises.map((e) => [e.id, e.name]));
  }
  return { machine: (id) => machines.get(id) ?? id, exercise: (id) => exercises.get(id) ?? id };
}
