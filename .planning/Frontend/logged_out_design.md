# Logged-Out Screen

> **Scope.** Structure, layout and copy for the logged-out screen. Colour, type and surfaces are in
> `design-system.md`, which supersedes §2 below. Sign-in is wired; the prompt is not.

Structural reference: `.planning/refs/landing-ui-ref.png` (Taskly AI). Visual reference:
`.planning/refs/silk-chat-hero-preview.jsx`.

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

> **Superseded.** The flat-black monochrome palette below was replaced by the silk palette in
> `design-system.md` §3. The `.dark` block quoted here is kept only as a record of what the screen
> used to be — do not paste it back into `globals.css`. §3 (Layout), §4 (Copy) and §5
> (Interactivity boundary) of this document are still current.


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
│ [VX] VULX               [Log in] [Sign up]   │  h-16, no border, over silk
├──────────────────────────────────────────────┤
│                                              │
│         VULX is your personal                │  text-4xl→5xl, normal, dim
│         AI Website Creator                   │  text-5xl→6xl, medium
│                                              │
│   Describe what you want. Watch it build.    │  text-sm, muted
│                                              │
│   ┌────────────────────────────────────┐     │  max-w-2xl, bg-surface
│   │ Ask VULX to build a portfolio…|    │     │  rounded-3xl, opaque, min-h-21
│   │                              [ ↑ ] │     │
│   └────────────────────────────────────┘     │
│                                              │
│      ( chips )  ( chips )  ( chips )         │  text-xs, pills
└──────────────────────────────────────────────┘
```

- Page: `relative isolate min-h-screen bg-background flex flex-col`, silk behind at `-z-10`.
- Below the top bar: one centred column, `flex-1`, vertically centred, `max-w-2xl`.
- No scroll at ≥ 720px viewport height.
- Vertical rhythm: headline → subhead `mt-5`; subhead → prompt `mt-9`; prompt → chips `mt-4`.
- **One box, and nothing above it.** The prompt input is the only bordered surface. An earlier build
  wrapped it in a second panel carrying a meta line (`Describe it once…` / `• Ready`); both the
  panel and that line are gone.
- **No centre app icon.** An earlier build put the VX tile above the headline; it was removed. The
  mark appears only in the top bar.

## 4. Copy

| Slot | Text |
| --- | --- |
| Wordmark | `VULX`, after the VX monogram tile |
| Headline | `VULX is your personal AI Website Creator` |
| Subhead | `Describe what you want. Watch it build itself.` |
| Prompt hint | `Ask VULX to build ` + a cycling tail, typed via `TypingText` (not a placeholder) |
| Buttons | `Log in` (ghost pill) · `Sign up` (white pill) |
| Modal, signup | `Create your account` / `Start building in seconds.` |
| Modal, login | `Welcome back` / `Sign in to continue.` |
| Modal button | `Continue with Google` |

## 5. Interactivity boundary

The prompt is still interactive-not-functional; **auth is now live** (`BeginAccountAuth` → redirect,
plus `/auth/callback` and the session query).

| Action | Behaviour now | Behaviour later |
| --- | --- | --- |
| Click the input / send button | Opens the auth modal | Focuses the textarea; submits the prompt |
| Click a suggestion chip | Opens the auth modal | Prefills the prompt |
| Click `Log in` / `Sign up` | Opens the auth modal, heading matches | unchanged |
| Click `Continue with Google` | **Wired** — `BeginAccountAuth`, then redirect | unchanged |
| Escape / overlay click | Closes the modal | unchanged |

The prompt textarea is **`readOnly`**, because "press the input, modal pops up" and "type freely"
are mutually exclusive. When the generation flow lands, drop `readOnly`, remove the `TypingText`
overlay, and move the modal trigger to submit.

## 6. Out of scope

Light mode · `TextShimmer` and `ThinkingBar` (reserved for "Generating response…") · attachments.
The left rail, conversations, the `Welcome back` hero and the model picker all belong to the
logged-in screen — see `logged_in_design.md`.

## 7. Known snags — all resolved

The four snags this document tracked are fixed: the `createUser({name})` compile error is gone, and
the three vendored-component overrides (`PromptInput`'s `bg-background`, `PromptInputTextarea`'s
`text-primary`, and the `Textarea`'s `dark:bg-input/30`) now live as standing guidance in
`design-system.md` §6, alongside a fourth found since.
