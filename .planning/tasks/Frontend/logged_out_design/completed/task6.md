# Task 6 — Assemble and wire

**Goal.** Put tasks 2–5 together, add the only state on the page, and render it. After this the
screen is done and clickable.

## 1. Create `app/src/components/landing/logged-out-screen.tsx`

```tsx
"use client";

import { useState } from "react";
import { AuthDialog, type AuthMode } from "@/components/auth/auth-dialog";
import { Hero } from "./hero";
import { HeroPrompt } from "./hero-prompt";
import { TopBar } from "./top-bar";

export function LoggedOutScreen() {
  const [authOpen, setAuthOpen] = useState(false);
  const [authMode, setAuthMode] = useState<AuthMode>("signup");

  const openAuth = (mode: AuthMode) => {
    setAuthMode(mode);
    setAuthOpen(true);
  };

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col">
      <TopBar onAuth={openAuth} />

      <main className="flex flex-1 flex-col items-center justify-center px-6 pb-24">
        <Hero />
        <HeroPrompt onAuth={openAuth} />
      </main>

      <AuthDialog open={authOpen} onOpenChange={setAuthOpen} mode={authMode} />
    </div>
  );
}
```

## 2. Replace `app/src/app/page.tsx`

```tsx
import { LoggedOutScreen } from "@/components/landing/logged-out-screen";

const Page = () => {
  return <LoggedOutScreen />;
};

export default Page;
```

## Notes

- Two pieces of state, both local. Nothing else on this page holds any.
- `pb-24` on `<main>` pulls the column slightly above true centre. Optically centred beats
  mathematically centred — dead-centre content reads as sitting low.
- `setAuthMode` before `setAuthOpen` so the heading is correct on the first paint. React batches
  both, so there is no flash either way, but the order documents the intent.
- The mode only ever changes the copy. Do not add branching behaviour here.

## Verify

```bash
docker compose exec app npm run build
docker compose exec app npm run lint
```

Then walk it in the browser at `https://local.app.vulx.ai`:

| Check | Expected |
| --- | --- |
| Page ground | Pure black, white text |
| Input box | Visibly darker-than-black card, **not** an empty outline — if it looks like just a border, override 1 in task 4 is missing |
| Inside the input | No lighter rectangle behind the placeholder — if there is one, override 3 in task 4 is missing |
| Click the input | Modal opens, heading "Create your account" |
| Click the send arrow | Same modal |
| Click `Sign up` | Same modal |
| Click `Log in` | Modal opens, heading "Welcome back" |
| Escape, overlay click, X | Modal closes |
| Tab from the top of the page | Reaches Log in → Sign up → textarea → send button, each with a visible ring |
| Enter on the focused textarea | Modal opens |
| Type in the textarea | Nothing happens — it is `readOnly` |
| `Continue with Google` | Nothing happens. Correct for this phase. |
| 375px width | Headline wraps and stays centred, no horizontal scroll |

## Done when

Every row above passes and the console is clean.
