# Messages Model — Backend

> The chat thread inside a project. This document owns **the `messages` table and its vertical
> slice** — migration, domain entity, repository, service, the `ListMessages` and `SendMessage`
> RPCs, and wiring.
>
> **Chat mode ships complete here. Build mode does not** — this document defines the message rows a
> Build run writes and then hands execution to `build-orchestration.md`. See §9.4 for why that split
> is forced rather than chosen.
>
> Depends on `project-model.md` (the FK target and the ownership check) and `ai-service-client.md`
> (one route). Verified against the code on `feature/messages` (2026-09-08). Status markers follow
> `ARCHITECTURE.md`: ✅ built · 🟡 partial · ⛔ not built.

## ⏸ Phase split — MVP vs Polish

**This document's Chat/Build split is inverted by the MVP.** It was written assuming Chat ships
first and Build waits for Temporal. The demo is the opposite: Build is the whole point, and Chat is
not needed at all.

| § | Content | Phase |
| --- | --- | --- |
| §1–§4 | the table, domain, repository, the CTE | **MVP** |
| §5 | the AI service is one-shot | **MVP** as context — the demo is one message, so it does not bite |
| §6 | proto: both enums, both RPCs | **MVP** |
| §7 | service, `List`, unpaginated threads | **MVP** |
| §7 `ChatResponder` port | | Polish |
| §8 | **Send — Chat mode** | **Polish** — needs `Query`, which the MVP client does not implement |
| §8.2, §8.3 | message-survives-failure, validation | **MVP** — they apply to Build unchanged |
| §9 | **Send — Build mode** | **MVP**, implemented synchronously — see §9.0 below |
| §9.1–§9.4 | why it was deferred, and the async contract | Polish rationale, kept |
| §10–§12 | handler, wiring, dependencies, verifying | **MVP** |

### §9.0 The MVP Build flow

Synchronous, inline, no Temporal, no `status` column, no polling:

```
1. project := projectService.Get(ctx, userID, projectID)      // ownership → NotFound if not theirs
2. validate body, mode, provider                              // §8.3, unchanged
3. userMsg := messageRepo.Create(User, BUILD, provider, body) // CTE bumps projects.updated_at
4. sandboxID := project.SandboxID()
   if sandboxID == "" {                                       // first Build on this project
       info := sandbox.CreateSandbox(ctx)
       projectRepo.UpdateSandbox(ctx, projectID, info.ID, info.URL)
       sandboxID = info.ID
   }
5. resp := codeAgent.RunCodeAgent(ctx, provider, sandboxID, body)   // minutes; blocks
6. asstMsg := messageRepo.Create(Assistant, BUILD, provider, resp.Summary)
7. return userMsg, asstMsg
```

**`resp.Files` and `resp.Commands` are discarded.** They are what `project-codebase-model.md`
persists, and that document is Polish. `resp.Summary` is parsed from the model's own output, not
from the callback, so this works even before
`.planning/AI-Service/agent-result-capture.md` lands.

MVP ports in `application/services/message_service.go` — declared here because this is the consumer:

```go
type SandboxCreator interface {
    CreateSandbox(ctx context.Context) (*aiservice.SandboxInfo, error)
}

type CodeAgentRunner interface {
    RunCodeAgent(ctx context.Context, provider domain.AIProvider, sandboxID, message string) (*aiservice.CodeAgentResult, error)
}
```

Narrower than `ai-service-client.md` §6's `SandboxRunner` on purpose: `WriteFiles` and `RunCommand`
have no MVP caller. `MessageService` also takes `ProjectRepository` directly for `UpdateSandbox`,
alongside the `*ProjectService` it already composes for the ownership check.

### Three MVP limitations, accepted deliberately

**A dead sandbox is unrecoverable.** Step 4 reuses `sandbox_id` whenever it is non-empty — it has no
way to tell a live sandbox from an expired one, because the AI service returns 500 for both
(`ai-service-client.md` §3.3). So a second Build after expiry fails with an opaque error and the
project is stuck. Mitigated by raising the sandbox timeout
(`.planning/AI-Service/sandbox-service.md` §4) so a demo outlives its own sandbox; fixed properly by
`project-codebase-model.md`'s `RefreshSandbox`.

**`SendMessage` does not return the preview URL.** The sandbox is created inside step 4, but
`SendMessageResponse` carries only the two messages (§6.2). The client learns the URL by refetching
`GetProject` after the send — which it wants to do anyway, since `updated_at` changed and the sidebar
needs reordering. Adding `preview_url` to the response would save a round trip; it is deliberately
not done, because the field would become redundant the moment `RefreshSandbox` exists.

**The call blocks for minutes.** Fine for `curl` and fine for a demo. It is also the thing that will
make you want `build-orchestration.md`, and that wanting is the right trigger — not before.


## 1. Why this exists

`ChatPanel` renders `mock.ts`'s `Message[]` fixture. `ARCHITECTURE.md` §10 lists the message model
as ⛔. This is the table that replaces the fixture and the RPC pair that feeds it.

A message is **immutable once written**. Nothing edits or deletes one; a thread only ever grows.
That is why the table has no `updated_at` semantics worth reasoning about and no soft-delete, and it
is what makes the whole feature small.

## 2. The table

### 2.1 `00005_add_messages.sql`

Numbers are sequential across the directory and must not gap. `project-model.md` takes `00004`, and
**this takes `00005`** — `project-codebase-model.md` is Polish and takes `00006` whenever it lands.
Never edit an applied migration; `make nuke` to reset.

```sql
-- +migrate Up
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL,
    mode VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_project_created
    ON messages (project_id, created_at ASC, id ASC);

-- +migrate Down
DROP TABLE IF EXISTS messages CASCADE;
```

### 2.2 Column by column

| Column | Why it is here |
| --- | --- |
| `project_id` | The thread's owner. `CASCADE` because a deleted project's messages are meaningless — though delete is out of scope (§13) |
| `role` | `user` / `assistant`. Drives which side of the thread renders it |
| `mode` | `chat` / `build`. **Per message, not per project** — the composer's `ToggleGroup` is re-read on every send and a user asks Chat questions mid-Build thread. `mock.ts`'s `Message` already carries it |
| `provider` | The **historical record** of which model produced this message. Distinct from `projects.provider`, which is the current selection — `project-model.md` §3 |
| `body` | The prose. For an assistant Build message this is the agent's `summary`, never its file list — §9.4 |
| `created_at` | The only ordering that matters. `ASC` within a thread |

### 2.3 What is deliberately absent

- **No `files_written`.** An assistant Build message does not record which paths it touched.
  Contents and paths live in `project_codebases` (`project-codebase-model.md`), and the thread
  renders the agent's `summary`. Adding a paths array would be a second, weaker copy of data that
  already has a home.
- **No `status`.** A Build message is written before the agent finishes, so a
  pending/complete/failed state will be needed — but its shape depends on whether the result is
  polled or streamed, which is undecided. It lands as its own migration in
  `build-orchestration.md` rather than being guessed at here. §9.4.
- **No `token_count` / `cost`.** `credits.md` decides whether a Build is debited and what it is
  measured in. Adding a column before that decision would pre-commit the unit.
- **No `parent_id` / threading.** A project has one linear thread.

### 2.4 Notes on the DDL

- **`NOT NULL` on the timestamps diverges from `users` and `user_auth_providers`.** Same reason as
  `project-model.md` §2.3: the value is read positionally and a `NULL` is unusable. The existing
  tables are not being changed.
- **`role`, `mode`, `provider` are `VARCHAR`, not Postgres `ENUM`s** — following
  `user_auth_providers.provider`, which stores `"google"` as text and parses it in the domain. A
  Postgres enum needs a migration to add a value.
- **`body` is `NOT NULL`.** An agent that returns empty output still yields a placeholder summary
  from the AI service (`ARCHITECTURE.md` §7.5), so there is no legitimate empty case. A failed run
  is a body that says so, not a null.
- **The index matches the only access path** — one project's thread, oldest first. §7.2.

## 3. Domain

### 3.1 `domain/message.go` — new

Follows `domain/user.go`: private fields, nil-safe accessors, a validating `NewX` and a
non-validating `RestoreX` (data already in the database is trusted — `ARCHITECTURE.md` §5.5).

```go
type Message struct {
    id        uuid.UUID
    projectID uuid.UUID
    role      MessageRole
    mode      ChatMode
    provider  AIProvider
    body      string
    createdAt time.Time
    updatedAt time.Time
}
```

- `NewMessage(projectID uuid.UUID, role MessageRole, mode ChatMode, provider AIProvider, body string) (*Message, error)` —
  rejects `uuid.Nil`, an empty body, and any unspecified enum. Sentinels `ErrMessageBodyEmpty`,
  `ErrMessageProjectEmpty`, `ErrMessageRoleUnspecified`, `ErrMessageModeUnspecified`, all
  `ErrorTypeInvalid`, declared at the top of the file as `user.go` does.
- `RestoreMessage(...)` takes `role`, `mode` and `provider` as **strings** and parses them,
  mirroring `RestoreUserFromProvider`. That keeps the persistence row struct free of domain enum
  types.
- Accessors: `ID`, `ProjectID`, `Role`, `Mode`, `Provider`, `Body`, `CreatedAt`, `UpdatedAt`.

`MessageRole` and `ChatMode` are declared in this same file as `int` enums with `iota`, a `String()`
method and a `ParseX` function — the `LoginProvider` pattern verbatim, including an
`...Unspecified` zero value whose `String()` is `"unspecified"`.

| Enum | Values | Wire strings |
| --- | --- | --- |
| `MessageRole` | `Unspecified`, `User`, `Assistant` | `user`, `assistant` |
| `ChatMode` | `Unspecified`, `Chat`, `Build` | `chat`, `build` |

`AIProvider` is **not** redeclared — `project-model.md` §4.2 already creates
`domain/ai_provider.go`, and its `String()` is the AI-service URL path segment.

## 4. Persistence

New file `postgres/message_repository.go`, following `user_repository.go`: package-private row
struct with `db` tags, `ToDomain()`, a constructor taking `*sqlx.DB`, hand-written SQL, errors
wrapped as `*domain.Error`.

```go
type Message struct {
    ID        uuid.UUID `db:"id"`
    ProjectID uuid.UUID `db:"project_id"`
    Role      string    `db:"role"`
    Mode      string    `db:"mode"`
    Provider  string    `db:"provider"`
    Body      string    `db:"body"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```

### 4.1 Methods

| Method | Signature | Notes |
| --- | --- | --- |
| `Create` | `(ctx, *domain.Message) (*domain.Message, error)` | also bumps the parent's `updated_at` — §4.2. Enums written via `.String()` |
| `FindByProject` | `(ctx, projectID uuid.UUID) ([]*domain.Message, error)` | `ORDER BY created_at ASC, id ASC`. Unpaginated — §7.2 |

Two methods. Nothing updates or deletes a message (§1).

### 4.2 `Create` bumps the project in the same statement

The sidebar orders projects by *last activity*, and inserting a message is the activity — but it
does not move the parent row on its own. Postgres can do both in one statement with a
data-modifying CTE:

```sql
WITH inserted AS (
    INSERT INTO messages (project_id, role, mode, provider, body, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    RETURNING id, project_id, role, mode, provider, body, created_at, updated_at
), bumped AS (
    UPDATE projects SET updated_at = NOW() WHERE id = $1
)
SELECT * FROM inserted;
```

A data-modifying CTE executes exactly once even when nothing references it, so `bumped` runs despite
being unreferenced. Both sub-statements see the same snapshot and commit together.

**This is decided — do not substitute two statements from the service.** The alternative
(`messageRepo.Create` followed by a separate `projectRepo.Touch`) is non-atomic, and this repo has
no transaction helper to reach for, so a failure between them leaves a message whose project never
moved up the sidebar.

**The one cost, stated so it is not discovered by surprise:** this puts an `UPDATE projects` inside
`message_repository.go`, which contradicts the "every statement touching `projects` lives in
`project_repository.go`" rationale in `project-model.md` §5.1. Someone grepping for
`UPDATE projects` will find this one hit outside that file. It is the accepted exception, and
`project-model.md` §5.1 names it as such from the other side.

**Knock-on:** `ProjectRepository` has **no `Touch` method**. It would have no caller — refresh
deliberately does not bump (`project-codebase-model.md` §9.1) and `UpdateTitle` bumps inline. It is
already absent from `project-model.md` §5.1 and §6; do not add it back.

### 4.3 Error mapping

`Create` can fail with a foreign-key violation if the project vanished between the ownership check
and the insert. Map `23503` (`foreign_key_violation`) to `ErrorTypeNotFound`, not `Internal` — it
means the project is gone, which is a client-visible fact.

`FindByProject` returns an **empty slice, not `NotFound`**, for a project with no messages. A
project created from Home before its first send legitimately has an empty thread; that is a normal
state the Workspace renders as an empty composer, not an error.

## 5. The AI service is one-shot — read this before designing the flow

**Verified, and it changes what Chat mode can honestly promise.**

```python
class AIRequest(BaseModel):
    message: str          # api/routes/models/ai_models.py

class AICodeAgentRequest(BaseModel):
    message: str
```

Both endpoints take **a single string**. There is no history parameter, no message list, no
conversation id. So today, every Chat question and every Build instruction is answered with no
knowledge of anything earlier in the thread. The UI renders a conversation; the model is not having
one.

It cannot be fixed on the Python side by bolting memory onto the agent, either: the executors are
`@lru_cache()` singletons in `api/dependencies.py`, one per provider **for the whole process**
(`ARCHITECTURE.md` §7.2). Any memory attached there would be shared across every user of the
service. The isolation is correct; the continuity has to come from the request.

Three options, and the recommendation:

| Option | Cost | Verdict |
| --- | --- | --- |
| **(a) Ship one-shot for MVP** | Each turn is independent. "Make it blue" after "build a hero" will not know what "it" is | **Recommended.** The thread still persists and renders correctly; only the model's memory is missing |
| (b) Go concatenates prior turns into `message` | No AI-service change, but invents an ad-hoc transcript format on the wrong side of the boundary and blows the token budget silently | Rejected |
| (c) `AIRequest` takes `messages: List[{role, content}]` | The correct fix. Changes the Pydantic models, both service classes, and the Go client structs | The follow-up — belongs in `ai-service-client.md`, not here |

Ship (a), **and say so in the UI copy or an issue** — a user whose follow-up is ignored will read it
as a bug, because it looks exactly like one. When (c) lands, `SendMessage` gains a
`FindByProject` call before the AI call and nothing else in this document changes.

## 6. Proto

> **`.planning/Proto/messages-proto.md` is authoritative for everything in this section**, and
> `project-proto.md` §2 carries the shared conventions. Reproduced here so this document reads on its
> own; if the two differ, those win. Note `AiProvider` (proto) vs `domain.AIProvider` (Go) — the
> spelling difference is deliberate, `project-proto.md` §2.3.

### 6.1 `enums.proto` — modified

```proto
enum MessageRole {
  MESSAGE_ROLE_UNSPECIFIED = 0;
  MESSAGE_ROLE_USER = 1;
  MESSAGE_ROLE_ASSISTANT = 2;
}

enum ChatMode {
  CHAT_MODE_UNSPECIFIED = 0;
  CHAT_MODE_CHAT = 1;
  CHAT_MODE_BUILD = 2;
}
```

Appended beside `LoginProvider` and the `AiProvider` added by `project-model.md` §7.1. Buf lint
`STANDARD` requires the enum-name prefix and a `_UNSPECIFIED` zero value, which is why the generated
constants are verbose and why the handler maps them to domain enums rather than passing them down.

`CHAT_MODE_CHAT` is redundant-looking and unavoidable — the prefix rule applies to every value.

### 6.2 `message_service.proto` — new

A separate service rather than more RPCs on `ProjectService`: messages are their own resource with
their own lifetime, and keeping them apart means the frontend can hold a separate React Query cache
keyed on the thread without invalidating the project list on every send.

```proto
syntax = "proto3";

package api.v1;

import "api/v1/enums.proto";
import "google/api/annotations.proto";

service MessageService {
  rpc ListMessages(ListMessagesRequest) returns (ListMessagesResponse) {
    option (google.api.http) = {
      get : "/api/v1/projects/{project_id}/messages"
    };
  }

  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse) {
    option (google.api.http) = {
      post : "/api/v1/projects/{project_id}/messages"
      body : "*"
    };
  }
}

message Message {
  string id = 1;
  string project_id = 2;
  MessageRole role = 3;
  ChatMode mode = 4;
  AiProvider provider = 5;
  string body = 6;
  string created_at = 7; // RFC3339
}

message ListMessagesRequest { string project_id = 1; }

message ListMessagesResponse { repeated Message messages = 1; }

message SendMessageRequest {
  string project_id = 1;
  string body = 2;
  ChatMode mode = 3;
  AiProvider provider = 4;
}

message SendMessageResponse {
  Message user_message = 1;
  Message assistant_message = 2;
}
```

### 6.3 Four proto decisions

- **`created_at` is an RFC3339 `string`.** Same as `project-model.md` §7.3 — Postgres generates the
  value, the proto only carries it, and the browser formats it. Always
  `t.UTC().Format(time.RFC3339)` in the converter, never `t.String()`.
- **No `updated_at` on the wire.** A message is immutable (§1); exposing a field that never differs
  from `created_at` invites the frontend to render it.
- **`SendMessage` returns both messages.** The client gets server-assigned ids and timestamps for
  the user's own message rather than inventing them, so an optimistic append can be reconciled
  rather than duplicated.
- **`ListMessages` has no pagination.** §7.2.

## 7. Application service

New file `application/services/message_service.go`.

```go
type MessageRepository interface {
    Create(ctx context.Context, message *domain.Message) (*domain.Message, error)
    FindByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Message, error)
}

// ⏸ Polish. The plain-LLM half of the AI-service client, used only by Chat mode (§8).
// Named for the use case, not the transport.
type ChatResponder interface {
    Query(ctx context.Context, provider domain.AIProvider, message string) (string, error)
}
```

**The MVP declares `SandboxCreator` and `CodeAgentRunner` here instead** — see §9.0. `ChatResponder`
arrives with Chat mode.

`MessageService` composes **`*ProjectService`**, not `ProjectRepository`, so the ownership check
lives in exactly one place (`project-model.md` §6.2). This mirrors `AccountService`, which composes
oauth + auth + user rather than reaching for repositories. It also takes `ProjectRepository`
directly for the one write `ProjectService` does not expose — `UpdateProvider` (§9.3).

### 7.1 `List(ctx, userID, projectID)`

Ownership-check through `projectService.Get`, then `FindByProject`. Returns an empty slice for an
empty thread (§4.3).

### 7.2 Threads are not paginated, on purpose

`FindByProject` loads the whole thread in one query. A project has a handful of messages and the
Workspace renders all of them; a cursor would be machinery with no caller, and the keyset codec
(`cursor.go`) is built for `DESC` listing, not `ASC` threads.

The consequence is real and stated rather than discovered later: a very long thread is one large
response, and Build messages can be verbose. Revisit when threads actually grow — the index already
supports a cursor when it is wanted.

**`GetProject` does not embed messages.** They are a separate RPC, so opening a project is two
requests. That is the right shape: the sidebar lists projects without ever loading a thread, and the
thread refetches after a send without re-fetching the project.

## 8. Send — Chat mode

> **⏸ Polish.** Chat needs `Query`, which the MVP client does not implement
> (`ai-service-client.md` phase split). The demo sends Build messages only.
>
> **§8.2 and §8.3 are MVP** and apply to Build unchanged — the user's message is persisted before the
> agent call and is not rolled back if it fails, and the same validation runs. Only §8.1's step 6
> (`chat.Query`) and step 5's early return differ.

### 8.1 The flow

```
1. project := projectService.Get(ctx, userID, projectID)     → ownership; NotFound if not theirs
2. validate: body non-empty and <= 10_000 chars; mode and provider not Unspecified
3. userMsg := messageRepo.Create(NewMessage(projectID, User, mode, provider, body))
4. if provider != project.Provider(): projectRepo.UpdateProvider(ctx, projectID, provider.String())
5. if mode == Build: return Unimplemented                    → §9.4
6. content := chat.Query(ctx, provider, body)                → POST /ai-service/v1/{provider}/query
7. asstMsg := messageRepo.Create(NewMessage(projectID, Assistant, mode, provider, content))
8. return userMsg, asstMsg
```

Every method wraps with `domain.WrapError("message service send", err)` so the innermost
`ErrorType` reaches `errorAdapter.ToConnectError` untouched.

### 8.2 The user's message is persisted before the AI call, and stays if it fails

Step 3 commits before step 6. If the LLM call fails, the user message is already in the thread and
**is not rolled back**. The RPC returns the error; the frontend refetches and shows the user's
message with no reply.

This is the right trade. Deleting what the user typed because a downstream service was unavailable
loses their words to protect a symmetry nobody asked for. An unanswered message is legible; a
vanished one is not.

The cost is a thread that can contain user messages with no assistant reply, which the frontend must
render without assuming they pair up. Worth stating in the frontend doc.

### 8.3 Validation

| Check | Limit | Error |
| --- | --- | --- |
| `body` non-empty after trim | — | `ErrorTypeInvalid` |
| `body` length | 10,000 chars | `ErrorTypeInvalid` |
| `mode` | not `CHAT_MODE_UNSPECIFIED` | `ErrorTypeInvalid` |
| `provider` | not `AI_PROVIDER_UNSPECIFIED` | `ErrorTypeInvalid` |

Unspecified enums are **rejected, not defaulted** — the composer always has a concrete selection, so
an unspecified value means a client bug and defaulting would hide it. Same rule as
`project-model.md` §3.

### 8.4 Provider drift

Step 4 keeps `projects.provider` in sync with the last provider the user actually sent with, so
reopening the Workspace restores their selection (`project-model.md` §3). It is a separate,
non-atomic statement; if it fails, log and continue — the message is what matters, and the selector
will simply show the previous provider.

## 9. Send — Build mode

> **⏸ MVP — but implemented synchronously, not as described below.** The flow is in the phase split
> at the top of this document (§9.0). §9.1–§9.4 record why it was originally deferred and what the
> asynchronous contract must look like; that reasoning is Polish and is kept for
> `build-orchestration.md` to inherit.

### 9.1 It cannot be synchronous with today's client

`AIService`'s `http.Client` is constructed once with `Timeout: 30 * time.Second`
(`outbound/ai_service/ai_service.go`) and shared by every call. A code-agent run over a sandbox is
seconds to minutes. Build over that client is a guaranteed timeout, not a slow success.

Raising the timeout is not the fix either: Connect over HTTP/1.1 through Caddy, a browser fetch, and
a user watching a spinner all have their own limits. Build needs to return immediately and report
progress separately.

**The MVP takes the narrower reading of this.** Once `ai-service-client.md` §5 replaces the shared
ceiling with per-call contexts, a two-minute request does succeed locally — browser → Caddy → Go has
no default timeout it trips. What remains true is the *user-experience* half: a multi-minute call
with no progress feedback beyond a spinner. That is a demo-acceptable cost, not a technical
blocker.

### 9.2 It needs a sandbox that probably does not exist

Build calls `POST /ai-service/v1/{provider}/{sandbox_id}/code`. `projects.sandbox_id` is nullable
and documented as usually stale (`project-model.md` §1.1). So Build must first ensure a live sandbox
— which is `RefreshSandbox` in `project-codebase-model.md` §9.1, itself a 10–30 second call.

### 9.3 What it does when it lands

Recorded here so `build-orchestration.md` inherits a contract rather than reinventing one:

1. Ownership check, validation, persist the user message — steps 1–4 above, unchanged.
2. Ensure a live sandbox (`RefreshSandbox`, or reuse if the user just refreshed).
3. Call the code agent.
4. `codebaseRepo.ApplyRun(projectID, resp.Files, filterInstalls(resp.Commands))` —
   `project-codebase-model.md` §7.1 and §8.
5. Persist the assistant message with **`resp.Summary` as the body** — never the agent's own account
   of which files it wrote. The summary is model prose; the file list is observed fact and belongs
   in `project_codebases`. That split is the trust boundary (`ARCHITECTURE.md` §7.5) and collapsing
   it here would undo the callback's entire purpose.
6. Debit credits, if `credits.md` says so.

### 9.4 Until then

> **⏸ Superseded by the MVP.** This section described the state where Build returns
> `CodeUnimplemented`. The MVP implements Build (§9.0), so neither the stub nor the disabled toggle
> applies — **Build is the mode the demo uses, and Chat is the one that does not work yet.** Kept
> because it is the correct behaviour for any intermediate state where the AI client is not ready.

`SendMessage` with `mode = CHAT_MODE_BUILD` returns `ErrorTypeUnimplemented` →
`connect.CodeUnimplemented`, after persisting the user message. The message rows, the enum, the
validation and the wire contract are all defined and testable; only execution is missing.

**The frontend's Build toggle should be visibly disabled until this lands** — an enabled control
that always errors is worse than an absent one.

## 10. Handler and wiring

### 10.1 `inbound/handlers/message.go` — new

Modelled on `handlers/user.go`, minus the superuser check — this is product surface, so the
hardcoded `alitsayyed@gmail.com` literal (`ARCHITECTURE.md` §11 #9) must not be copied into it.

```go
type MessageServiceHandler struct {
    apiv1connect.UnimplementedMessageServiceHandler

    messageService *services.MessageService
    authAdapter    *authAdapter.HTTPAuthAdapter
}
```

Both RPCs follow the same four steps: `authAdapter.User(ctx)` (omit those three lines and the RPC is
public), `uuid.Parse` the project id, call the service with `user.ID()`, map errors with
`errorAdapter.ToConnectError`.

A package-level `messageToProto(*domain.Message)` mirrors `userToProto`, plus total mapping helpers
for both directions of `MessageRole` and `ChatMode` — defaulting to unspecified. The domain enums
and the proto enums stay separate because the domain must not import generated code.
`providerToProto` already exists in `handlers/project.go` in the same package; reuse it.

### 10.2 `application/app.go` — modified

```go
// persistance
messageRepo := postgres.NewMessageRepository(db)

// business logic
messageService := services.NewMessageService(messageRepo, projectRepo, projectService, aiservice)

// handlers
messageServiceHandler := handlers.NewMessageServiceHandler(messageService, connectAuthAdapter)
```

and one entry in the `[]*vanguard.Service` slice:

```go
vanguard.NewService(apiv1connect.NewMessageServiceHandler(messageServiceHandler, interceptor)),
```

`aiservice` already exists at the top of `New` and satisfies `ChatResponder` once
`ai-service-client.md` lands. All `services.NewXService(...)` calls must sit **above** the
`services := []*vanguard.Service{...}` line, which shadows the imported `services` package.

## 11. Dependencies and ordering

| Depends on | For | Blocking? |
| --- | --- | --- |
| `project-model.md` | `projects` FK, `ProjectService.Get` ownership check, `ProjectRepository.UpdateProvider`, `domain.AIProvider`, `providerToProto` | **Yes** |
| `ai-service-client.md` | `Query` → `POST /{provider}/query`; `CallAI()` is an empty stub today (`ARCHITECTURE.md` §11 #2) | For Chat replies only |
| `project-codebase-model.md` | nothing — `ApplyRun`'s caller arrives with Build | No |
| `build-orchestration.md` | Build mode execution | No — §9.4 |

`ListMessages` and message persistence work with **no** AI-service dependency at all. If
`ai-service-client.md` is not ready, land this feature with a `ChatResponder` stub that returns a
fixed string: the table, thread rendering, ordering and ownership are then all verifiable, and the
real client drops in behind the port unchanged. Do not stub by returning an error — a responder that
fails is indistinguishable from a broken one in the logs.

## 12. Verifying

No test harness exists in this repo — manual only. Do not claim tests pass.

```bash
make gen
cd api && go build ./... && go vet ./...
cd app && npm run lint && npm run build     # make gen rewrites app/src/gen/
make nuke && make
docker compose exec sql psql -U postgres -d local -c '\d messages'
```

```bash
# create a project first (project-model.md), then:
curl -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' \
  -d '{"body":"what goes in a three-tier pricing page?","mode":"CHAT_MODE_CHAT","provider":"AI_PROVIDER_ANTHROPIC"}'

curl -b "jwt=$JWT" https://local.api.vulx.ai/api/v1/projects/$PID/messages
```

Six checks a happy-path `curl` will not surface:

1. **Unauthenticated is rejected.** `curl -i` without the cookie on both RPCs → `401`.
2. **Another user's project → `404`**, not `403`, and not the thread.
3. **The parent moved.** Note `projects.updated_at` before and after a send; it must change (§4.2).
   Then re-list projects and confirm the project jumped to the top of the sidebar order.
4. **A failed AI call keeps the user message.** Stop the `ai-service` container, send, expect an
   error, then `ListMessages` — the user message must be there with no reply (§8.2).
5. **Provider drift persists.** Send with a different provider than the project's, then
   `GET /api/v1/projects/{id}` and confirm `provider` changed (§8.4).
6. **Build is honestly unimplemented.** Send with `"mode":"CHAT_MODE_BUILD"` → `501`
   (`CodeUnimplemented`), and the user message is persisted (§9.4).

And the one behaviour that will look like a bug and is not: **ask a follow-up question that depends
on the previous turn.** The model will not know what you are referring to. That is §5, it is
expected, and it is fixed in `ai-service-client.md`, not here.

## 13. Out of scope

Build mode execution (`build-orchestration.md`) · conversation history sent to the model — §5 option
(c) (`ai-service-client.md`) · streaming and token-by-token rendering · editing, deleting or
regenerating a message · thread pagination (§7.2) · message search · attachments · credits
(`credits.md`) · a `status` column (§2.3) · tests.
