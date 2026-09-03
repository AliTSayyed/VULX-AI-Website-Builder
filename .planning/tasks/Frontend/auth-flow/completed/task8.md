# Task 8 — The OAuth callback route

**Goal.** `app/src/app/auth/callback/page.tsx` exchanges `code`/`state` for a session and lands the
user back on `/` signed in. **This is the task that closes the loop.**

## The file

```tsx
"use client";

import { Suspense, useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { ConnectError } from "@connectrpc/connect";
import { useAccountService } from "@/hooks/services/useAccountService";

function CallbackHandler() {
  const params = useSearchParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const account = useAccountService();
  const fired = useRef(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (fired.current) return;
    fired.current = true;

    const oauthError = params.get("error");
    if (oauthError) {
      setError(
        oauthError === "access_denied"
          ? "Sign-in was cancelled."
          : `Google returned an error: ${oauthError}`
      );
      return;
    }

    const code = params.get("code");
    const state = params.get("state");
    if (!code || !state) {
      setError("Invalid sign-in link.");
      return;
    }

    account
      .finishAccountAuth({ code, state })
      .then((res) => {
        queryClient.setQueryData(["profile"], res.profile ?? null);
        queryClient.invalidateQueries({ queryKey: ["profile"] });
        router.replace("/");
      })
      .catch((err) => {
        setError(
          err instanceof ConnectError
            ? err.rawMessage
            : "Could not complete sign-in."
        );
      });
  }, [account, params, queryClient, router]);

  if (error) {
    return (
      <div className="bg-background text-foreground flex min-h-screen flex-col items-center justify-center gap-4 px-6 text-center">
        <p className="text-muted-foreground text-sm text-balance">{error}</p>
        <a href="/" className="underline underline-offset-4">
          Back to VULX
        </a>
      </div>
    );
  }

  return <div className="bg-background min-h-screen" />;
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={<div className="bg-background min-h-screen" />}>
      <CallbackHandler />
    </Suspense>
  );
}
```

## The three traps in this file

### 1. `useRef` latch — non-negotiable

React 19 StrictMode fires effects twice in development. Right now both `FinishAccountAuth` calls
succeed, because OAuth state is replayable. **The moment
`tasks/Backend/auth-flow/task3.md` lands, the second call fails on a consumed state and the error it
sets paints over a login that actually worked.**

The latch is a `useRef`, not state — a `useState` guard does not update in time to stop the second
effect in the same commit.

Land task 8 before backend task 3, or both together. Never backend task 3 alone.

### 2. `<Suspense>` — or the production build fails

`useSearchParams()` opts the route into client rendering, and Next 15 requires a boundary above it.
Without one, `npm run build` fails with a prerender error — and only at build time, so `npm run dev`
gives you no warning. That is the entire reason the search-param reader is a separate inner
component.

### 3. `router.replace`, not `push`

`push` leaves the spent callback URL in history, so Back returns to a `code`/`state` pair that is
either consumed or about to be, and the user sees an error page from pressing Back after a
successful login.

## Other points

- **`setQueryData` then `invalidateQueries`.** `FinishAccountAuthResponse` already carries the
  `Profile`, so seeding the cache means `/` renders the logged-in screen with no second round trip;
  the invalidate then reconciles with the server. Both messages embed the same `api.v1.Profile`
  type, so the assignment needs no conversion.
- **The cookie is set by this response.** `FinishAccountAuth` attaches `Set-Cookie` on the API host
  (`handlers/account.go:51`), which the browser stores for `.vulx.ai` — but only because task 2 set
  `credentials: "include"`. If you land here signed out, check that first.
- **Handle `error` before `code`.** Google sends `?error=access_denied` with no code when the user
  cancels; looking for `code` first renders "Invalid sign-in link", which misdescribes what happened.
- **`err.rawMessage`, not `err.message`.** ConnectError's `message` is prefixed with the code
  (`[invalid_argument] …`), which is noise for a user-facing card.
- The loading state is the same black screen as the gate's splash, for the same reason (task 6).

## Done when

- `npm run build` (**not just `dev`**) and `npm run lint` pass.
- Clicking through Google's consent screen returns to `/` showing the account email and a `Log out`
  button.
- **A full browser refresh keeps you signed in.**
- DevTools → Application → Cookies shows `jwt` on `.vulx.ai` with `HttpOnly`, `Secure`,
  `SameSite=Lax`.
- Cancelling at Google shows "Sign-in was cancelled." with a working link home.
- Visiting `https://local.app.vulx.ai/auth/callback` directly shows "Invalid sign-in link."
- The Network tab shows **one** `FinishAccountAuth`, not two. If you see two, the latch is wrong.

## Notes

- `app/src/app/auth/callback/page.tsx` is a new nested route; no other routing config is needed.
- Do not add a spinner or `TextShimmer` here — see `design-system.md` §3.
