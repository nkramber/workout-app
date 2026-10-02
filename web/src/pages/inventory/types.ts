import type { GetCatalogResponse, Inventory } from "../../gen/workoutapp/v1/inventory_service_pb";

// View names the screen of the inventory that shows.
export type View =
  | { name: "list" }
  | { name: "add" }
  | { name: "edit"; machineId: string }
  | { name: "review"; machineId: string };

// ScreenProps holds what each screen gets: the catalog, the inventory
// after the last call, and the step to another screen.
export type ScreenProps = {
  catalog: GetCatalogResponse;
  inventory: Inventory | undefined;
  go: (view: View) => void;
};
