# Task 4 — Prompt input

**Goal.** `components/landing/hero-prompt.tsx`. The focal element. Uses prompt-kit's `PromptInput`
so it is the real component from day one — but `readOnly`, because clicking it opens the modal.

## Create `app/src/components/landing/hero-prompt.tsx`

```tsx
"use client";

import { ArrowUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  PromptInput,
  PromptInputActions,
  PromptInputTextarea,
} from "@/components/ui/prompt-input";

type HeroPromptProps = {
  onAuth: (mode: "login" | "signup") => void;
};

export function HeroPrompt({ onAuth }: HeroPromptProps) {
  const open = () => onAuth("signup");

  return (
    <PromptInput
      value=""
      onValueChange={() => {}}
      onSubmit={open}
      className="bg-card border-border mt-10 w-full max-w-3xl rounded-3xl p-3"
    >
      <PromptInputTextarea
        readOnly
        placeholder="Ask VULX to build..."
        aria-label="Sign up to start building"
        onClick={open}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            open();
          }
        }}
        className="text-foreground placeholder:text-muted-foreground min-h-32 cursor-pointer bg-transparent text-base dark:bg-transparent"
      />

      <PromptInputActions className="justify-end pt-1">
        <Button
          size="icon"
          className="size-9 rounded-full"
          onClick={open}
          aria-label="Sign up to start building"
        >
          <ArrowUp className="size-4" />
        </Button>
      </PromptInputActions>
    </PromptInput>
  );
}
```

## The three overrides, and why each is required

These are not preferences. Without them the component is visibly wrong:

1. **`bg-card` on the wrapper.** `PromptInput` hardcodes `bg-background`, which under our palette is
   the exact same pure black as the page. The box would render as a floating 1px outline with no
   body. `bg-card` puts it one step up the surface ladder.
2. **`text-foreground` on the textarea.** `PromptInputTextarea` hardcodes `text-primary`. Our
   `--primary` is white, so this happens to look right *today* — but it is coupling the text colour
   to the button colour, and the moment `--primary` changes the input goes with it. Pin it.
3. **`dark:bg-transparent` on the textarea.** The underlying shadcn `Textarea` carries
   `dark:bg-input/30`. Tailwind-merge treats `dark:bg-*` and `bg-*` as **different keys**, so the
   `bg-transparent` that `PromptInputTextarea` already passes does not override it. You get a
   lighter rectangle inside the card. The explicit `dark:` variant is what wins.

## Other notes

- `readOnly`, not `disabled`. `disabled` applies `opacity-60` and removes it from the tab order;
  `readOnly` keeps it focusable and full-contrast, which is what we want for a decorative surface
  that is really a button.
- `onKeyDown` handles Enter and Space so keyboard users get the same behaviour as a click. We
  intentionally do **not** use `onFocus` — that would fire the modal just from tabbing past.
- `value=""` with a no-op `onValueChange` keeps `PromptInput` controlled and permanently empty.
  Task 6 does not lift this into state; there is nothing to hold.
- `PromptInput` mounts its own `TooltipProvider`, so no provider is needed in `layout.tsx`.
- `PromptInput`'s wrapper has its own `onClick` that focuses the textarea. Harmless — focusing a
  readOnly textarea does nothing, and our handlers fire independently.

## Done when

Compiles and lints. Rendered in task 6 — check the three overrides visually there.
