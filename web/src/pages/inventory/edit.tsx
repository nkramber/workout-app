import { useState } from "react";

import { MachineState } from "../../gen/workoutapp/v1/inventory_service_pb";
import {
  addWeight,
  checkDumbbells,
  checkEstimate,
  dumbbellWeights,
  formatPounds,
  guessStep,
  parsePounds,
  rangeWeights,
  removeWeight,
  type Kind,
} from "../../lib/inventory";
import { useInventoryApi } from "../../lib/inventory-api";
import type { ScreenProps } from "./types";
import { ErrorText, field, primary, secondary, Title, useAction } from "./ui";

const text = (tenths: number | undefined) => (tenths ? formatPounds(tenths) : "");

// EditScreen enters the weights and the estimates of one machine. The
// form follows the kind of the machine:
//
//   - a machine or the cable station: the lightest weight, the heaviest
//     weight, and the step make the list, then the owner adds or removes
//     single weights (D-195),
//   - the dumbbells: the dumbbell set (D-155),
//   - a cardio machine: no weights and no estimate.
//
// Each exercise of the machine gets one optional estimate (D-192), from
// the lightest to the heaviest weight (D-198). The save makes a new
// machine a draft. A change of the weights makes a confirmed machine a
// draft again, and a change of the estimates alone keeps its state
// (D-193, D-200). After the save, the review screen shows the machine.
export function EditScreen({ catalog, inventory, go, machineId }: ScreenProps & { machineId: string }) {
  const api = useInventoryApi();
  const action = useAction();
  const machine = catalog.machines.find((m) => m.id === machineId);
  const stored = inventory?.machines.find((m) => m.machineId === machineId);
  const kind = (machine?.kind ?? "cardio") as Kind;
  const exercises = catalog.exercises.filter((e) => e.machineId === machineId);

  const storedList = stored?.weightsTenthLb ?? [];
  const [lightest, setLightest] = useState(
    text(stored?.dumbbells?.lightestTenthLb ?? storedList[0]),
  );
  const [heaviest, setHeaviest] = useState(
    text(stored?.dumbbells?.heaviestTenthLb ?? storedList[storedList.length - 1]),
  );
  const [step, setStep] = useState(text(stored?.dumbbells?.stepTenthLb ?? guessStep(storedList) ?? undefined));
  const [weights, setWeights] = useState<number[]>([...storedList]);
  const [single, setSingle] = useState("");
  const [estimates, setEstimates] = useState<Record<string, string>>(() =>
    Object.fromEntries((stored?.estimates ?? []).map((e) => [e.exerciseId, formatPounds(e.loadTenthLb)])),
  );

  if (!machine) {
    return (
      <div className="space-y-4">
        <Title onBack={() => go({ name: "list" })}>Unknown machine</Title>
        <p className="text-sm text-slate-400">The catalog has no machine with this id.</p>
      </div>
    );
  }

  const stack = kind === "machine" || kind === "cable";
  const dumbbells = kind === "dumbbell" ? checkDumbbells(parsePounds(lightest), parsePounds(heaviest), parsePounds(step)) : null;
  const loads = stack ? weights : dumbbells?.ok ? dumbbellWeights({
    lightestTenthLb: dumbbells.value.lightest,
    heaviestTenthLb: dumbbells.value.heaviest,
    stepTenthLb: dumbbells.value.step,
  }) : [];

  const makeList = () => {
    const r = rangeWeights(parsePounds(lightest), parsePounds(heaviest), parsePounds(step));
    if (r.ok) {
      setWeights(r.value);
      action.setError("");
    } else action.setError(r.error);
  };

  const addSingle = () => {
    const r = addWeight(weights, parsePounds(single));
    if (r.ok) {
      setWeights(r.value);
      setSingle("");
      action.setError("");
    } else action.setError(r.error);
  };

  const save = async () => {
    if (stack && weights.length === 0) {
      action.setError("Make the list of weights first.");
      return;
    }
    if (dumbbells && !dumbbells.ok) {
      action.setError(dumbbells.error);
      return;
    }
    const chosen: { exerciseId: string; loadTenthLb: number }[] = [];
    if (kind !== "cardio") {
      for (const e of exercises) {
        const r = checkEstimate(estimates[e.id] ?? "", loads);
        if (!r.ok) {
          action.setError(`${e.name}: ${r.error}`);
          return;
        }
        if (r.value !== null) chosen.push({ exerciseId: e.id, loadTenthLb: r.value });
      }
    }
    const ok = await action.run(() =>
      api.saveMachine({
        machineId,
        weightsTenthLb: stack ? weights : [],
        dumbbells:
          dumbbells?.ok
            ? {
                lightestTenthLb: dumbbells.value.lightest,
                heaviestTenthLb: dumbbells.value.heaviest,
                stepTenthLb: dumbbells.value.step,
              }
            : undefined,
        estimates: chosen,
      }),
    );
    if (ok) go({ name: "review", machineId });
  };

  const back = () => go(stored ? { name: "review", machineId } : { name: "add" });

  return (
    <div className="space-y-6">
      <Title onBack={back}>{machine.name}</Title>

      {stored?.state === MachineState.CONFIRMED && (
        <p className="text-sm text-slate-400" data-testid="draft-warning">
          This machine is confirmed. A change of the weights makes it a draft again, and you confirm it again on the
          review screen.
        </p>
      )}

      {kind === "cardio" ? (
        <p className="text-sm text-slate-400" data-testid="no-weights">
          A cardio machine has no weights.
        </p>
      ) : (
        <section className="space-y-3" aria-label="Weights">
          <h3 className="text-sm font-semibold text-slate-300">{kind === "dumbbell" ? "Dumbbell set" : "Weights"}</h3>
          <div className="grid grid-cols-3 items-end gap-2">
            <PoundField label="Lightest (lb)" value={lightest} onChange={setLightest} />
            <PoundField label="Heaviest (lb)" value={heaviest} onChange={setHeaviest} />
            <PoundField label="Step (lb)" value={step} onChange={setStep} />
          </div>
          {kind === "dumbbell" && dumbbells?.ok && (
            <p className="text-sm text-slate-400" data-testid="dumbbell-count">
              {loads.length} pairs of dumbbells.
            </p>
          )}
          {stack && (
            <>
              <button type="button" className={`${secondary} w-full`} onClick={makeList}>
                Make the list
              </button>
              <ul className="flex flex-wrap gap-2" aria-label="Weight list" data-testid="weight-list">
                {weights.map((w) => (
                  <li key={w}>
                    <button
                      type="button"
                      aria-label={`Remove ${formatPounds(w)} lb`}
                      onClick={() => setWeights(removeWeight(weights, w))}
                      className="min-h-11 rounded-lg border border-slate-700 px-3 text-sm text-slate-100 active:bg-slate-800"
                    >
                      {formatPounds(w)} <span className="text-slate-500">×</span>
                    </button>
                  </li>
                ))}
              </ul>
              {weights.length === 0 && <p className="text-sm text-slate-400">The list is empty.</p>}
              <div className="flex items-end gap-2">
                <div className="flex-1">
                  <PoundField label="One weight (lb)" value={single} onChange={setSingle} />
                </div>
                <button type="button" className={secondary} onClick={addSingle}>
                  Add
                </button>
              </div>
            </>
          )}
        </section>
      )}

      {kind !== "cardio" && exercises.length > 0 && (
        <section className="space-y-3" aria-label="Estimates">
          <h3 className="text-sm font-semibold text-slate-300">Estimates (optional)</h3>
          <p className="text-xs text-slate-500">
            The load that you think you can lift for each exercise. With no estimate, the plan starts at the lightest
            weight.
          </p>
          {exercises.map((e) => (
            <PoundField
              key={e.id}
              label={`${e.name} (lb)`}
              value={estimates[e.id] ?? ""}
              onChange={(v) => setEstimates({ ...estimates, [e.id]: v })}
            />
          ))}
        </section>
      )}

      <ErrorText>{action.error}</ErrorText>

      <button type="button" disabled={action.busy} className={`${primary} w-full`} onClick={() => void save()}>
        Save
      </button>
    </div>
  );
}

function PoundField({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="block">
      <span className="text-sm text-slate-400">{label}</span>
      <input type="text" inputMode="decimal" autoComplete="off" value={value} onChange={(e) => onChange(e.target.value)} className={field} />
    </label>
  );
}
