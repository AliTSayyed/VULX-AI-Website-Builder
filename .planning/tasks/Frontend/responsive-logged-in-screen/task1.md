# Task 1 — The shimmer keyframe and the time formatter

**Goal.** Two leaf utilities every later task depends on. No component changes.

## 1. `@keyframes shimmer` in `app/src/app/globals.css`

`components/ui/text-shimmer.tsx` applies `animate-[shimmer_4s_infinite_linear]` and animates
`background-position` over a `bg-size-[200%_auto]` gradient. **There is no `shimmer` keyframe in the
stylesheet, and Tailwind v4 does not generate one for an arbitrary animation name.** The component
has always rendered as static grey text with no error in the console — see `design-system.md` §8.

Add it next to `vx-rise` / `vx-fade`, in the same top-level position (not inside `@layer`):

```css
/* The generation flow's one motion: shimmer means "the model is working".
 * TextShimmer supplies the gradient and the 200% background size; this only
 * moves it. Without this rule the component renders static, silently. */
@keyframes shimmer {
  to {
    background-position: 200% center;
  }
}
```

Add `shimmer` to the existing `prefers-reduced-motion` block? **No.** That block zeroes
`.vx-rise` / `.vx-fade`, which are entrance animations — decorative. The shimmer is the only signal
that a multi-minute build is running; removing it would leave a reduced-motion user staring at
static text with no indication anything is happening. Leave it running.

### Done when

A scratch render of `<TextShimmer>Thinking</TextShimmer>` visibly sweeps. Remove the scratch render
before committing.

## 2. Create `app/src/lib/format-time.ts`

The fixtures hand-wrote `"2 hours ago"`. `Project.updated_at` is an RFC3339 string, so the sidebar
needs a formatter.

```ts
/*
 * Relative timestamps for the project list. Input is the RFC3339 string the API
 * sends (Project.created_at / updated_at), not a Date.
 */

const rtf = new Intl.RelativeTimeFormat("en", { numeric: "auto" });

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;

export function formatRelative(rfc3339: string): string {
  const then = new Date(rfc3339).getTime();
  if (Number.isNaN(then)) return "";

  const diff = Date.now() - then;
  if (diff < MINUTE) return "Just now";
  if (diff < HOUR) return rtf.format(-Math.floor(diff / MINUTE), "minute");
  if (diff < DAY) return rtf.format(-Math.floor(diff / HOUR), "hour");
  if (diff < WEEK) return rtf.format(-Math.floor(diff / DAY), "day");

  return new Date(then).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });
}
```

### Points that matter

- **`Number.isNaN` guard.** An empty string is a real possibility (`created_at` on an optimistic
  message that never round-tripped). `Invalid Date` would render as `"NaN minutes ago"`.
- **`numeric: "auto"`** is what produces `"yesterday"` instead of `"1 day ago"`.
- Past a week, fall back to an absolute date — `"9 days ago"` is less useful than `"Aug 31"`.
- No `"use client"`. It is a plain module; the components that call it are already client
  components.
- Do not memoise or put it on a ticking interval. The sidebar re-renders often enough, and a
  per-row timer for a demo is machinery with no payoff.

## Done when

- `npm run lint` and `npm run build` pass.
- Nothing imports `format-time.ts` yet.
