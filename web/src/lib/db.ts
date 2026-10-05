import { Dexie, type EntityTable } from "dexie";


// The offline store of the phone (D-62, D-77). Dexie on IndexedDB holds
// the local state and the outbox (REC-1, D-132). Each change and its
// outbox entry go into one transaction, so the phone never keeps a change
// that it can not sync, and never syncs a change that it did not keep.
// src/lib/sync.ts sends the outbox to the server (work area 6.3).

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

// A refused entry: an outbox entry that the server refused, with the
// code and the cause of the refusal (D-274). The phone moved it out of
// the outbox, and never sends it again, because a retry gets the same
// refusal. The owner dismisses it. The cause holds ids and numbers alone
// (D-80).
export type RefusedEntry = OutboxEntry & { code: string; message: string; refusedAt: string };

// The inventory entities of the outbox (D-272).
export const INVENTORY_ENTITIES: ReadonlySet<string> = new Set(["machine", "note"]);

// The offline copies of the answers of the server (D-250, D-278): the
// catalog, the inventory, the plan, and the profile. `json` is the JSON
// form of GetCatalogResponse, GetInventoryResponse, GetPlanResponse, or
// GetProfileResponse. The profile copy lets the profile gate of the app
// open with no connection. A copy
// changes only after a read of the server, and never syncs. The screens
// show the inventory copy with the inventory entries of the outbox on it.
export type CopyKey = "catalog" | "inventory" | "plan" | "profile";
export type Copy = { key: CopyKey; json: unknown; savedAt: string };

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
// (D-264). A plan of policy version 6 or earlier can have calibration
// sets, and calibrationLoads has one row for each weight of the machine.
// From version 7, firstSetCalibration tells that the first working set
// is the calibration: the other working sets get the weight that the
// owner logged for it (D-297, D-299). From version 8, followMaxTenthLb
// is the heaviest load that the other working sets can use after a
// heavier first set, and a lighter first set has no limit (D-306 to
// D-308). After an override of the owner,
// the sets are the sets of the override, and `override` keeps the
// recommendation and the reason as separate records (D-69, D-293).
export type WorkoutExercise = {
  exerciseId: string;
  name: string;
  restSeconds: number;
  calibrationSets: TargetSet[];
  calibrationLoads?: CalibrationLoads[];
  firstSetCalibration?: boolean;
  followMaxTenthLb?: number;
  workingSets: TargetSet[];
  weights: number[];
  override?: { reason: string; recommendedWorkingSets: TargetSet[] };
};

// The rest timer of a workout (D-59, D-270). The meta table holds it
// under the key REST_KEY, so it never syncs. The end time is a stored
// time in milliseconds, so the timer is correct after a screen lock and
// after a stop of the app.
export type RestTimer = { workoutId: string; endsAt: number };
export const REST_KEY = "rest";

// The generation of the history that the server gave last, from GetMe or
// DeleteHistory (D-315). The meta table holds it under this key. Each new
// workout carries it, and the server refuses a workout of an older
// generation.
export const HISTORY_GENERATION_KEY = "history-generation";

// historyGeneration gives the stored generation, or 0 when the phone has
// none.
export async function historyGeneration(store: WorkoutAppDB): Promise<number> {
  const m = await store.meta.get(HISTORY_GENERATION_KEY);
  return typeof m?.value === "number" ? m.value : 0;
}

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
  // The generation of the history at the start (D-315). A workout of an
  // older app has none, and it counts as 0.
  historyGeneration?: number;
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
  outbox!: EntityTable<OutboxEntry, "opId">;
  refused!: EntityTable<RefusedEntry, "opId">;
  copies!: EntityTable<Copy, "key">;
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
    // Version 3 adds the refused entries and the offline copies of the
    // sync (work area 6.3). It removes the settings of the skeleton, and
    // their outbox entries, because the server knows no such entity.
    this.version(3)
      .stores({ settings: null, refused: "opId, at", copies: "key" })
      .upgrade((tx) =>
        tx
          .table("outbox")
          .filter((e: OutboxEntry) => e.entity === "setting")
          .delete(),
      );
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

// pendingOutbox gives the outbox in the order of the op ids, which is the
// order of the sync (D-275).
export function pendingOutbox(store: WorkoutAppDB): Promise<OutboxEntry[]> {
  return withReopen(store, () => store.outbox.toArray());
}
