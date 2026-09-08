# Task 3 — Providers: React Query and the toaster

**Goal.** `QueryClientProvider` and `<Toaster />` are mounted so tasks 4–8 have somewhere to hang
server state and error messages.

`@tanstack/react-query` has been a dependency since the scaffold and has never been imported. Any
description of a "React Query cache" in this repo is aspirational until this task lands.

## 1. Create `app/src/app/providers.tsx`

```tsx
"use client";

import { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            retry: false,
            refetchOnWindowFocus: false,
            staleTime: 30_000,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}
```

### Why these defaults

- **`retry: false`** — the session query treats `unauthenticated` as a *result*, not a failure.
  Retrying it three times would triple every anonymous page load for no benefit.
- **`refetchOnWindowFocus: false`** — a tab-focus refetch of the profile buys nothing here and
  makes the network log noisy while debugging the round trip.
- **`staleTime: 30_000`** — the profile does not change minute to minute. Keeps a remount from
  re-hitting the API.
- **`useState(() => new QueryClient())`, not a module-level constant.** A module-level client is
  shared across every request the Next server handles, so one user's cached profile could be
  served to another. The lazy `useState` initializer creates one client per browser session and is
  the pattern TanStack documents for the App Router. *(This deviates from the older
  `logged_out_screen.md` Phase 3 note, deliberately.)*

## 2. Wire it in `app/src/app/layout.tsx`

Wrap `children` and add the toaster. The `dark` class and the VULX metadata already landed in
Phase 0 — leave both alone.

```tsx
import { Providers } from "./providers";
import { Toaster } from "@/components/ui/sonner";

// …inside RootLayout's return:
<body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
  <Providers>{children}</Providers>
  <Toaster theme="dark" position="top-center" />
</body>
```

**Pass `theme="dark"` explicitly.** `components/ui/sonner.tsx` reads `useTheme()` from
`next-themes`, and we deliberately have no `ThemeProvider` mounted (`design-system.md` §1.8), so it
falls back to `"system"` — which renders a light toast on a black page for anyone whose OS is in
light mode. This is a call-site override, exactly the pattern §1.6 asks for; do not edit
`sonner.tsx`.

`layout.tsx` stays a server component. `Providers` carries its own `"use client"`, and `Toaster`
already has one.

## Done when

- `npm run build` and `npm run lint` pass.
- The logged-out screen renders exactly as before — no visual change.
- React Query DevTools are **not** installed. Skip them; the whole cache is one key.

## Notes

- Do not add a `ThemeProvider`. Forced dark is a decision, not an oversight — revisit only when a
  light palette actually exists.
