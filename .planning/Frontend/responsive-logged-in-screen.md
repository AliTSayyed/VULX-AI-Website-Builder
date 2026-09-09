# Responsive Logged-In Screen

> **Responsive to the backend**, not to breakpoints. This document is about replacing
> `components/session/mock.ts` with the RPCs that now exist, so the Home and Workspace screens run
> on real projects, real messages and a real sandbox URL. Mobile layout is explicitly out of scope
> (§8) — the demo is a desktop browser.
>
> Structure, colour, type and motion are already decided: `logged_in_design.md` (structure) and
> `design-system.md` (tokens). Neither is being redesigned here. Status markers follow
> `ARCHITECTURE.md`: ✅ built · 🟡 partial · ⛔ not built.

## 1. What is actually changing

`LoggedInScreen` ships today as a complete static shell over `CONVERSATIONS` in `mock.ts`. Every
piece of the layout stays. What changes is where the data comes from, plus three things the fixture
build could not express: a project that does not exist yet, a request that takes minutes, and a
preview URL that only appears after the first build.

| Piece | Today | After |
| --- | --- | --- |
| Sidebar list | `CONVERSATIONS` fixture | `ListProjects` ✅ |
| Opening a project | `find()` over the fixture | `GetProject` + `ListMessages` ✅ |
| Home / composer submit | local `setState` | `CreateProject` and/or `SendMessage` ✅ |
| Thread | fixture messages | `ListMessages` + optimistic user bubble |
| Waiting for the agent | nothing — the composer cleared and did nothing | shimmer in the thread, overlay on the preview |
| Preview | a paragraph of placeholder text | `<iframe src={project.previewUrl}>` |
| Timestamps | hand-written strings (`"2 hours ago"`) | formatter over RFC3339 `updated_at` |
| Provider / mode on Home | not present | same two controls as the composer |

`mock.ts` is deleted at the end. Nothing may import it once this is done.

## 2. Data layer

Five hooks, one file each, alongside the existing `useAccountService` / `useUserService` one-liner
pattern. React Query owns all server state; component state owns only what is not on the server
(which view is open, the composer's draft text, the two selectors).

```
hooks/services/useProjectService.ts   ProjectService client
hooks/services/useMessageService.ts   MessageService client
hooks/useProjects.ts                  useProjects()   -> ListProjects
                                      useProject(id)  -> GetProject
                                      useCreateProject()
hooks/useMessages.ts                  useMessages(projectId) -> ListMessages
                                      useSendMessage()
```

| Query key | Source | Notes |
| --- | --- | --- |
| `["projects"]` | `ListProjects({ limit: 100 })` | `token`/`has_more` ignored — see §8 |
| `["project", id]` | `GetProject({ id })` | `enabled: !!id`. **This is where `previewUrl` comes from.** |
| `["messages", projectId]` | `ListMessages({ projectId })` | `staleTime: 0` — a thread must be fresh on open |

After a successful send, exactly three cache operations run:

1. `invalidateQueries(["messages", projectId])` — replaces the optimistic bubble with the persisted
   pair.
2. `invalidateQueries(["project", projectId])` — **this is how the preview URL arrives.**
   `SendMessageResponse` carries only the two messages; the sandbox lives on the project. Documented
   as an accepted trade-off in `Polish.md` → *Known runtime trade-offs*.
3. `invalidateQueries(["projects"])` — the message insert bumped `projects.updated_at` (the CTE in
   `message_repository.go`), so the sidebar reorders.

**Enums cross the boundary as enums.** `AiProvider.ANTHROPIC` / `ChatMode.BUILD` from
`@apiv1/enums_pb`, not the fixture's `"anthropic"` / `"build"` strings. One module
(`components/session/providers.ts`) owns the enum → label map, because the labels diverge from the
wire values on purpose (Claude / OpenAI / Gemini vs. `ANTHROPIC` / `OPENAI` / `GOOGLE`).

**No request timeout.** `createConnectTransport` is left without `timeoutMs`, and none is passed per
call. A Build send blocks for the whole agent run — minutes — because `SendMessage` is synchronous
today (`Polish.md` → *Async Build*). Anything that caps it turns a working build into a failed one.

## 3. Two entries, one flow

The fixture build had three paths into the Workspace (Home submit, New build, opening a row) and two
of them fabricated a conversation locally. With a real backend there is only one way a project comes
into existence — `CreateProject(first_prompt, provider)` — and it needs a prompt. So **Home submit
and New build converge**: both land in the Workspace, and the first send is what creates the project.

```
                    ┌──────────────────────────────┐
   New build ──────►│  Workspace, view = DRAFT      │  empty thread, no preview,
                    │  no project id yet            │  composer live
                    └───────────────┬───────────────┘
                                    │ first send
   Home submit ─────────────────────┤
   (prompt + mode + provider)       ▼
                          CreateProject(first_prompt, provider)
                                    │  project.id
                                    ▼
                    ┌──────────────────────────────┐
                    │  view = PROJECT(id)           │  row appears in the sidebar,
                    └───────────────┬───────────────┘  thread is real
                                    │ SendMessage(id, body, mode, provider)
                                    ▼
                              generating (§4)
```

Represent the open target as a discriminated union, not a nullable id:

```ts
type View =
  | { kind: "home" }
  | { kind: "draft" }                 // Workspace, no project yet
  | { kind: "project"; id: string };
```

`{ kind: "home" }` renders `HomeView`; the other two render the Workspace. `draft` and `project`
must render the *same* Workspace components — a draft is a project with no id, no messages and no
preview, not a fourth screen.

**The view switches to `project` the moment `CreateProject` resolves**, while the send is still in
flight. That is what makes the sidebar row and the thread appear immediately instead of after the
agent finishes. Home submit therefore carries all three values (`prompt`, `mode`, `provider`) up to
the screen, which means Home's prompt box needs the same two selectors as the composer — see §6.

## 4. The generating state

This is the part the static build had no way to show, and the part the demo is judged on.

A send is in flight from the moment the user hits submit until `SendMessage` resolves. Two places
react, and they must be driven by the *same* boolean so they can never disagree:

**In the thread.** The user's message appears instantly as an optimistic bubble (`onMutate` writes a
temporary message into `["messages", projectId]`; the invalidate on success replaces it). Below it,
where the assistant reply will land, a shimmer line:

> Thinking… → Building… → Creating…

cycling every ~2.4s through a single `const` list, wrapped in `TextShimmer`. **`TextShimmer` is
silently broken today** — it applies `animate-[shimmer_4s_infinite_linear]` and there is no
`@keyframes shimmer` in `globals.css`; Tailwind v4 will not generate one for an arbitrary animation
name, and it fails by rendering static rather than by erroring. The keyframe is the first task in
this feature for that reason (`design-system.md` §8).

**On the preview.** The iframe area takes a full-bleed overlay — **opaque `bg-surface`, no
`backdrop-blur`** — with the same cycling shimmer line centred in it, and the URL pill in the chrome
row reads `Generating…` instead of a URL. When the send resolves, the overlay fades out and the URL
pill shows the real preview URL.

The overlay is the one place it is tempting to reach for glass, and it is the worst place for it.
`design-system.md` rule 3 forbids translucent surfaces because contrast becomes a moving target over
the animated silk; behind *this* overlay sits an arbitrary generated website that may be white, dark
or animating — a backdrop we control even less than the silk. Opaque holds the measured
17.05 / 10.12 / 7.64 ladder. §3 also notes `backdrop-blur-*` behind an opaque fill is a no-op that
still costs a paint.

**No iframe refresh.** The sandbox runs `next dev` with hot reload and the HMR websocket lives
*inside* the iframe, so the preview repaints itself as the agent writes files. Bumping a `key` or
re-setting `src` would tear down that connection and reload a page that is already correct. The only
transition on the iframe is the overlay coming off.

**The shimmer must clear on failure.** No designed error surface is in scope (§8), but a mutation
that rejects has to clear the pending state and surface a toast — otherwise a failed build shimmers
forever, which is worse than an error message.

## 5. Preview URL lifecycle

```
project has no sandbox      →  "No preview yet" placeholder, URL pill empty
first Build send in flight  →  overlay + "Generating…"
send resolves               →  GetProject invalidated → previewUrl arrives → iframe loads
later sends                 →  overlay again over the *live* iframe, which keeps its URL
```

`previewUrl` is `""` (not null) when absent — the proto has no optional, and the column is
`NOT NULL DEFAULT ''`. Test it as `project.previewUrl || null`, never `!= null`.

The URL is already scheme-qualified: `ai-service/api/routes/sandbox.py` prepends `https://` because
E2B's `get_host()` returns a bare host, which fails silently as an `<iframe src>`. The frontend does
no URL construction.

`sandbox`-shaped chrome buttons (reload, open in new tab) are currently rendered disabled and wired
to nothing. Wire them: reload re-sets the iframe `src` (a *user-initiated* reload is legitimate;
§4's rule is only about automatic ones), open-in-new-tab is an `<a target="_blank" rel="noreferrer">`.

## 6. Composer rules

The mode toggle and the provider select move out of `ChatPanel` into one shared
`ComposerControls`, used by both the Workspace composer and the Home prompt box. The two must stay
identical — the same argument that put the prompt box's classes in `prompt-styles.ts`.

- **Locked while a send is in flight.** Textarea disabled, submit disabled, both selectors disabled.
  `SendMessage` is synchronous and a second concurrent send against the same sandbox is undefined
  behaviour. This is the one input rule that is not cosmetic.
- **Provider seeds from the open project** (`project.provider`), falling back to Anthropic on Home
  and on a draft. Note the known staleness: `SendMessage` does not write the provider back to the
  project (`Polish.md` → *`Project.provider` goes stale*), so switching provider mid-thread will not
  survive a reopen. Do not work around it here.
- **Mode defaults to Build.** Chat stays selectable — `SendMessage(CHAT)` returns `Unimplemented`
  and will toast — because gating it is not worth doing before Chat exists.
- Enter submits, Shift+Enter newlines (already the `PromptInput` behaviour).
- The thread scrolls to the bottom when the message list changes and when a send starts.

## 7. Sidebar

- Rows come from `ListProjects`, already ordered `updated_at DESC` by the backend. Do not re-sort.
- `Project.title` is the truncated first prompt, generated server-side. The frontend does not derive
  titles.
- Relative timestamps come from one `lib/format-time.ts` helper over RFC3339
  (`Just now` / `2 hours ago` / `Yesterday` / `4 days ago` / an absolute date past a week).
- The open project's row is marked active.
- **Empty state**: a single muted "No projects yet" line where the list would be. A new account has
  no projects, and an empty `SidebarMenu` renders as a blank void.
- `New build` no longer fabricates a row — it sets `view = draft`. The row appears when
  `CreateProject` resolves.

## 8. Deliberately not in scope

Everything here is a conscious omission for the MVP demo, not an oversight.

| Not doing | Why | Owner |
| --- | --- | --- |
| Dead-sandbox / outage handling | A 500 covers both cases; the frontend cannot tell them apart | `Polish.md` → *Known runtime trade-offs* |
| Disabling Chat mode | It fails loudly; not worth gating before Chat exists | `Polish.md` → *Chat mode* |
| Streaming / progress detail | `SendMessage` is one blocking call — there is nothing to stream | `Polish.md` → *Async Build* |
| `ListProjects` pagination | `limit: 100` is past any demo account's project count | — |
| Routing (`/projects/[id]`) | The open project is local state; a refresh returns to Home | — |
| Mobile breakpoints | The demo is a desktop browser; the fixed rail/preview split assumes width | `logged_in_design.md` §8 |
| Credit spend | Displayed, never debited | `Polish.md` |
| File tree tab | Needs codebase persistence first | `Polish.md` → *Codebase persistence* |

## 9. Files touched

```
app/src/
  app/globals.css                          + @keyframes shimmer
  lib/format-time.ts                       NEW  relative time over RFC3339
  hooks/services/useProjectService.ts      NEW
  hooks/services/useMessageService.ts      NEW
  hooks/useProjects.ts                     NEW  useProjects / useProject / useCreateProject
  hooks/useMessages.ts                     NEW  useMessages / useSendMessage
  components/session/providers.ts          NEW  AiProvider <-> label, ChatMode labels
  components/session/generating.tsx        NEW  cycling TextShimmer line
  components/session/composer-controls.tsx NEW  mode toggle + provider select, shared
  components/session/project-list.tsx      was conversation-list.tsx
  components/session/chat-panel.tsx        real messages, optimistic bubble, locked composer
  components/session/home-view.tsx         + ComposerControls, submit carries mode/provider
  components/session/preview-pane.tsx      real iframe, overlay, wired chrome buttons
  components/session/logged-in-screen.tsx  the View union, create -> send orchestration
  components/session/mock.ts               DELETED
```

## 10. What tells you it worked

Not "it compiles". The demo path, in one sitting, in a browser:

1. Log in with an account that has no projects → sidebar reads "No projects yet", Home shows
   `Welcome back,` and a prompt box with both selectors.
2. Type a prompt, submit → the Workspace opens, your message is in the thread, the shimmer is
   cycling, the preview shows the overlay and `Generating…`, the composer is locked, and a new row
   is in the sidebar with a real title.
3. The agent finishes (minutes) → the assistant summary lands in the thread, the overlay lifts, and
   the generated site renders in the iframe.
4. Send a second prompt → the same generating state over the *existing* preview; when it resolves
   the site has changed without the iframe reloading.
5. `← Back` → Home. The project is in the list with a relative timestamp. Reopen it → the full
   thread and the same live preview come back.
6. `New build` → empty Workspace, no project row yet; the first send creates one.
