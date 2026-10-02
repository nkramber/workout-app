import { create, type MessageInitShape } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";

import { GetProfileResponseSchema, ProfileService, type ProfileSchema } from "../gen/workoutapp/v1/profile_service_pb";

// useProfileApi gives the save of the profile. The save is one direct call
// to the API, with no offline copy and no outbox, as for the inventory
// (D-196). The answer holds the stored profile, so the hook puts it in the
// query cache of GetProfile, and the app and the screens read it from
// there. A failed call throws, and the screen shows the error.
export function useProfileApi() {
  const transport = useTransport();
  const queryClient = useQueryClient();
  return useMemo(() => {
    const client = createClient(ProfileService, transport);
    const key = createConnectQueryKey({
      schema: ProfileService.method.getProfile,
      transport,
      input: {},
      cardinality: "finite",
    });
    return {
      saveProfile: async (profile: MessageInitShape<typeof ProfileSchema>) => {
        const res = await client.saveProfile({ profile });
        queryClient.setQueryData(key, create(GetProfileResponseSchema, { profile: res.profile }));
      },
    };
  }, [transport, queryClient]);
}
