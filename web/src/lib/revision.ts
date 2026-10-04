import type { Plan, PlannedExercise } from "../gen/workoutapp/v1/plan_service_pb";
import type { OutboxEntry } from "./db";
import { setText } from "./plan";

// The next targets after a workout (work area 7.1). When the sync applies
// a finished workout, the server revises the plan (D-292): the rules give
// the next target of each exercise that the workout logged, and Luna
// writes the reason alone (D-288). The sync then reads the plan again,
// so the offline copy of the plan holds the revision (D-278).

// NextTarget is one revised exercise for the screen: its name, each
// target set as text, and its reason. fromLuna is true when the reason is
// the reason of Luna.
export type NextTarget = { exerciseId: string; name: string; sets: string[]; reason: string; fromLuna: boolean };

// NextState is the state of the next targets of a finished workout:
// "revised" when the plan copy holds the revision of the workout,
// "waiting" while the workout waits for the sync, and "none" when the
// sync ended and the plan has no revision of the workout, for example
// after a new plan.
export type NextState = { kind: "revised"; targets: NextTarget[] } | { kind: "waiting" } | { kind: "none" };

// WAITING_TEXT tells the owner that the targets come after the sync
// (D-292).
export const WAITING_TEXT = "The next targets show after the sync. The phone syncs when it is online.";

// NONE_TEXT tells the owner that the plan has no revision of the workout.
export const NONE_TEXT = "The plan has no new targets from this workout.";

// NO_CHANGE_TEXT tells the owner that the revision changed no target.
export const NO_CHANGE_TEXT = "This workout logged no exercise of the plan, so no target changed.";

// waitsForSync tells whether an outbox entry belongs to a workout: its
// header, or one of its sets or cardio logs.
export function waitsForSync(e: Pick<OutboxEntry, "entity" | "entityId" | "payload">, workoutId: string): boolean {
  if (e.entity === "workout") return e.entityId === workoutId;
  if (e.entity === "set" || e.entity === "cardio") return (e.payload as { workoutId?: string } | null)?.workoutId === workoutId;
  return false;
}

// nextTargets gives each revised exercise of the plan, in the order of
// the revision. Each session of the plan that holds an exercise has the
// same new target (D-290), so the first one gives it.
export function nextTargets(plan: Plan): NextTarget[] {
  const out: NextTarget[] = [];
  for (const id of plan.lastRevision?.exerciseIds ?? []) {
    let found: PlannedExercise | undefined;
    for (const s of plan.sessions) {
      found = s.exercises.find((e) => e.exerciseId === id);
      if (found) break;
    }
    if (!found) continue;
    out.push({
      exerciseId: id,
      name: found.name,
      sets: [
        ...found.calibrationSets.map((c) => `Calibration: ${setText(c, true)}`),
        ...found.workingSets.map((w, i) => `${i === 0 && found.firstSetCalibration ? "Calibration: " : ""}${setText(w, false)}`),
      ],
      reason: found.reason,
      fromLuna: found.reasonSource === "luna",
    });
  }
  return out;
}

// nextState gives the state of the next targets of a finished workout.
// waiting is true while an entry of the workout waits in the outbox, or
// while a sync runs, because the sync reads the plan after the outbox.
export function nextState(plan: Plan | undefined, workoutId: string, waiting: boolean): NextState {
  if (plan?.lastRevision?.workoutId === workoutId) return { kind: "revised", targets: nextTargets(plan) };
  if (waiting) return { kind: "waiting" };
  return { kind: "none" };
}

// revisedAt gives the time of the last revision of a plan copy, the JSON
// form of GetPlanResponse, or "" for none.
export function revisedAt(json: unknown): string {
  const plan = (json as { plan?: { lastRevision?: { revisedAt?: unknown } } } | null | undefined)?.plan;
  const at = plan?.lastRevision?.revisedAt;
  return typeof at === "string" ? at : "";
}
