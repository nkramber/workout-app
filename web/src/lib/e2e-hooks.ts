import { db, pendingOutbox, saveSetting, type OutboxEntry, type Setting } from "./db";

// The handles of the browser tests. Only a build in the e2e mode loads
// this file (src/main.tsx). Phase 2 has no screen that makes a change, so
// the tests make the change through this handle, in the real IndexedDB
// of each engine.
declare global {
  interface Window {
    workoutAppE2E?: {
      saveSetting: (id: string, value: unknown) => Promise<OutboxEntry>;
      pendingOutbox: () => Promise<OutboxEntry[]>;
      settings: () => Promise<Setting[]>;
    };
  }
}

export function install(): void {
  window.workoutAppE2E = {
    saveSetting: (id, value) => saveSetting(db, id, value),
    pendingOutbox: () => pendingOutbox(db),
    settings: () => db.settings.toArray(),
  };
}
