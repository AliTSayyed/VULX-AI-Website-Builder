# Task 5 — `generating.tsx`, the cycling shimmer line

**Goal.** One component that says "the model is working", used in two places (the thread in task 6,
the preview overlay in task 8) so they can never drift apart.

## Create `app/src/components/session/generating.tsx`

```tsx
"use client";

import { useEffect, useState } from "react";
import { TextShimmer } from "@/components/ui/text-shimmer";
import { cn } from "@/lib/utils";

/*
 * Shimmer means "the model is working" — design-system.md §8 reserves it for
 * exactly this flow. The words cycle because a Build takes minutes and a single
 * frozen label starts to read as a hung request after the first thirty seconds.
 *
 * The list is intentionally vague. The backend sends no progress at all
 * (SendMessage is one blocking call — Polish.md, "Async Build"), so these are
 * honest about the phase and dishonest about nothing.
 */
const PHASES = ["Thinking", "Building", "Creating"];

const INTERVAL = 2400;

export function GeneratingLine({ className }: { className?: string }) {
  const [i, setI] = useState(0);

  useEffect(() => {
    const t = setInterval(() => setI((n) => (n + 1) % PHASES.length), INTERVAL);
    return () => clearInterval(t);
  }, []);

  return (
    <TextShimmer className={cn("text-[13px]", className)} duration={3}>
      {`${PHASES[i]}…`}
    </TextShimmer>
  );
}
```

## Points that matter

- **This depends on task 1.** Without `@keyframes shimmer` in `globals.css` the text renders as flat
  grey and the whole generating state looks broken rather than alive. Verify the sweep visually, not
  by reading the code.
- **Interval, not per-word timers.** `2400ms` against `TextShimmer`'s `duration={3}` sweep means the
  word changes mid-sweep, which reads as continuous motion rather than a stutter. Do not sync them.
- **`clearInterval` in the cleanup.** This component unmounts every time a build finishes; a leaked
  interval per build is a real leak over a demo session.
- **One ellipsis character (`…`), not three dots.** Consistent with the rest of the UI copy.
- Do not add `onStop` / a cancel affordance. `components/ui/thinking-bar.tsx` ships one, and there is
  nothing to cancel — the request is a single blocking RPC with no abort path on the backend. Do not
  use `ThinkingBar` here for the same reason.
- Keep the phase list to these three words unless the demo script changes. Adding "Deploying" or
  "Testing" would claim work the agent does not do.

## Done when

- `npm run lint` and `npm run build` pass.
- A scratch render shows the words cycling **and** the gradient sweeping across each one. Remove the
  scratch render before committing.
