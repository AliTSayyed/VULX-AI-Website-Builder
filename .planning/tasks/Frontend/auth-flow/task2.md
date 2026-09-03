# Task 2 — Transport: real base URL and credentials

**Goal.** RPCs go to `https://local.api.vulx.ai` and carry the session cookie. Without this, every
call in tasks 4–8 is anonymous no matter how well the login works.

## 1. Rewrite `app/src/hooks/services/useServiceClient.ts`

Today it hardcodes `http://localhost:8080`, derives `useBinaryFormat` from comparing that constant
against itself (always `false`), and passes no credentials.

```ts
/*
 * This file sets up the rpc connection to the backend.
 * The base URL comes from NEXT_PUBLIC_API_URL, inlined at build time.
 */

import { useMemo } from "react";
import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { GenService } from "@bufbuild/protobuf/codegenv2";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "https://local.api.vulx.ai";

export function useServiceClient<T extends GenService<any>>(
  service: T
): Client<T> {
  const transport = useMemo(
    () =>
      createConnectTransport({
        baseUrl: BASE_URL,
        useBinaryFormat: false,
        fetch: (input, init) =>
          fetch(input, { ...init, credentials: "include" }),
      }),
    []
  );
  return useMemo(() => createClient(service, transport), [service, transport]);
}
```

### Why each change

- **`credentials: "include"` is the load-bearing one.** `local.app.vulx.ai` → `local.api.vulx.ai` is
  same-*site* but cross-*origin*, and `fetch` omits cookies cross-origin by default. Without it the
  browser will neither store the `Set-Cookie` from `FinishAccountAuth` nor send the cookie on
  `GetUserProfile`, and the login will appear to succeed and then immediately "log you out".
- **`process.env.NEXT_PUBLIC_API_URL` must be written as a full static member expression.** Next
  inlines it textually at build time; destructuring `process.env` or building the key dynamically
  yields `undefined` in the browser bundle.
- **`useBinaryFormat: false` unconditionally.** JSON is what the transcoded REST routes and the
  Swagger UI already speak, and it is what makes the Network tab readable while debugging this flow.
  The old ternary was dead code.
- **Keep both `useMemo`s.** Recreating the transport per render defeats connection reuse.

## 2. Create `app/src/hooks/services/useAccountService.ts`

One-liner, copying `useUserService.ts` exactly:

```ts
import { AccountService } from "@/gen/api/v1/account_service_pb";
import { useServiceClient } from "./useServiceClient";

export function useAccountService() {
  return useServiceClient(AccountService);
}
```

## Done when

- `npm run build` and `npm run lint` pass in `app/`.
- The screen is visually identical — nothing calls these hooks yet.
- In the browser console on `https://local.app.vulx.ai`:

  ```js
  await fetch("https://local.api.vulx.ai/healthz").then(r => r.status)   // 200
  ```

  A CORS error here means step 5 of task 1 (`APP_URL` in `.api-env`) is missing or the API was not
  restarted after the change.

## Notes

- `useUserService.ts` picks up the new base URL for free. It is currently unused; leave it.
- `ARCHITECTURE.md` §9 says the transport "sends no credentials option, which will need attention
  when cookie-authenticated calls are wired up" — this task is that attention. Update that sentence
  in `ARCHITECTURE.md` when the branch merges.
