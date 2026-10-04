import { create, toJson } from "@bufbuild/protobuf";

import {
  CardioEntrySchema,
  SetEntrySchema,
  WorkoutHeaderSchema,
} from "../gen/workoutapp/v1/workout_service_pb";
import type { Plan } from "../gen/workoutapp/v1/plan_service_pb";
import {
  OUTBOX_SCHEMA_VERSION,
  REST_KEY,
  withReopen,
  type CardioRecord,
  type OutboxEntry,
  type RestTimer,
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

// The reps in reserve that the owner taps for a working set (D-249).
// "4+" logs 4.
export const RIR_CHOICES: readonly { label: string; value: number }[] = [
  { label: "0", value: 0 },
  { label: "1", value: 1 },
  { label: "2", value: 2 },
  { label: "3", value: 3 },
  { label: "4+", value: 4 },
];

// The reps in reserve of a calibration set, so the owner can give each
// result of the calibration table (D-268). "6+" logs 6.
export const CALIBRATION_RIR_CHOICES: readonly { label: string; value: number }[] = [
  { label: "0", value: 0 },
  { label: "1", value: 1 },
  { label: "2", value: 2 },
  { label: "3", value: 3 },
  { label: "4", value: 4 },
  { label: "5", value: 5 },
  { label: "6+", value: 6 },
];

// rirChoices gives the reps in reserve of a kind of set.
export function rirChoices(kind: "working" | "calibration"): readonly { label: string; value: number }[] {
  return kind === "calibration" ? CALIBRATION_RIR_CHOICES : RIR_CHOICES;
}

// rirText gives the label of a logged reps in reserve: "4+" for a working
// set at 4 or more, and "6+" for a calibration set at 6 or more.
export function rirText(kind: "working" | "calibration", rir: number): string {
  const top = kind === "calibration" ? 6 : 4;
  return rir >= top ? `${top}+` : String(rir);
}

// PREVIEW_SECONDS is the time of the preview of the next machine before
// the automatic advance (D-269).
export const PREVIEW_SECONDS = 10;

// REST_STEP_SECONDS is the change of one tap of "-15 s" or "+15 s"
// (D-270).
export const REST_STEP_SECONDS = 15;

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

// headerPayload gives the whole header of a workout. It holds the target
// of each exercise that the owner saw at the start (D-291), so the
// server reads the history of an exercise after a new plan or a revision
// too.
function headerPayload(w: WorkoutRecord): unknown {
  const sets = (list: TargetSet[]) => list.map((s) => ({ reps: s.reps, loadTenthLb: s.loadTenthLb, rirTarget: s.rirTarget }));
  return toJson(
    WorkoutHeaderSchema,
    create(WorkoutHeaderSchema, {
      date: w.date,
      plan: { planCreatedAt: w.planCreatedAt, sessionIndex: w.sessionIndex },
      skippedExerciseIds: w.skippedExerciseIds,
      endedEarly: w.endedEarly,
      finished: w.finished,
      targets: w.exercises.map((e) => ({
        exerciseId: e.exerciseId,
        restSeconds: e.restSeconds,
        calibrationSets: sets(e.calibrationSets),
        workingSets: sets(e.workingSets),
        recommendedWorkingSets: e.override ? sets(e.override.recommendedWorkingSets) : [],
        overrideReason: e.override?.reason ?? "",
        firstSetCalibration: e.firstSetCalibration ?? false,
      })),
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
// targets. An exercise with an override of the owner gets the sets of the
// override, and keeps the recommendation and the reason (D-69, D-293).
// An expired override gives the recommendation.
export function workoutExercises(plan: Plan, sessionIndex: number, weights: (exerciseId: string) => number[]): WorkoutExercise[] {
  const session = plan.sessions[sessionIndex];
  if (!session) throw new RangeError(`the plan has no session ${sessionIndex}`);
  return session.exercises.map((e) => {
    const sets = (list: typeof e.workingSets): TargetSet[] =>
      list.map((s) => ({ reps: s.reps, loadTenthLb: s.loadTenthLb, rirTarget: s.rirTarget }));
    // An expired override no longer has the check of the policy for the
    // date, so the recommendation applies (D-294, D-295).
    const o = e.override && !e.override.expired ? e.override : undefined;
    const calibrationSets = sets(o ? o.calibrationSets : e.calibrationSets);
    const workingSets = sets(o ? o.workingSets : e.workingSets);
    let list = [...new Set(weights(e.exerciseId))].filter((w) => w > 0).sort((a, b) => a - b);
    if (list.length === 0) {
      list = [...new Set([...calibrationSets, ...workingSets].map((s) => s.loadTenthLb))].sort((a, b) => a - b);
    }
    const out: WorkoutExercise = { exerciseId: e.exerciseId, name: e.name, restSeconds: e.restSeconds, calibrationSets, workingSets, weights: list };
    if (o) out.override = { reason: o.reason, recommendedWorkingSets: sets(e.workingSets) };
    if (e.firstSetCalibration) out.firstSetCalibration = true;
    if (calibrationSets.length > 0 && e.calibrationLoads.length > 0) {
      out.calibrationLoads = e.calibrationLoads.map((c) => ({
        weight: c.weightTenthLb,
        down: c.downTenthLb,
        keep: c.keepTenthLb,
        upOne: c.upOneTenthLb,
        upTwo: c.upTwoTenthLb,
      }));
    }
    return out;
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

// skipExercise records the skip of an exercise (D-63, D-170): the new
// header and its outbox entry go in one transaction. The planned sets
// with no log are skipped work. A later log of a set on the exercise
// removes the skip, because the server reads a logged set first.
export function skipExercise(store: WorkoutAppDB, workoutId: string, exerciseId: string, now: Date = new Date()): Promise<WorkoutRecord> {
  return withReopen(store, () =>
    store.transaction("rw", store.workouts, store.outbox, async () => {
      const w = await openWorkout(store, workoutId);
      if (!w.exercises.some((e) => e.exerciseId === exerciseId)) throw new RangeError(`the workout has no exercise ${exerciseId}`);
      if (w.skippedExerciseIds.includes(exerciseId)) return w;
      const next: WorkoutRecord = { ...w, skippedExerciseIds: [...w.skippedExerciseIds, exerciseId] };
      await store.workouts.put(next);
      await store.outbox.add(entry("workout", w.id, w.version, headerPayload(next), now.toISOString(), now));
      return next;
    }),
  );
}

export type SetChange = { reps: number; weightTenthsLb: number; rir: number; pain?: number; note?: string };

// editSet changes a logged set of an open workout (D-63). The set keeps
// its id, its kind, and its time, so it keeps its place in the order of
// the sets. The new state and its outbox entry go in one transaction.
export function editSet(store: WorkoutAppDB, setId: string, change: SetChange, now: Date = new Date()): Promise<SetRecord> {
  return withReopen(store, () =>
    store.transaction("rw", store.workouts, store.sets, store.outbox, async () => {
      const old = await store.sets.get(setId);
      if (!old) throw new WorkoutClosedError();
      await openWorkout(store, old.workoutId);
      const s: SetRecord = {
        ...old,
        reps: change.reps,
        weightTenthsLb: change.weightTenthsLb,
        rir: change.rir,
        note: (change.note ?? "").trim(),
      };
      delete s.pain;
      if (change.pain !== undefined) s.pain = change.pain;
      await store.sets.put(s);
      await store.outbox.add(entry("set", s.id, s.version, setPayload(s), now.toISOString(), now));
      return s;
    }),
  );
}

// startRest starts the rest timer of a workout (D-59): it ends seconds
// after now. A new timer replaces the old one.
export function startRest(store: WorkoutAppDB, workoutId: string, seconds: number, now: Date = new Date()): Promise<void> {
  const timer: RestTimer = { workoutId, endsAt: now.getTime() + Math.max(0, seconds) * 1000 };
  return withReopen(store, async () => {
    await store.meta.put({ key: REST_KEY, value: timer });
  });
}

// restTimer gives the rest timer of a workout, or null.
export function restTimer(store: WorkoutAppDB, workoutId: string): Promise<RestTimer | null> {
  return withReopen(store, async () => {
    const t = (await store.meta.get(REST_KEY))?.value as RestTimer | undefined;
    return t && t.workoutId === workoutId ? t : null;
  });
}

// adjustRest moves the end of the rest timer by seconds (D-270). The bound
// of D-172 applies to the target alone, so the owner can make the rest
// longer than 180 seconds. The remaining time never goes below 0.
export function adjustRest(store: WorkoutAppDB, workoutId: string, seconds: number, now: Date = new Date()): Promise<void> {
  return withReopen(store, () =>
    store.transaction("rw", store.meta, async () => {
      const t = (await store.meta.get(REST_KEY))?.value as RestTimer | undefined;
      if (!t || t.workoutId !== workoutId) return;
      const endsAt = Math.max(now.getTime(), Math.max(t.endsAt, now.getTime()) + seconds * 1000);
      await store.meta.put({ key: REST_KEY, value: { workoutId, endsAt } satisfies RestTimer });
    }),
  );
}

// dismissRest stops the rest timer (D-59).
export function dismissRest(store: WorkoutAppDB): Promise<void> {
  return withReopen(store, () => store.meta.delete(REST_KEY));
}

// restLeft gives the whole seconds of rest that remain at now, 0 or
// more. It reads the stored end time, so it is correct after a screen
// lock (D-270).
export function restLeft(t: RestTimer, now: number): number {
  return Math.max(0, Math.ceil((t.endsAt - now) / 1000));
}

// clockText gives a time in seconds as "m:ss".
export function clockText(seconds: number): string {
  const s = Math.max(0, Math.floor(seconds));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
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

// openExercises gives each exercise that the owner did not skip and
// that has a planned set with no log. "Finish now" ends these early
// (D-63).
export function openExercises(w: WorkoutRecord, logged: readonly SetRecord[]): WorkoutExercise[] {
  return w.exercises.filter((e) => !w.skippedExerciseIds.includes(e.exerciseId) && nextSet(e, logged) !== null);
}

// finishWorkout ends a workout (D-63). Each exercise with no logged set
// is skipped. The workout ended early when an exercise that the owner did
// not skip has a planned set with no log, as with "finish now". The new
// header and its outbox entry go in one transaction, and the rest timer
// stops.
export function finishWorkout(store: WorkoutAppDB, workoutId: string, now: Date = new Date()): Promise<WorkoutRecord> {
  return withReopen(store, () =>
    store.transaction("rw", [store.workouts, store.sets, store.outbox, store.meta], async () => {
      const w = await openWorkout(store, workoutId);
      const sets = await store.sets.where("workoutId").equals(workoutId).toArray();
      const logged = new Set(sets.map((s) => s.exerciseId));
      const skipped = w.exercises.map((e) => e.exerciseId).filter((id) => !logged.has(id));
      const endedEarly = openExercises(w, sets).length > 0;
      const done: WorkoutRecord = { ...w, skippedExerciseIds: skipped, endedEarly, finished: true };
      await store.workouts.put(done);
      await store.outbox.add(entry("workout", w.id, w.version, headerPayload(done), now.toISOString(), now));
      await store.meta.delete(REST_KEY);
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
// fromCalibration is true when the calibration gave the load of the
// working set (D-267, D-299). calibrates is true for the first working
// set of the first-set calibration (D-297).
export type NextSet = {
  kind: "working" | "calibration";
  number: number;
  of: number;
  target: TargetSet;
  fromCalibration: boolean;
  calibrates: boolean;
};

// calibrationLoad gives the load of the working sets after a logged
// calibration set, from the loads of the policy (D-150, D-267): 2 or
// fewer reps in reserve or a pain rating of 1 or more go down, 3 or 4
// keep the load, 5 goes up one step, and 6 or more go up two steps. The
// table applies to the weight that the owner logged, because the owner
// can change the weight before the log (D-249). It gives null when the
// plan has no row for that weight: a plan of policy version 3, or a
// weight that the machine did not have at the time of the plan. The
// policy code of `go/internal/policy/calibrate.go` holds the same table.
export function calibrationLoad(e: WorkoutExercise, s: Pick<SetRecord, "weightTenthsLb" | "rir" | "pain">): number | null {
  const c = e.calibrationLoads?.find((r) => r.weight === s.weightTenthsLb);
  if (!c) return null;
  if ((s.pain !== undefined && s.pain >= 1) || s.rir <= 2) return c.down;
  if (s.rir >= 6) return c.upTwo;
  if (s.rir === 5) return c.upOne;
  return c.keep;
}

// nextSet gives the next planned set of an exercise, or null when each
// planned set has a log. After the calibration set of a plan of policy
// version 6 or earlier, each working set gets the load of the
// calibration table (D-267). With the first-set calibration, each later
// working set gets the weight that the owner logged for the first set
// (D-299).
export function nextSet(e: WorkoutExercise, logged: readonly SetRecord[]): NextSet | null {
  const mine = logged.filter((s) => s.exerciseId === e.exerciseId);
  const cals = mine.filter((s) => s.kind === "calibration");
  if (cals.length < e.calibrationSets.length) {
    const n = cals.length;
    return { kind: "calibration", number: n + 1, of: e.calibrationSets.length, target: e.calibrationSets[n], fromCalibration: false, calibrates: false };
  }
  const work = mine.filter((s) => s.kind === "working");
  if (work.length < e.workingSets.length) {
    const n = work.length;
    const set = { kind: "working" as const, number: n + 1, of: e.workingSets.length, target: e.workingSets[n] };
    if (e.firstSetCalibration) {
      if (n === 0) return { ...set, fromCalibration: false, calibrates: true };
      return { ...set, target: { ...set.target, loadTenthLb: work[0].weightTenthsLb }, fromCalibration: true, calibrates: false };
    }
    const load = cals.length > 0 ? calibrationLoad(e, cals[0]) : null;
    if (load === null) return { ...set, fromCalibration: false, calibrates: false };
    return { ...set, target: { ...set.target, loadTenthLb: load }, fromCalibration: true, calibrates: false };
  }
  return null;
}

// currentExercise gives the first exercise that the owner did not skip
// and that has a planned set with no log, or null when none remains.
export function currentExercise(w: WorkoutRecord, logged: readonly SetRecord[]): WorkoutExercise | null {
  return openExercises(w, logged)[0] ?? null;
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

// MAX_INT32 is the largest value of an `int32` field of the contract.
export const MAX_INT32 = 2_147_483_647;

// MAX_CARDIO_MINUTES is the longest cardio log, so that its duration in
// seconds fits the `int32` field `duration_seconds`.
export const MAX_CARDIO_MINUTES = Math.floor(MAX_INT32 / 60);

// parseMiles reads a distance such as "2.5" as tenths of a mile (D-165).
// An empty text gives undefined. A text that is not 0 or more with one
// decimal place or none, or that does not fit the `int32` field, gives
// null.
export function parseMiles(text: string): number | undefined | null {
  const t = text.trim();
  if (t === "") return undefined;
  const m = /^(\d+)(?:\.(\d))?$/.exec(t);
  if (!m) return null;
  const tenths = BigInt(m[1]) * 10n + BigInt(m[2] ?? "0");
  return tenths <= BigInt(MAX_INT32) ? Number(tenths) : null;
}

// stepMinutes gives the minutes of a cardio log after a tap of plus or
// minus: 1 or more (D-123), and MAX_CARDIO_MINUTES or fewer.
export function stepMinutes(current: number, dir: 1 | -1): number {
  return Math.min(MAX_CARDIO_MINUTES, Math.max(1, current + dir));
}

// parseLevel reads the resistance level, a whole number of 0 or more. An
// empty text gives undefined, and a bad text gives null.
export function parseLevel(text: string): number | undefined | null {
  const t = text.trim();
  if (t === "") return undefined;
  if (!/^\d+$/.test(t)) return null;
  const n = Number(t);
  return Number.isSafeInteger(n) && n <= MAX_INT32 ? n : null;
}

// loggedText gives one logged set, such as "8 reps at 20 lb, 2 in reserve".
export function loggedText(s: Pick<SetRecord, "kind" | "reps" | "weightTenthsLb" | "rir" | "pain">): string {
  const rir = rirText(s.kind, s.rir);
  const kind = s.kind === "calibration" ? "Calibration: " : "";
  const pain = s.pain !== undefined ? `, pain ${s.pain}` : "";
  return `${kind}${s.reps} ${s.reps === 1 ? "rep" : "reps"} at ${formatPounds(s.weightTenthsLb)} lb, ${rir} in reserve${pain}`;
}
