# Task 4 — `useSession()`

**Goal.** One hook that answers "is there a session, and whose". Every gate in the app reads this
and nothing else.

## Why it has to be a network call

The JWT is `HttpOnly`, so JavaScript can never read it. The only way the browser learns whether it
is signed in is to ask the API and inspect the answer:

| Result | Meaning |
| --- | --- |
| `GetUserProfile` resolves | logged in — and the profile (email, credits) arrives for free |
| rejects with Connect code `unauthenticated` | logged out |
| rejects with anything else | treat as logged out, but surface a toast |

This works on today's backend. The interceptor's error-flattening defect
(`tasks/Backend/auth-flow/task2.md`) only touches *authenticated* requests; an anonymous call takes
the interceptor's `return next(ctx, req)` branch, so `unauthenticated` arrives intact.

**Not Next.js middleware.** The Next server has no Ed25519 key and cannot verify the signature — it
could only check that a cookie exists, which is not authentication. Client-side gating is the
correct answer until there is a server-side session concept.

## Create `app/src/hooks/useSession.ts`

```ts
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
```

### Points that matter

- **`unauthenticated` is caught and returned as `null`, not thrown.** Being logged out is a normal
  outcome, not an error. If it throws, React Query marks the query failed and every consumer has to
  re-derive "failed how?".
- **`queryKey: ["profile"]`** — hardcode this exact key. Tasks 5 and 8 invalidate it by name, and a
  typo produces a UI that silently never updates.
- **Everything else rethrows** so a real outage is visible rather than being rendered as a
  successful logout.
- `retry: false` comes from the provider defaults in task 3; do not repeat it here.
- Profile fields are camelCased by protoc-gen-es (`firstName`, `lastName`, `email`), and
  **`credits` is a `bigint`** — `String(profile.credits)` to render it.

## Done when

- `npm run build` and `npm run lint` pass.
- Nothing imports the hook yet, so the screen is unchanged.
- Temporarily calling it from a client component shows one `GetUserProfile` request in the Network
  tab returning 401 with a Connect `unauthenticated` body while logged out. Remove the temporary
  call before committing.

## Notes

- The `error` field is returned for task 6, which surfaces it as a toast. Do not toast from inside
  the hook — it would fire on every consumer mount.
