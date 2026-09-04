# Auth Flow — Backend

> Companion to `.planning/Frontend/auth-flow.md`. That document owns the round trip, the two-host
> topology, and the browser side. **This one owns the Go API.** Verified against the code on
> `feature/UI` (2026-09-03).

## 1. What the backend already does correctly

The Google OAuth path is built and works end to end. Nothing in the login flow needs new Go code —
the work below is three defect fixes, not features.

| Concern | Where | Status |
| --- | --- | --- |
| `BeginAccountAuth` → state in Redis (10m TTL) → Google consent URL | `oauth_service.go:69` | ✅ |
| `FinishAccountAuth` → state check → code exchange → profile → user reconcile → JWT | `handlers/account.go:42`, `account_service.go` | ✅ |
| Cookie written as `HttpOnly; Secure; SameSite=Lax; Domain=.vulx.ai; Path=/` | `adapters/auth/auth.go:91` | ✅ **no change needed** |
| `AccountLogout` → Redis denylist + `ClearJWTCookie` | `handlers/account.go:56` | ✅ |
| `GetUserProfile` → `authAdapter.User(ctx)`, returns `CodeUnauthenticated` when anonymous | `handlers/account.go:72` | ✅ |
| CORS allowlist `[API_URL, APP_URL]` with `AllowCredentials: true` | `adapters/security/security.go` | ✅ config-only |

### 1.1 Why the cookie writer needs no change

`SetJWTCookie` carries a `// TODO MUST UPDATE DOMAIN` comment. **That TODO is about production DNS,
not local development.** Under the two-host layout (`local.app.vulx.ai` + `local.api.vulx.ai`) every
attribute it already writes is correct:

- `Domain=.vulx.ai` — the API host domain-matches it, so the browser stores the cookie and the app
  host shares it.
- `Secure` — Caddy terminates real TLS, so the attribute is satisfiable.
- `SameSite=Lax` — both hosts sit under the registrable domain `vulx.ai`, so they are same-*site*.

It is still cross-*origin*, which is why CORS and `credentials: "include"` both stay in play. Do not
touch this function while wiring the login flow.

### 1.2 Why the frontend can detect "logged out" today

The interceptor is advisory: no valid cookie means it falls through to `return next(ctx)` with an
unmodified context, and `GetUserProfile` then returns `authAdapter.User`'s error verbatim. So an
anonymous `GetUserProfile` reaches the browser as a clean `unauthenticated`, **which is the entire
basis of the frontend session gate**, and it works before any fix below lands.

## 2. Defects — all fixed

Three, all pre-existing, all listed in `ARCHITECTURE.md` §11. Tasks now in
`.planning/tasks/Backend/auth-flow/completed/`.

| # | Task | Defect | Blocking? | Status |
| --- | --- | --- | --- | --- |
| 1 | `task1.md` | `GetJWTCookie` kept only the **last** cookie in the header (§11 #4) | Not then, but was a live regression risk | ✅ landed |
| 2 | `task2.md` | Interceptor flattened every authenticated error to `CodeInvalidArgument` (§11 #3) | No — but made logout/profile failures unreadable | ✅ landed — as a client-caused-vs-server-caused split, not a blanket pass-through (avoids leaking infra error detail to the browser) |
| 3 | `task3.md` | OAuth state was replayable for its full 10-minute TTL (§11 #5) | No — and had a **hard ordering constraint**, see below | ✅ landed, after the frontend latch |

## 3. The one cross-cutting constraint — satisfied

**Backend task 3 landed after frontend task 8**, in the required order.

Before the fix, OAuth state survived its exchange, so React 19 StrictMode's double-fired effect
called `FinishAccountAuth` twice and *both* calls succeeded. Once the Redis delete landed, a second
call would fail on a consumed state — but the frontend callback route
(`tasks/Frontend/auth-flow/completed/task8.md`) already latches its effect with a `useRef`, so the
second StrictMode invocation never reaches `FinishAccountAuth` at all. No phantom bug materialized.

Everything else on the backend was independent of the frontend order.

## 4. Explicitly out of scope

Role column to replace the hardcoded superuser email (§11 #9) · renaming `services/cahce.go` (§11
#10) · the Go→AI-service defects (§11 #1, #2) · production cookie domain · a second OAuth provider ·
tests.
