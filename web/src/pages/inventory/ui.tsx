import { useState, type ReactNode } from "react";

import { MachineState, type InventoryMachine } from "../../gen/workoutapp/v1/inventory_service_pb";
import { changeErrorText } from "../../lib/errors";
import { formatPounds } from "../../lib/inventory";

// The shared parts of the inventory screens. The profile screen and the
// profile gate of the app use them too. The styles copy the sign-in page
// and the home screen. Each button is 44 px high or more.

export const field = "mt-1 block w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-base";
export const primary =
  "min-h-11 rounded-lg bg-sky-600 px-4 font-medium text-white active:bg-sky-700 disabled:opacity-40";
export const secondary =
  "min-h-11 rounded-lg border border-slate-700 px-4 font-medium text-slate-100 active:bg-slate-800 disabled:opacity-40";
export const danger =
  "min-h-11 rounded-lg border border-red-800 px-4 font-medium text-red-300 active:bg-red-950 disabled:opacity-40";

export function Title({ children, onBack }: { children: ReactNode; onBack?: () => void }) {
  return (
    <div className="flex items-center gap-3">
      {onBack && (
        <button type="button" onClick={onBack} className={`${secondary} shrink-0`}>
          Back
        </button>
      )}
      <h2 className="text-base font-semibold text-slate-100">{children}</h2>
    </div>
  );
}

export function StateBadge({ state }: { state: MachineState }) {
  const confirmed = state === MachineState.CONFIRMED;
  return (
    <span
      data-testid="machine-state"
      className={`shrink-0 rounded px-2 py-0.5 text-xs font-medium ${
        confirmed ? "bg-emerald-950 text-emerald-300" : "bg-amber-950 text-amber-300"
      }`}
    >
      {confirmed ? "Confirmed" : "Draft"}
    </span>
  );
}

export function ErrorText({ children, testId = "change-error" }: { children: ReactNode; testId?: string }) {
  if (!children) return null;
  return (
    <p role="alert" className="text-sm text-red-400" data-testid={testId}>
      {children}
    </p>
  );
}

// useAction runs one change at a time. It keeps the busy state, so the
// screen turns off its buttons during a call, and the error text of the
// last failed call.
export function useAction() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const run = async (fn: () => Promise<void>): Promise<boolean> => {
    setBusy(true);
    setError("");
    try {
      await fn();
      return true;
    } catch (err) {
      setError(changeErrorText(err));
      return false;
    } finally {
      setBusy(false);
    }
  };
  return { busy, error, setError, run };
}

// weightSummary gives the weights of a machine of the inventory in one
// line: the range and the count of a stack, the range and the step of the
// dumbbell set, or no weights for a cardio machine.
export function weightSummary(m: InventoryMachine): string {
  if (m.dumbbells) {
    const d = m.dumbbells;
    return `${formatPounds(d.lightestTenthLb)} to ${formatPounds(d.heaviestTenthLb)} lb, step ${formatPounds(d.stepTenthLb)} lb`;
  }
  const w = m.weightsTenthLb;
  if (w.length === 0) return "No weights";
  if (w.length === 1) return `${formatPounds(w[0])} lb, 1 weight`;
  return `${formatPounds(w[0])} to ${formatPounds(w[w.length - 1])} lb, ${w.length} weights`;
}
