import { useState } from "react";

import { useOfflineInventory } from "../../lib/inventory-api";
import { noCopyText } from "../../lib/sync";
import { engine, useSyncWhenMissing } from "../../lib/sync-engine";
import { AddScreen } from "./add";
import { EditScreen } from "./edit";
import { ListScreen } from "./list";
import { ReviewScreen } from "./review";
import type { View } from "./types";
import { ErrorText, secondary, Title } from "./ui";

// The screens of the equipment inventory (work area 4.1):
//
//   - list: each machine with its state, and each note,
//   - add: the catalog list by kind (D-202), and the text entry (D-191),
//   - edit: the weights and the estimates of one machine (D-192, D-195),
//   - review: the stored machine, its confirmation, and its removal
//     (D-193, D-201).
//
// The screens read the offline copies of the catalog and the inventory,
// with the changes that wait in the outbox on them, so they work with no
// connection. Each change goes into the outbox (D-250, D-272). With no
// copy yet, the page asks the sync for one.
export function InventoryPage({ onBack }: { onBack: () => void }) {
  const data = useOfflineInventory();
  const sync = useSyncWhenMissing(data === null);
  const [view, setView] = useState<View>({ name: "list" });

  if (data === undefined) return <p className="text-slate-400">Loading…</p>;
  if (data === null) {
    if (sync.running || !sync.error) return <p className="text-slate-400">Loading…</p>;
    return (
      <div className="space-y-4">
        <Title onBack={onBack}>Equipment</Title>
        <ErrorText testId="load-error">{noCopyText(sync.error)}</ErrorText>
        <button type="button" className={secondary} onClick={() => void engine.syncNow()}>
          Try again
        </button>
      </div>
    );
  }

  const props = { catalog: data.catalog, inventory: data.inventory, pending: data.pending, go: setView };
  switch (view.name) {
    case "list":
      return <ListScreen {...props} onBack={onBack} />;
    case "add":
      return <AddScreen {...props} />;
    case "edit":
      return <EditScreen key={view.machineId} {...props} machineId={view.machineId} />;
    case "review":
      return <ReviewScreen key={view.machineId} {...props} machineId={view.machineId} />;
  }
}
