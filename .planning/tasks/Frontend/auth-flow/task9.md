# Task 9 — Verification pass

**Goal.** Confirm the whole loop against a fresh stack, and write down what changed.

Nothing here is a code change. If a check fails, it names the task to go back to.

## 1. Cold start

```bash
make nuke   # optional but it is the honest test — it also drops the Caddy CA
make
```

After a `nuke`, **re-trust the Caddy CA** (task 1, step 4). The volume is gone and so is the root
certificate.

## 2. Checklist

| # | Check | Fails → |
| --- | --- | --- |
| 1 | `curl -s https://local.api.vulx.ai/healthz` returns 200 **without `-k`** | task 1 |
| 2 | `https://local.app.vulx.ai` renders the logged-out screen, green padlock, no console errors | task 1 |
| 3 | `npm run build` and `npm run lint` pass in `app/` | the task that broke them |
| 4 | Editing `hero.tsx` hot-reloads through Caddy without a manual refresh | task 1 |
| 5 | `Log in`, `Sign up`, the prompt box and the chips all open the same modal; Escape and the overlay close it | Phase 0 regression |
| 6 | `Continue with Google` reaches a real consent screen — no `redirect_uri_mismatch` | task 1 §5/§7, task 7 |
| 7 | Approving returns to `/`, showing the email and a `Log out` button | task 8 |
| 8 | **A full refresh keeps the session** | task 2 (`credentials`) |
| 9 | DevTools → Cookies: `jwt` on `.vulx.ai`, `HttpOnly`, `Secure`, `SameSite=Lax` | task 1, task 2 |
| 10 | Exactly one `FinishAccountAuth` in the Network tab | task 8 latch |
| 11 | Cancelling at Google → callback error state with a working link home | task 8 |
| 12 | `Log out` returns to the logged-out screen; a refresh stays there | task 5 |
| 13 | Deleting the `jwt` cookie by hand and refreshing returns to the logged-out screen | task 4 |
| 14 | Adding a second cookie on `.vulx.ai` by hand does not log you out | `tasks/Backend/auth-flow/task1.md` |
| 15 | Stopping the API and reloading shows the logged-out screen plus an error toast, not a blank page | task 6 |

Check 14 is the regression test for the cookie-parse defect and **only passes once backend task 1 is
merged**. If backend work is deferred, record it as a known failure rather than deleting the row.

## 3. Update the docs in the same change

- **`ARCHITECTURE.md` §9 (Frontend)** — it currently says there is "no `/auth/callback` route, no
  login UI … no protected-route logic", that React Query "is installed but not imported anywhere",
  and that the transport "hardcodes `http://localhost:8080`" and "sends no credentials option". All
  four are now false. §9 moves from 🟡 toward ✅ for the auth surface specifically; the editor,
  preview and dashboard are still ⛔.
- **`ARCHITECTURE.md` §11** — strike whichever of #3, #4, #5 the backend tasks closed, and strike #7
  (the demo `page.tsx`), closed by task 6.
- **`ARCHITECTURE.md` §10** — the status table's "Frontend beyond a transport" row.
- **`.planning/Frontend/auth-flow.md` §5** — points at the backend tasks; note which have landed.
- **`README.md`** — the two-host setup, `/etc/hosts`, CA trust, and the new `app/.env.local`. Anyone
  cloning this repo cannot log in without those, and none of them are discoverable from the code.
- Move `.planning/tasks/Frontend/auth-flow/*` into a `completed/` subdirectory, matching what was
  done for `logged_out_design/`.

## 4. Then

Move the finished task files, commit, and open the PR against `main`. The next feature is the
generation flow, which is where `TextShimmer` and `ThinkingBar` finally get used — and where the
missing `@keyframes shimmer` in `globals.css` (`design-system.md` §3) has to be fixed first.
