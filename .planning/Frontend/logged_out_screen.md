# Logged-Out Screen — Overview

> **Goal.** `make` brings up the whole stack; visiting `https://local.app.vulx.ai` with no cookie
> (or an invalid one) lands on a static VULX marketing screen. "Log in" and "Sign up" open a
> modal with a Google button, which completes real OAuth against our Go API. Once signed in,
> the same route renders a placeholder screen whose only content is a centred **Log out**
> button. Logging out returns to the logged-out screen.

**Branch:** `feature/UI`.

| Part | Status |
| --- | --- |
| The static screen — tokens, top bar, hero, prompt input, auth modal | ✅ built |
| The auth flow — Caddy hosts, OAuth round trip, session gate, logged-in screen | ⛔ planned, see the task trees below |

## Where the detail lives

| File | Covers |
| --- | --- |
| `logged_out_design.md` | The static screen: colour, layout, copy, component specs |
| `auth-flow.md` | The OAuth round trip, the two-host topology, cookie mechanics |
| `design-system.md` | Standing rules that outlive one screen; prompt-kit notes and known issues |
| `.planning/tasks/Frontend/logged_out_design/completed/` | How the static screen was built (done) |
| `.planning/tasks/Frontend/auth-flow/` | The auth flow, nine sequential tasks |
| `.planning/tasks/Backend/auth-flow/` | The three Go defect fixes the flow depends on |

## Decisions locked

| Question | Decision |
| --- | --- |
| Topology | **Two HTTPS hosts through Caddy.** `local.app.vulx.ai` → `app:3000`, `local.api.vulx.ai` → `api:8080`. |
| Cookie strategy | **No Go change.** `Domain=.vulx.ai; Secure; SameSite=Lax` is already correct for this topology. |
| Auth UX | Both buttons open **one shared modal**; the Google button does a **full-page redirect**. No popup. |
| Landing shape | Thin top bar, centred hero, real `PromptInput`. No side panel, no provider row. |
| Visual style | **Monochrome: black background, white text.** Forced dark; no light mode, no toggle yet. |
| Session detection | Call `GetUserProfile` and read the Connect code. The JWT is `HttpOnly`, so nothing else is possible client-side. |
| prompt-kit | `PromptInput` is used in the hero. `TextShimmer` / `ThinkingBar` are **reserved for the generation flow** and do not appear on these screens. |
| Logged-in screen | Deliberately bare: centred **Log out** button, with the account email above it. |

## Target layout

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

## Explicitly out of scope

Hero prompt draft persistence across the OAuth redirect (deferred) · light mode and a theme toggle ·
a real prompt submit / generation flow · `TextShimmer` and `ThinkingBar` usage · dashboard, editor,
preview pane · `/pricing`, `/privacy`, `/terms` · a second OAuth provider · credits spend UI ·
production DNS and certificates · any Go→AI-service work · tests.
