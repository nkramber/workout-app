import { create, toJson } from "@bufbuild/protobuf";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";

import {
  GetPlanResponseSchema,
  PlanService,
  type GetPlanResponse,
  type Plan,
  type PlanProgress,
} from "../gen/workoutapp/v1/plan_service_pb";
import { db } from "./db";
import { keepCopy, syncBeforePlan } from "./sync";
import { engine } from "./sync-engine";

// The events of RequestPlan and of ExcludeExercise have the same form.
type PlanEvent = { event: { case: "progress"; value: PlanProgress } | { case: "plan"; value: Plan } | { case: undefined } };

// usePlanApi gives the two calls that make a plan. Each call reads the
// server stream: it gives each progress event to onProgress, then puts
// the plan of the last event in the query cache of GetPlan (D-237). A
// stream with no plan at its end throws. After the end of a call, the
// hook reads GetPlan again, so the screen shows the plan and the
// exclusions that the server holds, after a failure too. The signal
// stops the stream, and the server then saves nothing.
//
// Each call first runs a sync, so the plan reads each change of the
// inventory that waited in the outbox (D-272). When an inventory change
// still waits after the sync, the call stops with InventoryNotSyncedError,
// and the plan does not change. The new plan goes into the
// offline copy of the plan too (D-278).
export function usePlanApi() {
  const transport = useTransport();
  const queryClient = useQueryClient();
  return useMemo(() => {
    const client = createClient(PlanService, transport);
    const key = createConnectQueryKey({ schema: PlanService.method.getPlan, transport, input: {}, cardinality: "finite" });

    const read = async (events: AsyncIterable<PlanEvent>, onProgress: (p: PlanProgress) => void) => {
      try {
        for await (const res of events) {
          const ev = res.event;
          if (ev.case === "progress") onProgress(ev.value);
          if (ev.case === "plan") {
            const plan = ev.value;
            const next = create(GetPlanResponseSchema, {
              plan,
              exclusions: queryClient.getQueryData<GetPlanResponse>(key)?.exclusions ?? [],
            });
            queryClient.setQueryData(key, next);
            await keepCopy(db, "plan", toJson(GetPlanResponseSchema, next));
            return plan;
          }
        }
        throw new ConnectError("the stream ended with no plan", Code.Internal);
      } finally {
        void queryClient.invalidateQueries({ queryKey: key });
      }
    };

    return {
      requestPlan: async (today: string, onProgress: (p: PlanProgress) => void, signal?: AbortSignal) => {
        await syncBeforePlan(db, () => engine.syncNow());
        return read(client.requestPlan({ today }, { signal }), onProgress);
      },
      excludeExercise: async (
        today: string,
        exerciseId: string,
        reason: string,
        onProgress: (p: PlanProgress) => void,
        signal?: AbortSignal,
      ) => {
        await syncBeforePlan(db, () => engine.syncNow());
        return read(client.excludeExercise({ today, exerciseId, reason: reason.trim() }, { signal }), onProgress);
      },
    };
  }, [transport, queryClient]);
}
