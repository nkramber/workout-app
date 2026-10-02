import { create, type MessageInitShape } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";

import {
  ConfirmMachineRequestSchema,
  GetInventoryResponseSchema,
  InventoryService,
  SaveMachineRequestSchema,
  type Inventory,
} from "../gen/workoutapp/v1/inventory_service_pb";

// useInventoryApi gives the changes of the inventory. Each change is one
// direct call to the API, with no offline copy and no outbox (D-196). The
// answer of each call holds the inventory after the change, so the hook
// puts it in the query cache of GetInventory, and each screen reads it
// from there. A failed call throws, and the screen shows the error.
export function useInventoryApi() {
  const transport = useTransport();
  const queryClient = useQueryClient();
  return useMemo(() => {
    const client = createClient(InventoryService, transport);
    const key = createConnectQueryKey({
      schema: InventoryService.method.getInventory,
      transport,
      input: {},
      cardinality: "finite",
    });
    const keep = (inventory: Inventory | undefined) => {
      if (inventory) queryClient.setQueryData(key, create(GetInventoryResponseSchema, { inventory }));
    };
    return {
      saveMachine: async (req: MessageInitShape<typeof SaveMachineRequestSchema>) =>
        keep((await client.saveMachine(req)).inventory),
      confirmMachine: async (req: MessageInitShape<typeof ConfirmMachineRequestSchema>) =>
        keep((await client.confirmMachine(req)).inventory),
      removeMachine: async (machineId: string) => keep((await client.removeMachine({ machineId })).inventory),
      saveNote: async (text: string) => keep((await client.saveNote({ text })).inventory),
      removeNote: async (id: string) => keep((await client.removeNote({ id })).inventory),
      // reload reads the inventory again, for example after the server
      // refused a confirmation because the weights changed (D-201).
      reload: () => queryClient.invalidateQueries({ queryKey: key }),
    };
  }, [transport, queryClient]);
}
