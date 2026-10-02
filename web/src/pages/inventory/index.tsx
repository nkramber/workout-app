import { useQuery } from "@connectrpc/connect-query";
import { useState } from "react";

import { InventoryService } from "../../gen/workoutapp/v1/inventory_service_pb";
import { loadErrorText } from "../../lib/errors";
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
// The catalog comes from GetCatalog, so the server holds the one copy of
// it. Each change is a direct call to the API (D-196).
export function InventoryPage({ onBack }: { onBack: () => void }) {
  const catalog = useQuery(InventoryService.method.getCatalog, {}, { staleTime: Infinity });
  const inventory = useQuery(InventoryService.method.getInventory, {});
  const [view, setView] = useState<View>({ name: "list" });

  const error = catalog.error ?? inventory.error;
  if (error) {
    return (
      <div className="space-y-4">
        <Title onBack={onBack}>Equipment</Title>
        <ErrorText testId="load-error">{loadErrorText(error)}</ErrorText>
        <button
          type="button"
          className={secondary}
          onClick={() => {
            void catalog.refetch();
            void inventory.refetch();
          }}
        >
          Try again
        </button>
      </div>
    );
  }
  if (!catalog.data || !inventory.data) return <p className="text-slate-400">Loading…</p>;

  const props = { catalog: catalog.data, inventory: inventory.data.inventory, go: setView };
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
