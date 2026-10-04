import { fromJson, type JsonValue } from "@bufbuild/protobuf";
import { useLiveQuery } from "dexie-react-hooks";
import { useEffect, useId, useState, type ReactNode } from "react";

import { GetPlanResponseSchema, type Plan } from "../gen/workoutapp/v1/plan_service_pb";
import { db, withReopen, type CardioRecord, type SetRecord, type WorkoutExercise, type WorkoutRecord } from "../lib/db";
import { available, formatPounds } from "../lib/inventory";
import { useOfflineInventory } from "../lib/inventory-api";
import { restText, setText } from "../lib/plan";
import { PAIN_WARNING, SYMPTOMS, symptomWarning } from "../lib/symptoms";
import { noCopyText } from "../lib/sync";
import { engine, useSyncWhenMissing } from "../lib/sync-engine";
import type { WakeStatus } from "../lib/wake-lock";
import {
  activeWorkout,
  adjustRest,
  clockText,
  currentExercise,
  dismissRest,
  doneSessions,
  editSet,
  finishWorkout,
  logCardio,
  loggedText,
  logSet,
  MAX_NOTE_CHARS,
  nextSession,
  nextSet,
  noteLength,
  openExercises,
  parseLevel,
  parseMiles,
  PREVIEW_SECONDS,
  REST_STEP_SECONDS,
  restLeft,
  restTimer,
  rirChoices,
  skipExercise,
  startRest,
  startWorkout,
  stepMinutes,
  stepReps,
  stepWeight,
  workoutCardio,
  workoutSets,
  type NextSet,
} from "../lib/workout";
import { danger, ErrorText, field, primary, secondary, Title } from "./inventory/ui";

// The workout screen of work area 6.1. The owner starts a workout from the
// next session of the plan, or from another session (D-248). The phone
// keeps each log first, with its outbox entry (D-62, D-132), so a workout
// needs no network after its start. The screen uses large targets and few
// taps (D-71), and each cue is visual alone (D-58).
export function WorkoutPage({ onBack, wake }: { onBack: () => void; wake: WakeStatus }) {
  // null while the store loads, and undefined when no workout is open.
  const active = useLiveQuery(() => activeWorkout(db), [], null);
  if (active === null) return <p className="text-slate-400">Loading…</p>;
  if (active) return <ActiveWorkout key={active.id} workout={active} wake={wake} onBack={onBack} />;
  return <StartWorkout onBack={onBack} />;
}

// StartWorkout reads the offline copies of the plan, the catalog, and the
// inventory, so a workout starts with no connection (D-278). Then it
// starts the next session, or the session that the owner picks. With no
// copy yet, it asks the sync for one.
function StartWorkout({ onBack }: { onBack: () => void }) {
  const planCopy = useLiveQuery(() => withReopen(db, async () => (await db.copies.get("plan")) ?? null), [], undefined);
  const inventory = useOfflineInventory();
  const sync = useSyncWhenMissing(planCopy === null || inventory === null);
  const plan = planCopy ? fromJson(GetPlanResponseSchema, planCopy.json as JsonValue, { ignoreUnknownFields: true }) : null;
  const createdAt = plan?.plan?.createdAt ?? "";
  const done = useLiveQuery(() => (createdAt ? doneSessions(db, createdAt) : Promise.resolve([])), [createdAt], null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  if (planCopy === undefined || inventory === undefined || done === null) return <p className="text-slate-400">Loading…</p>;
  if (!plan || !inventory) {
    if (sync.running || !sync.error) return <p className="text-slate-400">Loading…</p>;
    return (
      <div className="space-y-4">
        <Title onBack={onBack}>Workout</Title>
        <ErrorText testId="load-error">{noCopyText(sync.error)}</ErrorText>
        <button type="button" className={secondary} onClick={() => void engine.syncNow()}>
          Try again
        </button>
      </div>
    );
  }

  const p = plan.plan;
  if (!p || p.sessions.length === 0) {
    return (
      <div className="space-y-4">
        <Title onBack={onBack}>Workout</Title>
        <p className="text-slate-300" data-testid="no-plan">
          You have no plan yet. Make a plan on the plan screen first.
        </p>
      </div>
    );
  }

  const machineOf = new Map(inventory.catalog.exercises.map((e) => [e.id, e.machineId]));
  const machines = new Map(inventory.inventory.machines.map((m) => [m.machineId, m]));
  const weights = (exerciseId: string) => {
    const m = machines.get(machineOf.get(exerciseId) ?? "");
    return m ? available(m) : [];
  };

  const start = async (index: number) => {
    setBusy(true);
    setError("");
    try {
      await startWorkout(db, p, index, weights);
    } catch {
      // The live query shows a workout that is already open. Another
      // failure is a fault of the store of the phone.
      setError("The phone did not start the workout. Try again.");
      setBusy(false);
    }
  };

  const next = nextSession(p.sessions.length, done);
  const others = p.sessions.map((s, i) => ({ s, i })).filter(({ i }) => i !== next);

  return (
    <div className="space-y-6">
      <Title onBack={onBack}>Workout</Title>
      <ErrorText testId="start-error">{error}</ErrorText>

      <section className="space-y-3 rounded-lg border border-slate-800 p-3" data-testid="next-session">
        <p className="text-sm text-slate-400">Next session</p>
        <SessionSummary plan={p} index={next} />
        <button type="button" className={`${primary} min-h-14 w-full text-lg`} disabled={busy} onClick={() => void start(next)}>
          {`Start ${p.sessions[next].title}`}
        </button>
      </section>

      {others.length > 0 && (
        <section className="space-y-3" aria-label="Other sessions">
          <h3 className="text-base font-semibold text-slate-100">Other sessions</h3>
          {others.map(({ s, i }) => (
            <button key={i} type="button" className={`${secondary} w-full`} disabled={busy} onClick={() => void start(i)}>
              {`Start ${s.title}`}
            </button>
          ))}
        </section>
      )}
    </div>
  );
}

function SessionSummary({ plan, index }: { plan: Plan; index: number }) {
  const s = plan.sessions[index];
  return (
    <div className="space-y-1">
      <h3 className="text-base font-semibold text-slate-100">{s.title}</h3>
      <ul className="text-sm text-slate-300">
        {s.exercises.map((e) => (
          <li key={e.exerciseId}>{e.name}</li>
        ))}
        {s.cardio && <li>{`${s.cardio.name}, ${s.cardio.minutes} min`}</li>}
      </ul>
    </div>
  );
}

// The open dialog of the workout: the list of symptoms, a warning, the
// confirmation of "finish now" or of a skip, or the edit of a logged set.
type Dialog =
  | { kind: "symptoms" }
  | { kind: "warning"; text: string }
  | { kind: "finish"; open: number; unstarted: number }
  | { kind: "skip"; exercise: WorkoutExercise }
  | { kind: "edit"; set: SetRecord; exercise: WorkoutExercise };

// The preview of the next machine after the last set of an exercise
// (D-60, D-269). until is the end of the preview in milliseconds.
type Preview = { done: string; next: string; until: number };

// useNow gives the time in milliseconds. While `on` is true, it renders
// again 4 times each second and at each return to the front, so a timer
// reads the stored end time again after a screen lock (D-270).
function useNow(on: boolean): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!on) return;
    const tick = () => setNow(Date.now());
    tick();
    const t = setInterval(tick, 250);
    document.addEventListener("visibilitychange", tick);
    return () => {
      clearInterval(t);
      document.removeEventListener("visibilitychange", tick);
    };
  }, [on]);
  return now;
}

// ActiveWorkout is the open workout. It shows a notice when the phone
// refuses the screen wake lock of the app (D-265, D-283). A logged set
// starts the rest timer (D-59). After the last set of an exercise, the
// screen shows the next machine for PREVIEW_SECONDS, then advances
// (D-60, D-269). Each cue is visual alone, and the app sends no
// notification (D-58, D-61).
function ActiveWorkout({ workout: w, wake, onBack }: { workout: WorkoutRecord; wake: WakeStatus; onBack: () => void }) {
  const sets = useLiveQuery(() => workoutSets(db, w.id), [w.id], null);
  const cardio = useLiveQuery(() => workoutCardio(db, w.id), [w.id], null);
  const rest = useLiveQuery(() => restTimer(db, w.id), [w.id], null);
  const [selected, setSelected] = useState<string | null>(null);
  const [preview, setPreview] = useState<Preview | null>(null);
  const [dialog, setDialog] = useState<Dialog | null>(null);
  const [error, setError] = useState("");
  const now = useNow(rest !== null || preview !== null);
  const previewOn = preview !== null && now < preview.until;
  useEffect(() => {
    if (preview && !previewOn) setPreview(null);
  }, [preview, previewOn]);

  if (sets === null || cardio === null) return <p className="text-slate-400">Loading…</p>;

  // The exercise that the owner picked, while it has a set with no log.
  // Otherwise the first exercise that the owner did not skip, with a set
  // with no log.
  const picked = w.exercises.find((e) => e.exerciseId === selected);
  const exercise = picked && nextSet(picked, sets) ? picked : currentExercise(w, sets);
  const next = exercise ? nextSet(exercise, sets) : null;
  const open = openExercises(w, sets);
  const unstarted = open.filter((e) => !sets.some((s) => s.exerciseId === e.exerciseId)).length;
  const previewNext = previewOn ? w.exercises.find((e) => e.exerciseId === preview.next) : undefined;

  // onLogged starts the rest timer, and starts the preview after the
  // last set of an exercise when another exercise remains.
  const onLogged = (done: WorkoutExercise, s: SetRecord) => {
    void startRest(db, w.id, done.restSeconds).catch(() => {});
    const after = sets.some((x) => x.id === s.id) ? sets : [...sets, s];
    if (nextSet(done, after) !== null) return;
    const following = currentExercise(w, after);
    if (!following) return;
    setSelected(following.exerciseId);
    setPreview({ done: done.name, next: following.exerciseId, until: Date.now() + PREVIEW_SECONDS * 1000 });
  };

  const finish = async () => {
    setError("");
    try {
      await finishWorkout(db, w.id);
      onBack();
    } catch {
      setDialog(null);
      setError("The phone did not save the end of the workout. Try again.");
    }
  };

  const skip = async (e: WorkoutExercise) => {
    setError("");
    setDialog(null);
    try {
      await skipExercise(db, w.id, e.exerciseId);
      setSelected(null);
    } catch {
      setError("The phone did not save the skip. Try again.");
    }
  };

  const restAction = (work: () => Promise<void>) => void work().catch(() => setError("The phone did not save the rest timer. Try again."));

  return (
    <div className="space-y-6">
      <Title onBack={onBack}>{w.title}</Title>
      {wake.state === "off" && (
        <p className="text-sm text-amber-300" data-testid="wake-off">
          The screen can turn off. Tap the screen to try again.{" "}
          <span className="text-amber-400/80" data-testid="wake-error">{`(${wake.error})`}</span>
        </p>
      )}
      <ErrorText testId="workout-error">{error}</ErrorText>

      <button type="button" className={`${danger} w-full`} onClick={() => setDialog({ kind: "symptoms" })}>
        Report a symptom
      </button>

      {rest && (
        <RestCard
          left={restLeft(rest, now)}
          onLess={() => restAction(() => adjustRest(db, w.id, -REST_STEP_SECONDS))}
          onMore={() => restAction(() => adjustRest(db, w.id, REST_STEP_SECONDS))}
          onDismiss={() => restAction(() => dismissRest(db))}
        />
      )}

      {preview && previewNext ? (
        <NextPreview
          done={preview.done}
          next={previewNext}
          sets={sets}
          seconds={Math.max(1, Math.ceil((preview.until - now) / 1000))}
          onGo={() => setPreview(null)}
        />
      ) : exercise && next ? (
        <SetLogger
          key={`${exercise.exerciseId}-${next.kind}-${next.number}`}
          workoutId={w.id}
          exercise={exercise}
          next={next}
          onLogged={(s) => onLogged(exercise, s)}
          onSkip={() => setDialog({ kind: "skip", exercise })}
          onWarn={(text) => setDialog({ kind: "warning", text })}
          onError={setError}
        />
      ) : (
        <p className="rounded-lg border border-emerald-800 bg-emerald-950 p-3 text-emerald-100" data-testid="sets-done">
          No planned set remains.
        </p>
      )}

      <ExerciseList
        workout={w}
        sets={sets}
        current={previewOn ? undefined : exercise?.exerciseId}
        onPick={(id) => {
          setPreview(null);
          setSelected(id);
        }}
        onEdit={(set, e) => setDialog({ kind: "edit", set, exercise: e })}
      />

      {w.cardio && (
        <CardioCard
          workoutId={w.id}
          cardio={w.cardio}
          logs={cardio}
          onWarn={(text) => setDialog({ kind: "warning", text })}
          onError={setError}
        />
      )}

      <button
        type="button"
        className={`${secondary} w-full`}
        onClick={() => (open.length > 0 ? setDialog({ kind: "finish", open: open.length, unstarted }) : void finish())}
      >
        {open.length > 0 ? "Finish now" : "Finish workout"}
      </button>

      {dialog?.kind === "symptoms" && (
        <Modal title="Report a symptom" onClose={() => setDialog(null)}>
          <div className="space-y-2">
            {SYMPTOMS.map((s) => (
              <button
                key={s.id}
                type="button"
                className={`${secondary} w-full text-left`}
                onClick={() => setDialog({ kind: "warning", text: symptomWarning(s) })}
              >
                {s.label}
              </button>
            ))}
            <button type="button" className={`${secondary} w-full`} onClick={() => setDialog(null)}>
              Cancel
            </button>
          </div>
        </Modal>
      )}
      {dialog?.kind === "warning" && (
        <Modal title="Warning">
          <p className="text-lg font-semibold text-red-200" data-testid="warning-text">
            {dialog.text}
          </p>
          <div className="flex flex-col gap-3">
            <button type="button" className={`${primary} min-h-14`} onClick={() => setDialog(null)}>
              Continue the workout
            </button>
            <button type="button" className={`${danger} min-h-14`} onClick={() => void finish()}>
              Finish now
            </button>
          </div>
        </Modal>
      )}
      {dialog?.kind === "finish" && (
        <Modal title="Finish now?" onClose={() => setDialog(null)}>
          <p className="text-slate-200" data-testid="finish-text">
            {finishText(dialog.open, dialog.unstarted)}
          </p>
          <div className="flex flex-col gap-3">
            <button type="button" className={`${danger} min-h-14`} onClick={() => void finish()}>
              Finish now
            </button>
            <button type="button" className={`${secondary} min-h-14`} onClick={() => setDialog(null)}>
              Cancel
            </button>
          </div>
        </Modal>
      )}
      {dialog?.kind === "skip" && (
        <Modal title={`Skip ${dialog.exercise.name}?`} onClose={() => setDialog(null)}>
          <p className="text-slate-200" data-testid="skip-text">
            Each set with no log counts as skipped, and the next session repeats the target of this exercise.
          </p>
          <div className="flex flex-col gap-3">
            <button type="button" className={`${danger} min-h-14`} onClick={() => void skip(dialog.exercise)}>
              Skip
            </button>
            <button type="button" className={`${secondary} min-h-14`} onClick={() => setDialog(null)}>
              Cancel
            </button>
          </div>
        </Modal>
      )}
      {dialog?.kind === "edit" && (
        <Modal title={`Edit: ${dialog.exercise.name}`} onClose={() => setDialog(null)}>
          <EditSet
            set={dialog.set}
            exercise={dialog.exercise}
            onDone={() => setDialog(null)}
            onWarn={(text) => setDialog({ kind: "warning", text })}
          />
        </Modal>
      )}
    </div>
  );
}

// finishText gives the text of the confirmation of "finish now" (D-63).
function finishText(open: number, unstarted: number): string {
  const first = `${open} ${open === 1 ? "exercise has" : "exercises have"} a set with no log, so the workout ends early.`;
  if (unstarted === 0) return first;
  return `${first} ${unstarted === 1 ? "An exercise" : `${unstarted} exercises`} with no logged set ${unstarted === 1 ? "counts" : "count"} as skipped.`;
}

// RestCard shows the rest timer (D-59, D-270). It reads the stored end
// time, so it is correct after a screen lock. At 0 the card changes
// color and shows "Rest done", with no sound and no vibration (D-58).
function RestCard({ left, onLess, onMore, onDismiss }: { left: number; onLess: () => void; onMore: () => void; onDismiss: () => void }) {
  const done = left === 0;
  const control = "min-h-14 rounded-lg border text-lg font-semibold active:bg-slate-800";
  return (
    <section
      className={`space-y-3 rounded-lg border p-3 ${done ? "border-emerald-400 bg-emerald-900" : "border-slate-700"}`}
      aria-label="Rest timer"
      data-testid="rest-timer"
      data-done={done}
    >
      <div className="flex items-baseline justify-between">
        <span className={`text-lg font-semibold ${done ? "text-emerald-50" : "text-slate-300"}`} data-testid="rest-state">
          {done ? "Rest done" : "Rest"}
        </span>
        <span role="timer" className="text-5xl font-semibold tabular-nums text-slate-50" data-testid="rest-left">
          {clockText(left)}
        </span>
      </div>
      <div className="grid grid-cols-3 gap-2">
        <button type="button" aria-label="15 seconds less rest" className={`${control} border-slate-600 text-slate-100`} onClick={onLess}>
          −15 s
        </button>
        <button type="button" aria-label="15 seconds more rest" className={`${control} border-slate-600 text-slate-100`} onClick={onMore}>
          +15 s
        </button>
        <button type="button" className={`${control} border-slate-600 text-slate-100`} onClick={onDismiss}>
          Dismiss
        </button>
      </div>
    </section>
  );
}

// NextPreview shows the next machine after the last set of an exercise
// (D-60, D-269). "Go now" advances at once.
function NextPreview({
  done,
  next,
  sets,
  seconds,
  onGo,
}: {
  done: string;
  next: WorkoutExercise;
  sets: SetRecord[];
  seconds: number;
  onGo: () => void;
}) {
  const first = nextSet(next, sets);
  return (
    <section className="space-y-3 rounded-lg border border-emerald-700 bg-emerald-950 p-3" aria-label="Next machine" data-testid="next-preview">
      <p className="font-semibold text-emerald-100" data-testid="preview-done">{`${done} is done.`}</p>
      <div>
        <p className="text-xl font-semibold text-slate-50" data-testid="preview-next">{`Next: ${next.name}`}</p>
        {first && (
          <p className="text-sm text-slate-300" data-testid="preview-target">
            {`${first.kind === "calibration" ? "Calibration set" : "Set"} ${first.number} of ${first.of}: ${setText(first.target, first.kind === "calibration")}.`}
          </p>
        )}
      </div>
      <p className="text-sm text-slate-400" data-testid="preview-seconds">{`The next machine shows in ${seconds} s.`}</p>
      <button type="button" className={`${primary} min-h-14 w-full`} onClick={onGo}>
        Go now
      </button>
    </section>
  );
}

// SetLogger logs the next set of an exercise (D-249). The reps and the
// weight come from the target, and the plus and minus buttons change
// them. A tap on the reps in reserve logs the set. A calibration set
// offers 0 to 6+ (D-268), and after it the working sets get the load of
// the calibration table (D-267). Pain and a note are optional, behind
// one tap (D-57, D-162).
function SetLogger({
  workoutId,
  exercise,
  next,
  onLogged,
  onSkip,
  onWarn,
  onError,
}: {
  workoutId: string;
  exercise: WorkoutExercise;
  next: NextSet;
  onLogged: (s: SetRecord) => void;
  onSkip: () => void;
  onWarn: (text: string) => void;
  onError: (text: string) => void;
}) {
  const [reps, setReps] = useState(next.target.reps);
  const [weight, setWeight] = useState(next.target.loadTenthLb);
  const [more, setMore] = useState(false);
  const [pain, setPain] = useState<number | undefined>(undefined);
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const calibration = next.kind === "calibration";
  const tooLong = noteLength(note) > MAX_NOTE_CHARS;

  const log = async (rir: number) => {
    setBusy(true);
    onError("");
    try {
      const s = await logSet(db, workoutId, { exerciseId: exercise.exerciseId, kind: next.kind, reps, weightTenthsLb: weight, rir, pain, note });
      onLogged(s);
      if (pain !== undefined && pain >= 1) onWarn(PAIN_WARNING);
    } catch {
      onError("The phone did not save the set. Try again.");
      setBusy(false);
    }
  };

  return (
    <section className="space-y-4 rounded-lg border border-sky-800 p-3" aria-label="Log a set" data-testid="set-logger">
      <div>
        <h3 className="text-lg font-semibold text-slate-100" data-testid="logger-exercise">
          {exercise.name}
        </h3>
        <p className="text-slate-300" data-testid="set-label">
          {`${calibration ? "Calibration set" : "Set"} ${next.number} of ${next.of}`}
        </p>
        <p className="text-sm text-slate-400" data-testid="set-target">
          {`Target: ${setText(next.target, calibration)}. Rest ${restText(exercise.restSeconds)}.`}
        </p>
        {next.fromCalibration && (
          <p className="text-sm text-sky-200" data-testid="calibration-note">
            The calibration set gave this load.
          </p>
        )}
      </div>

      <Stepper
        label="Reps"
        value={String(reps)}
        testId="reps"
        less="Fewer reps"
        more="More reps"
        onLess={() => setReps(stepReps(reps, -1))}
        onMore={() => setReps(stepReps(reps, 1))}
      />
      <Stepper
        label="Weight"
        value={`${formatPounds(weight)} lb`}
        testId="weight"
        less="Lighter"
        more="Heavier"
        onLess={() => setWeight(stepWeight(exercise.weights, weight, -1))}
        onMore={() => setWeight(stepWeight(exercise.weights, weight, 1))}
      />

      <div className="space-y-2">
        <p className="text-sm font-semibold text-slate-100">
          {calibration ? "Stop at 3 to 4 reps in reserve. Tap the reps in reserve to log the set." : "Tap the reps in reserve to log the set."}
        </p>
        <div className={`grid gap-2 ${calibration ? "grid-cols-4" : "grid-cols-5"}`}>
          {rirChoices(next.kind).map((c) => (
            <button
              key={c.label}
              type="button"
              aria-label={`${c.label} in reserve`}
              className="min-h-16 rounded-lg bg-sky-600 text-2xl font-semibold text-white active:bg-sky-700 disabled:opacity-40"
              disabled={busy || tooLong}
              onClick={() => void log(c.value)}
            >
              {c.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap gap-3">
        <button type="button" className={secondary} aria-expanded={more} onClick={() => setMore(!more)}>
          {more ? "Hide pain and note" : "Add pain or a note"}
        </button>
        <button type="button" className={secondary} onClick={onSkip}>
          Skip this exercise
        </button>
      </div>
      {more && (
        <div className="space-y-3">
          <PainPicker value={pain} onChange={setPain} />
          <NoteField value={note} onChange={setNote} />
        </div>
      )}
    </section>
  );
}

// EditSet changes a logged set (D-63). The set keeps its kind, so the
// reps in reserve are the choices of its kind.
function EditSet({
  set,
  exercise,
  onDone,
  onWarn,
}: {
  set: SetRecord;
  exercise: WorkoutExercise;
  onDone: () => void;
  onWarn: (text: string) => void;
}) {
  const [reps, setReps] = useState(set.reps);
  const [weight, setWeight] = useState(set.weightTenthsLb);
  const [rir, setRir] = useState(Math.min(set.rir, set.kind === "calibration" ? 6 : 4));
  const [pain, setPain] = useState<number | undefined>(set.pain);
  const [note, setNote] = useState(set.note);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const tooLong = noteLength(note) > MAX_NOTE_CHARS;
  const rirId = useId();

  const save = async () => {
    setBusy(true);
    setError("");
    try {
      await editSet(db, set.id, { reps, weightTenthsLb: weight, rir, pain, note });
      if (pain !== undefined && pain >= 1 && pain !== set.pain) onWarn(PAIN_WARNING);
      else onDone();
    } catch {
      setError("The phone did not save the change. Try again.");
      setBusy(false);
    }
  };

  return (
    <div className="space-y-4" data-testid="edit-set">
      <Stepper
        label="Reps"
        value={String(reps)}
        testId="edit-reps"
        less="Fewer reps"
        more="More reps"
        onLess={() => setReps(stepReps(reps, -1))}
        onMore={() => setReps(stepReps(reps, 1))}
      />
      <Stepper
        label="Weight"
        value={`${formatPounds(weight)} lb`}
        testId="edit-weight"
        less="Lighter"
        more="Heavier"
        onLess={() => setWeight(stepWeight(exercise.weights, weight, -1))}
        onMore={() => setWeight(stepWeight(exercise.weights, weight, 1))}
      />
      <div className="space-y-2" role="group" aria-labelledby={rirId}>
        <p id={rirId} className="text-sm font-semibold text-slate-100">
          Reps in reserve
        </p>
        <div className={`grid gap-2 ${set.kind === "calibration" ? "grid-cols-4" : "grid-cols-5"}`}>
          {rirChoices(set.kind).map((c) => (
            <button
              key={c.label}
              type="button"
              aria-label={`${c.label} in reserve`}
              aria-pressed={rir === c.value}
              className={`min-h-14 rounded-lg border text-xl font-semibold ${
                rir === c.value ? "border-sky-400 bg-sky-700 text-white" : "border-slate-700 text-slate-100"
              }`}
              onClick={() => setRir(c.value)}
            >
              {c.label}
            </button>
          ))}
        </div>
      </div>
      <PainPicker value={pain} onChange={setPain} />
      <NoteField value={note} onChange={setNote} />
      <ErrorText testId="edit-error">{error}</ErrorText>
      <div className="flex flex-col gap-3">
        <button type="button" className={`${primary} min-h-14`} disabled={busy || tooLong} onClick={() => void save()}>
          Save the change
        </button>
        <button type="button" className={`${secondary} min-h-14`} onClick={onDone}>
          Cancel
        </button>
      </div>
    </div>
  );
}

function Stepper({
  label,
  value,
  testId,
  less,
  more,
  onLess,
  onMore,
}: {
  label: string;
  value: string;
  testId: string;
  less: string;
  more: string;
  onLess: () => void;
  onMore: () => void;
}) {
  const step = "min-h-14 min-w-14 rounded-lg border border-slate-600 text-2xl font-semibold text-slate-100 active:bg-slate-800";
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-slate-300">{label}</span>
      <div className="flex items-center gap-3">
        <button type="button" aria-label={less} className={step} onClick={onLess}>
          −
        </button>
        <span className="min-w-20 text-center text-xl font-semibold text-slate-100" data-testid={testId}>
          {value}
        </span>
        <button type="button" aria-label={more} className={step} onClick={onMore}>
          +
        </button>
      </div>
    </div>
  );
}

// PainPicker gives the optional pain rating from 0 to 10 (D-162). A tap
// on the picked rating removes it, so no value means no report.
function PainPicker({ value, onChange }: { value: number | undefined; onChange: (v: number | undefined) => void }) {
  const id = useId();
  return (
    <div className="space-y-2" role="group" aria-labelledby={id}>
      <p id={id} className="text-sm font-semibold text-slate-100">
        Pain (optional), from 0 (no pain) to 10
      </p>
      <div className="grid grid-cols-6 gap-2">
        {Array.from({ length: 11 }, (_, n) => (
          <button
            key={n}
            type="button"
            aria-label={`Pain ${n}`}
            aria-pressed={value === n}
            className={`min-h-11 rounded-lg border text-base font-medium ${
              value === n ? "border-amber-400 bg-amber-900 text-amber-50" : "border-slate-700 text-slate-100"
            }`}
            onClick={() => onChange(value === n ? undefined : n)}
          >
            {n}
          </button>
        ))}
      </div>
    </div>
  );
}

function NoteField({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const count = noteLength(value);
  return (
    <label className="block">
      <span className="text-sm font-semibold text-slate-100">Note (optional)</span>
      <textarea rows={2} value={value} onChange={(e) => onChange(e.target.value)} className={field} />
      <span className={`text-xs ${count > MAX_NOTE_CHARS ? "text-red-400" : "text-slate-500"}`}>
        {count} of {MAX_NOTE_CHARS} characters
      </span>
    </label>
  );
}

// ExerciseList shows each exercise with its logged sets. A tap picks the
// exercise of the next log, and a skipped exercise can be picked again.
// Each logged set has an edit button (D-63).
function ExerciseList({
  workout,
  sets,
  current,
  onPick,
  onEdit,
}: {
  workout: WorkoutRecord;
  sets: SetRecord[];
  current: string | undefined;
  onPick: (exerciseId: string) => void;
  onEdit: (s: SetRecord, e: WorkoutExercise) => void;
}) {
  return (
    <section className="space-y-3" aria-label="Exercises">
      <h3 className="text-base font-semibold text-slate-100">Exercises</h3>
      {workout.exercises.map((e) => {
        const mine = sets.filter((s) => s.exerciseId === e.exerciseId);
        const planned = e.calibrationSets.length + e.workingSets.length;
        const open = nextSet(e, sets) !== null;
        const skipped = workout.skippedExerciseIds.includes(e.exerciseId) && open;
        return (
          <div key={e.exerciseId} className="space-y-1 border-t border-slate-800 pt-2" data-testid="workout-exercise" data-exercise-id={e.exerciseId}>
            <button
              type="button"
              className={`min-h-11 w-full rounded-lg px-3 text-left ${
                e.exerciseId === current ? "bg-slate-800 text-slate-50" : "text-slate-200 active:bg-slate-800"
              }`}
              disabled={!open}
              aria-current={e.exerciseId === current}
              onClick={() => onPick(e.exerciseId)}
            >
              <span className="font-medium">{e.name}</span>{" "}
              <span className="text-sm text-slate-400" data-testid="exercise-count">{`${mine.length} of ${planned} sets`}</span>
              {skipped && (
                <span className="ml-2 text-sm text-amber-300" data-testid="exercise-skipped">
                  Skipped
                </span>
              )}
            </button>
            {mine.length > 0 && (
              <ul className="space-y-1 px-3 text-sm text-slate-300" data-testid="logged-sets">
                {mine.map((s, i) => (
                  <li key={s.id} className="flex items-center justify-between gap-2">
                    <span data-testid="logged-text">{loggedText(s)}</span>
                    <button
                      type="button"
                      aria-label={`Edit set ${i + 1} of ${e.name}`}
                      className="min-h-11 min-w-14 rounded-lg border border-slate-700 px-3 text-slate-100 active:bg-slate-800"
                      onClick={() => onEdit(s, e)}
                    >
                      Edit
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        );
      })}
    </section>
  );
}

// CardioCard logs the cardio of the session (D-123): the duration and the
// effort rating, with the distance, the resistance level, pain, and a
// note behind one tap.
function CardioCard({
  workoutId,
  cardio,
  logs,
  onWarn,
  onError,
}: {
  workoutId: string;
  cardio: NonNullable<WorkoutRecord["cardio"]>;
  logs: CardioRecord[];
  onWarn: (text: string) => void;
  onError: (text: string) => void;
}) {
  const [minutes, setMinutes] = useState(cardio.minutes);
  const [effort, setEffort] = useState<number | null>(null);
  const [more, setMore] = useState(false);
  const [distance, setDistance] = useState("");
  const [level, setLevel] = useState("");
  const [pain, setPain] = useState<number | undefined>(undefined);
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const effortId = useId();

  if (logs.length > 0) {
    return (
      <section className="space-y-1 rounded-lg border border-slate-800 p-3" aria-label="Cardio" data-testid="cardio-card">
        <h3 className="text-base font-semibold text-slate-100">{cardio.name}</h3>
        {logs.map((c) => (
          <p key={c.id} className="text-sm text-slate-300" data-testid="cardio-logged">
            {`Logged: ${Math.round(c.durationSeconds / 60)} min, effort ${c.effort} of 10`}
          </p>
        ))}
      </section>
    );
  }

  const miles = parseMiles(distance);
  const resistance = parseLevel(level);
  const bad = miles === null || resistance === null || noteLength(note) > MAX_NOTE_CHARS;

  const log = async () => {
    if (effort === null || miles === null || resistance === null) return;
    setBusy(true);
    onError("");
    try {
      await logCardio(db, workoutId, {
        exerciseId: cardio.exerciseId,
        durationSeconds: minutes * 60,
        effort,
        distanceTenthsMi: miles,
        resistance,
        pain,
        note,
      });
      if (pain !== undefined && pain >= 1) onWarn(PAIN_WARNING);
    } catch {
      onError("The phone did not save the cardio. Try again.");
      setBusy(false);
    }
  };

  return (
    <section className="space-y-4 rounded-lg border border-slate-800 p-3" aria-label="Cardio" data-testid="cardio-card">
      <div>
        <h3 className="text-base font-semibold text-slate-100">{cardio.name}</h3>
        <p className="text-sm text-slate-400">{`Target: ${cardio.minutes} min.`}</p>
      </div>
      <Stepper
        label="Minutes"
        value={String(minutes)}
        testId="cardio-minutes"
        less="Fewer minutes"
        more="More minutes"
        onLess={() => setMinutes(stepMinutes(minutes, -1))}
        onMore={() => setMinutes(stepMinutes(minutes, 1))}
      />
      <div className="space-y-2" role="group" aria-labelledby={effortId}>
        <p id={effortId} className="text-sm font-semibold text-slate-100">
          Effort, from 1 (easy) to 10 (hardest)
        </p>
        <div className="grid grid-cols-5 gap-2">
          {Array.from({ length: 10 }, (_, i) => i + 1).map((n) => (
            <button
              key={n}
              type="button"
              aria-label={`Effort ${n}`}
              aria-pressed={effort === n}
              className={`min-h-12 rounded-lg border text-lg font-medium ${
                effort === n ? "border-sky-400 bg-sky-800 text-white" : "border-slate-700 text-slate-100"
              }`}
              onClick={() => setEffort(n)}
            >
              {n}
            </button>
          ))}
        </div>
      </div>

      <button type="button" className={secondary} aria-expanded={more} onClick={() => setMore(!more)}>
        {more ? "Hide the optional fields" : "Add distance, level, pain, or a note"}
      </button>
      {more && (
        <div className="space-y-3">
          <TextField label="Distance in miles (optional)" value={distance} onChange={setDistance} invalid={miles === null} inputMode="decimal">
            Give a number such as 2.5.
          </TextField>
          <TextField label="Resistance level (optional)" value={level} onChange={setLevel} invalid={resistance === null} inputMode="numeric">
            Give a whole number.
          </TextField>
          <PainPicker value={pain} onChange={setPain} />
          <NoteField value={note} onChange={setNote} />
        </div>
      )}

      <button type="button" className={`${primary} min-h-14 w-full`} disabled={busy || effort === null || bad} onClick={() => void log()}>
        Log cardio
      </button>
    </section>
  );
}

function TextField({
  label,
  value,
  onChange,
  invalid,
  inputMode,
  children,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  invalid: boolean;
  inputMode: "decimal" | "numeric";
  children: ReactNode;
}) {
  return (
    <label className="block">
      <span className="text-sm font-semibold text-slate-100">{label}</span>
      <input type="text" inputMode={inputMode} value={value} onChange={(e) => onChange(e.target.value)} className={field} aria-invalid={invalid} />
      {invalid && <span className="text-xs text-red-400">{children}</span>}
    </label>
  );
}

// Modal shows a dialog over the workout. It needs a choice, so a tap
// outside it does nothing.
function Modal({ title, onClose, children }: { title: string; onClose?: () => void; children: ReactNode }) {
  const id = useId();
  return (
    <div className="fixed inset-0 z-20 flex items-end justify-center bg-black/70 p-4 sm:items-center">
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={id}
        className="max-h-full w-full max-w-md space-y-4 overflow-y-auto rounded-lg border border-slate-700 bg-slate-900 p-4"
        onKeyDown={(e) => {
          if (e.key === "Escape" && onClose) onClose();
        }}
      >
        <h3 id={id} className="text-lg font-semibold text-slate-100">
          {title}
        </h3>
        {children}
      </div>
    </div>
  );
}
