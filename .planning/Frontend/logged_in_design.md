# Logged-In Screen

> Colour, type, surfaces and motion come from `design-system.md` — this covers structure only.
> Status markers follow `ARCHITECTURE.md`: ✅ built · 🟡 partial · ⛔ not built.

Two screens behind one shell. **Home** is where you start something; **Workspace** is where you
build it. The sidebar is the constant, and it is the thing that changes meaning between them.

## 1. Home — no conversation selected

```
┌────┬─────────────────────────────────────────────┐
│ [»]│  [VX] VULX                        [avatar]  │  same top bar, avatar replaces auth
│    ├─────────────────────────────────────────────┤
│ ●  │                                             │
│ ●  │            Welcome back,                    │  text-4xl→5xl, normal, dim
│ ●  │            Tony                             │  text-5xl→6xl, medium
│    │                                             │
│    │   ┌─────────────────────────────────────┐   │  the logged-out input, unchanged
│    │   │ Describe the site you want to build │   │  STATIC placeholder, not TypingText
│    │   │                             [ ↑ ]   │   │
│    │   └─────────────────────────────────────┘   │
│    │                                             │
└────┴─────────────────────────────────────────────┘
```

- **Reuses `HeroPrompt` verbatim** except: no `TypingText` overlay, a real `placeholder`, `readOnly`
  dropped, and submit sends instead of opening the auth modal. Both the `PROMPTS` cycle and the
  suggestion chips are logged-out devices — a returning user does not need to be told what the box
  is for, and their own history is in the rail beside it.
- **Headline is the two-tier pattern** from `design-system.md` §2: `Welcome back,` dim on line 1,
  the name bright on line 2.
- Name is `Profile.first_name` ✅. Google may not return it — fall back to the email local part,
  never render `Welcome back,` with a blank second line.
- Silk runs behind both screens, as on the landing page. On Home it is fully visible behind the
  hero; in Workspace it shows as the gutter around the preview card.

## 2. Workspace — a conversation is open

```
┌──────────────┬──────────────────────────────────┐
│ [«] Back     │ [»] preview.vulx.ai/s/ab12 ⟳  ↗  │  toggle lives in this row,
│              │                                  │  so both panels share a top edge
│  chat thread │                                  │
│  ▸ user      │                                  │
│  ▸ assistant │        <iframe>                  │  sandbox url, resizable split
│  ▸ user      │        live preview              │
│              │                                  │
│ ┌──────────┐ │                                  │
│ │ message… │ │                                  │
│ │ [model▾] │ │  provider selector               │
│ │ [chat|⚒] │ │  mode toggle                     │
│ └──────────┘ │                                  │
└──────────────┴──────────────────────────────────┘
```

- The left rail **becomes the chat thread** for the open conversation. Same panel, different
  content — not a second sidebar. `[« Back]` returns it to the conversation list.
- The rail widens (`--sidebar-width` 16rem → 26rem) to hold a thread. It is **not** draggable —
  the vendored sidebar has no resize handle. See §8.1.
- **Collapse differs by screen.** Home collapses to an icon rail; Workspace collapses `offcanvas`,
  all the way out. A chat thread has no icon representation — squeezing one into 3rem was a bug, not
  a state. `SidebarRail` plus the toggle in the preview chrome are the ways back.
- **Opening a conversation forces the rail open.** Otherwise a click from a collapsed icon rail flips
  `icon` → `offcanvas` on the same frame as the content swap, which reads as a glitch. The rail is a
  controlled `open`/`onOpenChange` pair for this reason.
- The composer reuses the same `PromptInput` shell, so the input a user learned on Home is the input
  they keep using.

## 3. The two selectors

These are not new concepts to invent — each maps to a route that already exists.

| Control | Options | Maps to |
| --- | --- | --- |
| Model | `openai` · `google` · `anthropic` | the `{provider}` path segment ✅ |
| Mode | **Chat** — talk, no writes | `POST /{provider}/query` → `{content}` ✅ |
| | **Build** — writes to the codebase | `POST /{provider}/{sandbox_id}/code` ✅ |

Label the providers by their product names (OpenAI, Gemini, Claude) but send the exact segments
`openai` / `google` / `anthropic` — the routes reject `gemini` and `claude`.

**Build mode must look consequential.** It mutates a real codebase and spends credits; Chat does
neither. Mode is the more important of the two selectors, so give it a segmented toggle
(`toggle-group.tsx`) rather than burying it in a dropdown next to the model `select.tsx`.

## 4. Sidebar

`sidebar.tsx` is vendored and does collapse/expand, keyboard shortcut and persistence already.
Use `variant="sidebar"` — **not** `floating` or `inset`.

- **The rail is ground, not a card.** Flush to the viewport left, top and bottom, with one hairline
  on its only exposed edge. `--sidebar` sits at `L 0.16`, a step below the `L 0.18` of the panels
  that float on the silk, so the chrome recedes and the content advances. `floating` insets it by
  `p-2` on all four sides, which is the gap we do not want, and it needs padding arithmetic to
  collapse cleanly; `variant="sidebar"` collapses to a flat `3rem`.
- The only gutter is between the rail and the preview card, where the silk shows through.
- Collapsed: icon rail. Expanded: conversation list, newest first, title + relative time.
- Bottom: email, credits, log out. Credits are read-only today 🟡 — display them, do not imply a
  balance that changes.

## 5. What this needs that does not exist

The screen is mostly blocked on backend work, and the blockers are load-bearing — not polish.

| Need | Status | Consequence if ignored |
| --- | --- | --- |
| Conversation / message model | ⛔ | The sidebar has nothing to list and the thread has nothing to render. This is the whole screen. |
| Sandbox persistence and reuse | ⛔ | `POST /sandbox/` returns `{id, url}` but nothing stores it. Reopening a conversation cannot restore its preview — it would spin up a fresh, empty sandbox. |
| Go API → AI service | 🟡 | Wrong URL and an ignored error that panics; `CallAI()` is an empty stub. See `ARCHITECTURE.md` §11.1–2. |
| Credit spend | 🟡 | Granted and displayed, never debited. Build mode is free today. |

**The browser cannot call the AI service directly** — its CORS allows only `http://api:8080`. Every
prompt, in both modes, goes through the Go API. Do not design a flow that assumes otherwise.

## 6. Endpoints this screen implies

Read off the static build. `Profile` and logout already exist; everything below does not.
`{provider}` is `openai` / `google` / `anthropic`.

| UI | RPC needed | Backing call |
| --- | --- | --- |
| Sidebar list | `ListConversations` → `{id, title, updated_at, preview_url}[]` | new table |
| `New build` / Home submit | `CreateConversation(first_message, mode, provider)` | new table |
| Opening a conversation | `GetConversation(id)` → conversation + messages | new table |
| Composer, **Chat** mode | `SendMessage(conversation_id, body, provider, mode=CHAT)` | `POST /{provider}/query` ✅ |
| Composer, **Build** mode | same, `mode=BUILD` | `POST /{provider}/{sandbox_id}/code` ✅ |
| Preview URL bar | `GetPreview(conversation_id)` → `{url, status}` | `POST /sandbox/` ✅, **not persisted** ⛔ |
| Preview reload / open | none — client-side once a URL exists | — |
| Footer credits | already on `Profile` 🟡 | debit on Build ⛔ |

Two shapes to decide before writing the proto:

1. **`SendMessage` is not request/response.** Build runs an agent over a sandbox; that is seconds to
   minutes. It needs to return immediately with a message id and stream or poll for the result,
   which is what the Temporal workflow was meant for.
2. **`preview_url` belongs on the conversation, not the message.** One sandbox per conversation,
   created lazily on the first Build message — otherwise every opened conversation costs a sandbox.

## 7. Decisions still open

1. **Preview lifecycle.** Sandboxes are not free and not persisted. Spin up on conversation open, or
   on first Build message? What does the pane show before one exists, and after one dies? The build
   assumes *lazily, on first Build message*, and renders the no-sandbox state.
2. **Chat mode with no sandbox.** Chat needs no `sandbox_id`, so a conversation can exist before any
   code does. Is Home's first message always Build, or does the user pick?
3. **Streaming.** `{content}` is a single response. Token streaming changes the thread component and
   is what `TextShimmer` / `ThinkingBar` are reserved for (`design-system.md` §8).
4. **File tree.** ARCHITECTURE §1 promises "generated files *and* a preview". This design shows only
   the preview; a files tab is deferred until the sandbox routes are reachable from Go.

## 8. Known limits of the static build

1. **The chat/preview split is not draggable.** The chat lives in the vendored sidebar, which has a
   fixed width and no resize handle. If dragging matters, the chat has to move out of the sidebar
   into a `ResizablePanelGroup` alongside the preview, and the conversation list becomes a separate
   narrow rail — a different shell from the one in §2.
2. **The composer clears and does nothing.** Appending a user message with no reply would read as a
   bug rather than as a stub.
3. **Home submit is the one live flow** — it creates a local draft conversation and opens the
   Workspace on it, so the no-sandbox preview state is reachable.
