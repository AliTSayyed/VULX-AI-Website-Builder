# Task 5 — Logged-in screen

**Goal.** `app/src/components/session/logged-in-screen.tsx` — the account email and a **Log out**
button, centred. Deliberately bare; this is a placeholder for the dashboard, not a draft of it.

Building it before the gate (task 6) means the gate has something real to render.

## Create `app/src/components/session/logged-in-screen.tsx`

```tsx
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
```

### Points that matter

- **`setQueryData(["profile"], null)` then invalidate.** The `setQueryData` flips the UI to the
  logged-out screen immediately; the invalidate refetches to confirm the server agrees. Invalidating
  alone leaves the old profile on screen for the length of the round trip.
- **No navigation on logout.** The gate in task 6 watches `["profile"]` and re-renders on its own.
  Calling `router.push("/")` from `/` would be a no-op that reads like it does something.
- **`profile` is a required prop.** The gate only renders this component when it holds one, so the
  component never has to handle a missing profile.
- Do not render credits yet. They are granted but never spent (`ARCHITECTURE.md` §10), so showing a
  number implies a feature that does not exist.
- `AccountLogout` is server-side idempotent — the handler returns success when no cookie is present
  (`handlers/account.go:57`), so a double click cannot error.

## Done when

- `npm run build` and `npm run lint` pass.
- The component is unreachable — nothing renders it until task 6. That is expected.

## Notes

- `components/session/` is a new directory. It holds post-login UI, as distinct from
  `components/landing/`.
- Until `tasks/Backend/auth-flow/task2.md` lands, a genuinely failing logout arrives as
  `invalid_argument` rather than its real code. The toast copy is intentionally generic so it stays
  honest either way.
