# VULX Design System

Tokens live in `app/src/app/globals.css`. Screen structure and copy live in `logged_out_design.md`
(its §2 palette is superseded by this file; its §3–§5 are current) and `logged_in_design.md`.

Reference: `.planning/refs/silk-chat-hero-preview.jsx`.

## 1. Rules

1. **Near-monochrome, one hue — plus one documented accent.** Hue `245`, chroma `0.005–0.014` —
   reads as cool grey, not blue. The saturated tokens are `--destructive`, `--mint` (unused, reserved
   for a status dot), and `--accent-blue` (`#569CD6`). Any *further* hue needs a reason written here
   first. `--accent-blue` is a flat sRGB value, deliberately **not** derived from the hue-245 OKLCH
   ladder — it needs to read as a distinct accent, not another shade of the ground. It marks
   selected/active state as a colour cue layered on top of rule 2's contrast cue, rather than
   replacing it: the composer border (`chat-panel.tsx`, `home-view.tsx`, `hero-prompt.tsx`), the
   active icon in the Chat/Build mode toggle, the "Build"/"Chat" message-row icon, the
   generating row's Build icon (`generating.tsx`), the project-list row dot, and the logged-out top
   bar's "Log in" outline. See §3.
2. **Emphasis is contrast, not colour.** The primary button is near-white on dark.
3. **No translucent surfaces, no alpha on text.** Over an animated background alpha makes contrast a
   *moving target* — see §3. Every colour that carries or backs text is solid.
4. **No drop shadows, and no internal dividers.** Separation comes from the surface ladder, one
   hairline per element, and space. A panel gets a border; the rows inside it do not. Use
   `shadow-none` when overriding a vendored component that ships `shadow-xs`/`shadow-sm`.
5. Visible `focus-visible` ring on every interactive element. **Exception:** the Chat/Build mode
   toggle (`chat-panel.tsx`) drops it at explicit request (`focus-visible:ring-0`) — keyboard-focus
   users currently get no visible indicator there. Revisit before shipping if that trade-off doesn't
   hold.
6. `text-balance` on every headline and subhead.
7. **`src/components/ui/` is vendored** (shadcn + prompt-kit), as is `src/components/animate-ui/`.
   Restyle via tokens and call-site `className`, never by editing those files. See §6 for the trap.
8. No external font or icon host beyond Geist. `lucide-react` for icons; brand marks are
   hand-authored SVG in `src/components/brand/`.
9. Forced dark via `className="dark"` on `<html>`. The `:root` light block stays so light mode is a
   revision, not a rewrite.

## 2. Typography

Geist Sans throughout, wired in `layout.tsx`. `font-medium` is the heaviest weight — no `font-bold`.
`tracking-tight` on display sizes and the wordmark only.

| Role | Size | Colour |
| --- | --- | --- |
| Display line 1 | `text-4xl sm:text-5xl` `font-normal` | `text-foreground-dim` |
| Display line 2 | `text-5xl sm:text-6xl` `font-medium` | `text-foreground` |
| Dialog title | `text-xl` `font-medium` | `text-foreground` |
| Body / input | `text-[15px]` | `text-foreground` |
| Subhead | `text-sm` | `text-muted-foreground` |
| Chip | `text-xs` | `text-foreground-dim` |
| Meta, footer | `text-[11.5px]` | `text-muted-foreground` |

**The two-tier headline is the signature** — a muted first line, a brighter heavier second line under
it, from the reference's `h1` + nested `<strong>`.

## 3. Colour

Ink `#f5f4f8`, body `rgba(230,228,235,·)`, surfaces `rgba(255,255,255,0.05)` on a `0.14` hairline,
status dot `#b7f5c8` — **every alpha resolved to an opaque token.**

The weave is steel blue `rgb(102,120,144)`, recoloured from the reference's violet
`rgb(123,116,129)` and **luminance-matched** to it (Y `0.183`, L\* `49.8`). Hold that luminance on any
future recolour; picking by eye cost ~5 L\* on the first attempt, straight out of the 11.5px text.

### Surface ladder

| Token | Utility | Dark value | Use |
| --- | --- | --- | --- |
| `--surface` | `bg-surface` | `oklch(0.18 0.009 245)` | Panel and chip fill; the preview pane's generating overlay |
| `--surface-2` | `bg-surface-2` | `oklch(0.24 0.01 245)` | Hover fill; logo tile; inset chrome |
| `--hairline` | `border-hairline` | `oklch(0.42 0.01 245)` | Every panel edge |
| `--hairline-strong` | `border-hairline-strong` | `oklch(0.52 0.012 245)` | Hover edge |
| `--foreground-dim` | `text-foreground-dim` | `oklch(0.8 0.012 245)` | De-emphasised type |
| `--mint` | `bg-mint` | `oklch(0.9 0.09 155)` | **Unused.** Reserved for a status dot. |
| `--accent-blue` | `bg-accent-blue` / `text-accent-blue` / `border-accent-blue` | `#569CD6` (same in light and dark — a fixed sRGB value, not part of the ladder) | Selected/active-state marker. See rule 1. |

**These were translucent, and it was a readability bug.** Panels at `rgba(255,255,255,0.05)` over the
weave meant contrast *oscillated* as ribbons flowed under them — `muted-foreground` swung 7.06 → 2.12,
chips 9.32 → 3.45. Opaque holds them constant at **17.05 / 10.12 / 7.64** (foreground / dim / muted).

`--hairline` at `L 0.42` is what defines a panel edge: the opaque fill alone gives only 1.07–1.47
against the silk, since the panel is darker than a bright ribbon and lighter than a dark valley.
Below `L 0.34` the border disappears into the ribbon.

Do **not** add `backdrop-blur-*` — behind an opaque fill it is a no-op that still costs a paint.

`--card` and `--popover` are opaque at `L 0.18` — the same material as `--surface`, so every panel
that floats on the silk reads as one substance. A dialog stays readable over a bright ribbon.
`--sidebar` sits one step lower at `L 0.16`: the rail is chrome flush to the viewport, not a card,
and the tone step is what makes it recede while content advances.
`--primary` is near-white on near-black, so the default `Button` is already the white pill we want.

## 4. The silk background

`src/components/landing/silk-background.tsx`, a port of the reference's canvas shader.

- **Quarter-scale buffer, upscaled by the compositor.** The reference walks every second pixel of a
  full-size `ImageData` per frame; the field has no high-frequency detail, so this looks identical
  for ~16x less work. Capped at 30fps, paused on `visibilitychange`, single static frame under
  `prefers-reduced-motion`.
- Peak thread `rgb(107,126,152)`, mean intensity ≈ `0.62`.
- **The middle scrim stop is load-bearing.** The canvas vignette already crushes the weave top and
  bottom, so the footer and top bar measure 5.3–8.3 unaided. The centre band is the brightest part of
  the page *and* the only place text sits on silk with no panel under it (the hero). At the original
  `via-black/10` the subhead measured **2.38:1**; at `/55` it holds **5.19:1**. Do not lighten it.
- Renders `absolute inset-0` and measures its parent, so the parent needs `relative isolate`.
- **Used on both screens.** Dropping it from the logged-in screen made that screen read as a
  different product. Panels float on it as opaque rounded surfaces — the same language as the
  landing page's prompt box.

## 5. Motion

- **`vx-rise`** (12px up + fade, `0.9s`, `cubic-bezier(0.16, 1, 0.3, 1)`), staggered by inline
  `animationDelay`; a no-op under `prefers-reduced-motion`. Ladder on the logged-out screen:
  top bar `0` → headline `0` → subhead `200` → prompt panel `400` → chips `500` → footer `600`.
- **`vx-fade`** (opacity only, `0.22s`, `0.08s` delay) for content that swaps inside a panel whose
  size is also changing. Movement would fight the width transition underneath it, and the delay lets
  the width mostly settle so the two read as one gesture. Key the wrapper on the view so React
  remounts and replays it.
- The vendored sidebar animates width at `duration-200 ease-linear`. Retimed at the call site to
  `300ms` on the app's easing, matched to `vx-fade`, on both `[data-slot=sidebar-container]` and
  `[data-slot=sidebar-gap]` — miss the gap and the rail and the content slide apart.
- **`TypingText`** (`animate-ui/primitives/texts/typing`) cycles the prompt hint. It cannot be a
  native `placeholder` — that would render under the animation — so the textarea drops its
  placeholder, keeps `aria-label` as its accessible name, and the overlay is `aria-hidden` +
  `pointer-events-none` so clicks still reach the textarea. The overlay's `top-2 left-3 leading-6`
  mirrors the textarea's own padding and line box; change one, change the other.
  Only the tail cycles — `Ask VULX to build` is static so the hint stays readable. The loop is
  endless, so `usePrefersReducedMotion` swaps in a single static prompt when that is set; the
  `PROMPTS` list is kept distinct from the suggestion chips sitting right below it.

## 6. Overriding vendored components (the tailwind-merge trap)

tailwind-merge treats `bg-x` and `dark:bg-x` as **different keys**, so a bare override loses to a
`dark:`-prefixed base class — in the only mode we ship. Every `dark:` on the base needs a `dark:` on
the override.

| Component | Base ships | Override needs |
| --- | --- | --- |
| `Textarea` (via `PromptInputTextarea`) | `dark:bg-input/30` | `bg-transparent dark:bg-transparent` |
| `Button variant="ghost"` | `dark:hover:bg-accent/50` | `dark:hover:bg-surface-2` |
| `Button variant="outline"` | `dark:bg-input/30`, `dark:border-input`, `dark:hover:bg-input/50` | `dark:bg-surface`, `dark:border-hairline`, `dark:hover:bg-surface-2` |
| `PromptInput` wrapper | `bg-background` (same as the page ground) | an explicit fill — `bg-surface` |
| `AlertDialogContent` | `sm:max-w-lg` | `sm:max-w-sm`, to match `DialogContent`'s width — confirm modals (e.g. logout) and the auth modal must read as the same system. |
| `ToggleGroupItem` (default variant) | `data-[state=on]:bg-accent` fill | `data-[state=on]:bg-transparent` — selection is carried by `text-accent-blue` alone (see rule 1), not a background fill. |

## 7. What changed from the monochrome system

The old rules mandated pure-neutral greys on flat black and dropped glassmorphism from the Taskly
reference (`.planning/refs/landing-ui-ref.png`). We reversed the hue call — the ground now carries a
steel-blue cast the surfaces inherit.

On glassmorphism the old rules were **right, for a reason they did not state**. We adopted the silk
reference's translucent surfaces, measured them, and reverted to opaque: translucency over a *static*
ground is a fixed ratio you check once; over an *animated* one it is a ratio that moves under the
reader. The silk is what makes glass unaffordable here.

## 8. Open items

- **`TextShimmer`'s keyframe now exists.** `@keyframes shimmer` was added to `globals.css` next to
  `vx-rise`/`vx-fade`, so the component sweeps correctly. It is in use in the generation flow via
  `GeneratingLine` (`components/session/generating.tsx`), shared by the chat thread and the preview
  overlay — shimmer stays reserved for exactly that, meaning "the model is working." `ThinkingBar`
  stays unused: it ships `onStop`/`stopLabel`, which assume a cancellable stream, and `SendMessage`
  is one blocking RPC with no abort path.
