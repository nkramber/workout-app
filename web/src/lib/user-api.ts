import { createClient } from "@connectrpc/connect";
import { useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";

import { UserService } from "../gen/workoutapp/v1/user_service_pb";
import { db } from "./db";
import { DELETE_CONFIRMATION, deleteAllData } from "./history";
import { engine } from "./sync-engine";

// useUserApi gives the deletion of all data (D-314, D-315). It is one
// direct call to the API, with no outbox, so it needs a connection.
// After the call, each query of the cache reads the server again, so no
// screen shows a deleted plan. A failed call throws, and the screen
// shows the error.
export function useUserApi() {
  const transport = useTransport();
  const queryClient = useQueryClient();
  return useMemo(() => {
    const client = createClient(UserService, transport);
    return {
      deleteAllData: async (): Promise<number> => {
        const deleted = await deleteAllData(db, {
          sync: () => engine.syncNow(),
          deleteOnServer: async () => {
            const res = await client.deleteHistory({ confirmation: DELETE_CONFIRMATION });
            return { deleted: res.deletedWorkouts, generation: res.historyGeneration };
          },
        });
        await queryClient.invalidateQueries();
        return deleted;
      },
    };
  }, [transport, queryClient]);
}
