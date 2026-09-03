# Tasks — Logged-Out Screen, Static Design

Spec: `.planning/Frontend/logged_out_design.md`. Do these in order; each one leaves the app in a
state that builds and can be looked at.

| # | Task | Touches |
| --- | --- | --- |
| 1 | [Tokens and shell](task1.md) | `globals.css`, `layout.tsx`, `page.tsx` |
| 2 | [Top bar](task2.md) | `components/landing/top-bar.tsx` |
| 3 | [Hero](task3.md) | `components/landing/hero.tsx` |
| 4 | [Prompt input](task4.md) | `components/landing/hero-prompt.tsx` |
| 5 | [Auth modal](task5.md) | `components/auth/auth-dialog.tsx`, `google-icon.tsx` |
| 6 | [Assemble and wire](task6.md) | `components/landing/logged-out-screen.tsx`, `page.tsx` |
| 7 | [Optional polish](task7.md) | chips + footer line |

**Ground rules for every task**

- Nothing calls the network. No `fetch`, no Connect client, no session.
- `src/components/ui/` is vendored (shadcn + prompt-kit). Never edit those files — override with
  `className` at the call site.
- After each task: `docker compose exec app npm run lint` should stay clean.
