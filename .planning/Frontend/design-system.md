# VULX Design System — Standing Rules

> **The palette, layout, and component specs live in `logged_out_design.md`.** That document is the
> current source of truth and supersedes the violet-accent palette this file used to carry. What is
> left here is the material that outlives any one screen.

## 1. Rules to carry forward

1. **Monochrome.** Black ground, white type, neutral greys (chroma `0`). The only saturated token is
   `--destructive`, because errors must read as errors. Introducing a hue needs a reason written
   down here first.
2. **Emphasis comes from contrast, not colour.** The primary button is white on black. That is the
   strongest signal available and it costs nothing.
3. **Separation comes from the surface ladder and alpha borders, never drop shadows.** Shadows are
   close to invisible on a black ground and cost paint.
4. Every interactive element gets a visible `focus-visible` ring.
5. `text-balance` on every headline and subhead.
6. **`src/components/ui/` is vendored** — generated shadcn plus prompt-kit. Restyle through tokens
   and `className` overrides at the call site, never by editing those files. The three overrides in
   task 4 are the pattern to follow.
7. No external image, font, or icon host beyond the Geist fonts already in `layout.tsx`.
   `lucide-react` is the icon source.
8. Forced dark via `className="dark"` on `<html>`. No `next-themes` provider until a light palette
   actually exists — leave the `:root` light block in `globals.css` in place so adding one later is
   a revision rather than a rewrite.

## 2. Typography

Geist Sans, already wired in `layout.tsx`. `font-medium` for display sizes, not `font-bold` — at
60px on black, bold reads heavy.

The Taskly reference (`.planning/images/landing-ui-ref.png`) uses a serif for its display text, and
that single choice is most of what makes it distinctive. We passed on it to keep the first build
simple. It remains the cheapest available upgrade if the screen ends up feeling generic: one
`next/font` import, applied to the headline and subhead only, leaving Geist for all functional UI.

## 3. prompt-kit: `TextShimmer` and `ThinkingBar`

**Reserved for the generation flow.** Shimmer means "the model is working" and will carry
"Generating response…" once prompts actually run. Spending it on a hero headline or a sign-in
spinner would strip that meaning before the feature needing it exists. Neither component appears on
the logged-out screen.

**Both are currently broken, and it is not obvious.** `text-shimmer.tsx` applies
`animate-[shimmer_4s_infinite_linear]`, but there is **no `@keyframes shimmer` anywhere in
`globals.css`**, and Tailwind v4 does not generate one for an arbitrary animation name. They render
completely static today — no error, no warning. Add before building the generation flow:

```css
@keyframes shimmer {
  to {
    background-position: 200% center;
  }
}
```

That pairs with the component's existing `bg-size-[200%_auto]` and its inline `animationDuration`,
which is what makes the `duration` prop take effect. `ThinkingBar` composes `TextShimmer`, so the
same block fixes both.

`ThinkingBar` also ships `onStop` / `stopLabel` props ("Answer now") — it assumes an interruptible
stream, which implies cancellation support in the API. Worth remembering when that flow is designed.
