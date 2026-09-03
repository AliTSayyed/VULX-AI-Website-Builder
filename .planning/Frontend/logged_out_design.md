# Logged-Out Screen — Static Design (Phase 0)

> **Scope.** The visual build only. Black background, white text, hero, prompt input, auth modal.
> Interactive but not functional: the modal opens and closes, nothing touches the network.
> Build order lives in `.planning/tasks/Frontend/logged_out_design/`.

Reference image: `.planning/images/landing-ui-ref.png` (Taskly AI). We follow its **structure**,
not its colours or its ornament.

## 1. What we take, and what we drop

| From the reference | Decision |
| --- | --- |
| Left icon rail | **Dropped.** No side panel when logged out. It arrives with login, and it is not built in this phase. |
| Top bar: wordmark left, actions right | **Kept.** Actions become `Log in` / `Sign up`. |
| Centered column, generous vertical air | **Kept.** This is the main thing we are copying. |
| Large serif headline + small subhead | **Structure kept, serif dropped.** Geist Sans, already wired. |
| Big rounded input as the focal element | **Kept.** The single most important element on the page. |
| Controls inside the input (`+`, sliders, paperclip) | **Dropped.** Send button only. Nothing else maps to a feature that exists. |
| Suggestion chips + footer disclaimer | **Optional**, task 7. Easy to drop. |
| Orange cloud gradient, glassmorphism | **Dropped.** Flat black. |

## 2. Colour

Black background, white text, no accent hue. Greys are pure neutral (chroma `0`) — this is a
monochrome screen on purpose, and the only saturated token is `--destructive`, kept because error
states need to read as errors.

The primary button is **white with black text**. On a black page that is the strongest possible
emphasis and it costs no new colour.

Replace the `.dark { … }` block in `app/src/app/globals.css`:

```css
.dark {
  --radius: 0.625rem;

  --background: oklch(0 0 0);
  --foreground: oklch(1 0 0);

  --card:              oklch(0.16 0 0);
  --card-foreground:   oklch(1 0 0);
  --popover:           oklch(0.16 0 0);
  --popover-foreground:oklch(1 0 0);

  --muted:             oklch(0.22 0 0);
  --muted-foreground:  oklch(0.70 0 0);
  --secondary:         oklch(0.22 0 0);
  --secondary-foreground: oklch(1 0 0);
  --accent:            oklch(0.26 0 0);
  --accent-foreground: oklch(1 0 0);

  --primary:            oklch(1 0 0);
  --primary-foreground: oklch(0 0 0);

  --border: oklch(1 0 0 / 10%);
  --input:  oklch(1 0 0 / 14%);
  --ring:   oklch(1 0 0 / 35%);

  --destructive: oklch(0.63 0.21 25);

  --chart-1: oklch(0.98 0 0);
  --chart-2: oklch(0.80 0 0);
  --chart-3: oklch(0.62 0 0);
  --chart-4: oklch(0.44 0 0);
  --chart-5: oklch(0.28 0 0);

  --sidebar:                    oklch(0.10 0 0);
  --sidebar-foreground:         oklch(1 0 0);
  --sidebar-primary:            oklch(1 0 0);
  --sidebar-primary-foreground: oklch(0 0 0);
  --sidebar-accent:             oklch(0.26 0 0);
  --sidebar-accent-foreground:  oklch(1 0 0);
  --sidebar-border:             oklch(1 0 0 / 10%);
  --sidebar-ring:               oklch(1 0 0 / 35%);
}
```

`--muted-foreground` at L 0.70 on pure black clears 4.5:1, so it is safe for the subhead. Leave the
`:root` light block untouched — deleting it makes adding light mode a rewrite instead of a revision.

Forced dark: `className="dark"` on `<html>`. No `next-themes` provider until a light palette exists.

## 3. Layout

```
┌──────────────────────────────────────────────┐
│ ◆ VULX                  [Log in] [Sign up]   │  h-14, border-b
├──────────────────────────────────────────────┤
│                                              │
│                                              │
│      VULX is your personal AI                │  text-4xl→6xl, font-medium
│         Website Creator                      │  tracking-tight, text-balance
│                                              │
│   Describe what you want. Watch it build.    │  text-lg, muted
│                                              │
│   ┌────────────────────────────────────┐     │
│   │ Ask VULX to build...               │     │  max-w-3xl, bg-card
│   │                                    │     │  rounded-3xl, min-h-40
│   │                              [ ↑ ] │     │
│   └────────────────────────────────────┘     │
│                                              │
└──────────────────────────────────────────────┘
```

- Page: `min-h-screen bg-background text-foreground flex flex-col`.
- Everything below the top bar is one centred column, `flex-1`, vertically centred, `max-w-3xl`.
- No scroll at ≥ 720px viewport height.
- Vertical rhythm: headline → subhead `mt-4`; subhead → input `mt-10`.

## 4. Copy

| Slot | Text |
| --- | --- |
| Wordmark | `VULX` |
| Headline | `VULX is your personal AI Website Creator` |
| Subhead | `Describe what you want. Watch it build itself.` |
| Placeholder | `Ask VULX to build...` |
| Buttons | `Log in` (ghost) · `Sign up` (default) |
| Modal, signup | `Create your account` / `Start building in seconds.` |
| Modal, login | `Welcome back` / `Sign in to continue.` |
| Modal button | `Continue with Google` |

## 5. Interactivity boundary

Interactive, not functional. Everything below is local React state; **nothing calls the network.**

| Action | Behaviour now | Behaviour later |
| --- | --- | --- |
| Click the input | Opens the auth modal | Focuses the textarea and lets you type |
| Click the send button | Opens the auth modal | Submits the prompt |
| Click `Log in` / `Sign up` | Opens the auth modal, heading matches | unchanged |
| Click `Continue with Google` | Nothing. Not wired. | `BeginAccountAuth` → redirect |
| Escape / overlay click | Closes the modal | unchanged |

The prompt textarea is **`readOnly`** in this phase, because "press the input, modal pops up" and
"type freely" are mutually exclusive. When auth lands, drop `readOnly` and move the modal trigger
to submit.

## 6. Out of scope

The left rail and conversations · any network call, session, or cookie · the OAuth round trip ·
`/auth/callback` · the logged-in screen and the `Hi {name}` hero · light mode · `TextShimmer` and
`ThinkingBar` (reserved for "Generating response…") · a model picker · attachments.

## 7. Known snags the tasks handle

1. `app/src/app/page.tsx` calls `createUser({name})`, a field that does not exist on the generated
   type — **the app does not compile today.** Task 1 replaces it.
2. `PromptInput`'s wrapper is `bg-background`, identical to our page ground, so the box would be an
   invisible outline. Task 4 overrides it to `bg-card`.
3. `PromptInputTextarea` hardcodes `text-primary`, which under this palette is **white on near-black
   — invisible placeholder contrast issues aside, typed text takes the primary colour**. Task 4
   overrides to `text-foreground`.
4. The underlying shadcn `Textarea` carries `dark:bg-input/30`. Tailwind-merge treats `dark:bg-*`
   and `bg-*` as different keys, so `bg-transparent` alone does **not** win. Task 4 adds an explicit
   `dark:bg-transparent`.
