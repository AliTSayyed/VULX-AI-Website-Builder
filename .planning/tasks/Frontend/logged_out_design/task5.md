# Task 5 — Auth modal

**Goal.** `components/auth/auth-dialog.tsx`. Opens, closes, looks finished. The Google button does
**nothing** — it is wired in a later phase.

## 1. Create `app/src/components/auth/google-icon.tsx`

Inline SVG. No remote asset, no `next/image` domain config, no licence question.

```tsx
export function GoogleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden focusable="false">
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.27-4.74 3.27-8.1Z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.65l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84A11 11 0 0 0 12 23Z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.11a6.6 6.6 0 0 1 0-4.22V7.05H2.18a11 11 0 0 0 0 9.9l3.66-2.84Z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1a11 11 0 0 0-9.82 6.05l3.66 2.84C6.71 7.29 9.14 5.38 12 5.38Z"
      />
    </svg>
  );
}
```

## 2. Create `app/src/components/auth/auth-dialog.tsx`

```tsx
"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { GoogleIcon } from "./google-icon";

export type AuthMode = "login" | "signup";

type AuthDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: AuthMode;
};

const COPY: Record<AuthMode, { title: string; description: string }> = {
  signup: {
    title: "Create your account",
    description: "Start building in seconds.",
  },
  login: {
    title: "Welcome back",
    description: "Sign in to continue.",
  },
};

export function AuthDialog({ open, onOpenChange, mode }: AuthDialogProps) {
  const copy = COPY[mode];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="text-xl tracking-tight">{copy.title}</DialogTitle>
          <DialogDescription>{copy.description}</DialogDescription>
        </DialogHeader>

        <Button variant="outline" size="lg" className="mt-2 w-full">
          <GoogleIcon className="size-4" />
          Continue with Google
        </Button>

        <p className="text-muted-foreground mt-2 text-center text-xs">
          By continuing you agree to our Terms and Privacy Policy.
        </p>
      </DialogContent>
    </Dialog>
  );
}
```

## Notes

- **`variant="outline"`, not the default.** The Google mark carries the recognition; a solid white
  fill fights it and makes the modal look like it has two primary actions.
- `mode` changes **only the heading and sub-copy**. There is no behavioural difference — the
  backend has a single create-or-login path — and the UI must not imply one.
- `Button` already applies `gap-2` and auto-sizes a leading `svg`, so the icon needs no wrapper.
- The dialog is fully controlled. It owns no state; task 6 owns it.
- No pending state, no error state, no `onClick`. Those land with the auth phase — leaving the
  button inert is deliberate, not an oversight.

## Done when

Compiles and lints. Opened for the first time in task 6.
