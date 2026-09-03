# Task 6 — The route gate

**Goal.** `/` renders the logged-out screen, the logged-in screen, or a neutral splash, decided by
`useSession()`.

## Rewrite `app/src/app/page.tsx`

Phase 0 left it rendering `<LoggedOutScreen />` unconditionally:

```tsx
"use client";

import { useEffect } from "react";
import { toast } from "sonner";
import { LoggedOutScreen } from "@/components/landing/logged-out-screen";
import { LoggedInScreen } from "@/components/session/logged-in-screen";
import { useSession } from "@/hooks/useSession";

const Page = () => {
  const { profile, status, error } = useSession();

  useEffect(() => {
    if (error) {
      toast.error("Could not reach VULX. Some features may not work.");
    }
  }, [error]);

  if (status === "loading") {
    return <div className="bg-background min-h-screen" />;
  }

  if (status === "authed" && profile) {
    return <LoggedInScreen profile={profile} />;
  }

  return <LoggedOutScreen />;
};

export default Page;
```

### Points that matter

- **The splash is an empty black screen, not a spinner.** It is visible for one round trip on a warm
  session. A spinner that flashes for 80 ms reads as jank; a black rectangle on a black-background
  app reads as the page still painting. If the wait turns out to be long enough to feel broken,
  revisit — but do not reach for `TextShimmer`, which stays reserved for the generation flow
  (`design-system.md` §3).
- **Splash first, always.** Rendering `<LoggedOutScreen />` while the query is pending means a
  logged-in user sees the marketing page flash before their dashboard on every load.
- **The `error` toast lives here, not in the hook** — one mount, one toast. The hook could have
  several consumers later.
- **`status === "authed" && profile`** narrows `Profile | null` to `Profile` for the prop. The `&&`
  is redundant at runtime (the status is derived from the same value) but it is what convinces
  TypeScript, and it costs nothing.
- `"use client"` is required — hooks and state. The page stops being server-rendered, which is fine:
  there is nothing to prerender that does not depend on the session.

## Done when

- `npm run build` and `npm run lint` pass.
- Logged out (the only state reachable so far): brief black frame, then the logged-out screen. The
  Network tab shows one `GetUserProfile` returning `unauthenticated`.
- Stopping the API and reloading shows the logged-out screen plus the error toast — not a blank
  page and not an unhandled rejection.
- The modal still opens from `Log in`, `Sign up`, the prompt box, and the suggestion chips. The
  Google button still does nothing; that is task 7.

## Notes

- `logged-out-screen.tsx`, `top-bar.tsx`, `hero.tsx`, `hero-prompt.tsx` and `suggestion-chips.tsx`
  are all untouched by this task.
- This deletes the last of the `create-next-app` scaffold from `page.tsx`, closing
  `ARCHITECTURE.md` §11 #7 (the demo page that called `createUser({name})` against a proto with no
  `name` field). Note it in the commit.
