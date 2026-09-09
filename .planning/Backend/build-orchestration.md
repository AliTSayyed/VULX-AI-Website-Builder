# Build Orchestration — Backend

> ⚠️ **This is a scoping document, not a specification. Rewrite it before implementing.**
>
> It exists now for one reason: four other documents deferred decisions to this one, and those
> deferrals need somewhere to live so they are not lost. What follows is the **collection point** —
> what is already settled, what is still open, and what the other docs are counting on.
>
> The open decisions in §3 are deliberately unmade. They will be made better once the Chat path is
> real and the thread's behaviour can be observed. Do not resolve them from this document.
>
> Written 2026-09-08 against `feature/messages`. Everything here is ⛔ **not built**.
>
> ## ⏸ Phase: POLISH — the MVP implements Build synchronously instead
>
> **The MVP does not use Temporal, a `status` column, or polling.** `SendMessage` with
> `mode = BUILD` runs the whole thing inline and blocks until the agent finishes
> (`messages-model.md` §9). Locally that works: browser → Caddy → Go has no timeout a two-minute
> request trips, once `ai-service-client.md`'s 30-second client ceiling is gone.
>
> That is §3.1's falsification check answering itself early — for a single synchronous flow Temporal
> earns nothing. **Keep the connection and the worker** (`ai-service-client.md` §7.1) and revisit this
> document when the blocking call becomes the thing you want to fix, which will be when the frontend
> needs progress feedback beyond a spinner.


## 1. What this will own

Making the Build toggle work end to end: `SendMessage` with `mode = CHAT_MODE_BUILD` currently
persists the user message and returns `CodeUnimplemented` (`messages-model.md` §9.4). This document
replaces that with real execution.

Concretely, when written it will cover the workflow, the `status` column and its migration, the
status-reporting RPC, the retry and timeout policy, and the failure semantics.

## 2. What is already settled — inherited, not up for redesign

### 2.1 The step sequence (`messages-model.md` §9.3)

1. Ownership check, validation, persist the user message — identical to Chat mode, unchanged.
2. Ensure a live sandbox. `projects.sandbox_id` is nullable and usually stale
   (`project-model.md` §1.1), so this is `RefreshSandbox`'s flow — create, replay files, replay the
   collapsed `npm install` (`project-codebase-model.md` §9.1). 10–30 seconds on its own.
3. Call the code agent: `POST /ai-service/v1/{provider}/{sandbox_id}/code`.
4. `codebaseRepo.ApplyRun(projectID, resp.Files, filterInstalls(resp.Commands))` —
   `project-codebase-model.md` §7.1 and §8.
5. Persist the assistant message.
6. Debit credits, if `credits.md` says so.

### 2.2 The trust boundary — the one rule that must not be broken

**The assistant message body is `resp.Summary`. It is never the agent's own account of which files
it wrote.**

`summary` is model prose; `files` and `commands` are observed fact, captured by
`CodeAgentCallBack` and stored in `project_codebases`. That split is the entire purpose of the
callback (`ARCHITECTURE.md` §7.5). Collapsing the two — rendering the agent's narration as if it
were a file list, or storing narration as the record of what changed — undoes it.

Related: `messages` has **no `files_written` column** and is not gaining one
(`messages-model.md` §2.3). Paths live in `project_codebases`.

### 2.3 The client does not retry (`ai-service-client.md` §2)

`AIService` makes one attempt and reports what happened. **Retry policy is this document's job** —
set it on the Temporal activity, or hand-roll it if Temporal goes. A retry loop inside the client
nested in an activity retry would multiply calls to a provider billed per token.

`RunCodeAgent` also deliberately has no client-side deadline; it inherits the caller's context
(`ai-service-client.md` §5). Setting that deadline is this document's job too.

### 2.4 Temporal's current state

After `ai-service-client.md` lands: the connection, the worker and `StopWorkers` survive;
`user_workflow.go` is deleted; `RegisterWorkers` registers nothing. That is deliberate scaffolding
waiting for this document — not dead code, and not a hint about what the workflow should look like.

`CodeAgentRunner` is listed as a port `build-orchestration.md` declares
(`ai-service-client.md` §6), satisfied by `AIService.RunCodeAgent`.

## 3. What is open — resolve these when writing the real version

### 3.1 First: does Temporal survive?

`ai-service-client.md` §7.3 set this as an explicit falsification condition rather than leaving it
unexamined a second time:

> **If the Build flow collapses to a single activity, remove Temporal.**

Its value comes from checkpointing *completed* steps across a multi-step pipeline — not redoing a
30-second sandbox rebuild because the agent call failed. For one long call it buys only retry and
visibility, and it does **not** resume an in-flight HTTP request: a crash mid-agent-run means paying
for that run twice either way.

§2.1 looks multi-step, which is the case for keeping it. Confirm that against the real
implementation rather than against this sketch. Everything else in this document branches on the
answer.

### 3.2 Poll or stream

`SendMessage(BUILD)` must return immediately with a message id. How does the client learn it
finished?

- A `GetBuildStatus`-style RPC the frontend polls, or
- a server-streaming RPC.

**This is the decision that shapes everything else at once** — the proto, the `status` column, the
frontend's thread component, and whether `TextShimmer` / `ThinkingBar` (`design-system.md` §8) ever
get used. Make it first, after §3.1.

### 3.3 The `status` column

Migration `00007`, on `messages`. Deliberately deferred by `messages-model.md` §2.3 and
`messages-proto.md` §4.7 because its shape depends on §3.2 — adding it earlier would pre-commit the
answer. Adding the proto field later is a backward-compatible field addition.

### 3.4 Failure semantics

Four cases, none decided:

| Case | Question |
| --- | --- |
| Agent call fails | Retry how many times? Does the user see a failed assistant message, or none? |
| Sandbox dies mid-run | Rebuild and retry the whole thing, or fail the message? |
| Workflow crashes **after** the agent succeeded | The expensive one — tokens spent, nothing recorded. Is the response checkpointed before `ApplyRun`? |
| `ApplyRun` fails after a successful agent run | The codebase and the sandbox have now diverged |

### 3.5 Credits

Whether a Build is debited, at what point in the sequence, and what happens when the run fails after
the debit. Owned by `credits.md`, consumed here.

## 4. Dependencies

| Depends on | For |
| --- | --- |
| `ai-service-client.md` | `RunCodeAgent`, and the 30s client timeout removed |
| `project-codebase-model.md` | `ApplyRun`, and the ensure-live-sandbox flow |
| `messages-model.md` | the `messages` table, `SendMessage`, and the §2.1 contract |
| `.planning/AI-Service/` gate-2 callback fix | `files` / `commands` not silently dropping real writes (`project-codebase-model.md` §4) |
| `credits.md` | §3.5 |

**Everything else can ship without this.** Projects list, threads render, Chat replies, previews
refresh. This document is the last thing, and the product is demonstrable without it — which is why
it is also the one with the most freedom to change shape.

## 5. Frontend consequence, for now

**The Build toggle should be visibly disabled until this lands** (`messages-proto.md` §5). An
enabled control that always returns `501` is worse than an absent one.
