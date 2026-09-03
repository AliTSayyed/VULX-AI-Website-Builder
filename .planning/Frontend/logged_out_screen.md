# Logged-Out Screen — Implementation Plan

> **Goal.** `make` brings up the whole stack; visiting `https://local.app.vulx.ai` with no cookie
> (or an invalid one) lands on a static VULX marketing screen. "Log in" and "Sign up" open a
> modal with a Google button, which completes real OAuth against our Go API. Once signed in,
> the same route renders a placeholder screen whose only content is a centred **Log out**
> button. Logging out returns to the logged-out screen.

**Status:** planned, not started. **Branch:** `feature/UI`.

## Companion documents

| File | Covers |
| --- | --- |
| `logged_out_design.md` | **Phase 0.** The static screen: colour, layout, copy, component specs |
| `auth-flow.md` | The OAuth round trip, the two-host topology, cookie mechanics, and backend changes |
| `design-system.md` | Standing rules that outlive one screen; prompt-kit notes and known issues |

## Decisions locked

| Question | Decision |
| --- | --- |
| Topology | **Two HTTPS hosts through Caddy.** `local.app.vulx.ai` → `app:3000`, `local.api.vulx.ai` → `api:8080`. |
| Cookie strategy | **No Go change.** `Domain=.vulx.ai; Secure; SameSite=Lax` is already correct for this topology. |
| Auth UX | Both buttons open **one shared modal**; the Google button does a **full-page redirect**. No popup. |
| Landing shape | Thin top bar, centred hero, real `PromptInput`. No side panel, no provider row. See `logged_out_design.md`. |
| Visual style | **Monochrome: black background, white text.** Forced dark; no light mode, no toggle yet. |
| prompt-kit | `PromptInput` is used in the hero. `TextShimmer` / `ThinkingBar` are **reserved for the generation flow** ("Generating response…") and do not appear on these screens. |
| Logged-in screen | Deliberately bare: centred **Log out** button, with the account email above it. |

### Why this topology works

Both hosts sit under the registrable domain `vulx.ai`, which makes all four constraints line up
at once — this is the thing that was broken in every other arrangement:

| Constraint | Why it passes |
| --- | --- |
| Cookie `Domain=.vulx.ai` | `local.api.vulx.ai` domain-matches `.vulx.ai`, so the browser stores it |
| Cookie `Secure` | Caddy terminates TLS, so the API is genuinely HTTPS |
| Cookie `SameSite=Lax` | "Site" ignores scheme-host-port detail below the registrable domain — app and api are same-site |
| Google redirect URI | `https://local.app.vulx.ai/auth/callback` is HTTPS on a domain you own |

It is cross-*origin* (different subdomain), so CORS still applies and the transport still needs
`credentials: "include"`. `SecurityAdapterCors` already sets `AllowCredentials: true` against an
explicit `APP_URL` allowlist, so this is an env change, not a code change.

### Target layout

```
┌──────────────────────────────────────────────┐
│ ◆ VULX                  [Log in] [Sign up]   │
├──────────────────────────────────────────────┤
│                                              │
│          Build a website by asking.          │
│     Describe it. Watch it build itself.      │
│                                              │
│   ┌────────────────────────────────────┐     │
│   │ Ask VULX to build...               │     │
│   │                              [ ↑ ] │     │
│   └────────────────────────────────────┘     │
│      (type freely; submit → auth modal)      │
│                                              │
│      OpenAI  ·  Gemini  ·  Claude            │
└──────────────────────────────────────────────┘
```

## How the session is determined

The JWT is `HttpOnly`, so JavaScript can never read it. The **only** way the frontend learns
whether it is logged in is to call `GetUserProfile` and inspect the result:

- resolves → logged in, and we already hold the profile (email, credits) for free
- rejects with Connect code `unauthenticated` → logged out
- rejects with anything else → treat as logged out, but surface a toast

This works on today's backend. The interceptor's error-flattening defect (ARCHITECTURE §11 #3)
only affects *authenticated* requests; an unauthenticated call takes the `return next(ctx)`
branch, so `GetUserProfile`'s `authAdapter.User(ctx)` error reaches the browser with its correct
`unauthenticated` code.

**Rejected alternative — Next.js middleware.** The Next server has no Ed25519 key and cannot
validate the signature; it could only check for cookie presence, which is not authentication.
Client-side gating is correct until there is a server-side session concept.

## Phases

### Phase 0 — Static design (done separately)

The whole visual build — black/white tokens, top bar, hero, prompt input, auth modal — is specced
in **`logged_out_design.md`** and broken into steps in
**`.planning/tasks/Frontend/logged_out_design/`**. Interactive but not functional: the modal opens
and closes, nothing touches the network.

Phases 3–6 below assume those components already exist and only wire them to real calls.

### Phase 1 — Infrastructure (blocking, mostly config)

1. **`Caddyfile`** — replace the single `local.vulx.ai` block:

   ```caddyfile
   {
       admin off
   }

   local.api.vulx.ai {
       reverse_proxy api:8080
       tls internal
   }

   local.app.vulx.ai {
       reverse_proxy app:3000
       tls internal
   }
   ```

   Both upstreams resolve by compose service name on the shared `vulx` network. Caddy starts
   before `app`, which is fine — `reverse_proxy` retries per request rather than at boot.

2. **`/etc/hosts`** — replace the old `local.vulx.ai` line:

   ```
   127.0.0.1 local.api.vulx.ai
   127.0.0.1 local.app.vulx.ai
   ```

3. **Trust Caddy's internal CA** — otherwise Google's redirect lands on a certificate warning
   and the OAuth round trip visibly dies. Commands in `auth-flow.md` §3.3.

4. **`../AI-Website-Builder-Secrets/.api-env`** — see `auth-flow.md` §3.4:
   `APP_URL=https://local.app.vulx.ai`, `API_URL=https://local.api.vulx.ai`,
   `REDIRECT_URL=https://local.app.vulx.ai/auth/callback`.

5. **Google Cloud Console** — add `https://local.app.vulx.ai/auth/callback` as an Authorized
   redirect URI, byte-identical to `REDIRECT_URL`.

6. **`app/.env.local`** — `NEXT_PUBLIC_API_URL=https://local.api.vulx.ai`.

**Exit criteria.** `https://local.app.vulx.ai` serves Next with a green padlock and working hot
reload; `curl https://local.api.vulx.ai/healthz` returns 200 without `-k`.

### Phase 2 — Backend fixes

Much smaller than previously planned — the cookie writer needs no change at all.

1. **`grpc/adapters/auth/auth.go:70-76`** — `GetJWTCookie` resets `token` to `""` on every
   non-`jwt` cookie, so the session survives only if `jwt` is parsed last (ARCHITECTURE §11 #4).
   `return cookie.Value` on match instead. One line, real bug, fix it now.
2. **`grpc/adapters/auth/interceptor.go:26`** — return the handler error unwrapped instead of
   re-wrapping every authenticated error as `CodeInvalidArgument` (§11 #3).
3. **`oauth_service.go:125`** — delete `provider:<state>` after a successful exchange (§11 #5).
   Pairs with the callback latch in Phase 7; see `auth-flow.md` §4.3.

### Phase 3 — App shell and transport

| File | Change |
| --- | --- |
| `app/src/hooks/services/useServiceClient.ts` | Base URL from `NEXT_PUBLIC_API_URL`, fallback `https://local.api.vulx.ai`. Custom `fetch` setting `credentials: "include"`. Drop the `binaryFormat` ternary — JSON always. |
| `app/src/hooks/services/useAccountService.ts` | **New.** One-liner wrapping `AccountService`, copying `useUserService.ts`. |
| `app/src/app/providers.tsx` | **New.** `"use client"` — `QueryClientProvider` with a module-level client (`retry: false`, `refetchOnWindowFocus: false`). React Query is installed but has never been mounted. |
| `app/src/app/layout.tsx` | Wrap children in `<Providers>` and add `<Toaster />`. The `dark` class and metadata already landed in Phase 0. |
| `app/next.config.ts` | Add `allowedDevOrigins` if the dev server complains about the non-localhost host. |

### Phase 4 — Session hook and route gate

| File | Change |
| --- | --- |
| `app/src/hooks/useSession.ts` | **New.** `useQuery({ queryKey: ["profile"] })` calling `getUserProfile({})`. Returns `{ profile, status: "loading" \| "authed" \| "anon" }`, mapping Connect `unauthenticated` to `anon` rather than an error. |
| `app/src/app/page.tsx` | **Rewrite.** Phase 0 leaves it rendering `<LoggedOutScreen />` unconditionally. Make it `"use client"`, read `useSession()`, and render splash / `<LoggedOutScreen />` / `<LoggedInScreen />`. |

Render a neutral splash while `status === "loading"` so the landing page never flashes before the
dashboard on a warm session.

### Phase 5 — Logged-out screen — wire it up

Every component already exists from Phase 0 (`logged_out_design.md`). This phase changes behaviour
only — no new files, no restyling.

| File | Change |
| --- | --- |
| `hero-prompt.tsx` | Drop `readOnly` and the `onClick`/`onKeyDown` modal triggers from the textarea. Lift `value` into real state so people can type. Move the modal trigger to `onSubmit` and the send button. Persist the draft to `sessionStorage` under `vulx.draft` so it survives the OAuth redirect. |
| `logged-out-screen.tsx` | Own the draft state alongside the existing modal state. |
| `top-bar.tsx` | Unchanged. |
| `hero.tsx` | Unchanged. |

### Phase 6 — Auth modal — wire it up

`app/src/components/auth/auth-dialog.tsx` already exists from Phase 0 with the correct markup and
copy. This phase gives the Google button a handler and adds the two states it does not yet have.

- `mode` still changes **only the heading and sub-copy**. The backend has a single create-or-login
  path (`AccountService.FinishAuth`), so there is no behavioural difference and we must not imply one.
- Handler: `beginAccountAuth({ loginProvider: LoginProvider.GOOGLE })` →
  `window.location.assign(res.loginUrl)`. Stay pending after the promise resolves; the navigation
  ends the state.
- Failure: inline destructive text **and** a sonner toast. Do not close the modal.

### Phase 7 — OAuth callback route

`app/src/app/auth/callback/page.tsx`. Detail in `auth-flow.md` §4.

- Wrap in `<Suspense>` — `useSearchParams()` requires it in Next 15 or `npm run build` fails.
- Handle `?error=access_denied` (user cancelled) before looking for `code`/`state`.
- **Latch the effect with `useRef`.** React 19 StrictMode double-fires effects; once Phase 2 item 3
  lands, the second `FinishAccountAuth` fails on a consumed state and paints an error over a
  login that actually succeeded.
- Success → `invalidateQueries({ queryKey: ["profile"] })`, then `router.replace("/")`, replaying
  any stashed draft. `replace`, not `push`, so Back does not return to a spent callback URL.
- Failure → centred card with the message and a link home.

### Phase 8 — Logged-in placeholder

`app/src/components/session/logged-in-screen.tsx` — centred column: muted email line, then a
`Log out` button. On click: `accountLogout({})` → `invalidateQueries(["profile"])`. The gate in
`page.tsx` flips back on its own; no navigation needed.

## Acceptance criteria

1. `make` starts the stack; `https://local.app.vulx.ai` renders the logged-out screen, green padlock, no console errors.
2. `npm run build` and `npm run lint` pass in `app/` (they do not today — §11 #7).
3. Hot reload works through Caddy — editing `hero.tsx` updates without a manual refresh.
4. Clicking `Log in`, `Sign up`, or submitting the prompt box opens the same modal; Escape and the overlay close it.
5. `Continue with Google` reaches a real Google consent screen for our client ID — no `redirect_uri_mismatch`.
6. Approving returns to `/auth/callback`, which lands on `/` showing the email and a Log out button.
7. **A full browser refresh keeps the session.**
8. DevTools → Application → Cookies shows `jwt` on `.vulx.ai` with `HttpOnly`, `Secure`, `SameSite=Lax`.
9. Cancelling at Google returns to a callback error state with a working link home.
10. `Log out` returns to the logged-out screen, and a refresh stays there.
11. Deleting the `jwt` cookie by hand and refreshing returns to the logged-out screen.
12. Setting a second cookie on `.vulx.ai` by hand does not log you out — the §11 #4 regression test.
13. Text typed into the hero prompt survives the OAuth round trip and is still there afterwards.

## Explicitly out of scope

Light mode and a theme toggle · a real prompt submit / generation flow · `TextShimmer` and
`ThinkingBar` usage (reserved for "Generating response…") · dashboard, editor, preview pane ·
`/pricing`, `/privacy`, `/terms` · a second OAuth provider · credits spend UI · production DNS and
certificates · any Go→AI-service work · tests (no harness exists in this repo yet).
