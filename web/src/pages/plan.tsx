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
  planErrorText,
  progressText,
  reasonLength,
  restText,
  setText,
} from "../lib/plan";
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
// and an exclusion (D-252).
export function PlanPage({ onBack }: { onBack: () => void }) {
  const plan = useQuery(PlanService.method.getPlan, {});
  // Each read of the plan goes into the offline copy, so the workout
  // starts with no connection (D-278).
  useEffect(() => {
    if (plan.data) void keepCopy(db, "plan", toJson(GetPlanResponseSchema, plan.data));
  }, [plan.data]);

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
}: {
  session: PlanSession;
  index: number;
  busy: boolean;
  target: Target | null;
  onOpen: (exerciseId: string) => void;
  onCancel: () => void;
  onExclude: (exerciseId: string, reason: string) => void;
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

function ExerciseCard({
  exercise,
  busy,
  open,
  onOpen,
  onCancel,
  onExclude,
}: {
  exercise: PlannedExercise;
  busy: boolean;
  open: boolean;
  onOpen: () => void;
  onCancel: () => void;
  onExclude: (reason: string) => void;
}) {
  const [reason, setReason] = useState("");
  const count = reasonLength(reason);

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

      {!open && (
        <button type="button" className={secondary} disabled={busy} onClick={onOpen}>
          Exclude
        </button>
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
