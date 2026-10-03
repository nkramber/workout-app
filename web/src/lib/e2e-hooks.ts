import { db, pendingOutbox, type OutboxEntry, type RefusedEntry, type SetRecord, type WorkoutRecord } from "./db";

// The handles of the browser tests. Only a build in the e2e mode loads
// this file (src/main.tsx). The tests read the tables of the real
// IndexedDB of each engine here.
declare global {
  interface Window {
    workoutAppE2E?: {
      pendingOutbox: () => Promise<OutboxEntry[]>;
      refused: () => Promise<RefusedEntry[]>;
      workouts: () => Promise<WorkoutRecord[]>;
      sets: () => Promise<SetRecord[]>;
    };
  }
}

export function install(): void {
  window.workoutAppE2E = {
    pendingOutbox: () => pendingOutbox(db),
    refused: () => db.refused.toArray(),
    workouts: () => db.workouts.toArray(),
    sets: () => db.sets.toArray(),
  };
}
