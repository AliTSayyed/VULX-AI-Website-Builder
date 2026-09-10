# Task 7 — `HomeView` gets the two selectors

**Goal.** The Home prompt box carries the same mode toggle and model select as the Workspace
composer, and its submit hands all three values up. Home and the composer become the same control.

## Why

Home's submit is what runs `CreateProject(first_prompt, provider)` (task 10), so the provider has to
be chosen *before* the project exists — it cannot be picked afterwards in the Workspace, because the
project is created with one. Mode comes along for the same trip: the first send carries it.

## Changes to `app/src/components/session/home-view.tsx`

```tsx
type HomeViewProps = {
  name: string;
  onStart: (input: { prompt: string; mode: ChatMode; provider: AiProvider }) => void;
  disabled?: boolean;
};
```

- Local state for `mode` (default `ChatMode.BUILD`) and `provider` (`DEFAULT_PROVIDER`).
- `PromptInputActions` changes from `justify-end` to `justify-between`, with `<ComposerControls />`
  on the left and the existing submit button on the right — the same arrangement as the Workspace
  composer, which is the whole point.
- `submit()` trims and calls `onStart({ prompt, mode, provider })`. Keep the "empty prompt does
  nothing" guard.
- `disabled` (passed while a create/send is in flight) locks the textarea, the button and the
  controls.

Everything else stays: the two-tier `Welcome back,` headline, `vx-rise` with the `200ms` delay on the
prompt panel, `PROMPT_BOX` + `PROMPT_TEXTAREA` from `prompt-styles.ts`, the `border-accent-blue`
override, no `TypingText`.

## Points that matter

- **Keep the box the same size.** `PROMPT_TEXTAREA` sets `min-h-21`; the controls row adds height
  below it. That is fine — do not shrink the textarea to compensate. Home's box is deliberately
  larger than the Workspace's (`min-h-16`); that difference is intentional and stays.
- **The submit button stays `size-9`** here, not the composer's `size-8`. Same reason.
- `LoggedInScreen`'s `start()` must be updated to the new signature in this task or the build breaks.
  It can keep fabricating a local conversation and ignore `mode`/`provider` for now — task 10
  replaces the body.
- No `SuggestionChips`, no `PROMPTS` cycle. Those are logged-out devices (`logged_in_design.md` §1).

## Done when

- `npm run lint` and `npm run build` pass.
- Home shows the prompt box with the toggle and select beneath it, visually matching the Workspace
  composer's controls, and submitting still opens the Workspace as before.
