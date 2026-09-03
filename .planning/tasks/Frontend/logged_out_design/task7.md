# Task 7 — Optional polish

**Optional.** The screen is complete without this. Do it if it feels bare; skip it without
consequence. Both pieces come from the reference layout.

## 1. Suggestion chips

`app/src/components/landing/suggestion-chips.tsx`:

```tsx
"use client";

const SUGGESTIONS = [
  "A portfolio for a photographer",
  "A pricing page with three tiers",
  "A landing page for a SaaS app",
];

type SuggestionChipsProps = {
  onSelect: () => void;
};

export function SuggestionChips({ onSelect }: SuggestionChipsProps) {
  return (
    <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
      {SUGGESTIONS.map((s) => (
        <button
          key={s}
          type="button"
          onClick={onSelect}
          className="border-border text-muted-foreground hover:border-ring/50 hover:text-foreground focus-visible:ring-ring rounded-full border px-4 py-2 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
        >
          {s}
        </button>
      ))}
    </div>
  );
}
```

In `logged-out-screen.tsx`, after `<HeroPrompt />`:

```tsx
<SuggestionChips onSelect={() => openAuth("signup")} />
```

They open the modal like everything else. Once the prompt is real, a chip should prefill the
textarea instead.

**Worth knowing:** a blank prompt box is the biggest drop-off point on a screen like this — people
do not know what to type. Chips are the cheapest fix. That is the argument for doing this one.

## 2. Footer line

At the end of the outer `div` in `logged-out-screen.tsx`, after `</main>`:

```tsx
<footer className="text-muted-foreground shrink-0 pb-6 text-center text-xs">
  VULX can make mistakes. Check generated code before shipping.
</footer>
```

Then drop `pb-24` from `<main>` to `pb-12` so the column does not sit too high.

## Done when

Both render, chips open the modal, and nothing overflows at 375px.
