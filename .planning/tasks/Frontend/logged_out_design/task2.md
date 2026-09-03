# Task 2 — Top bar

**Goal.** `components/landing/top-bar.tsx`. Wordmark left, two buttons right. No logic — it takes
a callback and calls it.

## Create `app/src/components/landing/top-bar.tsx`

```tsx
"use client";

import { Button } from "@/components/ui/button";

type TopBarProps = {
  onAuth: (mode: "login" | "signup") => void;
};

export function TopBar({ onAuth }: TopBarProps) {
  return (
    <header className="border-border h-14 shrink-0 border-b">
      <div className="mx-auto flex h-full max-w-6xl items-center justify-between px-6">
        <div className="flex items-center gap-2">
          <div className="bg-foreground size-5 rounded-md" aria-hidden />
          <span className="font-medium tracking-tight">VULX</span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            className="hidden sm:inline-flex"
            onClick={() => onAuth("login")}
          >
            Log in
          </Button>
          <Button size="sm" onClick={() => onAuth("signup")}>
            Sign up
          </Button>
        </div>
      </div>
    </header>
  );
}
```

## Notes

- The mark is a plain white rounded square standing in for a logo. Swap it for real artwork later
  — do not import an image or reach for an icon font.
- `Log in` hides below `sm`. `Sign up` opens the same modal anyway, so nothing is unreachable.
- `shrink-0` matters: the parent is a flex column and the bar must not compress.

## Done when

Nothing renders it yet — that is task 6. Confirm it compiles and lints:

```bash
docker compose exec app npm run lint
```
