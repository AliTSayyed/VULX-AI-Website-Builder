# Polish — what the MVP deliberately skips

> The MVP backend is done (`feature/messages`): `project-proto`, `project-model`,
> `ai-service-client`, `sandbox-service` §4, `messages-proto`, `messages-model` §9 (synchronous
> Build) are all built, verified, and merged. This file replaces `MVP.md` as the index of what's
> left for a real product — each item names the doc that owns the real design work, if one exists.
> Frontend RPC wiring landed as `responsive-logged-in-screen.md` — tracked in `ARCHITECTURE.md`
> §9/§10, not here.

## Backend

**Async Build** — `build-orchestration.md`. `SendMessage(BUILD)` currently blocks the HTTP request
for the full agent run (minutes), synchronously, no progress feedback beyond a spinner. Needs: a
poll-vs-stream decision (shapes everything else), a `status` column (migration `00007`, deferred
because its shape depends on that decision), failure semantics for four unresolved cases (agent
fails / sandbox dies mid-run / workflow crashes after the agent succeeds but before persisting /
`ApplyRun` fails after success), and a credit-debit point. Open falsification check carried over from
`ai-service-client.md`: if Build stays one activity, remove Temporal rather than keep it as
unused-but-present infra.

**Codebase persistence & replay** — `project-codebase-model.md`, `project-codebase-proto.md`. Agent
writes are currently discarded (`resp.Files`/`resp.Commands` decoded, never stored) — a dead sandbox
means the code is gone. Needs a `project_codebases` table (JSONB files+commands), a `RefreshSandbox`
RPC that replays the stored files into a fresh sandbox, size/path guards (1MB/file, 10MB total, 512
char paths), and `npm install` replay handling. Known limits even once built: no delete tool yet (see
below) means the map only grows; text-only files; refresh isn't cost-idempotent (no reaping of the
old dead sandbox).

**Chat mode** — `messages-model.md` §8. MVP inverted the doc's original phase split: Build shipped,
Chat didn't. `SendMessage(CHAT_MODE_CHAT)` returns `Unimplemented` today (message still persists).
Needs the `ChatResponder`/`Query` port on the AI-service client (`ai-service-client.md`) and the
`/{provider}/query` route wired through.

**`Project.provider` goes stale** — spec'd in `messages-model.md` §8.1/§8.4 as a step that applies to
every send regardless of mode (`if provider != project.Provider(): projectRepo.UpdateProvider(...)`),
but the derived task doc's Build-flow pseudocode omitted it, and it shipped without it. Found live
during MVP testing; deliberate call to defer rather than patch in after the fact. Effect: switching
providers mid-thread doesn't update what the Workspace restores as "current selection" on reopen.
Fix is small — one `ProjectRepository.UpdateProvider` method (mirrors `UpdateSandbox`'s shape),
called from `MessageService.Send`.

**Project renaming** — `project-model.md`. No `RenameProject` RPC — a generated title can't be
edited afterward even though it's wrong sometimes.

**LLM-generated project titles are a hack, not the originally-planned design.**
`ProjectService.Create` now generates titles via a detached goroutine calling the new
`ProjectTitler` port — `POST /openai/query`, hardcoded to OpenAI regardless of the project's own
selected provider (`ARCHITECTURE.md` §5.6). This diverges from the idea on the table when this was
still unbuilt (a dedicated free/cheap model, e.g. Gemini Flash, via its own route) in two ways
worth revisiting: it reuses the general `/query` route and whatever `OPENAI_MODEL` happens to be
configured to, rather than a purpose-built cheap model — so a deployment with no
`OPENAI_API_KEY` silently and permanently falls back to the truncated-prompt title for every
project — and the provider choice is hardcoded, not configurable. Any failure (timeout, outage,
empty response) fails silently: the provisional title just stands forever, logged as a warning
server-side, never surfaced to the user.

**AI-service client completeness** — `ai-service-client.md`. `Query` (needed by Chat + titles),
`WriteFiles`/`RunCommand` (needed by codebase replay) aren't implemented on the Go client. Revisit
together, since Query unblocks two other items above.

## AI-service (Python)

**Agent result capture is unreliable** — `agent-result-capture.md`. The callback's success test is a
substring match (`"error" in output`), which both false-drops real writes (a file whose *content*
mentions "error") and misses real failures (`npm ERR!` doesn't contain the literal word). Fix: a
`TOOL_STATUS: ok/failed` marker line + exit codes on `TerminalInfo`. Blocks `project-codebase-model.md`
— storing unreliable capture data is worse than not storing it.

**No conversation memory** — `conversation-history.md`. Every message (Chat or Build) is answered as
if it were the first; a second Build message edits the same sandbox correctly but doesn't understand
prior turns. Fix: thread `history` through `ChatPromptTemplate`+`MessagesPlaceholder` (string-concat
was rejected — causes duplicate builds), token-budget truncation (~10 turns), Go's `SendMessage`
passing prior messages via `FindByProject`.

**Agent capabilities are incomplete** — `agent-capabilities.md`. No delete tool (a file removed by
the agent resurrects on the next replay once `project-codebase-model.md` exists). Paths aren't
canonicalized (relative vs. absolute produces duplicate map keys, non-deterministic replay) — fix is
`posixpath.normpath` in Python. `npm install` mutates `package.json` with nothing capturing the
result.

**Sandbox-service hardening** — `sandbox-service.md` (§4, the timeout, is done). The `command` route
still takes a query param instead of a JSON body. Every failure returns a blanket 500 — no
distinction between a dead sandbox (404), E2B being down (502), or a bad path (400), which is why Go
can't tell a dead sandbox from an outage either (see below). Keep-alive / active sandbox reaping is
explicitly out of scope forever — manual refresh (`RefreshSandbox`) is the intended fix, not
background lifecycle management.

## Frontend

`logged_in_design.md` open questions, now that `responsive-logged-in-screen.md` has wired the RPCs to
the screen: whether a sandbox spins up on opening a project or waits for the first Build message; the
poll-vs-stream decision above driving how the thread UI shows a pending reply; a file tree tab (needs
`project-codebase-model.md` first); whether the chat/preview split becomes a draggable
`ResizablePanelGroup` instead of the fixed split it is today.

**The title-generation reveal is a bespoke, one-off polling mechanism, not a general pattern.**
`useProject`'s `pollForTitle` option, `LoggedInScreen`'s `pendingTitleId`, and `ChatPanel`'s
snapshot-and-cross-fade logic exist solely to notice an async title update in an app that
otherwise never polls or refetches spontaneously (`refetchOnWindowFocus: false`). If a second
async-backend-update-needs-frontend-polling need ever shows up (e.g. async Build's progress, above),
this ad hoc per-feature approach won't scale — worth generalizing into a small reusable
"poll until this field changes, bounded" hook rather than copying the pattern a second time.

## Known runtime trade-offs (accepted, not bugs)

- **A dead sandbox is indistinguishable from an outage.** The AI service returns a plain 500 for
  both, so `SendMessage` can't tell them apart and just fails with an opaque error. Sandbox-service
  hardening (proper status codes) is the prerequisite; `RefreshSandbox` is the actual fix.
- **`SendMessage` doesn't return the sandbox/preview URL** — the frontend must refetch `GetProject`
  after a send to see it.
- **No credits/payments plan exists yet.** Referenced in passing (`build-orchestration.md` §3.5,
  `project-model.md` §6.3) but has no dedicated planning doc — needs one before Build gains a debit
  point.
