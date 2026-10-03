import type { GetCatalogResponse, Inventory } from "../../gen/workoutapp/v1/inventory_service_pb";
import type { Pending } from "../../lib/inventory-api";

// View names the screen of the inventory that shows.
export type View =
  | { name: "list" }
  | { name: "add" }
  | { name: "edit"; machineId: string }
  | { name: "review"; machineId: string };

// ScreenProps holds what each screen gets: the catalog, the inventory of
// the phone, the items with a change that waits to sync, and the step to
// another screen.
export type ScreenProps = {
  catalog: GetCatalogResponse;
  inventory: Inventory | undefined;
  pending: Pending;
  go: (view: View) => void;
};
