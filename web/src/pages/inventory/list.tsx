import { sortByName } from "../../lib/inventory";
import { useInventoryApi } from "../../lib/inventory-api";
import type { ScreenProps } from "./types";
import { ErrorText, primary, secondary, StateBadge, Title, localErrorText, useAction, WaitingBadge, weightSummary } from "./ui";

// ListScreen shows each machine of the inventory with its state, and each
// note. The machines show A to Z by name (D-205). A tap on a machine opens
// its review screen. A plan reads the confirmed machines alone, and no
// plan reads a note (D-191, D-193). An item with a change that waits in
// the outbox shows "Waiting to sync".
export function ListScreen({ catalog, inventory, pending, go, onBack }: ScreenProps & { onBack: () => void }) {
  const api = useInventoryApi();
  const action = useAction(localErrorText);
  const names = new Map(catalog.machines.map((m) => [m.id, m.name]));
  const machines = sortByName(inventory?.machines ?? [], names);
  const notes = inventory?.notes ?? [];

  return (
    <div className="space-y-6">
      <Title onBack={onBack}>Equipment</Title>

      <button type="button" className={`${primary} w-full`} onClick={() => go({ name: "add" })}>
        Add a machine
      </button>

      <section className="space-y-2" aria-label="Machines">
        <h3 className="text-sm font-semibold text-slate-300">Machines</h3>
        {machines.length === 0 && <p className="text-sm text-slate-400">No machine yet.</p>}
        <ul className="space-y-2">
          {machines.map((m) => (
            <li key={m.machineId}>
              <button
                type="button"
                data-testid={`machine-${m.machineId}`}
                onClick={() => go({ name: "review", machineId: m.machineId })}
                className="flex min-h-11 w-full items-center justify-between gap-3 rounded-lg border border-slate-800 px-3 py-2 text-left active:bg-slate-800"
              >
                <span className="min-w-0">
                  <span className="block text-slate-100">{names.get(m.machineId) ?? m.machineId}</span>
                  <span className="block text-sm text-slate-400">{weightSummary(m)}</span>
                </span>
                <span className="flex shrink-0 flex-col items-end gap-1">
                  <StateBadge state={m.state} />
                  {pending.machines.has(m.machineId) && <WaitingBadge />}
                </span>
              </button>
            </li>
          ))}
        </ul>
      </section>

      <section className="space-y-2" aria-label="Notes">
        <h3 className="text-sm font-semibold text-slate-300">Notes</h3>
        <p className="text-xs text-slate-500">A note is a text that matched no machine of the catalog. No plan uses it.</p>
        {notes.length === 0 && <p className="text-sm text-slate-400">No note.</p>}
        <ul className="space-y-2">
          {notes.map((n) => (
            <li
              key={n.id}
              data-testid="note"
              className="flex items-center justify-between gap-3 rounded-lg border border-slate-800 px-3 py-2"
            >
              <span className="min-w-0 break-words text-slate-100">
                {n.text}
                {pending.notes.has(n.id) && (
                  <span className="ml-2 align-middle">
                    <WaitingBadge />
                  </span>
                )}
              </span>
              <button
                type="button"
                disabled={action.busy}
                className={`${secondary} shrink-0`}
                aria-label={`Remove the note ${n.text}`}
                onClick={() => void action.run(() => api.removeNote(n.id))}
              >
                Remove
              </button>
            </li>
          ))}
        </ul>
      </section>

      <ErrorText>{action.error}</ErrorText>
    </div>
  );
}
