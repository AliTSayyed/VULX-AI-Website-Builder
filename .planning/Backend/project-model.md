# Project Model — Backend

> A **project** is the durable thing a user builds: an owned, named record with a codebase, a chat
> thread, and a (usually dead) sandbox pointer. This document owns **the `projects` table and its
> full vertical slice** — migration, domain entity, repository, service, handler, proto, wiring.
>
> It does **not** own the codebase or the thread. Those are two more tables in two sibling
> documents: `project-codebase-model.md` (the JSONB file map + `npm install` replay) and
> `messages-model.md`. Both take a FK to `projects.id`; neither changes anything defined here.
>
> Verified against the code on `feature/messages` (2026-09-08). Status markers follow
> `ARCHITECTURE.md`: ✅ built · 🟡 partial · ⛔ not built.

## ⏸ Phase split — MVP vs Polish

**Nearly all of this document is MVP.** A project is where the sandbox pointer lives, so the demo
cannot happen without the table, the domain type, the repository and three RPCs. What defers is the
LLM-generated title and renaming.

| § | Content | Phase |
| --- | --- | --- |
| §1–§4 | why, the table, the `provider` column, domain entities | **MVP** |
| §5.1 `Create`, `FindByID`, `FindAllByUser`, `UpdateSandbox` | | **MVP** |
| §5.1 `UpdateTitle`, `UpdateProvider` | | Polish |
| §5.2–§5.4 | ownership, pagination, the cursor caveat | **MVP** |
| §6.1 `List`, `Get`, `Create` | | **MVP** |
| §6.1 `Rename` | | Polish |
| §6.2 ownership check | | **MVP** — every id-taking method |
| §6.3 provisional title, empty-prompt rejection | | **MVP** |
| §6.3 LLM titling — the goroutine, `WithoutCancel`, sanitising | | Polish |
| §6.4 | why empty projects cannot exist | **MVP** |
| §7 `ListProjects`, `GetProject`, `CreateProject` | | **MVP** |
| §7 `RenameProject` | | Polish |
| §8, §10, §11 | handler, wiring, files, verifying | **MVP**, minus the Polish rows |

### `UpdateSandbox` is MVP, and its caller changes

§5.1 says its caller lives in `project-codebase-model.md`. **In the MVP the caller is
`messages-model.md` §9**: `SendMessage` creates a sandbox on the project's first Build message and
stores the id and URL on the project row. That is the entire sandbox lifecycle in the MVP — no
refresh, no replay, no expiry handling.

`sandbox_id` and `preview_url` stay exactly as specified in §2.1. They are still "last known, usually
stale" — the MVP just has no way to recover when they go stale, which is acceptable for a demo that
outlives its own sandbox (`.planning/AI-Service/sandbox-service.md` §4 raises the timeout so it does).

### The titler is stubbed, not called

§9 already describes this as the smaller sequencing step; the MVP takes it. `ProjectTitler` is
declared as written, and satisfied by a trivial in-package implementation that returns the
provisional title unchanged.

So the MVP has **no LLM call in `CreateProject`** — no goroutine, no `context.WithoutCancel`, no
output sanitising, and none of the failure modes those bring. Projects are titled by truncating the
first prompt to ~60 characters, which is legible and permanent until Polish lands.

That also means `CreateProject` has **no dependency on `ai-service-client.md`** in the MVP. §9's
ordering discussion resolves to: build this document first, on its own.


## 1. Why this exists, and what "project" means

`ARCHITECTURE.md` §10 lists the conversation/message model as ⛔ — nothing exists. The entire
logged-in screen reads `app/src/components/session/mock.ts`. This feature is the first table that
fixture is replaced by, and it is the one the sidebar lists.

**The naming changed deliberately.** The frontend design docs and the fixture say *conversation*.
That name describes the chat log, which is the least durable thing in the record. What a user owns
is a **project**: the code survives, the sandbox does not, and the chat is the history of how the
project got built. Everything server-side — table, domain type, service, handler, proto, REST path —
uses `project` from the start. The frontend keeps saying "conversation" until it is wired; renaming
`mock.ts`, `ConversationList` and `logged_in_design.md` is a separate, later change and is **not**
blocked by this one.

### 1.1 The reframe that makes the sandbox tolerable

**Postgres is the truth; the E2B sandbox is a disposable materialization of it.**

Sandboxes expire after five minutes with no keep-alive (`ARCHITECTURE.md` §7.6). If the sandbox is
the codebase, every project dies in five minutes. Inverting it — persisting the agent's file writes
and rebuilding a sandbox on demand — makes expiry a non-event.

The consequence for *this* table: `sandbox_id` and `preview_url` are a **cache pointer that is
usually stale**, not state. They are nullable, they are written on refresh, and nothing may treat a
non-empty `preview_url` as proof that a sandbox is alive.

There is deliberately **no `sandbox_expires_at` column and no lazy-recreate logic**. The user drives
recreation with an explicit refresh button, so the API never has to reason about liveness — refresh
unconditionally creates a new sandbox and replays. That decision deletes a column and an entire
class of expiry bookkeeping. `project-codebase-model.md` owns the refresh RPC itself.

## 2. The table

### 2.1 `00004_add_projects.sql`

Migrations are `go:embed`ed and run by `sql-migrate` on API boot (`postgres.NewDb`); a failure
panics the API. Never edit an applied file — `make nuke` to reset.

```sql
-- +migrate Up
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    sandbox_id VARCHAR(255),
    preview_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_projects_user_updated
    ON projects (user_id, updated_at DESC, id DESC);

-- +migrate Down
DROP TABLE IF EXISTS projects CASCADE;
```

### 2.2 Column by column

| Column | Why it is here |
| --- | --- |
| `user_id` | **Load-bearing, not bookkeeping.** There is no public-route allowlist and no per-route config (`ARCHITECTURE.md` §6.4) — a handler is protected only because it calls `authAdapter.User(ctx)`. Authorisation is therefore a column comparison inside the service, and the repository has to return enough to make it. See §5.2. |
| `title` | Provisional at creation, overwritten once by an LLM, renameable by the user forever after. §6. |
| `provider` | The provider **currently selected** for this project. Restored on reload and reused by refresh, so a user who built with Claude does not silently come back on OpenAI. §3. |
| `sandbox_id`, `preview_url` | Last known sandbox. Nullable, usually stale. §1.1. |
| `updated_at` | What the sidebar sorts and paginates on — *last activity*, not creation. §5.3. |

### 2.3 Notes on the DDL

- **`NOT NULL` on the timestamps diverges from `users` and `user_auth_providers`,** which declare
  `DEFAULT NOW()` without it. Intentional: keyset pagination reads `updated_at` off the last row of
  a page, and a `NULL` there yields an unusable cursor. The existing tables are not being changed.
- **The index matches the only access path** — a user's projects, newest-activity-first. The
  trailing `id` mirrors the `ORDER BY` tiebreaker.
- **`provider` is `VARCHAR`, not a Postgres `ENUM`.** This follows `user_auth_providers.provider`,
  which stores `"google"` as text and parses it in the domain (`ParseLoginProvider`). A Postgres
  enum needs a migration to add a value; adding a fourth model provider should be a Go and proto
  change, not DDL.
- **Postgres owns timestamp generation.** `DEFAULT NOW()` on insert, `NOW()` in every `UPDATE`.
  Go never sets `created_at` or `updated_at` — which is why `Create` and `UpdateTitle` use
  `RETURNING` and re-hydrate the row rather than trusting what they sent.
- **`citext` is enabled (`00001`) but unused.** Nothing here needs it.
- **No `deleted_at`.** Delete is out of scope (§10).

## 3. The `provider` column

Two different questions get confused here, so both answers are written down:

| Question | Answered by | Lives in |
| --- | --- | --- |
| *Which model is this project set to right now?* | `projects.provider` | this document |
| *Which model produced this particular message?* | `messages.provider` | `messages-model.md` |

They are not redundant. The first is a sticky UI preference restored when the Workspace opens and
consumed by the refresh path; the second is an immutable historical record, because a user can
switch providers mid-thread.

- **Set at creation** from `CreateProjectRequest.provider`. An `AI_PROVIDER_UNSPECIFIED` is rejected
  as `ErrorTypeInvalid` rather than silently defaulted — the composer always has a concrete
  selection, so an unspecified value means a client bug, and defaulting would hide it.
- **Updated when the user sends with a different provider.** The repository method `UpdateProvider`
  is defined here (§5.1); its only caller lives in `messages-model.md`. Defining it here keeps all
  `projects` SQL in one file.
- `NOT NULL`, so every read is total and no caller handles a missing provider.

## 4. Domain

Follow `domain/user.go` exactly: private fields, nil-safe accessors, a `NewX` that validates and a
`RestoreX` that does not — data already in the database is trusted (`ARCHITECTURE.md` §5.5).

### 4.1 `domain/project.go` — new

```go
type Project struct {
    id         uuid.UUID
    userID     uuid.UUID
    title      string
    provider   AIProvider
    sandboxID  string      // "" until a sandbox has been created
    previewURL string      // "" until a sandbox has been created
    createdAt  time.Time
    updatedAt  time.Time
}
```

- `NewProject(userID uuid.UUID, title string, provider AIProvider) (*Project, error)` — rejects
  `uuid.Nil`, an empty title, and `AIProviderUnspecified`. Sentinels `ErrProjectTitleEmpty`,
  `ErrProjectUserEmpty`, `ErrProjectProviderUnspecified`, all `ErrorTypeInvalid`, declared at the
  top of the file as `user.go` does.
- `RestoreProject(...)` takes `provider` as a **string** and parses it, mirroring
  `RestoreUserFromProvider`, which takes `providerName string` and calls `ParseLoginProvider`. That
  keeps the persistence row struct free of domain enum types.
- Accessors: `ID`, `UserID`, `Title`, `Provider`, `SandboxID`, `PreviewURL`, `CreatedAt`,
  `UpdatedAt`, plus `HasSandbox() bool`.
- **Nullable columns map to the zero string, not `*string`.** The domain should not make callers
  reason about three states when two suffice; `HasSandbox()` reads better at the call site than a
  nil check, and it is the only question anyone asks of those two fields.

### 4.2 `domain/ai_provider.go` — new

`AIProvider` gets its own file for the same reason `LoginProvider` lives in `login.go` rather than
`user.go`: it is referenced by more than the entity that carries it. `messages-model.md` stores it
per message, and the AI-service client uses its `String()` as a **URL path segment**.

An `int` enum with `iota`, a `String()` method and a `ParseAIProvider` function — the
`LoginProvider` pattern verbatim, including an `AIProviderUnspecified` zero value whose `String()`
is `"unspecified"`.

| Value | Wire string |
| --- | --- |
| `AIProviderUnspecified` | `unspecified` |
| `AIProviderOpenAI` | `openai` |
| `AIProviderGoogle` | `google` |
| `AIProviderAnthropic` | `anthropic` |

**Those three strings are exact and non-negotiable.** The FastAPI routes mount at
`/ai-service/v1/{provider}/...` and reject `gemini` and `claude` (`ARCHITECTURE.md` §7.1). The
frontend labels them "Gemini" and "Claude"; that label belongs in the UI and must never reach this
enum.

## 5. Persistence

New file `postgres/project_repository.go`, following `user_repository.go`: a package-private row
struct with `db` tags, a `ToDomain()` method, a constructor taking `*sqlx.DB`, hand-written SQL, and
every error wrapped as a `*domain.Error` with a meaningful type.

```go
type Project struct {
    ID         uuid.UUID      `db:"id"`
    UserID     uuid.UUID      `db:"user_id"`
    Title      string         `db:"title"`
    Provider   string         `db:"provider"`
    SandboxID  sql.NullString `db:"sandbox_id"`
    PreviewURL sql.NullString `db:"preview_url"`
    CreatedAt  time.Time      `db:"created_at"`
    UpdatedAt  time.Time      `db:"updated_at"`
}
```

`sql.NullString` is confined to this struct; `ToDomain()` flattens `.String` (empty when invalid)
before calling `RestoreProject`. The nullability stops at the persistence boundary — §4.1.

### 5.1 Methods

| Method | Signature | Notes |
| --- | --- | --- |
| `Create` | `(ctx, *domain.Project) (*domain.Project, error)` | `RETURNING` every column, re-hydrated through `ToDomain()` so the DB-assigned timestamps win |
| `FindByID` | `(ctx, id uuid.UUID) (*domain.Project, error)` | **Not user-scoped** — §5.2 |
| `FindAllByUser` | `(ctx, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error)` | keyset paged on `updated_at` — §5.3 |
| `UpdateTitle` | `(ctx, id uuid.UUID, title string) (*domain.Project, error)` | also bumps `updated_at` |
| `UpdateProvider` | `(ctx, id uuid.UUID, provider string) error` | caller lives in `messages-model.md` — §3 |
| `UpdateSandbox` | `(ctx, id uuid.UUID, sandboxID, previewURL string) error` | caller lives in `project-codebase-model.md` |

The last two methods have no caller when this feature lands; they are here so that every statement
touching `projects` lives in one file, which is the point of the repository.

**There is deliberately no `Touch`.** The sidebar orders by last activity, so sending a message must
move the parent row — but that bump happens inside `messageRepo.Create`, as a data-modifying CTE in
the same statement as the insert (`messages-model.md` §4.2). A standalone
`UPDATE projects SET updated_at = NOW()` would have no caller: refresh deliberately does not bump
(`project-codebase-model.md` §9.1) and `UpdateTitle` bumps inline.

That is the one accepted exception to "every statement touching `projects` lives in this file" —
atomicity won over the file boundary. Someone grepping for `UPDATE projects` will find one hit
outside this repository, in `message_repository.go`, and it is intentional.

### 5.2 `FindByID` is not user-scoped, and `NotFound` is not free

`FindByID` returns the project including its `user_id` and does **not** filter by user. The service
compares and returns `ErrorTypeNotFound` when they differ (§6.2). Filtering inside the query would
make "not yours" and "does not exist" indistinguishable to the service — fine for the response,
which should be identical either way, but it also erases the difference from the logs, where it is
exactly what you want when debugging.

Both `FindByID` and `UpdateTitle` check `errors.Is(err, sql.ErrNoRows)` and return
`ErrorTypeNotFound`. That follows `user_repository.go`'s `FindByEmail`, **not** its `FindByID`,
which maps every failure including `ErrNoRows` to `ErrorTypeInternal`. The existing method is not
being fixed here; the inconsistency is noted so it is not copied.

This matters downstream: `errorAdapter.ToConnectError` maps `NotFound` → `CodeNotFound`. Without the
distinction, opening a stale project id would surface to the browser as a generic `Internal` and the
frontend could not tell a deleted project from a broken database.

### 5.3 Pagination — the one thing that is not a copy of `user_repository.go`

The sidebar renders a relative time meaning *last activity* and orders newest-first, so
`FindAllByUser` orders by `updated_at DESC, id DESC` and the cursor must encode `updated_at`.
Encoding `created_at` while ordering by `updated_at` makes the `WHERE` and the `ORDER BY` disagree,
and pages then overlap or skip rows.

`cursor.go` already does the right thing — the token is an opaque base64 RFC3339Nano timestamp and
nothing binds it to any particular column. It just names the value wrong:

```go
func encodeToken(createdAt time.Time) string   // ← the parameter name is the only problem
```

**Rename the parameter to `ts` in `encodeToken` and `decodeToken` and update the doc comment.** Zero
behaviour change, no call-site churn (`user_repository.go` passes positionally), and it stops the
next reader from concluding the project list is sorted wrongly. Do not add a second cursor codec.

Two conventions carried over unchanged: query `limit+1` rows, trim to `limit`, take the token from
the last kept row; and derive `HasMore` from the token being non-empty rather than counting.

### 5.4 Inherited limit: the cursor has no tiebreaker

`FindAll` orders by `created_at DESC, id DESC` but encodes only the timestamp, so rows sharing an
exact timestamp can be dropped or repeated across a page boundary. `FindAllByUser` inherits this
with `updated_at`.

Two projects touched in the same nanosecond is not an MVP concern, but the bug is being knowingly
duplicated and should not be discovered later as a surprise. Fixing it means encoding
`(timestamp, id)` in the token and a composite `WHERE`, which changes `user_repository.go` too. Out
of scope here — recorded, not hidden.

## 6. Application service

New file `application/services/project_service.go`. Ports are declared **here**, in the file of the
consumer, not in a `ports/` package (`ARCHITECTURE.md` §5.2).

```go
type ProjectRepository interface {
    Create(ctx context.Context, project *domain.Project) (*domain.Project, error)
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
    FindAllByUser(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error)
    UpdateTitle(ctx context.Context, id uuid.UUID, title string) (*domain.Project, error)
    UpdateProvider(ctx context.Context, id uuid.UUID, provider string) error
    UpdateSandbox(ctx context.Context, id uuid.UUID, sandboxID, previewURL string) error
}

// Narrow on purpose. The method is Query, not GenerateTitle: the client is transport only,
// and the titling prompt is a product decision that stays in this service — see §6.3.
type ProjectTitler interface {
    Query(ctx context.Context, provider domain.AIProvider, message string) (string, error)
}
```

`ProjectTitler` is implemented by the AI-service client and is **the one hard dependency this
feature has on another document** — see §6.3 and §9. It is deliberately identical in shape to
`ChatResponder` in `message_service.go`; two consumers each declaring the narrow interface they
need, satisfied by one type, is ordinary interface segregation, not duplication.

### 6.1 Methods

| Method | Behaviour |
| --- | --- |
| `List(ctx, userID, limit, token)` | `limit = utils.Clamp(limit, 10, 100)` exactly as `UserService.List` does, then `FindAllByUser` |
| `Get(ctx, userID, id)` | `FindByID`, then the ownership check (§6.2) |
| `Create(ctx, userID, firstPrompt, provider)` | §6.3 |
| `Rename(ctx, userID, id, title)` | ownership check, trim, reject empty, clamp to 255, `UpdateTitle` |

Every method wraps with `domain.WrapError("project service <verb>", err)`, so the innermost
`ErrorType` propagates untouched to the single place that maps it to a transport code.

### 6.2 The ownership check

```go
project, err := s.projectRepo.FindByID(ctx, id)
if err != nil {
    return nil, domain.WrapError("project service get", err)
}
if project.UserID() != userID {
    return nil, domain.NewError(domain.ErrorTypeNotFound, fmt.Errorf("project %s not found", id))
}
```

**`NotFound`, not `PermissionDenied`.** `PermissionDenied` confirms the id exists and belongs to
someone else, which is an enumeration oracle. The user cannot distinguish the two cases and does not
need to; the server log can, because the repository already returned the real owner.

Every method that takes a project id runs this check. It is three lines and there is no interceptor
that will do it — omit it and the RPC leaks other users' projects.

### 6.3 Create, and how the title gets made

`CreateProject` does **not** block on an LLM call.

1. Validate the provider (reject unspecified — §3) and the prompt (reject empty).
2. Derive a **provisional title**: the first prompt, whitespace-collapsed and truncated to ~60
   characters on a word boundary. This is a pure function, always succeeds, and is the permanent
   fallback.
3. `domain.NewProject(userID, provisionalTitle, provider)` → `projectRepo.Create` → **return
   immediately.**
4. Generate the real title in the background and `UpdateTitle` when it lands.

Blocking would put a full LLM round trip in front of the user's first screen for a cosmetic string.
The Workspace should open instantly with a readable title that later improves.

**Background, not durable.** A goroutine, not Temporal. Durability buys nothing here — a lost title
degrades to the provisional one, which is already acceptable output. Temporal's value is the build
pipeline (`build-orchestration.md`), not this.

One thing this **will** get wrong if written naively: the RPC's context is cancelled the moment the
handler returns, so the goroutine must not inherit it.

```go
go func() {
    ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
    defer cancel()
    ...
}()
```

`context.WithoutCancel` keeps any request-scoped values while detaching the cancellation. Without
it, the title call is killed mid-flight every time and the feature silently never works.

**Generated exactly once, at creation.** That single rule is what makes rename work with no extra
column: a manual rename always wins, because generation never runs again. Do not add a
`title_generated` flag.

**Sanitise the output.** An LLM asked for a title will sometimes return a quoted string, a trailing
period, a newline, or a whole sentence. Strip surrounding quotes and newlines, collapse whitespace,
truncate to 60 characters, and **fall back to the provisional title** on error, empty output, or
anything that survives sanitising as an empty string. A three-paragraph "title" would wreck the
sidebar layout. The column is `VARCHAR(255)`; the service clamps well before that.

**Not billed.** Title generation is system overhead, not user work. `credits.md` does not need to
account for it.

### 6.4 Rejecting an empty prompt is what stops empty projects existing

Step 1's "reject an empty prompt" is not routine validation — it is the whole mechanism by which a
user who opens **+ New build** and types nothing leaves nothing behind.

`first_prompt` is what the title is derived from, so a project without one has nothing to name it.
There is therefore no API call that can produce an empty project, and no server-side machinery is
needed to clean one up: no `is_draft` column, no TTL sweep, no background reaper.

**The frontend obligation this creates:** *+ New build must not call `CreateProject`.* It is a
purely client-side action that opens an empty Workspace with **no project id**. The project comes
into existence on the first send, so both entry points run the same two-call sequence:

```
first send  →  CreateProject(first_prompt, provider)  →  SendMessage(project_id, body, …)
later sends →  SendMessage(project_id, body, …)
```

Home's hero box and + New build differ only in which screen the composer sits on. Until the first
send there is no sidebar row and `RefreshSandbox` is unavailable, both of which are correct — the
project does not exist yet.

**The one window that can still leave an empty project:** `CreateProject` succeeds and `SendMessage`
fails before persisting. It is narrow — `SendMessage` commits the user message *before* the AI call
(`messages-model.md` §8.2), so an AI-service outage still leaves the message; only a transport error
or a failed insert opens the gap. Accepted, with a single client-side retry on the send. It is the
cost of keeping project creation and message creation as separate RPCs with separate purposes, which
is deliberate — folding them together would make `ProjectService` depend on `MessageRepository`.

Which provider titles the project: **the one the user selected**, passed straight through. Zero new
config, and it keeps the AI-service call on a code path the user already exercised. A fixed cheap
model would be cheaper but adds a config key and a second failure mode for a call that already has a
safe fallback.

## 7. Proto

> **`.planning/Proto/project-proto.md` is authoritative for everything in this section.** It carries
> the buf-lint conventions, field-number discipline, REST annotation mechanics and codegen details
> that apply across all three proto documents. What follows is the same contract, reproduced so this
> document reads on its own — if the two ever differ, that one wins.
>
> **The proto enum is `AiProvider`, while the Go domain type is `AIProvider`.** That is not a typo:
> `protoc-gen-es` computes the prefix to strip with a snake-caser that inserts an underscore before
> every capital, so `AIProvider` would expect `A_I_PROVIDER_` and leave the frontend writing
> `AIProvider.AI_PROVIDER_OPENAI`. `AiProvider` expects `AI_PROVIDER_` and strips to
> `AiProvider.OPENAI`. Verified against `@bufbuild/protobuf`'s `registry.js` — see
> `project-proto.md` §2.3. Go keeps the initialism because Go convention says so; the handler already
> maps between the two enums.

### 7.1 `enums.proto` — modified

```proto
enum AiProvider {
  AI_PROVIDER_UNSPECIFIED = 0;
  AI_PROVIDER_OPENAI = 1;
  AI_PROVIDER_GOOGLE = 2;
  AI_PROVIDER_ANTHROPIC = 3;
}
```

Appended to the existing file alongside `LoginProvider`. Buf lint `STANDARD` requires the
`ENUM_NAME_UPPER_SNAKE_CASE` prefix and a zero value named `_UNSPECIFIED`, which is why the
generated Go constants are ugly and why the handler maps them to `domain.AIProvider` rather than
passing them down. `messages-model.md` reuses this enum and adds `ChatMode` beside it.

### 7.2 `project_service.proto` — new

```proto
syntax = "proto3";

package api.v1;

import "api/v1/enums.proto";
import "google/api/annotations.proto";

service ProjectService {
  rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse) {
    option (google.api.http) = {
      get : "/api/v1/projects"
    };
  }

  rpc GetProject(GetProjectRequest) returns (GetProjectResponse) {
    option (google.api.http) = {
      get : "/api/v1/projects/{id}"
    };
  }

  rpc CreateProject(CreateProjectRequest) returns (CreateProjectResponse) {
    option (google.api.http) = {
      post : "/api/v1/projects"
      body : "*"
    };
  }

  rpc RenameProject(RenameProjectRequest) returns (RenameProjectResponse) {
    option (google.api.http) = {
      patch : "/api/v1/projects/{id}"
      body : "*"
    };
  }
}

message Project {
  string id = 1;
  string title = 2;
  AiProvider provider = 3;
  string sandbox_id = 4;
  string preview_url = 5;
  string created_at = 6; // RFC3339
  string updated_at = 7; // RFC3339
}

message ListProjectsRequest {
  int64 limit = 1;
  string token = 2;
}

message ListProjectsResponse {
  repeated Project projects = 1;
  string token = 2;
  bool has_more = 3;
}

message GetProjectRequest { string id = 1; }

message GetProjectResponse { Project project = 1; }

message CreateProjectRequest {
  string first_prompt = 1;
  AiProvider provider = 2;
}

message CreateProjectResponse { Project project = 1; }

message RenameProjectRequest {
  string id = 1;
  string title = 2;
}

message RenameProjectResponse { Project project = 1; }
```

### 7.3 Four proto decisions worth stating

- **`limit`/`token`, not `page_size`/`page_token`.** Matches `ListUsersRequest` and the
  `domain.Page[T]` field names. Consistency with the repo beats consistency with Google's AIP here,
  since one convention is already established and generated into the TS client.
- **`user_id` is not in `Project` and not in any request.** It comes from the JWT via
  `authAdapter.User(ctx)`. Accepting it from the client would be an authorisation bug wearing a
  field name.
- **Timestamps are RFC3339 `string`s, not `google.protobuf.Timestamp`.** Postgres generates the
  values (§2.3); the proto only has to carry them. A plain string keeps the repo free of
  well-known-type imports, and the browser reads it with `new Date(s)` — no `timestampDate()` helper
  from `@bufbuild/protobuf/wkt`, no `{seconds, nanos}` unwrapping in the sidebar.
  The trade is real and small: the schema no longer enforces that these are instants, so the
  **format is a convention the handler must hold up** — always `t.UTC().Format(time.RFC3339)`, never
  `t.String()`, which emits Go's `2006-01-02 15:04:05.999999999 -0700 MST` layout that `new Date()`
  parses inconsistently across browsers. RFC3339 always carries its offset, so this survives the
  binary wire format that `useServiceClient.ts` switches to once a non-local `NEXT_PUBLIC_API_URL`
  is set.
  The API sends an **absolute instant and nothing else** — formatting it as "2 hours ago" is the
  browser's job. Worth stating because the fixture invites the opposite: `mock.ts` declares
  `updatedAt: "2 hours ago"` as a literal string, so a naive wiring would have the server render
  relative time. That would be wrong twice over — it bakes the reader's clock and locale into a
  cached response, and it goes stale the moment the row is cached or the tab is left open.
- **`PATCH` is the first non-GET/POST verb in this repo.** `google.api.http` and Vanguard both
  support it, but nothing here has exercised it. If transcoding misbehaves, the fallback is
  `post: "/api/v1/projects/{id}/rename"` — the Connect/gRPC path is unaffected either way.

### 7.4 Regeneration

`make gen` fans out to three trees, **none of which may be hand-edited** — Go stubs under
`inbound/grpc/gen/`, TS clients under `app/src/gen/`, and the OpenAPI doc served at `/docs/`. It
invokes `app/node_modules/.bin/protoc-gen-es`, so the frontend's npm install must have happened
first. `make plint` formats the protos with clang-format; run it before committing.

## 8. Handler and wiring

### 8.1 `inbound/handlers/project.go` — new

Modelled on `handlers/user.go`, minus the superuser check — projects are product surface, not admin
endpoints, so the hardcoded `alitsayyed@gmail.com` literal (`ARCHITECTURE.md` §11 #9) must **not**
be copied into this file.

Each method is the same four steps:

```go
user, err := authAdapter.User(ctx)   // omit these three lines and the RPC is public
if err != nil {
    return nil, err                  // already a connect.CodeUnauthenticated error
}
```

then parse the id (`uuid.Parse`, mapped through `errorAdapter.ToConnectError`), call the service
with `user.ID()`, and map the result with a package-level `projectToProto(*domain.Project)`
converter mirroring `userToProto`. That converter is the **only** place timestamps are formatted:
`CreatedAt: project.CreatedAt().UTC().Format(time.RFC3339)` and the same for `UpdatedAt` (§7.3). Handlers never construct a `connect.Error` for a domain failure —
`errorAdapter.ToConnectError` is the single translation point.

Two mapping helpers live in this file, both total and both defaulting to unspecified:
`providerToProto(domain.AIProvider) apiv1.AIProvider` and its inverse. They mirror the
`LoginProvider` mapping already in `handlers/account.go`. The domain enum and the proto enum are
kept separate on purpose — the domain must not import generated code.

### 8.2 `application/app.go` — modified

Four lines, in the existing order and under the existing comment banners:

```go
// persistance
projectRepo := postgres.NewProjectRepository(db)

// business logic
projectService := services.NewProjectService(projectRepo, aiservice)

// handlers
projectServiceHandler := handlers.NewProjectServiceHandler(projectService, connectAuthAdapter)
```

and one entry in the `[]*vanguard.Service` slice:

```go
vanguard.NewService(apiv1connect.NewProjectServiceHandler(projectServiceHandler, interceptor)),
```

The `aiservice` variable already exists at the top of `New` — it is passed as the `ProjectTitler`.
Note the existing shadowing quirk in this file: the local `services := []*vanguard.Service{...}`
shadows the imported `services` package, so **every `services.NewXService(...)` call must appear
above that line.** It already does; keep it that way.

## 9. Dependency on `ai-service-client.md`

`ProjectTitler.Query` is satisfied by `AIService.Query`, hitting
`POST /ai-service/v1/{provider}/query` → `{content}`. That route is ✅ built and needs no sandbox.
The Go client that calls it is not: `CallAI()` is an empty stub returning `nil`
(`ARCHITECTURE.md` §11 #2).

**This service sends a finished prompt, not a title request.** Building the instruction that turns
the user's first message into a title — and clamping, stripping and falling back on the reply — is
§6.3's job and stays here. The client contributes no prompt text of its own
(`ai-service-client.md` §2).

**This is the only ordering constraint in the feature.** Everything else — migration, domain,
repository, service, proto, handler, wiring — is independent of it.

Sequence it either way:

- **`ai-service-client.md` first**, and this feature lands complete; or
- **this feature first**, with a trivial in-package `ProjectTitler` implementation that returns the
  provisional title unchanged. Projects then work end to end with truncated titles, and the LLM
  titler is dropped in behind the existing interface with no change to anything above it.

The second is the smaller step and the port makes it free. Do not stub by returning an error — a
titler that fails is indistinguishable from a broken one in the logs.

## 10. Files

**Created**

| File | Contents |
| --- | --- |
| `persistence/postgres/migrations/00004_add_projects.sql` | §2.1 |
| `domain/project.go` | `Project`, `NewProject`, `RestoreProject`, accessors, `HasSandbox` |
| `domain/ai_provider.go` | `AIProvider`, `String()`, `ParseAIProvider` |
| `persistence/postgres/project_repository.go` | row struct, `ToDomain`, 6 methods (§5.1) |
| `application/services/project_service.go` | `ProjectRepository` + `ProjectTitler` ports, `ProjectService`, 4 methods |
| `inbound/handlers/project.go` | 4 RPCs, `projectToProto`, provider mapping helpers |
| `proto/api/v1/project_service.proto` | §7.2 |

**Modified**

| File | Change |
| --- | --- |
| `proto/api/v1/enums.proto` | append `AiProvider` (§7.1) |
| `persistence/postgres/cursor.go` | rename the `createdAt` parameter to `ts` in both functions; update the comment (§5.3) |
| `application/app.go` | repo + service + handler construction, one `vanguard.NewService` entry (§8.2) |

**Generated by `make gen` — never hand-edited**

`api/internal/infrastructure/inbound/grpc/gen/` · `app/src/gen/` ·
`api/internal/infrastructure/inbound/http/handlers/openapi.yaml`

**Untouched on purpose:** `domain/sandbox.go` — it models a `uuid` primary key while E2B returns an
opaque provider string, so it does not fit; leave it unused rather than reshaping it here ·
`postgres/user_repository.go` (§5.2, §5.4) · `handlers/user.go` · the frontend (§1).

## 11. Verifying

There is no test harness in this repo — manual only. Do not claim tests pass.

```bash
make gen                              # after editing the protos
cd api && go build ./... && go vet ./...
cd app && npm run lint && npm run build   # make gen rewrites app/src/gen/
```

Apply the migration — it runs on API boot, and a bad one panics the API:

```bash
make nuke                             # migrations are append-only; a half-applied 00004 must not linger
make
docker compose logs api | grep -i migrat
docker compose exec sql psql -U postgres -d local -c '\d projects'
```

Exercise the RPCs over REST through Caddy. All four are protected, so the cookie is required —
log in through the browser first and reuse its `jwt`, or drive it with `curl -b`:

```bash
# unauthenticated must be rejected — this is the check most likely to be silently missing
curl -i https://local.api.vulx.ai/api/v1/projects                    # expect 401

curl -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects \
  -H 'content-type: application/json' \
  -d '{"first_prompt":"a portfolio for a photographer","provider":"AI_PROVIDER_ANTHROPIC"}'

curl -b "jwt=$JWT" https://local.api.vulx.ai/api/v1/projects
curl -b "jwt=$JWT" -X PATCH https://local.api.vulx.ai/api/v1/projects/<id> \
  -H 'content-type: application/json' -d '{"title":"Renamed"}'
```

Three things to check that a happy-path `curl` will not surface:

1. **Ownership.** Log in as a second Google account and `GET /api/v1/projects/<first user's id>`.
   Expect `404`, not `403` and not the project.
2. **The title actually updates.** Create a project, then re-list ~2 seconds later — the title must
   have changed from the truncated prompt. If it never changes, the goroutine is inheriting the
   cancelled request context (§6.3).
3. **`AI_PROVIDER_UNSPECIFIED` is rejected** on create with `400`, not defaulted.

## 12. Out of scope

Messages and the chat thread (`messages-model.md`) · the codebase JSONB, `npm install` replay and the
refresh RPC (`project-codebase-model.md`) · the AI-service client itself (`ai-service-client.md`) ·
deleting or archiving a project · sharing a project between users · credits (`credits.md`) ·
streaming · fixing the cursor tiebreaker (§5.4) or `user_repository.go`'s `FindByID` error mapping
(§5.2) · renaming the frontend's "conversation" vocabulary (§1) · tests.
