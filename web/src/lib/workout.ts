import { create, toJson } from "@bufbuild/protobuf";

import {
  CardioEntrySchema,
  SetEntrySchema,
  WorkoutHeaderSchema,
} from "../gen/workoutapp/v1/workout_service_pb";
import type { Plan } from "../gen/workoutapp/v1/plan_service_pb";
import {
  OUTBOX_SCHEMA_VERSION,
  withReopen,
  type CardioRecord,
  type OutboxEntry,
  type SetRecord,
  type TargetSet,
  type WorkoutAppDB,
  type WorkoutExercise,
  type WorkoutRecord,
} from "./db";
import { formatPounds } from "./inventory";
import { localDate } from "./plan";
import { nextId } from "./uuidv7";

// The workout log of the phone (work area 6.1). Each change writes the
// entity and its outbox entry in one Dexie transaction (D-132). The
// payload of an entry is the JSON form of the message of
// `proto/workoutapp/v1/workout_service.proto` that holds the whole new
// state of the entity, so the sync of PR-32 sends it with no change.

// MAX_NOTE_CHARS is the limit of the note of a set or of a cardio log
// (D-261).
export const MAX_NOTE_CHARS = 280;

// The reps in reserve that the owner taps (D-249). "4+" logs 4.
export const RIR_CHOICES: readonly { label: string; value: number }[] = [
  { label: "0", value: 0 },
  { label: "1", value: 1 },
  { label: "2", value: 2 },
  { label: "3", value: 3 },
  { label: "4+", value: 4 },
];

// WorkoutInProgressError: a workout is open, so the phone starts no other
// one (D-252).
export class WorkoutInProgressError extends Error {
  constructor() {
    super("a workout is in progress");
    this.name = "WorkoutInProgressError";
  }
}

// WorkoutClosedError: the workout does not exist, or it ended, so it
// takes no log.
export class WorkoutClosedError extends Error {
  constructor() {
    super("the workout is not open");
    this.name = "WorkoutClosedError";
  }
}

function entry(entity: string, entityId: string, baseVersion: number, payload: unknown, at: string, now: Date): OutboxEntry {
  return {
    opId: nextId(now.getTime()),
    entity,
    entityId,
    baseVersion,
    payload,
    at,
    attempts: 0,
    schemaVersion: OUTBOX_SCHEMA_VERSION,
  };
}

function headerPayload(w: WorkoutRecord): unknown {
  return toJson(
    WorkoutHeaderSchema,
    create(WorkoutHeaderSchema, {
      date: w.date,
      plan: { planCreatedAt: w.planCreatedAt, sessionIndex: w.sessionIndex },
      skippedExerciseIds: w.skippedExerciseIds,
      endedEarly: w.endedEarly,
      finished: w.finished,
    }),
  );
}

function setPayload(s: SetRecord): unknown {
  return toJson(
    SetEntrySchema,
    create(SetEntrySchema, {
      workoutId: s.workoutId,
      exerciseId: s.exerciseId,
      kind: s.kind,
      reps: s.reps,
      weightTenthsLb: BigInt(s.weightTenthsLb),
      rir: s.rir,
      pain: s.pain,
      note: s.note,
    }),
  );
}

function cardioPayload(c: CardioRecord): unknown {
  return toJson(
    CardioEntrySchema,
    create(CardioEntrySchema, {
      workoutId: c.workoutId,
      exerciseId: c.exerciseId,
      durationSeconds: c.durationSeconds,
      effort: c.effort,
      distanceTenthsMi: c.distanceTenthsMi,
      resistance: c.resistance,
      pain: c.pain,
      note: c.note,
    }),
  );
}

// activeWorkout gives the open workout, or undefined. The phone holds one
// open workout at most.
export function activeWorkout(store: WorkoutAppDB): Promise<WorkoutRecord | undefined> {
  return withReopen(store, () => store.workouts.filter((w) => !w.finished).first());
}

// doneSessions gives the index of each session of a plan that a finished
// workout started from (D-248).
export function doneSessions(store: WorkoutAppDB, planCreatedAt: string): Promise<number[]> {
  return withReopen(store, async () =>
    (await store.workouts.filter((w) => w.finished && w.planCreatedAt === planCreatedAt).toArray()).map((w) => w.sessionIndex),
  );
}

// nextSession gives the next session of a plan that the owner did not do
// yet (D-248): the first session with the fewest finished workouts. So
// after the last session of the week, the first session comes again.
export function nextSession(count: number, done: readonly number[]): number {
  let best = 0;
  let least = Infinity;
  for (let i = 0; i < count; i++) {
    const n = done.filter((d) => d === i).length;
    if (n < least) {
      least = n;
      best = i;
    }
  }
  return best;
}

// workoutExercises copies the exercises of a plan session, with the
// weights of each machine (D-264). An exercise with no known weights gets
// the loads of its targets, so its buttons still step between real
// targets.
export function workoutExercises(plan: Plan, sessionIndex: number, weights: (exerciseId: string) => number[]): WorkoutExercise[] {
  const session = plan.sessions[sessionIndex];
  if (!session) throw new RangeError(`the plan has no session ${sessionIndex}`);
  return session.exercises.map((e) => {
    const sets = (list: typeof e.workingSets): TargetSet[] =>
      list.map((s) => ({ reps: s.reps, loadTenthLb: s.loadTenthLb, rirTarget: s.rirTarget }));
    const calibrationSets = sets(e.calibrationSets);
    const workingSets = sets(e.workingSets);
    let list = [...new Set(weights(e.exerciseId))].filter((w) => w > 0).sort((a, b) => a - b);
    if (list.length === 0) {
      list = [...new Set([...calibrationSets, ...workingSets].map((s) => s.loadTenthLb))].sort((a, b) => a - b);
    }
    return { exerciseId: e.exerciseId, name: e.name, restSeconds: e.restSeconds, calibrationSets, workingSets, weights: list };
  });
}

// startWorkout starts a workout from a session of the plan (D-248). It
// writes the workout and the outbox entry of its header in one
// transaction. It refuses a second open workout.
export async function startWorkout(
  store: WorkoutAppDB,
  plan: Plan,
  sessionIndex: number,
  weights: (exerciseId: string) => number[],
  now: Date = new Date(),
): Promise<WorkoutRecord> {
  const exercises = workoutExercises(plan, sessionIndex, weights);
  const session = plan.sessions[sessionIndex];
  return withReopen(store, () =>
    store.transaction("rw", store.workouts, store.outbox, async () => {
      if (await store.workouts.filter((w) => !w.finished).first()) throw new WorkoutInProgressError();
      const at = now.toISOString();
      const w: WorkoutRecord = {
        id: nextId(now.getTime()),
        date: localDate(now),
        planCreatedAt: plan.createdAt,
        sessionIndex,
        title: session.title,
        exercises,
        cardio: session.cardio
          ? { exerciseId: session.cardio.exerciseId, name: session.cardio.name, minutes: session.cardio.minutes }
          : null,
        skippedExerciseIds: [],
        endedEarly: false,
        finished: false,
        startedAt: at,
        version: 0,
      };
      await store.workouts.add(w);
      await store.outbox.add(entry("workout", w.id, 0, headerPayload(w), at, now));
      return w;
    }),
  );
}

async function openWorkout(store: WorkoutAppDB, workoutId: string): Promise<WorkoutRecord> {
  const w = await store.workouts.get(workoutId);
  if (!w || w.finished) throw new WorkoutClosedError();
  return w;
}

export type SetInput = {
  exerciseId: string;
  kind: "working" | "calibration";
  reps: number;
  weightTenthsLb: number;
  rir: number;
  pain?: number;
  note?: string;
};

// logSet writes a set and its outbox entry in one transaction (D-249).
export function logSet(store: WorkoutAppDB, workoutId: string, input: SetInput, now: Date = new Date()): Promise<SetRecord> {
  return withReopen(store, () =>
    store.transaction("rw", store.workouts, store.sets, store.outbox, async () => {
      await openWorkout(store, workoutId);
      const at = now.toISOString();
      const s: SetRecord = {
        id: nextId(now.getTime()),
        workoutId,
        exerciseId: input.exerciseId,
        kind: input.kind,
        reps: input.reps,
        weightTenthsLb: input.weightTenthsLb,
        rir: input.rir,
        note: (input.note ?? "").trim(),
        at,
        version: 0,
      };
      if (input.pain !== undefined) s.pain = input.pain;
      await store.sets.add(s);
      await store.outbox.add(entry("set", s.id, 0, setPayload(s), at, now));
      return s;
    }),
  );
}

export type CardioInput = {
  exerciseId: string;
  durationSeconds: number;
  effort: number;
  distanceTenthsMi?: number;
  resistance?: number;
  pain?: number;
  note?: string;
};

// logCardio writes a cardio log and its outbox entry in one transaction
// (D-123).
export function logCardio(store: WorkoutAppDB, workoutId: string, input: CardioInput, now: Date = new Date()): Promise<CardioRecord> {
  return withReopen(store, () =>
    store.transaction("rw", store.workouts, store.cardio, store.outbox, async () => {
      await openWorkout(store, workoutId);
      const at = now.toISOString();
      const c: CardioRecord = {
        id: nextId(now.getTime()),
        workoutId,
        exerciseId: input.exerciseId,
        durationSeconds: input.durationSeconds,
        effort: input.effort,
        note: (input.note ?? "").trim(),
        at,
        version: 0,
      };
      if (input.distanceTenthsMi !== undefined) c.distanceTenthsMi = input.distanceTenthsMi;
      if (input.resistance !== undefined) c.resistance = input.resistance;
      if (input.pain !== undefined) c.pain = input.pain;
      await store.cardio.add(c);
      await store.outbox.add(entry("cardio", c.id, 0, cardioPayload(c), at, now));
      return c;
    }),
  );
}

// finishWorkout ends a workout. Each exercise with no logged set is
// skipped, and then the workout ended early, as with "finish now" (D-63).
// The new header and its outbox entry go in one transaction.
export function finishWorkout(store: WorkoutAppDB, workoutId: string, now: Date = new Date()): Promise<WorkoutRecord> {
  return withReopen(store, () =>
    store.transaction("rw", store.workouts, store.sets, store.outbox, async () => {
      const w = await openWorkout(store, workoutId);
      const logged = new Set((await store.sets.where("workoutId").equals(workoutId).toArray()).map((s) => s.exerciseId));
      const skipped = w.exercises.map((e) => e.exerciseId).filter((id) => !logged.has(id));
      const done: WorkoutRecord = { ...w, skippedExerciseIds: skipped, endedEarly: skipped.length > 0, finished: true };
      await store.workouts.put(done);
      await store.outbox.add(entry("workout", w.id, w.version, headerPayload(done), now.toISOString(), now));
      return done;
    }),
  );
}

// workoutSets gives the sets of a workout in the order of their time.
export function workoutSets(store: WorkoutAppDB, workoutId: string): Promise<SetRecord[]> {
  return withReopen(store, () => store.sets.where("workoutId").equals(workoutId).sortBy("at"));
}

// workoutCardio gives the cardio logs of a workout in the order of their
// time.
export function workoutCardio(store: WorkoutAppDB, workoutId: string): Promise<CardioRecord[]> {
  return withReopen(store, () => store.cardio.where("workoutId").equals(workoutId).sortBy("at"));
}

// The next set of an exercise: the calibration sets first, then the
// working sets (D-150). `number` counts from 1 inside its kind.
export type NextSet = { kind: "working" | "calibration"; number: number; of: number; target: TargetSet };

// nextSet gives the next planned set of an exercise, or null when each
// planned set has a log.
export function nextSet(e: WorkoutExercise, logged: readonly SetRecord[]): NextSet | null {
  const mine = logged.filter((s) => s.exerciseId === e.exerciseId);
  const cal = mine.filter((s) => s.kind === "calibration").length;
  if (cal < e.calibrationSets.length) {
    return { kind: "calibration", number: cal + 1, of: e.calibrationSets.length, target: e.calibrationSets[cal] };
  }
  const work = mine.filter((s) => s.kind === "working").length;
  if (work < e.workingSets.length) {
    return { kind: "working", number: work + 1, of: e.workingSets.length, target: e.workingSets[work] };
  }
  return null;
}

// currentExercise gives the first exercise with a planned set that has
// no log, or null when each planned set has a log.
export function currentExercise(w: WorkoutRecord, logged: readonly SetRecord[]): WorkoutExercise | null {
  return w.exercises.find((e) => nextSet(e, logged) !== null) ?? null;
}

// stepWeight gives the next heavier (+1) or lighter (-1) weight of the
// list of the machine (D-264). A weight that is not on the list moves to
// the nearest weight of the list in that direction. At an end of the
// list, the weight stays.
export function stepWeight(list: readonly number[], current: number, dir: 1 | -1): number {
  if (dir > 0) return list.find((w) => w > current) ?? current;
  return [...list].reverse().find((w) => w < current) ?? current;
}

// stepReps gives the reps after a tap of plus or minus. Reps are 0 or
// more (D-164).
export function stepReps(current: number, dir: 1 | -1): number {
  return Math.max(0, current + dir);
}

// noteLength counts the characters of a note after the trim, as the
// server does (D-261).
export function noteLength(note: string): number {
  return [...note.trim()].length;
}

// parseMiles reads a distance such as "2.5" as tenths of a mile (D-165).
// An empty text gives undefined. A text that is not 0 or more with one
// decimal place or none gives null.
export function parseMiles(text: string): number | undefined | null {
  const t = text.trim();
  if (t === "") return undefined;
  if (!/^\d+(\.\d)?$/.test(t)) return null;
  return Math.round(Number(t) * 10);
}

// parseLevel reads the resistance level, a whole number of 0 or more. An
// empty text gives undefined, and a bad text gives null.
export function parseLevel(text: string): number | undefined | null {
  const t = text.trim();
  if (t === "") return undefined;
  if (!/^\d+$/.test(t)) return null;
  const n = Number(t);
  return Number.isSafeInteger(n) && n <= 2_147_483_647 ? n : null;
}

// loggedText gives one logged set, such as "8 reps at 20 lb, 2 in reserve".
export function loggedText(s: Pick<SetRecord, "kind" | "reps" | "weightTenthsLb" | "rir" | "pain">): string {
  const rir = s.rir >= 4 ? "4+" : String(s.rir);
  const kind = s.kind === "calibration" ? "Calibration: " : "";
  const pain = s.pain !== undefined ? `, pain ${s.pain}` : "";
  return `${kind}${s.reps} ${s.reps === 1 ? "rep" : "reps"} at ${formatPounds(s.weightTenthsLb)} lb, ${rir} in reserve${pain}`;
}
