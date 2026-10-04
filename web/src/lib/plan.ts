import { Code, ConnectError } from "@connectrpc/connect";

import type { PlanProgress, PlannedSet } from "../gen/workoutapp/v1/plan_service_pb";
import { isNoConnection } from "./errors";
import { formatPounds, parsePounds } from "./inventory";
import { InventoryNotSyncedError } from "./sync";

export { localDate, planRequest } from "./today";

// The texts and the formats of the plan screen (work area 5.2). The owner
// chose the text of each progress step (D-239) and of each error (D-240).

// MAX_REASON_CHARS is the limit of the reason of an exclusion (D-228).
// The server trims the reason, then counts its characters.
export const MAX_REASON_CHARS = 200;

export function reasonLength(reason: string): number {
  return [...reason.trim()].length;
}

// The cause of a failed call, from the status that the next call event
// names (D-239). An unknown status gets the cause of "error".
const CAUSES: Record<string, string> = {
  malformed: "it was not in the right form",
  refusal: "Luna refused",
  incomplete: "it was cut off",
  timeout: "it took too long",
  error: "the call failed",
};

// progressText gives the text of one progress event of the stream
// (D-231, D-237).
export function progressText(p: Pick<PlanProgress, "step" | "attempt" | "maxAttempts" | "previousStatus">): string {
  switch (p.step) {
    case "call": {
      const attempt = `Try ${p.attempt} of ${p.maxAttempts}.`;
      if (p.attempt <= 1 || !p.previousStatus) return `Asking Luna for a plan. ${attempt}`;
      const cause = CAUSES[p.previousStatus] ?? CAUSES.error;
      return `The last answer of Luna was not valid (${cause}). Asking again. ${attempt}`;
    }
    case "check":
      return "Checking each set and load with the safety rules.";
    case "save":
      return "Saving the plan.";
    default:
      return "Making the plan.";
  }
}

export const NO_VALID_PLAN = "Luna gave no valid plan in 4 tries. Your plan did not change. Try again later.";
export const CAP_REACHED =
  "The monthly AI limit is reached. Your plan did not change. The limit resets on the first day of the month (UTC).";
export const NO_EXERCISE =
  "No confirmed machine gives an exercise that you can do. Your plan did not change. Confirm a machine, or change the injuries in your profile.";
export const NOT_EXCLUDED = "The exercise is not excluded.";

// NOT_SYNCED is the text of a plan request that stopped before its call,
// because an inventory change waits in the outbox (D-272).
const NOT_SYNCED = "Your equipment changes did not reach the server. Your plan did not change. Try again when the line above says Synced.";

// planErrorText gives the text of a failed plan request (D-230, D-240).
// The server gives UNAVAILABLE after the last failed call, and only after
// progress events. With no progress event, UNAVAILABLE comes from the
// network, so the text says that the API did not answer. A failed
// exclusion first says that the exercise is not excluded (D-234).
export function planErrorText(
  err: unknown,
  { exclude = false, sawProgress = false, isOnline }: { exclude?: boolean; sawProgress?: boolean; isOnline?: boolean } = {},
): string {
  const text = baseText(err, sawProgress, isOnline);
  return exclude ? `${NOT_EXCLUDED} ${text}` : text;
}

function baseText(err: unknown, sawProgress: boolean, isOnline: boolean | undefined): string {
  if (err instanceof InventoryNotSyncedError) return NOT_SYNCED;
  const code = ConnectError.from(err).code;
  if (code === Code.Unavailable && sawProgress && isOnline !== false) return NO_VALID_PLAN;
  if (isNoConnection(err, isOnline)) return "No connection. Try again when the phone is online.";
  switch (code) {
    case Code.ResourceExhausted:
      return CAP_REACHED;
    case Code.FailedPrecondition:
      return NO_EXERCISE;
    case Code.Aborted:
      return "The exclusions changed during the request. Your plan did not change. Try again.";
    case Code.DeadlineExceeded:
      return "The request took too long. Your plan did not change. Try again.";
    case Code.InvalidArgument:
      return "The server did not accept the request. Your plan did not change.";
    case Code.PermissionDenied:
      return "This account is not on the allowlist.";
    case Code.Unauthenticated:
      return "The API did not accept the sign-in.";
    case Code.Internal:
      return "The server failed. Your plan did not change.";
    default:
      return "The API did not answer.";
  }
}

// setText gives one target set. A calibration set has no RIR target
// (D-150).
export function setText(s: Pick<PlannedSet, "reps" | "loadTenthLb" | "rirTarget">, calibration: boolean): string {
  const base = `${s.reps} ${s.reps === 1 ? "rep" : "reps"} at ${formatPounds(s.loadTenthLb)} lb`;
  if (calibration) return base;
  return `${base}, ${s.rirTarget} ${s.rirTarget === 1 ? "rep" : "reps"} in reserve`;
}

// restText gives the rest between sets, such as "1 min 30 s".
export function restText(seconds: number): string {
  if (seconds < 60) return `${seconds} s`;
  const min = Math.floor(seconds / 60);
  const s = seconds % 60;
  return s === 0 ? `${min} min` : `${min} min ${s} s`;
}

// overrideErrorText gives the text of a failed change of an override
// (D-293). With no connection, nothing is saved: an override needs the
// policy check of the server (D-23).
export function overrideErrorText(err: unknown, isOnline?: boolean): string {
  if (isNoConnection(err, isOnline)) return "No connection. The change is not saved. Try again when the phone is online.";
  switch (ConnectError.from(err).code) {
    case Code.InvalidArgument:
      return "The policy did not accept the change. Each set needs 6 to 20 reps and a weight of the machine, and the reason can not be empty.";
    case Code.FailedPrecondition:
      return "The plan does not hold this exercise now. Read the plan again.";
    case Code.Aborted:
      return "A new plan replaced the plan. Read the plan again.";
    case Code.PermissionDenied:
      return "This account is not on the allowlist.";
    case Code.Unauthenticated:
      return "The API did not accept the sign-in.";
    case Code.Internal:
      return "The server failed. The change is not saved.";
    default:
      return "The API did not answer. The change is not saved.";
  }
}

// OverrideDraft is one working set of an override form, as text.
export type OverrideDraft = { reps: string; pounds: string };

// overrideSets reads the sets of an override form. It gives null when a
// set has no whole count of reps or no load in pounds. The server checks
// the bounds (D-23, D-293).
export function overrideSets(drafts: readonly OverrideDraft[]): { reps: number; loadTenthLb: number }[] | null {
  const out: { reps: number; loadTenthLb: number }[] = [];
  for (const d of drafts) {
    const reps = /^\d{1,2}$/.test(d.reps.trim()) ? Number(d.reps.trim()) : NaN;
    const load = parsePounds(d.pounds);
    if (!Number.isInteger(reps) || reps <= 0 || load === null) return null;
    out.push({ reps, loadTenthLb: load });
  }
  return out;
}
