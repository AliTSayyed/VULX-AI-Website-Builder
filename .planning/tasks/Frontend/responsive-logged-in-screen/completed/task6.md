# Task 6 — `ChatPanel` on real messages

**Goal.** The thread renders `ListMessages` instead of fixture messages, shows the generating state,
and locks the composer while a send is in flight. The panel owns its *query*; it does not own the
send.

## New props

```tsx
type ChatPanelProps = {
  /** null while the Workspace is on a draft — no project exists yet. */
  projectId: string | null;
  title: string;
  onBack: () => void;
  onSend: (input: { body: string; mode: ChatMode; provider: AiProvider }) => void;
  generating: boolean;
  /** Seeds the provider select from the open project. */
  defaultProvider?: AiProvider;
};
```

The `conversation` prop and the `Conversation` import from `mock.ts` are gone.

## What changes inside

1. `const { data: messages = [], isPending } = useMessages(projectId);` — the query is `enabled` only
   when `projectId` is non-null, so a draft simply has an empty thread.
2. Render from `Message` (`@apiv1/message_service_pb`): role is `MessageRole.USER` /
   `MessageRole.ASSISTANT`, mode is `ChatMode.BUILD` / `ChatMode.CHAT`. **Enum comparisons, not
   string comparisons** — the fixture's `m.role === "user"` no longer typechecks, which is the point.
   The two bubble styles stay exactly as they are.
3. Empty thread: keep the existing `No messages yet.` line, but only when `!isPending` and not
   generating — showing "No messages yet" under a running build is a contradiction.
4. **The generating block**, rendered after the last message whenever `generating` is true, in the
   assistant's own column so the reply lands where the shimmer was:

```tsx
{generating && (
  <div className="flex flex-col gap-1.5">
    <span className="text-muted-foreground flex items-center gap-1.5 px-1 text-[11px]">
      <Hammer className="text-accent-blue size-3" />
      Build
    </span>
    <GeneratingLine className="px-1" />
  </div>
)}
```

5. **Composer locked while generating.** `disabled` on the textarea, the submit button *and*
   `ComposerControls`. `SendMessage` is synchronous and two concurrent agent runs against one sandbox
   are undefined behaviour — this is the one input rule that is not cosmetic.
6. Submit calls `onSend({ body, mode, provider })` and clears the textarea immediately. The optimistic
   bubble from task 3 is what makes that safe.
7. **Seed the provider** from `defaultProvider` (via `providerOrDefault`). Because the value must
   change when a different project is opened, key the panel on the project in task 10
   (`<ChatPanel key={projectId ?? "draft"} …/>`) rather than syncing state in an effect.
8. **Scroll to the bottom** when `messages.length` changes and when `generating` flips true: a
   `ref` on a trailing sentinel `<div>` plus `scrollIntoView({ block: "end" })` in an effect.
   `ScrollArea` is a Radix viewport, so scrolling the sentinel is more reliable than setting
   `scrollTop` on a container you would first have to find.

## Points that matter

- **Do not fire the mutation here.** A draft has no project id, so the create-then-send sequence has
  to live one level up (task 10) where both halves can share one pending flag.
- `generating` is a prop, never derived locally. One boolean, two consumers (here and the preview).
- Assistant bodies are the agent's prose summary only — the file list is captured separately and
  never returned to the browser today (`ARCHITECTURE.md` §7.5). Render it as plain text; no markdown
  renderer is in scope.
- Message ordering comes from the backend (`ORDER BY created_at ASC, id ASC`). Do not sort.

## Expected transient

`LoggedInScreen` still passes fixture ids (`"c1"`), which do not exist in the database, so
`ListMessages` will reject and the thread will render empty until task 10 cuts the screen over. That
is expected. Pass `generating={false}` and a no-op `onSend` from the screen for now.

## Done when

- `npm run lint` and `npm run build` pass.
- Opening a fixture conversation shows an empty thread and a working composer; the Network tab shows
  one failing `ListMessages`.
- Temporarily hard-coding `generating` to `true` shows the cycling shimmer under the thread and a
  fully locked composer.
