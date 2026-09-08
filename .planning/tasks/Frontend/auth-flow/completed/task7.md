# Task 7 — Auth dialog: begin the OAuth flow

**Goal.** `Continue with Google` calls `BeginAccountAuth` and sends the browser to Google's consent
screen.

`app/src/components/auth/auth-dialog.tsx` already has the right markup and copy from Phase 0. This
task adds a handler and the two states it lacks: pending and failed.

## Rewrite the body of `auth-dialog.tsx`

Keep `AuthMode`, the `COPY` record, the props, and the layout exactly as they are. Add:

```tsx
"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { LoginProvider } from "@/gen/api/v1/enums_pb";
import { useAccountService } from "@/hooks/services/useAccountService";
// …existing imports

export function AuthDialog({ open, onOpenChange, mode }: AuthDialogProps) {
  const copy = COPY[mode];
  const account = useAccountService();
  const [failed, setFailed] = useState(false);

  const begin = useMutation({
    mutationFn: async () => {
      const res = await account.beginAccountAuth({
        loginProvider: LoginProvider.GOOGLE,
      });
      if (!res.loginUrl) {
        throw new Error("no login url returned");
      }
      window.location.assign(res.loginUrl);
    },
    onMutate: () => setFailed(false),
    onError: () => {
      setFailed(true);
      toast.error("Could not start sign-in. Please try again.");
    },
  });

  // …Dialog / DialogContent / DialogHeader unchanged

  <Button
    variant="outline"
    size="lg"
    className="mt-2 w-full"
    onClick={() => begin.mutate()}
    disabled={begin.isPending}
  >
    <GoogleIcon className="size-4" />
    {begin.isPending ? "Redirecting…" : "Continue with Google"}
  </Button>

  {failed && (
    <p className="text-destructive mt-2 text-center text-sm">
      Something went wrong. Please try again.
    </p>
  )}
}
```

### Points that matter

- **Stay pending after the promise resolves.** `window.location.assign` starts a navigation that
  takes a moment; the mutation never settles because the page is replaced. That is intentional — the
  button stays disabled instead of flickering back to "Continue with Google" mid-navigation.
- **Do not close the dialog on failure.** Closing it hides the error and the retry in one motion.
- **`mode` still changes only the heading and sub-copy.** The backend has a single create-or-login
  path (`AccountService.FinishAuth`), so there is no behavioural difference between "Log in" and
  "Sign up" and the UI must not imply one.
- **`LoginProvider.GOOGLE` comes from `@/gen/api/v1/enums_pb`** — the generated enum, value `1`.
  Never send `UNSPECIFIED`; `loginProviderToDomain` maps it to `LoginProviderUnspecified` and the
  registry lookup fails.
- **`failed` resets in `onMutate`** so a retry clears the previous error before the new attempt.
- Both `destructive` text **and** a toast, per the design decision: the toast survives if the dialog
  is dismissed, the inline text is where the eye already is.

## Done when

- `npm run build` and `npm run lint` pass.
- Clicking `Continue with Google` lands on a real Google consent screen for our client ID.
- **No `redirect_uri_mismatch`.** If you see one, `REDIRECT_URL` in `.api-env` and the Authorized
  redirect URI in Google Cloud Console are not byte-identical (task 1, steps 5 and 7).
- Approving at Google returns to `https://local.app.vulx.ai/auth/callback?code=…&state=…`, which is
  a 404 until task 8. **That 404 is the success condition for this task.**
- Stopping the API and clicking the button shows the inline error and the toast, with the dialog
  still open.

## Notes

- `BeginAccountAuth` is a public RPC — the handler never calls `authAdapter.User(ctx)`
  (`ARCHITECTURE.md` §6.4). No cookie is needed to reach it.
- The proto annotates it `GET /api/v1/account/auth/begin` for REST clients, but the Connect client
  POSTs to `/api.v1.AccountService/BeginAccountAuth`. Both are served by the same handler through
  Vanguard; nothing to configure.
