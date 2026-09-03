"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import type { Profile } from "@/gen/api/v1/account_service_pb";
import { useAccountService } from "@/hooks/services/useAccountService";

type LoggedInScreenProps = {
  profile: Profile;
};

export function LoggedInScreen({ profile }: LoggedInScreenProps) {
  const account = useAccountService();
  const queryClient = useQueryClient();

  const logout = useMutation({
    mutationFn: async () => {
      await account.accountLogout({});
    },
    onSuccess: () => {
      queryClient.setQueryData(["profile"], null);
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
    onError: () => {
      toast.error("Could not log out. Please try again.");
    },
  });

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col items-center justify-center gap-6 px-6">
      <p className="text-muted-foreground text-sm">{profile.email}</p>

      <Button
        size="lg"
        onClick={() => logout.mutate()}
        disabled={logout.isPending}
      >
        {logout.isPending ? "Logging out…" : "Log out"}
      </Button>
    </div>
  );
}
