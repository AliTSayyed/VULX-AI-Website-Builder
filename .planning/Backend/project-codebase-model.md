# Project Codebase Model — Backend

> The durable copy of what the agent built, and the machinery to replay it into a fresh sandbox.
> This document owns **the `project_codebases` table and its vertical slice** — migration, domain
> entity, repository, service, the `RefreshSandbox` RPC, and wiring.
>
> It depends on `project-model.md` (the FK target) and on `ai-service-client.md` (three routes it
> cannot work without). One correctness fix in the Python service rides alongside it — §4.
>
> Verified against the code on `feature/messages` (2026-09-08). Status markers follow
> `ARCHITECTURE.md`: ✅ built · 🟡 partial · ⛔ not built.
>
> ## ⏸ Phase: POLISH — none of this is in the MVP
>
> **The MVP does not store the codebase and does not replay it.** A demo runs inside one sandbox's
> lifetime: type a prompt, the agent writes into a live sandbox, you watch it in the iframe. Nothing
> needs to survive the sandbox dying.
>
> This whole document — the `project_codebases` table, the JSONB merge, the `npm install` filter, the
> replay flow and `RefreshSandbox` — is what turns that demo into a product that survives a second
> visit. Build it when a project needs to still exist tomorrow.
>
> **What the MVP does instead:** `SendMessage` creates a sandbox on the project's first Build message
> and stores `sandbox_id` / `preview_url` on the `projects` row (`project-model.md` §5.1
> `UpdateSandbox`). `resp.Files` and `resp.Commands` come back from the agent and are **discarded**.
> Reuse works within a session because the sandbox id is reused; it stops working when that sandbox
> expires, and the MVP has no recovery from that. Accepted — see `sandbox-service.md` §4, which
> raises the sandbox timeout so a demo outlives its own runtime.


## 1. The problem this solves

E2B sandboxes expire after five minutes with no keep-alive (`ARCHITECTURE.md` §7.6). Today the
sandbox **is** the codebase: the agent writes files into it, nothing else holds a copy, and when it
dies the app is gone. Reopening a project would spin up an empty sandbox.

The fix is to invert which side is authoritative:

> **Postgres is the truth. The sandbox is a disposable materialization of it.**

Persist every file the agent writes. When the user wants a preview, create a fresh sandbox from the
same template, replay the files into it, and hand back the URL. Expiry stops being a failure and
becomes a cache miss.

`project-model.md` §1.1 already assumes this — it is why `projects.sandbox_id` and
`projects.preview_url` are nullable and documented as *usually stale*. This document is the other
half.

**Recreation is user-driven, not automatic.** There is no lazy revival, no keep-alive, no
`expires_at`. A refresh button in the Workspace calls one RPC. That decision removes an entire class
of liveness bookkeeping, at the cost of the user occasionally waiting ~10–30 seconds. Accepted for
MVP.

## 2. What actually gets stored — much less than "the codebase"

This is the point most likely to be got wrong, so it is stated first.

**You are not storing the codebase. You are storing the diff from the template.**

The E2B image (`ai-service/sandbox-template/nextjs/`) already contains Node 21,
`create-next-app@15.3.3` and every shadcn component, pre-installed and moved to `/home/user/`. Every
sandbox starts from that image. Persisting it would mean persisting thousands of files that are
already baked in.

What varies per project is only what the agent wrote — and
`ai-service/services/agent_callback_service.py` captures exactly that set, observed rather than
narrated (`ARCHITECTURE.md` §7.5).

So the persisted artifact is a path → content map of roughly 5–40 files, tens of kilobytes. **That
size is what makes JSONB the right choice**; it would not be if this were a real filesystem.

Restore is therefore *template + replay*, not *filesystem restore*:

```
POST /sandbox/              → fresh sandbox from ats-nextjs-template  (all of shadcn, node_modules)
POST /sandbox/{id}/files    → write the stored map                     (the agent's work)
POST /sandbox/{id}/command  → npm install <collapsed deps>             (anything npm added)
```

## 3. The Go API already receives this data

Verified in `ai-service/api/routes/models/ai_models.py`:

```python
class AICodeAgentResponse(BaseModel):
    human_message: str
    summary: str                 # the model's prose — untrusted narration
    commands: List[str]          # observed
    files: Dict[str, str]        # observed: path -> FULL file content
```

`files` is a complete path→content map, not a list of paths. So the Build response carries
everything needed to persist, and no extra read-back call is required on the happy path.

Two properties of that map worth knowing before writing the merge:

- **It is cumulative within one run, not across runs.** `CodeAgentCallBack` is constructed fresh per
  request, so `updated_files` holds only that request's writes. Merging run-over-run is the Go
  side's job — §6.2.
- **Paths are whatever the agent typed.** The tool takes a free-form `path` and the system prompt
  says the project lives at `/home/user/`. Store paths **exactly as received** and replay them
  unchanged; do not normalise, prefix, or strip. Normalising risks writing to a different location
  than the agent did. The corresponding limit — an agent that is inconsistent about absolute vs.
  relative paths will produce two keys for one file — is recorded in §10.

## 4. Prerequisite: the callback's success test is not trustworthy

Everything below stores what `ai-service/services/agent_callback_service.py` reports. Pending tool
inputs are promoted to the result in exactly one place, behind two gates:

```python
tool_name = kwargs.get("name", "")                                   # gate 1 — verified sound
success = "failed to" not in output and "error" not in output        # gate 2 — unreliable
```

**Gate 1 is fine — verified against the pinned version, not assumed.** `langchain-core` 0.3.79
passes the tool name explicitly at `langchain_core/tools/base.py:897`
(`run_manager.on_tool_end(output, color=color, name=self.name, **kwargs)`), and `handle_event`
forwards it to the handler. `tool_name` is populated. Recorded here only so nobody re-litigates it.

The same read settles what `output` is: `_format_output` returns the raw `_run` string unless a
`tool_call_id` was supplied, and the classic `AgentExecutor` does not supply one. So `output` is a
plain `str` and gate 2 really is a substring test on prose — which is the problem.

**Gate 2 — substring matching on prose.** This is `ARCHITECTURE.md` §11 #6, and `npm install` is its
worst case. `SandboxCommandTool._run` splices the command's **full stdout *and* stderr** into the
string it returns, and npm writes deprecation notices to stderr routinely. Any package whose name or
output contains the substring `error` silently drops the command from `commands_executed`. Note the
asymmetry that makes it worse rather than better: `npm ERR!` is uppercase and does **not** match, so
genuine failures can be recorded as successes while harmless installs are discarded. A restore path
built on this is a coin flip.

*Fix:* `execute_terminal_command` should return an **exit code**, and the callback should branch on
that rather than on the rendered text. `SandboxService.execute_terminal_command` already receives a
`CommandResult` from the E2B SDK and throws away everything but `stdout`/`stderr`.

This is a correctness fix in the AI service, not a blocker on the schema or the Go code — the shape
of what gets stored does not change, only how often a real write is silently dropped. Land it
alongside this feature rather than ahead of it, and keep §14's capture check in the loop until it
does.

## 5. The table

### 5.1 `00006_add_project_codebases.sql`

Numbers are sequential across the whole directory and must not gap. `project-model.md` takes `00004`
and `messages-model.md` takes `00005` — both MVP — so **this takes `00006`**. Never edit an applied
migration; `make nuke` to reset.

```sql
-- +migrate Up
CREATE TABLE project_codebases (
    project_id UUID PRIMARY KEY,
    files JSONB NOT NULL DEFAULT '{}'::jsonb,
    commands JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS project_codebases CASCADE;
```

### 5.2 Why a separate table rather than columns on `projects`

One weak reason and one strong one.

*Weak:* `projects` is the hottest table — the sidebar lists it on every page load. A large TOASTed
`files` column there is one `SELECT *` away from loading every user's entire codebase to render a
list of titles. This repo always enumerates columns explicitly, so it is a footgun rather than a
live bug.

*Strong:* **applying a run becomes a single atomic statement.** Postgres merges JSONB objects
natively with `||`, which is exactly the semantics of "the agent wrote these files":

```sql
INSERT INTO project_codebases (project_id, files, commands)
VALUES ($1, $2::jsonb, $3::jsonb)
ON CONFLICT (project_id) DO UPDATE
SET files      = project_codebases.files || EXCLUDED.files,
    commands   = project_codebases.commands || EXCLUDED.commands,
    updated_at = NOW()
RETURNING files, commands;
```

No read-modify-write, no transaction, no partial state, and no row to pre-create when a project is
made. New paths are added, existing paths are overwritten by path, commands are appended in order.

`project_id` **is** the primary key, which enforces one-codebase-per-project structurally rather
than by convention and removes a surrogate id nothing would use.

### 5.3 Why not per-file rows, and why no versioning

Per-file rows (`PRIMARY KEY (project_id, path)`) only win if something needs to query or update an
individual file. Nothing does — every consumer reads the whole map: replay, a future file tree, a
future ZIP export.

Versioning is **explicitly out of scope**. `ARCHITECTURE.md` §10 lists version history as ⛔ and it
stays there. If it is wanted later it is additive: a `project_codebase_snapshots` table, with this
row remaining "current". Nothing here needs to change to allow it.

## 6. Domain

### 6.1 `domain/project_codebase.go` — new

```go
type ProjectCodebase struct {
    projectID uuid.UUID
    files     map[string]string
    commands  []string
    updatedAt time.Time
}
```

- `RestoreProjectCodebase(projectID uuid.UUID, files map[string]string, commands []string, updatedAt time.Time) *ProjectCodebase`
- `EmptyProjectCodebase(projectID uuid.UUID) *ProjectCodebase` — non-nil maps, used when no row
  exists yet. §7.2.
- Accessors `ProjectID`, `UpdatedAt`, `FileCount`, `IsEmpty`.

**`Files()` and `Commands()` return copies** (`maps.Clone` / `slices.Clone`), not the internal
values. This is the one place the `domain/user.go` pattern needs adapting: returning a scalar from
an accessor is safe, but returning a map or slice hands the caller a mutable reference straight into
the entity, defeating the "expose nothing, use methods" rule stated at the top of `user.go`. The
maps are tens of kilobytes and copied at most once per refresh, so the cost is irrelevant.

There is no `NewProjectCodebase` validating constructor. A codebase is never authored by a user —
it only ever arrives from an agent run or from the database — so there is no input to validate at
construction. Validation of an incoming run happens in the service, where the size caps live (§7.3).

## 7. Persistence

New file `postgres/project_codebase_repository.go`, following `user_repository.go`: package-private
row struct with `db` tags, `ToDomain()`, a constructor taking `*sqlx.DB`, hand-written SQL, errors
wrapped as `*domain.Error`.

```go
type ProjectCodebase struct {
    ProjectID uuid.UUID `db:"project_id"`
    Files     []byte    `db:"files"`      // raw JSONB
    Commands  []byte    `db:"commands"`   // raw JSONB
    UpdatedAt time.Time `db:"updated_at"`
}
```

`JSONB` scans into `[]byte`; `ToDomain()` unmarshals into `map[string]string` and `[]string`. Do not
reach for `sqlx`'s struct-scanning of JSON or a `pq.StringArray` — `commands` is a JSON array, not a
Postgres array, and the two are not interchangeable.

### 7.1 Methods

| Method | Signature | Notes |
| --- | --- | --- |
| `Get` | `(ctx, projectID uuid.UUID) (*domain.ProjectCodebase, error)` | `sql.ErrNoRows` → **empty codebase, not `NotFound`** — §7.2 |
| `ApplyRun` | `(ctx, projectID uuid.UUID, files map[string]string, commands []string) (*domain.ProjectCodebase, error)` | the upsert-merge in §5.2; caller lives in `messages-model.md` |

Two methods. Everything else the feature does happens in the service.

### 7.2 A project with no builds has an empty codebase, not a missing one

`Get` returns `EmptyProjectCodebase(projectID)` when no row exists. This is deliberate and it is the
opposite of the `NotFound` convention used for projects.

A row is only created on the first successful Build run. A project created from Home and used only
in Chat mode legitimately has no row, and that is a normal state — the Workspace should show "no
sandbox running", not an error. Returning `NotFound` would force every caller to distinguish
"never built" from "broken", and both would render identically anyway.

`Get` is called on a project the service has *already* ownership-checked, so a missing row cannot
mean "someone else's project".

### 7.3 Guards that belong in the service, not the schema

Postgres will accept a 200 MB JSONB document without complaint. A runaway agent writing a huge file
would then be replayed into every future sandbox. Before calling `ApplyRun`, the service enforces:

| Guard | Suggested limit | On breach |
| --- | --- | --- |
| Single file size | 1 MB | drop that file, log the path, keep the rest |
| Total codebase size | 10 MB | reject the run's file writes, log, keep commands |
| Path length | 512 chars | drop that file |

Log every drop. Silently discarding an agent's work is worse than the size problem it prevents, and
the log is the only place it will ever be visible.

## 8. Commands: filter on the way in

`commands` exists for exactly one reason — `npm install` breaks restore.

The system prompt forbids editing `package.json` directly and requires `npm install`. But npm
mutates `package.json` **inside the sandbox**; the agent never calls the write tool on it, so the
callback never captures it, so it is never persisted. Replaying files alone gives you code importing
a package that is not installed.

**Filter to `npm install` at write time, not at replay time.** The service inspects each command
from the run and stores only those matching `npm install` / `npm i` (trimmed, lowercased prefix
match). Everything else — `ls`, `cat`, `mkdir`, `rm` — is discarded before it reaches the database.

Two reasons this belongs on the write side:

1. `execute_sandbox_command` runs arbitrary shell. A stored log that must be sanitised before every
   replay is a log you can never fully trust; one that only ever contained `npm install` lines is
   safe by construction. Replaying a captured `rm -rf` into a fresh sandbox is not a hypothetical
   you want guarded by a filter someone might later forget to apply.
2. It keeps the column small and readable.

**Collapse on replay.** Twenty historical installs become one
`npm install pkg-a pkg-b pkg-c` — parse the package names out of the stored lines, dedupe preserving
order, and issue a single command. Twenty sequential `npm install` round trips would dominate the
refresh time on their own.

The alternative considered and rejected for MVP: read `package.json` back out of the sandbox after
each run (`GET /sandbox/{id}/file`) and store it as a normal file. More robust — it captures
whatever npm actually resolved rather than what was asked for — but it costs a round trip per run
and a dependency on another route. Worth doing as hardening once Build works; note it, do not build
it now.

## 9. Service and the refresh flow

New file `application/services/codebase_service.go`.

```go
type CodebaseRepository interface {
    Get(ctx context.Context, projectID uuid.UUID) (*domain.ProjectCodebase, error)
    ApplyRun(ctx context.Context, projectID uuid.UUID, files map[string]string, commands []string) (*domain.ProjectCodebase, error)
}

// The sandbox half of the AI-service client. Named for the use case, not the transport.
type SandboxRunner interface {
    CreateSandbox(ctx context.Context) (sandboxID string, url string, err error)
    WriteFiles(ctx context.Context, sandboxID string, files map[string]string) error
    RunCommand(ctx context.Context, sandboxID string, command string) error
}
```

`ProjectRepository` is **not** redeclared — it already exists in `project_service.go` in the same
`services` package (`ARCHITECTURE.md` §5.2 puts ports beside their consumer, and here the consumer
is the package). `CodebaseService` composes `*ProjectService` for the ownership-checked `Get`,
mirroring how `AccountService` composes oauth + auth + user rather than reaching for repositories
directly.

### 9.1 `RefreshSandbox(ctx, userID, projectID)`

```
1. project := projectService.Get(ctx, userID, projectID)   → ownership check, NotFound if not theirs
2. codebase := codebaseRepo.Get(ctx, projectID)            → empty is fine, not an error
3. sandboxID, url := sandbox.CreateSandbox(ctx)            → fresh sandbox from the template
4. if !codebase.IsEmpty():
       sandbox.WriteFiles(ctx, sandboxID, codebase.Files())
       if deps := collapseInstalls(codebase.Commands()); deps != "":
           sandbox.RunCommand(ctx, sandboxID, deps)
5. projectRepo.UpdateSandbox(ctx, projectID, sandboxID, url)
6. return url
```

Notes that are easy to get wrong:

- **Step 3 happens even for an empty codebase.** A project with no builds still gets a working
  template preview. That is the correct behaviour — the user sees a running Next.js app, not an
  error.
- **`RefreshSandbox` takes no provider and calls no LLM.** Writing files and running npm is
  mechanical. `projects.provider` (`project-model.md` §3) is restored for the *composer's* selector;
  it plays no part here.
- **Step 5 is last.** If any earlier step fails, the old (dead) pointer stays in the row rather than
  being replaced by a half-built sandbox the user would be shown as working.
- **`updated_at` on the project is not bumped.** Refreshing a preview is not authorship — it must
  not reorder the sidebar.
- **The dev server is already running.** `compile_page.sh` launches `next dev --turbopack` at
  sandbox start and polls until it compiles, so the URL is warm. Writing files afterwards is picked
  up by hot reload. Never issue `npm run dev`; the system prompt forbids it for the same reason.

### 9.2 Failure modes worth handling explicitly

| Failure | Response |
| --- | --- |
| `CreateSandbox` fails (E2B down, quota) | `ErrorTypeUnavailable`, project row untouched |
| `WriteFiles` fails midway | `ErrorTypeInternal`. The sandbox is abandoned and expires on its own in 5 minutes; do not attempt cleanup |
| `npm install` fails | **Log and continue.** Return the URL. A preview with a missing dependency is more useful than no preview, and the user can see the broken import |
| Codebase empty | Not a failure. §9.1 step 3 |

### 9.3 This is the slowest call in the app

Sandbox creation plus an `npm install` is comfortably 10–30 seconds. Two consequences:

1. **`AIService`'s `http.Client` is constructed with `Timeout: 30 * time.Second`**
   (`outbound/ai_service/ai_service.go`). That is a single client shared by every call, and it will
   cut off a slow refresh. `ai-service-client.md` must either raise it or use per-call contexts.
   This is a concrete, already-existing blocker, not a theoretical one.
2. It is synchronous for MVP, and the frontend must show a pending state on the refresh button. It
   is the second-best candidate for the Temporal treatment after Build itself — but it stays
   synchronous here on purpose.

## 10. Known limits — recorded, not hidden

1. **There is no delete tool.** The agent has list / read / write / execute-command only, so a file
   it removes with `rm` is still in the stored map and gets resurrected on the next refresh. The map
   can only grow.
2. **Text only.** `files` is `Dict[str, str]`; an image or any binary asset cannot round-trip. The
   agent has no way to produce one today, so this is a limit rather than a bug.
3. **Inconsistent paths make duplicate keys.** If the agent writes `app/page.tsx` in one run and
   `/home/user/app/page.tsx` in the next, both live in the map and both get replayed. Normalising
   would be worse (§3); the real fix is prompt-side.
4. **`npm install` captures the request, not the resolution.** A floating version range can resolve
   differently on replay than it did originally. The `package.json` read-back in §8 is the fix when
   it matters.
5. **No versioning.** Replay always produces the latest state; there is no way back to an earlier
   one. §5.3.
6. **Refresh is not idempotent in cost.** Every press creates a new sandbox. Nothing reaps the
   previous one — it expires on its own.

## 11. Proto

> **`.planning/Proto/project-codebase-proto.md` is authoritative for this section**, and
> `project-proto.md` §2 carries the shared conventions. There is no `body : "*"` on this RPC: every
> request field is bound to the path, so there is nothing left to put in a body — which means the
> `curl` in §14 needs no `-d`.

`RefreshSandbox` is added to the **existing** `ProjectService` rather than getting a service of its
own. It operates on a project, it is one RPC, and a second service would mean a second
`vanguard.NewService` entry and a second handler for no benefit.

### 11.1 `project_service.proto` — modified

```proto
  rpc RefreshSandbox(RefreshSandboxRequest) returns (RefreshSandboxResponse) {
    option (google.api.http) = {
      post : "/api/v1/projects/{id}/sandbox/refresh"
    };
  }
```

```proto
message RefreshSandboxRequest { string id = 1; }

message RefreshSandboxResponse {
  string sandbox_id = 1;
  string preview_url = 2;
}
```

`POST`, not `GET` — it creates a resource and costs money. The response deliberately does **not**
return the whole `Project`: the only fields that changed are these two, and returning the project
would invite the frontend to overwrite a title the user may have just renamed.

There is no `GetProjectFiles` RPC. A file tree is deferred (`logged_in_design.md` §7.4) and nothing
in the MVP screen reads the map — only the server replays it.

## 12. Handler and wiring

### 12.1 `inbound/handlers/project.go` — modified

`RefreshSandbox` joins the four RPCs from `project-model.md` on the existing
`ProjectServiceHandler`, which grows a second dependency:

```go
type ProjectServiceHandler struct {
    apiv1connect.UnimplementedProjectServiceHandler

    projectService  *services.ProjectService
    codebaseService *services.CodebaseService
    authAdapter     *authAdapter.HTTPAuthAdapter
}
```

Same four steps as every other handler: `authAdapter.User(ctx)` (omit it and the RPC is public),
`uuid.Parse` the id, call the service with `user.ID()`, map errors with
`errorAdapter.ToConnectError`.

### 12.2 `application/app.go` — modified

```go
// persistance
codebaseRepo := postgres.NewProjectCodebaseRepository(db)

// business logic
codebaseService := services.NewCodebaseService(codebaseRepo, projectRepo, projectService, aiservice)
```

and `codebaseService` is passed into `handlers.NewProjectServiceHandler(...)`.

`aiservice` already exists at the top of `New` and satisfies `SandboxRunner` once
`ai-service-client.md` lands. Both constructions must sit **above** the
`services := []*vanguard.Service{...}` line, which shadows the imported `services` package.

## 13. Dependencies and ordering

| Depends on | For | Blocking? |
| --- | --- | --- |
| AI-service callback gate-2 fix (§4) | not silently dropping real writes | No — but this feature is unreliable until it lands |
| `project-model.md` | `projects` FK, `ProjectService.Get`, `UpdateSandbox` | Yes |
| `ai-service-client.md` | `CreateSandbox` (fix `GET /sandbox/create` → `POST /ai-service/v1/sandbox/`, and the ignored error at `:18` that panics), `WriteFiles`, `RunCommand`, and the 30s client timeout (§9.3) | Yes |
| `messages-model.md` | the only caller of `ApplyRun` | No — refresh works without it |

Note the scope claim on the AI service: `ARCHITECTURE.md` §7.1 describes
`POST /sandbox/{id}/files` and `POST /sandbox/{id}/command` as *development affordances, not
intended for the Go API*. **This feature promotes both to first-class API surface.** Update that
sentence in `ARCHITECTURE.md` in the same change.

Until `messages-model.md` lands, `ApplyRun` has no caller and `files` stays empty — refresh then
produces a bare template preview. That is a coherent, demonstrable state, and it is the right place
to verify §14 steps 1–3 before Build exists.

## 14. Verifying

No test harness exists in this repo — manual only. Do not claim tests pass.

```bash
make gen
cd api && go build ./... && go vet ./...
make nuke && make
docker compose exec sql psql -U postgres -d local -c '\d project_codebases'
```

**Check capture first, directly against the AI service.** Everything downstream stores what comes
back here, so confirm it is non-empty before debugging anything on the Go side. Include a package
whose install output is noisy, since that is what gate 2 (§4) drops.

```bash
curl -X POST http://localhost:9999/ai-service/v1/anthropic/<sandbox_id>/code \
  -H 'content-type: application/json' \
  -d '{"message":"create app/page.tsx with a red heading, then npm install clsx"}'
# expect: files -> {"...page.tsx": "..."} AND commands -> ["npm install clsx"]
```

Then the Go path:

```bash
# empty codebase: must still return a working template preview
curl -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects/<id>/sandbox/refresh

# unauthenticated must be rejected
curl -i -X POST https://local.api.vulx.ai/api/v1/projects/<id>/sandbox/refresh   # 401

# another user's project
curl -b "jwt=$OTHER_JWT" -X POST https://local.api.vulx.ai/api/v1/projects/<id>/sandbox/refresh   # 404
```

Four checks a happy-path `curl` will not surface:

1. **The merge accumulates.** Call `ApplyRun` twice with overlapping paths (via psql or two Build
   runs) and confirm new paths are added, repeated paths overwritten, and commands appended in
   order.
2. **Replay reaches the sandbox.** After a refresh with a non-empty codebase, read a known file back
   with `GET /ai-service/v1/sandbox/{id}/file?path=...` and diff it against the stored JSONB.
3. **The command filter holds.** Confirm a run containing `ls -la` stores nothing, and one
   containing `npm install clsx` stores exactly that.
4. **`projects.sandbox_id` and `preview_url` actually change**, and `projects.updated_at` does
   **not** (§9.1).

## 15. Out of scope

Versioning and snapshots (§5.3) · code export / ZIP · a file tree RPC (§11.1) · automatic or lazy
sandbox recreation, keep-alive, expiry tracking (§1) · reaping abandoned sandboxes (§10.6) ·
`package.json` read-back hardening (§8) · binary assets (§10.2) · making refresh asynchronous
(§9.3) · credits (`credits.md`) · tests.
