# Frontend — Auth Flow Tasks

**Goal.** `make` brings up the stack. `https://local.app.vulx.ai` shows the logged-out screen built
in Phase 0. `Continue with Google` runs a real OAuth round trip against our Go API and comes back
signed in, at which point the same route renders a bare screen with the account email and a **Log
out** button. Logging out returns to the logged-out screen. A refresh preserves whichever state you
are in.

Design context: `.planning/Frontend/auth-flow.md` (round trip, topology, cookie mechanics).
Backend fixes live separately in `.planning/tasks/Backend/auth-flow/`.

## Tasks

| # | File | Change | Leaves the app |
| --- | --- | --- | --- |
| 1 | `task1.md` | **Infrastructure.** Caddy two hosts, `/etc/hosts`, CA trust, env vars, Google Console | serving over HTTPS on both hosts |
| 2 | `task2.md` | Transport: real base URL + `credentials: "include"`, `useAccountService` | compiling, unchanged visually |
| 3 | `task3.md` | `providers.tsx` — mount `QueryClientProvider` and `<Toaster />` | compiling, unchanged visually |
| 4 | `task4.md` | `useSession()` — the one source of truth for "am I logged in" | compiling; hook unused |
| 5 | `task5.md` | `logged-in-screen.tsx` — email + Log out button | compiling; screen unreachable |
| 6 | `task6.md` | `page.tsx` — the gate: splash / logged-out / logged-in | gated, still no way to log in |
| 7 | `task7.md` | Auth dialog — Google button calls `BeginAccountAuth` and redirects | able to reach Google |
| 8 | `task8.md` | `/auth/callback` — exchange, invalidate, redirect home | **logging in end to end** |
| 9 | `task9.md` | Verification pass — the full manual checklist | done |

## Order

Strictly sequential. Each task leaves `npm run build` passing, so a failure is always attributable
to the task you just finished rather than to something three steps back.

Task 1 is a hard prerequisite for everything — without the two Caddy hosts there is no HTTPS origin
for the cookie to be stored against, and no valid Google redirect URI.

**One cross-tree constraint:** `tasks/Backend/auth-flow/task3.md` (OAuth state delete) must not land
before task 8 here. Details in `.planning/Backend/auth-flow.md` §3.

## Standing rules

- `src/components/ui/` is vendored (shadcn + prompt-kit). Restyle at the call site, never by editing
  those files — `.planning/Frontend/design-system.md` §1.6.
- Monochrome, forced dark. No new hue without writing down why.
- `TextShimmer` / `ThinkingBar` stay reserved for the generation flow. A sign-in spinner does not
  get to spend them.

## Out of scope

Hero prompt draft persistence across the redirect (**deliberately deferred** — `hero-prompt.tsx`
stays exactly as Phase 0 built it) · light mode · generation flow · dashboard, editor, preview ·
`/pricing`, `/privacy`, `/terms` · a second OAuth provider · credits spend UI · production DNS ·
tests.
