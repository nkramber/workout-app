import { db, pendingOutbox, saveSetting, type OutboxEntry, type SetRecord, type Setting, type WorkoutRecord } from "./db";

// The handles of the browser tests. Only a build in the e2e mode loads
// this file (src/main.tsx). Phase 2 has no screen that makes a change, so
// the tests make the change through this handle, in the real IndexedDB
// of each engine. The tests of the workout read its tables here.
declare global {
  interface Window {
    workoutAppE2E?: {
      saveSetting: (id: string, value: unknown) => Promise<OutboxEntry>;
      pendingOutbox: () => Promise<OutboxEntry[]>;
      settings: () => Promise<Setting[]>;
      workouts: () => Promise<WorkoutRecord[]>;
      sets: () => Promise<SetRecord[]>;
    };
  }
}

export function install(): void {
  window.workoutAppE2E = {
    saveSetting: (id, value) => saveSetting(db, id, value),
    pendingOutbox: () => pendingOutbox(db),
    settings: () => db.settings.toArray(),
    workouts: () => db.workouts.toArray(),
    sets: () => db.sets.toArray(),
  };
}
