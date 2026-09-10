# Task 4 — `providers.ts` and the shared `ComposerControls`

**Goal.** One module owns the enum↔label mapping, and the mode toggle + provider select become one
component used by both composers. Pure refactor — the Workspace looks and behaves exactly as it does
now when this is done.

## 1. Create `app/src/components/session/providers.ts`

Replaces `mock.ts`'s `PROVIDERS` (still present; deleted in task 10).

```ts
import { AiProvider } from "@/gen/api/v1/enums_pb";

/*
 * Labels diverge from wire values on purpose. The AI service's routes are the
 * path segments openai / google / anthropic — it rejects "gemini" and "claude" —
 * but nobody calls the model "google". The Go API maps the enum to the segment;
 * the browser only ever sends the enum.
 */
export const PROVIDERS: { value: AiProvider; label: string }[] = [
  { value: AiProvider.ANTHROPIC, label: "Claude" },
  { value: AiProvider.OPENAI, label: "OpenAI" },
  { value: AiProvider.GOOGLE, label: "Gemini" },
];

export const DEFAULT_PROVIDER = AiProvider.ANTHROPIC;

/** A project created before a provider was recorded, or an unknown value, must
 *  still render something selectable rather than an empty trigger. */
export function providerOrDefault(p: AiProvider | undefined): AiProvider {
  return p && p !== AiProvider.UNSPECIFIED ? p : DEFAULT_PROVIDER;
}
```

## 2. Create `app/src/components/session/composer-controls.tsx`

Lift the `ToggleGroup` and the `Select` out of `chat-panel.tsx` **verbatim** — same class strings,
same icons, same sizes. The only additions are enum values and a `disabled` prop.

```tsx
"use client";

import { Hammer, MessageSquare } from "lucide-react";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Select, ... } from "@/components/ui/select";
import { AiProvider, ChatMode } from "@/gen/api/v1/enums_pb";
import { PROVIDERS } from "./providers";

type ComposerControlsProps = {
  mode: ChatMode;
  onModeChange: (mode: ChatMode) => void;
  provider: AiProvider;
  onProviderChange: (provider: AiProvider) => void;
  disabled?: boolean;
};
```

- `ToggleGroup`'s `value` is a string, so convert at the boundary: `String(mode)` out,
  `Number(v) as ChatMode` in. Same for the `Select` and `AiProvider`. Guard the empty string —
  `ToggleGroup` emits `""` when you deselect the active item, and `Number("")` is `0`
  (`UNSPECIFIED`), which the backend rejects. Keep the existing `(v) => v && ...` guard.
- `disabled` goes on **both** controls. Task 6 passes it while a send is in flight.
- Keep the wrapper `<div className="flex min-w-0 items-center gap-3">` inside this component so both
  call sites get identical spacing.

Then edit `chat-panel.tsx` to render `<ComposerControls />` in place of the inlined pair, and change
its local state from `mock.ts`'s string unions to `ChatMode` / `AiProvider`
(`useState<ChatMode>(ChatMode.BUILD)`, `useState<AiProvider>(DEFAULT_PROVIDER)`).

## Points that matter

- **Do not restyle anything.** The overrides in place are deliberate and documented in
  `design-system.md` §6 — the `data-[state=on]:bg-transparent` on the toggle items in particular
  (selection is carried by `text-accent-blue` alone, not a fill). tailwind-merge will silently drop a
  bare override that collides with a `dark:`-prefixed base class.
- **Chat stays selectable.** `SendMessage(CHAT_MODE_CHAT)` returns `Unimplemented`, and gating it is
  out of scope for the MVP (doc §8). Build is the default.
- `mock.ts`'s `PROVIDERS` and its `Provider` / `ChatMode` string types are now unused by the
  composer but still used by the rest of the fixture; leave the file alone until task 10.

## Done when

- `npm run lint` and `npm run build` pass.
- The Workspace's composer is pixel-identical to before, and both selectors still work.
