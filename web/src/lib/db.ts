import { Dexie, type EntityTable } from "dexie";

import { nextId } from "./uuidv7";

// The offline store of the phone (D-62, D-77). Dexie on IndexedDB holds
// the local state and the outbox (REC-1, D-132). Each change and its
// outbox entry go into one transaction, so the phone never keeps a change
// that it can not sync, and never syncs a change that it did not keep.
// The sync call of the outbox comes in PR-32.

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

// A target set of a workout, copied from the plan at the start, so the
// workout needs no network after its start (D-62). A load is in tenths
// of a pound (D-160). A calibration set has an RIR target of 0 (D-150).
export type TargetSet = { reps: number; loadTenthLb: number; rirTarget: number };

// The load of the working sets after a calibration set at `weight`, for
// each result of the calibration table, in tenths of a pound (D-267).
export type CalibrationLoads = { weight: number; down: number; keep: number; upOne: number; upTwo: number };

// An exercise of a workout: the targets of the plan session, and the
// weights of its machine at the start, for the plus and minus buttons
// (D-264). calibrationLoads has one row for each weight of the machine,
// and a plan of policy version 3 has none.
export type WorkoutExercise = {
  exerciseId: string;
  name: string;
  restSeconds: number;
  calibrationSets: TargetSet[];
  calibrationLoads?: CalibrationLoads[];
  workingSets: TargetSet[];
  weights: number[];
};

// The rest timer of a workout (D-59, D-270). The meta table holds it
// under the key REST_KEY, so it never syncs. The end time is a stored
// time in milliseconds, so the timer is correct after a screen lock and
// after a stop of the app.
export type RestTimer = { workoutId: string; endsAt: number };
export const REST_KEY = "rest";

// A workout on the phone (work area 6.1). It holds the session of the
// plan that it started from (D-248), and the state of the header of
// `proto/workoutapp/v1/workout_service.proto`. `version` is the last
// server version that the phone knows, and the sync of PR-32 sets it.
export type WorkoutRecord = {
  id: string;
  date: string;
  planCreatedAt: string;
  sessionIndex: number;
  title: string;
  exercises: WorkoutExercise[];
  cardio: { exerciseId: string; name: string; minutes: number } | null;
  skippedExerciseIds: string[];
  endedEarly: boolean;
  finished: boolean;
  startedAt: string;
  version: number;
};

// A logged set (D-57, D-164, D-249). No value of `pain` means no report
// (D-162).
export type SetRecord = {
  id: string;
  workoutId: string;
  exerciseId: string;
  kind: "working" | "calibration";
  reps: number;
  weightTenthsLb: number;
  rir: number;
  pain?: number;
  note: string;
  at: string;
  version: number;
};

// A cardio log (D-123, D-165). Each optional field has no value when the
// owner gives none.
export type CardioRecord = {
  id: string;
  workoutId: string;
  exerciseId: string;
  durationSeconds: number;
  effort: number;
  distanceTenthsMi?: number;
  resistance?: number;
  pain?: number;
  note: string;
  at: string;
  version: number;
};

export class WorkoutAppDB extends Dexie {
  settings!: EntityTable<Setting, "id">;
  outbox!: EntityTable<OutboxEntry, "opId">;
  meta!: EntityTable<Meta, "key">;
  workouts!: EntityTable<WorkoutRecord, "id">;
  sets!: EntityTable<SetRecord, "id">;
  cardio!: EntityTable<CardioRecord, "id">;

  constructor(name = "workout-app") {
    super(name);
    this.version(1).stores({
      settings: "id",
      outbox: "opId, at",
      meta: "key",
    });
    // Version 2 adds the workout log of work area 6.1. The earlier tables
    // and their data stay.
    this.version(2).stores({
      workouts: "id, startedAt",
      sets: "id, workoutId, at",
      cardio: "id, workoutId, at",
    });
  }
}

export const db = new WorkoutAppDB();

// Safari on iOS can close the database when the app goes to the
// background. withReopen opens it again and tries the work one more time
// (platform research, section 5.1).
const reopenErrors = new Set(["DatabaseClosedError", "InvalidStateError"]);

export async function withReopen<T>(store: WorkoutAppDB, work: () => Promise<T>): Promise<T> {
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
export function saveSetting(store: WorkoutAppDB, id: string, value: unknown, now: Date = new Date()): Promise<OutboxEntry> {
  return withReopen(store, () =>
    store.transaction("rw", store.settings, store.outbox, async () => {
      const at = now.toISOString();
      const baseVersion = (await store.settings.get(id))?.version ?? 0;
      await store.settings.put({ id, value, version: baseVersion, updatedAt: at });
      const entry: OutboxEntry = {
        opId: nextId(now.getTime()),
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

export function pendingOutbox(store: WorkoutAppDB): Promise<OutboxEntry[]> {
  return withReopen(store, () => store.outbox.orderBy("at").toArray());
}
