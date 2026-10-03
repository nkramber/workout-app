import { useState } from "react";

import { MachineState } from "../../gen/workoutapp/v1/inventory_service_pb";
import { dumbbellWeights, formatPounds } from "../../lib/inventory";
import { useInventoryApi } from "../../lib/inventory-api";
import type { ScreenProps } from "./types";
import { danger, ErrorText, primary, secondary, StateBadge, Title, localErrorText, useAction, WaitingBadge } from "./ui";

// ReviewScreen shows the stored machine: its identity, each weight, and
// each estimate. The owner confirms a draft here (D-193). The confirmation
// goes into the outbox with the weights that this screen shows. When the
// stored weights are different, the server refuses it, the machine shows
// as a draft again, and the line of the sync shows the refusal (D-201,
// D-273). A cardio machine has no weights, so its save confirms it
// (D-246). A cardio draft from before that rule gets a confirmation with
// no text about weights. The owner also removes the machine here.
export function ReviewScreen({ catalog, inventory, pending, go, machineId }: ScreenProps & { machineId: string }) {
  const api = useInventoryApi();
  const action = useAction(localErrorText);
  const [removing, setRemoving] = useState(false);
  const machine = catalog.machines.find((m) => m.id === machineId);
  const stored = inventory?.machines.find((m) => m.machineId === machineId);
  const back = () => go({ name: "list" });

  if (!machine || !stored) {
    return (
      <div className="space-y-4">
        <Title onBack={back}>{machine?.name ?? "Unknown machine"}</Title>
        <p className="text-sm text-slate-400">The inventory does not hold this machine.</p>
      </div>
    );
  }

  const shownWeights = [...stored.weightsTenthLb];
  const shownDumbbells = stored.dumbbells
    ? {
        lightestTenthLb: stored.dumbbells.lightestTenthLb,
        heaviestTenthLb: stored.dumbbells.heaviestTenthLb,
        stepTenthLb: stored.dumbbells.stepTenthLb,
      }
    : undefined;
  const exercises = catalog.exercises.filter((e) => e.machineId === machineId);
  const estimate = new Map(stored.estimates.map((e) => [e.exerciseId, e.loadTenthLb]));
  const draft = stored.state !== MachineState.CONFIRMED;
  const cardio = machine.kind === "cardio";

  const confirm = async () => {
    await action.run(() => api.confirmMachine({ machineId, weightsTenthLb: shownWeights, dumbbells: shownDumbbells }));
  };

  const remove = async () => {
    if (await action.run(() => api.removeMachine(machineId))) back();
  };

  return (
    <div className="space-y-6">
      <Title onBack={back}>{machine.name}</Title>

      <div className="flex items-center gap-2 text-sm">
        <span className="text-slate-400">State</span>
        <StateBadge state={stored.state} />
        {pending.machines.has(machineId) && <WaitingBadge />}
      </div>

      <section className="space-y-2" aria-label="Weights">
        <h3 className="text-sm font-semibold text-slate-300">{shownDumbbells ? "Dumbbell set" : "Weights"}</h3>
        {shownDumbbells ? (
          <p className="text-slate-100" data-testid="shown-weights">
            {formatPounds(shownDumbbells.lightestTenthLb)} to {formatPounds(shownDumbbells.heaviestTenthLb)} lb, step{" "}
            {formatPounds(shownDumbbells.stepTenthLb)} lb ({dumbbellWeights(shownDumbbells).length} pairs)
          </p>
        ) : shownWeights.length === 0 ? (
          <p className="text-slate-100" data-testid="shown-weights">
            {cardio ? "A cardio machine has no weights" : "No weights"}
          </p>
        ) : (
          <p className="break-words text-slate-100" data-testid="shown-weights">
            {shownWeights.map(formatPounds).join(", ")} lb
          </p>
        )}
      </section>

      {exercises.length > 0 && !cardio && (
        <section className="space-y-2" aria-label="Estimates">
          <h3 className="text-sm font-semibold text-slate-300">Estimates</h3>
          <ul className="space-y-1 text-sm">
            {exercises.map((e) => (
              <li key={e.id} className="flex justify-between gap-4 border-b border-slate-800 pb-1">
                <span className="text-slate-400">{e.name}</span>
                <span className="text-slate-100" data-testid={`estimate-${e.id}`}>
                  {estimate.has(e.id) ? `${formatPounds(estimate.get(e.id)!)} lb` : "No estimate"}
                </span>
              </li>
            ))}
          </ul>
        </section>
      )}

      {draft && (
        <p className="text-sm text-slate-400">
          {cardio
            ? "A plan uses this machine only after you confirm it."
            : "A plan uses this machine only after you confirm it. Read the weights above, then confirm them."}
        </p>
      )}

      <ErrorText>{action.error}</ErrorText>

      <div className="space-y-3">
        {draft && (
          <button type="button" disabled={action.busy} className={`${primary} w-full`} onClick={() => void confirm()}>
            {cardio ? "Confirm this machine" : "Confirm these weights"}
          </button>
        )}
        {!cardio && (
          <button
            type="button"
            disabled={action.busy}
            className={`${secondary} w-full`}
            onClick={() => go({ name: "edit", machineId })}
          >
            Change
          </button>
        )}
        {removing ? (
          <div className="space-y-2 rounded-lg border border-red-900 p-3">
            <p className="text-sm text-slate-100">Remove {machine.name} from the inventory?</p>
            <div className="flex gap-2">
              <button type="button" disabled={action.busy} className={`${danger} flex-1`} onClick={() => void remove()}>
                Yes, remove
              </button>
              <button type="button" className={`${secondary} flex-1`} onClick={() => setRemoving(false)}>
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <button type="button" disabled={action.busy} className={`${danger} w-full`} onClick={() => setRemoving(true)}>
            Remove the machine
          </button>
        )}
      </div>
    </div>
  );
}
