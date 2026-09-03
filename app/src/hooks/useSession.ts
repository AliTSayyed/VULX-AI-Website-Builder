"use client";

import { useQuery } from "@tanstack/react-query";
import { Code, ConnectError } from "@connectrpc/connect";
import type { Profile } from "@/gen/api/v1/account_service_pb";
import { useAccountService } from "./services/useAccountService";

export type SessionStatus = "loading" | "authed" | "anon";

export function useSession() {
  const account = useAccountService();

  const query = useQuery<Profile | null>({
    queryKey: ["profile"],
    queryFn: async () => {
      try {
        const res = await account.getUserProfile({});
        return res.profile ?? null;
      } catch (err) {
        if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
          return null;
        }
        throw err;
      }
    },
  });

  const status: SessionStatus = query.isPending
    ? "loading"
    : query.data
      ? "authed"
      : "anon";

  return { profile: query.data ?? null, status, error: query.error };
}
