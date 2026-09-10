# AI Service Client — Backend

> The Go→Python seam. This document owns `api/internal/infrastructure/outbound/ai_service/` in full:
> the five methods every other backend feature is waiting on, the request/response contract for each
> route, timeout and retry policy, and error mapping. It also deletes the demo Temporal workflow,
> because that workflow is a consumer of the method being changed.
>
> **This is the unblocking feature.** `project-model.md`, `messages-model.md` and
> `project-codebase-model.md` each declare a port that nothing implements until this lands.
>
> Verified against the code on `feature/messages` (2026-09-08). Status markers follow
> `ARCHITECTURE.md`: ✅ built · 🟡 partial · ⛔ not built.

## ⏸ Phase split — MVP vs Polish

The MVP demo is: type a prompt → a project is created → a sandbox starts → the agent writes code
into it → you watch it in the iframe. **Two methods carry that.** The other three exist for Chat,
LLM titles, and codebase replay, none of which the demo does.

| § | Content | Phase |
| --- | --- | --- |
| §1 defects 1, 2, 4 | wrong verb/path, swallowed error, the 30s ceiling | **MVP** |
| §1 defect 3 | `CallAI` stub → replaced by `llm.go` | **MVP** (for `RunCodeAgent`) |
| §2 | transport only, no retries, no prompts | **MVP** |
| §3.1 | route contract | **MVP** for `/sandbox/` and `/{provider}/{sandbox_id}/code` |
| §3.2, §3.3 | query-param `command`, blanket 500s | Polish — both only affect `RunCommand` |
| §3.4 | generated DTOs | Polish |
| §4 `CreateSandbox`, `RunCodeAgent` | | **MVP** |
| §4 `Query`, `WriteFiles`, `RunCommand` | | Polish |
| §4.1 shared `do` helper | | **MVP** |
| §5 timeouts | | **MVP** — the ceiling is why Build cannot work today |
| §6 ports | `CodeAgentRunner` only | **MVP** |
| §6.1 `ProjectTitler` | | Polish — the MVP stubs the titler |
| §7 Temporal | delete `user_workflow.go`, keep the connection | **MVP** |
| §8–§11 | | as marked above |

### The MVP client is two methods

```go
func (a *AIService) CreateSandbox(ctx context.Context) (*SandboxInfo, error)
func (a *AIService) RunCodeAgent(ctx context.Context, provider domain.AIProvider, sandboxID, message string) (*CodeAgentResult, error)
```

`SandboxInfo` and `CodeAgentResult` are unchanged from §4. The MVP **reads `CodeAgentResult.Summary`
and discards `Files` and `Commands`** — they are what `project-codebase-model.md` persists, and that
document is Polish.

### Two consequences worth knowing before you start

**The 30-second ceiling is the single reason Build cannot work today.** `http.Client.Timeout`
overrides any longer context deadline, and a code-agent run exceeds it every time. §5 is not
optional polish inside the MVP — it *is* the MVP.

**`CodeAgentRunner` moves.** §6 lists it as declared by `build-orchestration.md`, which is Polish.
In the MVP the port is declared in `application/services/message_service.go` alongside
`MessageRepository`, because `SendMessage` is what calls it (`messages-model.md` §9). Same interface,
same implementation, different declaring file — ports live with their consumer, and in the MVP the
consumer is the message service.

### The ordering constraint dissolves for the MVP

§6 of `.planning/AI-Service/sandbox-service.md` says to change the `command` route's binding before
writing the Go client. **That only applies to `RunCommand`, which the MVP does not implement.** So
for the MVP the two documents are independent, and `sandbox-service.md`'s only MVP content is its §4
sandbox lifetime.

## 0. Three decisions taken without asking — correct them cheaply if wrong

1. **DTOs are hand-maintained Go structs, not generated from FastAPI's OpenAPI schema.** §3.4 records
   the alternative and the trigger for switching.
2. **The gate-2 callback fix is *not* in this document.** It is a Python change to
   `agent_callback_service.py`; it belongs in `.planning/AI-Service/`. Named here as a dependency
   (§9), not owned.
3. **`ReadFile` and `ListFiles` are out of scope.** No Go caller needs them: the file tree is
   deferred and the `package.json` read-back is hardening, not MVP.

## 1. Current state — three defects, all verified

The package is 40 lines and none of it works.

| # | File | Defect |
| --- | --- | --- |
| 1 | `sandbox.go:16` | `a.client.Get(a.baseURL + "/sandbox/create")` — **wrong method and wrong path.** The route is `POST /sandbox/`. `AI_SERVICE_URL` already ends in `/ai-service/v1` (`config.go:68`), so the prefix is right and only the tail is wrong |
| 2 | `sandbox.go:18` | `if err != nil { }` — an **empty body with a TODO**. Execution falls through to `defer resp.Body.Close()` and nil-derefs, so a transport failure panics the API rather than returning an error |
| 3 | `openai_agent.go:8` | `func (a *AIService) CallAI() error { return nil }` — an empty stub. No Go code has ever reached an LLM endpoint |

Plus one that is not listed in `ARCHITECTURE.md` §11 and matters more than any of them:

| 4 | `ai_service.go:15` | `&http.Client{Timeout: 30 * time.Second}` — a **single shared client with a hard 30-second ceiling.** `http.Client.Timeout` overrides any longer context deadline, so a code-agent run (minutes) fails at 30s regardless of what the caller asks for. §5 |

Defect 1 has a live second victim: `temporal/user_workflow.go:58` calls `CreateSandbox` and then logs
`SandboxResponse.ID` *before* checking the returned error — a second nil-deref stacked on the first.

## 2. What the client is, and what it is not

> **Transport only. One attempt. No retries. No prompts. No domain decisions.**

This is the governing rule and every design choice below follows from it.

**No retries** because Temporal activities already provide retry with exponential backoff,
declaratively. A 3× retry inside the client nested in a 3× activity retry is nine calls to a
provider you pay per token. Retry policy belongs to the caller — `build-orchestration.md` sets it
for Build; the synchronous callers deliberately have none.

**No prompt construction.** The client sends the string it is given. This is why there is no
`GenerateTitle` method: the titling prompt is a product decision and lives in `ProjectService`
(§6.1).

**No domain entities in or out.** The client speaks plain structs with exported fields and JSON
tags. That is not incidental — it is what lets the same types cross a Temporal activity boundary
(§7.2).

## 3. The contract

### 3.1 Routes, exactly as they exist

Base URL is `http://ai-service:9999/ai-service/v1` (`config.go:68`), so every path below is relative
to that. `{provider}` is `openai` / `google` / `anthropic` — the exact segments, from
`domain.AIProvider.String()`.

| Method | Path | Request | Response |
| --- | --- | --- | --- |
| `POST` | `/sandbox/` | *(none)* | `{id, url}` |
| `POST` | `/sandbox/{id}/files` | `{"write_data":[{"path":"…","data":"…"}]}` | `{files_written_to, write_data}` |
| `POST` | `/sandbox/{id}/command?command=…` | **query param** | `{command, stdout, stderr}` |
| `POST` | `/{provider}/query` | `{"message":"…"}` | `{content}` |
| `POST` | `/{provider}/{sandbox_id}/code` | `{"message":"…"}` | `{human_message, summary, commands, files}` |

### 3.2 Three details that will silently break the client

**The trailing slash on `/sandbox/` is required.** FastAPI mounts it as `@router.post("/")` under
`prefix="/sandbox"`. Requesting `/sandbox` without the slash gets a 307 redirect; Go's default
client follows it and 307 does preserve the method, so it may appear to work — but it is an extra
round trip on a path that should be exact.

**`command` is a query parameter, not a body.** `execute_sandbox_command(sandbox_id: str, command:
str, ...)` declares a bare `str`, and FastAPI binds a non-Pydantic scalar that is not in the path as
a **query** parameter. So the request is `POST /sandbox/{id}/command?command=npm+install+clsx` with
an empty body. This is inconsistent with `/files` next door, which takes a JSON body, and it is
URL-length-limited. Use `url.Values` and never string-concatenate the command into the URL.

*Worth fixing on the Python side* — a `CommandRequest` body model would make the two sandbox
mutation routes symmetric. Out of scope here; the Go client matches what exists today. Note it in
`.planning/AI-Service/`.

**The file-write field is `data`, not `content`.** `WriteEntry` is `{path, data}`
(`services/models/sandbox_models.py`). Guessing `content` produces a 422 that reads like a routing
problem.

### 3.3 The service returns 500 for everything

Every route wraps its body in `try/except Exception` and re-raises `HTTPException(status_code=500,
detail=...)`. So:

- There is no 404 for a dead sandbox — `Sandbox.connect()` failing surfaces as a 500.
- There is no 4xx for a bad provider — an unknown `{provider}` is a routing 404 from FastAPI itself.
- 422 comes only from FastAPI's own request validation, which means **the Go client sent something
  malformed**. That is a bug in Go, not a user error, and should be logged loudly.

**Consequence for `RefreshSandbox`:** Go cannot distinguish "this sandbox expired" from "E2B is
down" from "the template is broken". All three are a 500 with prose in `detail`. Do not write logic
that branches on that distinction — it does not exist. Log `detail` server-side and treat the class
as `ErrorTypeUnavailable`.

**Never surface `detail` to the browser.** It contains upstream infrastructure text. The auth
interceptor already collapses non-client errors to a generic `CodeInternal` message
(`ARCHITECTURE.md` §11 #3), so this is contained by default — do not defeat it by copying `detail`
into a `connect.Error`.

### 3.4 Hand-maintained structs, and when to stop

`ARCHITECTURE.md` §4 calls this "the one seam with no shared schema" and notes it has already
drifted. This feature roughly doubles its surface area, so the choice is worth stating rather than
inheriting.

**Decision: hand-maintained Go structs**, with §3.1 as the written contract of record. Five
endpoints and eight structs is a size a person can hold, and the repo already carries two code
generators (buf, protoc-gen-es); a third tool, a committed spec file and a regeneration discipline
is real weight for this much surface.

**The alternative, for when that stops being true:** FastAPI serves `/openapi.json` for these exact
routes. `oapi-codegen -generate types` turns it into Go structs — *types only*, with the calls still
hand-written, because a fully generated client returns its own error types and would fight the
`domain.Error` taxonomy that every layer here depends on.

**Switch when any of these happens:** a drift bug reaches runtime a second time · the seam passes
~10 endpoints · a second consumer (not the Go API) appears. Until then, §3.1 and the Pydantic models
are kept in sync by hand and by review.

## 4. The Go API

`ai_service.go` keeps the struct and constructor; `sandbox.go` is rewritten; `openai_agent.go` is
deleted and replaced by `llm.go` (its name describes one provider, but the file serves all three).

```go
// llm.go
func (a *AIService) Query(ctx context.Context, provider domain.AIProvider, message string) (string, error)
func (a *AIService) RunCodeAgent(ctx context.Context, provider domain.AIProvider, sandboxID, message string) (*CodeAgentResult, error)

// sandbox.go
func (a *AIService) CreateSandbox(ctx context.Context) (*SandboxInfo, error)
func (a *AIService) WriteFiles(ctx context.Context, sandboxID string, files map[string]string) error
func (a *AIService) RunCommand(ctx context.Context, sandboxID, command string) (*CommandResult, error)
```

**Both result types live in `internal/domain/`, not in this package.** `MessageService` and
`CodebaseService` declare the ports that reference them, and a port in `application/services`
importing `infrastructure/outbound/ai_service` would point the dependency arrow outward. `domain`
imports nothing, so putting them there keeps every arrow inward. They are value objects — public
fields, no constructors, no accessors — deliberately unlike `domain/user.go`.

```go
type SandboxInfo struct {
    ID  string `json:"id"`
    URL string `json:"url"`
}

type CodeAgentResult struct {
    Summary  string            `json:"summary"`
    Commands []string          `json:"commands"`
    Files    map[string]string `json:"files"`
}

type CommandResult struct {
    Command string `json:"command"`
    Stdout  string `json:"stdout"`
    Stderr  string `json:"stderr"`
}
```

`RunCommand` returns stdout/stderr rather than just an error because the refresh path logs and
continues on a failed `npm install` (`project-codebase-model.md` §9.2) — it needs something to log.

`CodeAgentResult` deliberately omits `human_message`: it is the request echoed back, and the caller
already has it.

### 4.1 One private helper does all five

Every method is a thin wrapper over one `do` function that builds the request, sets
`Content-Type: application/json`, executes, checks the status, and decodes. Five near-identical
hand-rolled `http.NewRequestWithContext` blocks is where drift and forgotten `Body.Close()` calls
come from.

Non-negotiables inside it:

- `http.NewRequestWithContext` — never `http.NewRequest`. The context is the only thing that makes
  cancellation and per-call deadlines work.
- `defer resp.Body.Close()` **after** the error check, never before. That ordering is defect #2.
- On a non-2xx, decode `{"detail": …}`, log it with the status and the path, and return a
  `*domain.Error` — never the raw body.
- Always return `*domain.Error`. `domain.WrapError` assigns `ErrorTypeUnknown` to anything that is
  not already one, so a bare `fmt.Errorf` from here becomes `CodeUnknown` at the handler.

## 5. Timeouts

**Delete `Timeout: 30 * time.Second` from the shared client.** It is a ceiling that silently
overrides longer context deadlines, and it is the reason Build cannot work no matter how the caller
is configured.

Replace it with two layers:

```go
// ai_service.go — a backstop, not a policy
client: &http.Client{Timeout: 15 * time.Minute}

// each method — the actual policy
ctx, cancel := context.WithTimeout(ctx, opTimeout)
defer cancel()
```

| Operation | Deadline | Why |
| --- | --- | --- |
| `Query` | 60s | a plain LLM call; a title or a chat reply |
| `CreateSandbox` | 90s | E2B provisioning plus template boot |
| `WriteFiles` | 60s | one batched write |
| `RunCommand` | 5 min | `npm install` on a cold cache |
| `RunCodeAgent` | **none — inherits the caller's** | minutes, and the caller (a Temporal activity) owns the deadline |

The 15-minute client backstop exists specifically for `RunCodeAgent`. Temporal does **not** forcibly
kill an activity goroutine when `StartToCloseTimeout` expires — it marks the attempt failed and may
retry. Without a backstop, a hung connection leaks a goroutine and a socket for the process
lifetime.

## 6. Ports this satisfies

Declared by their consumers (`ARCHITECTURE.md` §5.2), all implemented by `*AIService`. Nothing new
is declared here — an outbound adapter does not own contracts.

| Port | Declared in | Method |
| --- | --- | --- |
| `ProjectTitler` | `project_service.go` | `Query` |
| `ChatResponder` | `message_service.go` | `Query` |
| `SandboxRunner` | `codebase_service.go` | `CreateSandbox`, `WriteFiles`, `RunCommand` |
| `CodeAgentRunner` | `build-orchestration.md` ⛔ | `RunCodeAgent` |

### 6.1 `ProjectTitler` changed shape — already applied to `project-model.md`

That document originally declared `GenerateTitle(ctx, provider, prompt) (string, error)`. It now
declares `Query(ctx, provider, message) (string, error)`, identical to `ChatResponder`. The edit is
made; this section records why.

The reason is §2: a `GenerateTitle` method would have to carry the titling prompt, which is a
product decision sitting inside a transport adapter. Move it up — `ProjectService` owns the prompt
template, the 60-character clamp, the quote/newline stripping and the fallback to the provisional
title (`project-model.md` §6.3, unchanged). The client just sends a string.

Two ports with the same method set, satisfied by one type, is ordinary Go interface segregation —
each consumer still declares what it needs.

## 7. Temporal

### 7.1 Delete `user_workflow.go`

The demo workflow is non-functional (`UseLlm` is empty), `StartUserWorkflow` blocks on
`workflowRun.Get` which makes the "async" orchestration synchronous (`ARCHITECTURE.md` §11 #8), its
`CreateSandbox` activity nil-derefs, and its `StartToCloseTimeout: 10 * time.Second` is shorter than
sandbox creation takes. Every one of those is a pattern worth not copying, and it must change
regardless because it consumes the method whose signature is changing.

Deleting it touches four files:

| File | Change |
| --- | --- |
| `outbound/temporal/user_workflow.go` | delete |
| `outbound/temporal/temporal.go` | `RegisterWorkers` takes no `*UserWorkflow`; keep the connection, the worker and `StopWorkers` |
| `services/user_service.go` | delete the `UserWorkflowService` port; `NewUserService(userRepo)` loses its second parameter. **`StartUserWorkflow` has no caller** — it is stored on the struct and never invoked, so nothing behavioural is lost |
| `application/app.go` | drop the `NewUserWorkflow` / `RegisterWorkers(userWorkflow)` lines and the `NewUserService` argument |

**Keep `temporal.go`, the client connection, the compose services and the health-check ordering.**
`build-orchestration.md` needs all of it. `RegisterWorkers` becomes a worker with nothing registered
yet — that is honest scaffolding, not dead code, and it keeps the shutdown path intact.

### 7.2 Why the client's DTOs are shaped for activities

Temporal's default data converter serializes activity arguments and results as JSON. Domain entities
have **private fields** (`domain/user.go`'s "expose nothing" rule), so `domain.Project` marshals to
`{}` — silently, with no error.

`SandboxInfo`, `CodeAgentResult` and `CommandResult` have exported fields and JSON tags, so they
cross an activity boundary unchanged. That is why `RunCodeAgent` returns one of them rather than a
domain type, and it is a property to preserve deliberately rather than rediscover.

### 7.3 Temporal's scope, recorded

Temporal exists to run **the Build pipeline and nothing else** — not title generation (a goroutine;
a lost title degrades to the provisional one), not `RefreshSandbox` (synchronous for MVP), not Chat.

And a falsification condition, so it does not sit unexamined a second time: **if
`build-orchestration.md` finds the Build flow collapses to a single activity, remove Temporal
then.** Its value comes from checkpointing completed steps across a multi-step pipeline; for one
long call it buys only retry and visibility, and it does not resume an in-flight HTTP request — a
crash mid-agent-run means paying for that run twice.

## 8. Files

**Modified**

| File | Change |
| --- | --- |
| `outbound/ai_service/ai_service.go` | client timeout → 15 min backstop (§5); shared `do` helper (§4.1) |
| `outbound/ai_service/sandbox.go` | rewrite: `CreateSandbox` (fix verb + path + the swallowed error), add `WriteFiles`, `RunCommand` |
| `outbound/temporal/temporal.go` | `RegisterWorkers` drops its parameter |
| `application/services/user_service.go` | delete `UserWorkflowService`; `NewUserService` loses a parameter |
| `application/services/project_service.go` | `ProjectTitler.GenerateTitle` → `Query` (§6.1) |
| `application/app.go` | drop the workflow wiring and the `NewUserService` argument |
| `.planning/Backend/project-model.md` | §6 port rename — **already applied** (§6.1) |
| `ARCHITECTURE.md` | §7.1's "only sandbox *create* is intended for the Go API" is no longer true — `/files` and `/command` are now first-class. §11 #1, #2 and #8 move to fixed |

**Created**

| File | Contents |
| --- | --- |
| `outbound/ai_service/llm.go` | `Query`, `RunCodeAgent`, `CodeAgentResult` |

**Deleted**

| File | Why |
| --- | --- |
| `outbound/ai_service/openai_agent.go` | the `CallAI` stub; replaced by `llm.go` |
| `outbound/temporal/user_workflow.go` | §7.1 |

## 9. Dependencies

| Depends on | For | Blocking? |
| --- | --- | --- |
| `domain/ai_provider.go` | `AIProvider.String()` as the `{provider}` path segment | **Yes** — created by `project-model.md` §4.2. If this feature lands first, create that one file here |
| `.planning/AI-Service/` gate-2 callback fix | `RunCodeAgent`'s `files`/`commands` not silently dropping real writes | No for the client; yes for anything that trusts the result (`project-codebase-model.md` §4) |
| E2B credentials + template id in `.ai-service-env` | any sandbox call | Yes, for manual verification |

Nothing depends on the order of the other three backend docs. This one can land first and stand
alone.

## 10. Verifying

No test harness exists in this repo — manual only. Do not claim tests pass.

```bash
cd api && go build ./... && go vet ./...     # deleting user_workflow.go must not leave a dangling reference
make nuke && make
docker compose logs api | grep -i temporal   # connection still established, no workflow registered
```

Exercise the Python routes directly first, so a Go failure is unambiguous:

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | tee /dev/stderr | jq -r .id)

curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/$SB/files \
  -H 'content-type: application/json' \
  -d '{"write_data":[{"path":"/home/user/app/page.tsx","data":"export default () => <h1>hi</h1>"}]}'

# command is a QUERY parameter — §3.2
curl -s -X POST "http://localhost:9999/ai-service/v1/sandbox/$SB/command?command=npm%20install%20clsx"

curl -s -X POST http://localhost:9999/ai-service/v1/anthropic/query \
  -H 'content-type: application/json' -d '{"message":"say hi"}'
```

Then the Go side, through whichever consumer has landed — `CreateProject` exercises `Query`,
`RefreshSandbox` exercises all three sandbox methods.

Five checks that a happy path will not surface:

1. **The 30s ceiling is really gone.** Run a command that takes ~90 seconds
   (`?command=sleep%2090`) through `RunCommand`. Before the fix this fails at 30s; after, it
   succeeds. This is the single most important check in the feature.
2. **A transport failure returns, it does not panic.** `docker compose stop ai-service`, then call
   any consumer. Expect a clean `Unavailable`, not a nil-deref stack trace in the API logs — that
   is defect #2.
3. **The trailing slash.** Confirm `CreateSandbox` hits `/sandbox/` and gets a 200 directly, not a
   307 followed by a 200.
4. **`write_data` field names.** A wrong key gives a 422, which is Go's bug — confirm a real write
   round-trips by reading the file back with
   `curl "…/sandbox/$SB/file?path=/home/user/app/page.tsx"`.
5. **Cancellation propagates.** Start a long `RunCommand` and cancel the client request; the Go
   goroutine must return promptly rather than running to the backstop.

## 11. Out of scope

`ReadFile` / `ListFiles` (§0.3) · retries, circuit breaking, rate limiting (§2 — the caller's job) ·
generated DTOs (§3.4) · the gate-2 callback fix (§0.2) · making the `command` route take a body
(§3.2 — Python-side) · conversation history in the request payload (`messages-model.md` §5) ·
streaming responses · the real Build workflow (`build-orchestration.md`) · tests.
