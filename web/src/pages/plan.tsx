import { toJson } from "@bufbuild/protobuf";
import { useQuery } from "@connectrpc/connect-query";
import { useLiveQuery } from "dexie-react-hooks";
import { useEffect, useRef, useState, type ReactNode } from "react";

import {
  GetPlanResponseSchema,
  PlanService,
  type Exclusion,
  type GuidanceItem,
  type Plan,
  type PlannedExercise,
  type PlanProgress,
  type PlanSession,
} from "../gen/workoutapp/v1/plan_service_pb";
import { db } from "../lib/db";
import { loadErrorText } from "../lib/errors";
import {
  localDate,
  MAX_REASON_CHARS,
  overrideErrorText,
  overrideSets,
  planErrorText,
  planRequest,
  progressText,
  reasonLength,
  restText,
  setText,
  type OverrideDraft,
} from "../lib/plan";
import { formatPounds } from "../lib/inventory";
import { usePlanApi } from "../lib/plan-api";
import { keepCopy } from "../lib/sync";
import { activeWorkout } from "../lib/workout";
import { danger, ErrorText, field, primary, secondary, Title } from "./inventory/ui";

// PlanPage is the plan screen (work area 5.2). It shows the plan of the
// API: for each session the warm-up, the work sets, the rest, the
// cardio of D-255, and the cool-down, then the mobility and recovery
// items, in text alone (D-44, D-73). It requests a new plan, and it
// excludes an exercise with an optional reason (D-48). While a request
// runs, the screen shows each progress step (D-231, D-239). A failed
// request shows its error, and the plan does not change (D-230, D-240).
// While a workout is open on the phone, the screen refuses a new plan
// and an exclusion (D-252). Each target is the target on the local date
// (D-294, D-295). The owner can change the load and the reps of a target
// for the next session, with a reason (D-69, D-293).
export function PlanPage({ onBack }: { onBack: () => void }) {
  const plan = useQuery(PlanService.method.getPlan, planRequest());
  // Each read of the plan goes into the offline copy, so the workout
  // starts with no connection (D-278). A stale plan of the cache waits
  // for its new read, because a sync can revise the plan (D-292), and the
  // old plan must not replace the new copy.
  useEffect(() => {
    if (plan.data && !plan.isStale) void keepCopy(db, "plan", toJson(GetPlanResponseSchema, plan.data));
  }, [plan.data, plan.isStale]);

  if (plan.error) {
    return (
      <div className="space-y-4">
        <Title onBack={onBack}>Plan</Title>
        <ErrorText testId="load-error">{loadErrorText(plan.error)}</ErrorText>
        <button type="button" className={secondary} onClick={() => void plan.refetch()}>
          Try again
        </button>
      </div>
    );
  }
  if (!plan.data) return <p className="text-slate-400">Loading…</p>;
  return <PlanView plan={plan.data.plan} exclusions={plan.data.exclusions} onBack={onBack} />;
}

// The exercise of an open exclusion form: the session and the exercise.
type Target = { session: number; exerciseId: string };

function PlanView({ plan, exclusions, onBack }: { plan?: Plan; exclusions: Exclusion[]; onBack: () => void }) {
  const api = usePlanApi();
  const [progress, setProgress] = useState<PlanProgress | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [target, setTarget] = useState<Target | null>(null);
  const abort = useRef<AbortController | null>(null);
  const top = useRef<HTMLDivElement>(null);
  const workout = useLiveQuery(() => activeWorkout(db), [], null);
  // While the store loads, the screen does not know of a workout, so it
  // refuses a change too.
  const locked = workout !== undefined;
  const blocked = busy || locked;

  // A close of the screen stops the stream, so the server saves nothing
  // (D-237).
  useEffect(() => () => abort.current?.abort(), []);

  const run = async (exclude: boolean, call: (onProgress: (p: PlanProgress) => void, signal: AbortSignal) => Promise<Plan>) => {
    if (locked) return;
    const ctl = new AbortController();
    abort.current = ctl;
    let sawProgress = false;
    setBusy(true);
    setError("");
    setProgress(null);
    // An exclusion starts low on the screen, so the screen goes to the
    // top, where the progress, the error, and the new plan show.
    top.current?.scrollIntoView({ block: "start" });
    try {
      await call((p) => {
        sawProgress = true;
        setProgress(p);
      }, ctl.signal);
      setTarget(null);
    } catch (err) {
      if (!ctl.signal.aborted) setError(planErrorText(err, { exclude, sawProgress }));
    } finally {
      if (!ctl.signal.aborted) {
        setBusy(false);
        setProgress(null);
      }
    }
  };

  const request = () => void run(false, (onProgress, signal) => api.requestPlan(localDate(), onProgress, signal));
  const exclude = (exerciseId: string, reason: string) =>
    void run(true, (onProgress, signal) => api.excludeExercise(localDate(), exerciseId, reason, onProgress, signal));
  const override: OverrideApi = {
    save: (exerciseId, sets, reason) => api.overrideTarget(localDate(), exerciseId, sets, reason),
    remove: (exerciseId) => api.removeOverride(localDate(), exerciseId),
  };

  return (
    <div ref={top} className="space-y-6">
      <Title onBack={busy ? undefined : onBack}>Plan</Title>

      {busy && progress && (
        <p role="status" className="rounded-lg border border-sky-800 bg-sky-950 p-3 text-sm text-sky-100" data-testid="plan-progress">
          {progressText(progress)}
        </p>
      )}
      <ErrorText testId="plan-error">{error}</ErrorText>
      {workout && (
        <p className="rounded-lg border border-amber-800 bg-amber-950 p-3 text-sm text-amber-100" data-testid="workout-lock">
          A workout is in progress. Finish it before you make a new plan or exclude an exercise.
        </p>
      )}

      {!plan && (
        <section className="space-y-3">
          <p className="text-slate-300" data-testid="no-plan">
            You have no plan yet. Luna makes a plan from your profile and your confirmed machines.
          </p>
          <button type="button" className={`${primary} w-full`} disabled={blocked} onClick={request}>
            Make a plan
          </button>
        </section>
      )}

      {plan && (
        <>
          <section className="space-y-3">
            <p className="text-slate-100" data-testid="plan-summary">
              {plan.summary}
            </p>
            <button type="button" className={`${secondary} w-full`} disabled={blocked} onClick={request}>
              Make a new plan
            </button>
          </section>

          {plan.sessions.map((s, i) => (
            <SessionCard
              key={i}
              session={s}
              index={i}
              busy={blocked}
              target={target}
              onOpen={(exerciseId) => setTarget({ session: i, exerciseId })}
              onCancel={() => setTarget(null)}
              onExclude={exclude}
              override={override}
            />
          ))}

          <GuidanceList items={plan.guidance} />
        </>
      )}

      <ExclusionList items={exclusions} />
    </div>
  );
}

function SessionCard({
  session,
  index,
  busy,
  target,
  onOpen,
  onCancel,
  onExclude,
  override,
}: {
  session: PlanSession;
  index: number;
  busy: boolean;
  target: Target | null;
  onOpen: (exerciseId: string) => void;
  onCancel: () => void;
  onExclude: (exerciseId: string, reason: string) => void;
  override: OverrideApi;
}) {
  return (
    <section
      className="space-y-4 rounded-lg border border-slate-800 p-3"
      aria-label={session.title}
      data-testid="plan-session"
    >
      <h3 className="text-base font-semibold text-slate-100">{session.title}</h3>
      <Part title="Warm-up" testId="warm-up">
        {session.warmUp?.text}
      </Part>

      {session.exercises.map((e) => (
        <ExerciseCard
          key={e.exerciseId}
          exercise={e}
          busy={busy}
          open={target?.session === index && target.exerciseId === e.exerciseId}
          onOpen={() => onOpen(e.exerciseId)}
          onCancel={onCancel}
          onExclude={(reason) => onExclude(e.exerciseId, reason)}
          override={override}
        />
      ))}

      {session.cardio && (
        <Part title="Cardio" testId="cardio">
          {`${session.cardio.name}, ${session.cardio.minutes} min`}
        </Part>
      )}
      <Part title="Cool-down" testId="cool-down">
        {session.coolDown?.text}
      </Part>
    </section>
  );
}

// OverrideApi saves and removes an override of the owner (D-293).
type OverrideApi = {
  save: (exerciseId: string, sets: { reps: number; loadTenthLb: number }[], reason: string) => Promise<Plan>;
  remove: (exerciseId: string) => Promise<Plan>;
};

function ExerciseCard({
  exercise,
  busy,
  open,
  onOpen,
  onCancel,
  onExclude,
  override,
}: {
  exercise: PlannedExercise;
  busy: boolean;
  open: boolean;
  onOpen: () => void;
  onCancel: () => void;
  onExclude: (reason: string) => void;
  override: OverrideApi;
}) {
  const [reason, setReason] = useState("");
  const count = reasonLength(reason);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const o = exercise.override;

  const remove = async () => {
    setSaving(true);
    setError("");
    try {
      await override.remove(exercise.exerciseId);
    } catch (err) {
      setError(overrideErrorText(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-2 border-t border-slate-800 pt-3" data-testid="plan-exercise" data-exercise-id={exercise.exerciseId}>
      <div className="flex items-baseline justify-between gap-3">
        <h4 className="font-medium text-slate-100">{exercise.name}</h4>
        {exercise.source === "rules" && (
          <span className="shrink-0 rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300" data-testid="rules-target">
            Rules target
          </span>
        )}
      </div>
      {o?.expired && (
        <p className="rounded-lg border border-amber-800 bg-amber-950 p-3 text-sm text-amber-100" data-testid="override-expired">
          Your change no longer applies. A missed week, a break, or a deload changed the target after you saved it, so the
          recommendation applies. Change the target again if you want.
        </p>
      )}
      {o && !o.expired && (
        <div className="space-y-1 rounded-lg border border-sky-800 bg-sky-950 p-3 text-sm text-sky-100" data-testid="override">
          <p className="font-semibold">Your change for the next session</p>
          <ul className="space-y-1">
            {o.calibrationSets.map((s, i) => (
              <li key={`c${i}`}>{`Calibration set ${i + 1}: ${setText(s, true)}`}</li>
            ))}
            {o.workingSets.map((s, i) => (
              <li key={`w${i}`}>{`Set ${i + 1}: ${setText(s, false)}`}</li>
            ))}
          </ul>
          <p data-testid="override-reason">{`Your reason: ${o.reason}`}</p>
        </div>
      )}
      {o && !o.expired && <p className="text-sm font-semibold text-slate-300">Recommendation</p>}
      <ul className="space-y-1 text-sm text-slate-200" data-testid="sets">
        {exercise.calibrationSets.map((s, i) => (
          <li key={`c${i}`}>{`Calibration set ${i + 1}: ${setText(s, true)}`}</li>
        ))}
        {exercise.workingSets.map((s, i) => (
          <li key={`w${i}`}>{`Set ${i + 1}: ${setText(s, false)}`}</li>
        ))}
      </ul>
      <p className="text-sm text-slate-400" data-testid="rest">
        {`Rest ${restText(exercise.restSeconds)} between sets.`}
      </p>
      {exercise.reason && <p className="text-sm text-slate-400">{exercise.reason}</p>}
      <ErrorText testId="override-error">{error}</ErrorText>

      {editing && (
        <OverrideForm
          exercise={exercise}
          busy={busy || saving}
          onSave={async (sets, why) => {
            setSaving(true);
            setError("");
            try {
              await override.save(exercise.exerciseId, sets, why);
              setEditing(false);
            } catch (err) {
              setError(overrideErrorText(err));
            } finally {
              setSaving(false);
            }
          }}
          onCancel={() => {
            setError("");
            setEditing(false);
          }}
        />
      )}

      {!open && !editing && (
        <div className="flex flex-wrap gap-3">
          <button type="button" className={secondary} disabled={busy || saving} onClick={() => setEditing(true)}>
            Change target
          </button>
          {o && (
            <button type="button" className={secondary} disabled={busy || saving} onClick={() => void remove()}>
              Use the recommendation
            </button>
          )}
          <button type="button" className={secondary} disabled={busy || saving} onClick={onOpen}>
            Exclude
          </button>
        </div>
      )}
      {open && (
        <div className="space-y-3 rounded-lg border border-slate-700 p-3" data-testid="exclude-form">
          <p className="text-sm text-slate-300">Luna makes a new plan without this exercise.</p>
          <label className="block">
            <span className="text-sm font-semibold text-slate-100">Reason (optional)</span>
            <span className="block text-sm text-slate-400">No AI model reads this text.</span>
            <textarea rows={2} value={reason} onChange={(e) => setReason(e.target.value)} className={field} />
            <span className={`text-xs ${count > MAX_REASON_CHARS ? "text-red-400" : "text-slate-500"}`}>
              {count} of {MAX_REASON_CHARS} characters
            </span>
          </label>
          <div className="flex flex-wrap gap-3">
            <button
              type="button"
              className={danger}
              disabled={busy || count > MAX_REASON_CHARS}
              onClick={() => onExclude(reason)}
            >
              Exclude and make a new plan
            </button>
            <button
              type="button"
              className={secondary}
              disabled={busy}
              onClick={() => {
                setReason("");
                onCancel();
              }}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

// OverrideForm changes the load and the reps of each working set for the
// next session, with a reason (D-69, D-293). The reps in reserve, the
// count of sets, and the rest stay as recommended. The server checks the
// change with the policy before it saves it (D-23).
function OverrideForm({
  exercise,
  busy,
  onSave,
  onCancel,
}: {
  exercise: PlannedExercise;
  busy: boolean;
  onSave: (sets: { reps: number; loadTenthLb: number }[], reason: string) => Promise<void>;
  onCancel: () => void;
}) {
  const live = exercise.override && !exercise.override.expired ? exercise.override : undefined;
  const start = live?.workingSets ?? exercise.workingSets;
  const [drafts, setDrafts] = useState<OverrideDraft[]>(() =>
    start.map((s) => ({ reps: String(s.reps), pounds: formatPounds(s.loadTenthLb) })),
  );
  const [reason, setReason] = useState(exercise.override?.reason ?? "");
  const count = reasonLength(reason);
  const sets = overrideSets(drafts);
  const change = (i: number, part: Partial<OverrideDraft>) => setDrafts((d) => d.map((x, j) => (j === i ? { ...x, ...part } : x)));

  return (
    <div className="space-y-3 rounded-lg border border-slate-700 p-3" data-testid="override-form">
      <p className="text-sm text-slate-300">
        Change the reps and the load of each set for the next session. The reps in reserve stay as recommended.
      </p>
      {drafts.map((d, i) => (
        <div key={i} className="grid grid-cols-[auto_1fr_1fr] items-end gap-3">
          <span className="pb-2 text-sm text-slate-300">{`Set ${i + 1}`}</span>
          <label className="block">
            <span className="text-xs text-slate-400">Reps</span>
            <input
              inputMode="numeric"
              value={d.reps}
              onChange={(e) => change(i, { reps: e.target.value })}
              className={field}
              aria-label={`Set ${i + 1} reps`}
            />
          </label>
          <label className="block">
            <span className="text-xs text-slate-400">Load (lb)</span>
            <input
              inputMode="decimal"
              value={d.pounds}
              onChange={(e) => change(i, { pounds: e.target.value })}
              className={field}
              aria-label={`Set ${i + 1} load`}
            />
          </label>
        </div>
      ))}
      <label className="block">
        <span className="text-sm font-semibold text-slate-100">Reason</span>
        <span className="block text-sm text-slate-400">No AI model reads this text.</span>
        <textarea rows={2} value={reason} onChange={(e) => setReason(e.target.value)} className={field} />
        <span className={`text-xs ${count > MAX_REASON_CHARS ? "text-red-400" : "text-slate-500"}`}>
          {count} of {MAX_REASON_CHARS} characters
        </span>
      </label>
      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          className={primary}
          disabled={busy || !sets || reason.trim() === "" || count > MAX_REASON_CHARS}
          onClick={() => sets && void onSave(sets, reason)}
        >
          Save the change
        </button>
        <button type="button" className={secondary} disabled={busy} onClick={onCancel}>
          Cancel
        </button>
      </div>
    </div>
  );
}

function GuidanceList({ items }: { items: GuidanceItem[] }) {
  if (items.length === 0) return null;
  const label = (kind: string) => (kind === "mobility" ? "Mobility" : kind === "recovery" ? "Recovery" : "Guidance");
  return (
    <section className="space-y-2" aria-label="Mobility and recovery" data-testid="guidance">
      <h3 className="text-base font-semibold text-slate-100">Mobility and recovery</h3>
      <ul className="space-y-2 text-sm text-slate-200">
        {items.map((g) => (
          <li key={g.id}>
            <span className="font-medium text-slate-100">{label(g.kind)}:</span> {g.text}
          </li>
        ))}
      </ul>
    </section>
  );
}

function ExclusionList({ items }: { items: Exclusion[] }) {
  if (items.length === 0) return null;
  return (
    <section className="space-y-2" aria-label="Excluded exercises" data-testid="exclusions">
      <h3 className="text-base font-semibold text-slate-100">Excluded exercises</h3>
      <ul className="space-y-2 text-sm">
        {items.map((x) => (
          <li key={x.exerciseId} data-testid="exclusion">
            <span className="text-slate-100">{x.name}</span>
            {x.reason && <span className="block text-slate-400">{x.reason}</span>}
          </li>
        ))}
      </ul>
    </section>
  );
}

function Part({ title, testId, children }: { title: string; testId: string; children: ReactNode }) {
  return (
    <div className="text-sm" data-testid={testId}>
      <span className="font-semibold text-slate-100">{title}:</span> <span className="text-slate-200">{children}</span>
    </div>
  );
}
