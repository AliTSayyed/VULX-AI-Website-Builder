# Task 3 — Hero

**Goal.** `components/landing/hero.tsx`. Headline and subhead. Pure presentation, no props.

## Create `app/src/components/landing/hero.tsx`

```tsx
export function Hero() {
  return (
    <div className="flex flex-col items-center text-center">
      <h1 className="text-4xl font-medium tracking-tight text-balance sm:text-5xl md:text-6xl">
        VULX is your personal AI Website Creator
      </h1>
      <p className="text-muted-foreground mt-4 max-w-xl text-lg text-balance">
        Describe what you want. Watch it build itself.
      </p>
    </div>
  );
}
```

## Notes

- `font-medium`, not `font-bold`. At 60px on pure black, bold reads heavy and cheap.
- `text-balance` on both lines — without it the headline breaks into a long line and an orphan.
- No `"use client"`. This renders on the server; only components with handlers or state need it.
- The headline is long. Check it at 375px width — it should wrap to three or four lines and stay
  centred, never overflow.

## Done when

Compiles and lints. Rendered in task 6.
