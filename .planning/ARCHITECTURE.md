# Architecture

> **How to use this document.** It describes the system as it exists in the code today, not as it
> is planned. Every section is marked with a status so an agent can tell scaffolding from working
> code before relying on it:
>
> - ✅ **Built** — implemented and wired into the running app.
> - 🟡 **Partial** — exists and runs, but incomplete or not connected end to end.
> - ⛔ **Not built** — named in the product vision, no code yet.
>
> `ARCHITECTURE.md` explains *why the system is shaped this way*. `CLAUDE.md` is the short
> operational brief (commands, gotchas). `README.md` is the setup guide. When code and this
> document disagree, the code wins — fix the document in the same change.

## 1. Product

A SaaS platform that turns natural-language prompts into frontend code, in the vein of Bolt and
Lovable. A user describes what they want, an LLM agent writes the code into a live E2B sandbox
running Next.js, and the user sees both the generated files and a working preview URL. Three model
providers are selectable (OpenAI, Google Gemini, Anthropic Claude). Usage is metered in credits.

**Where the product actually is:** the demo path is done and verified end to end, frontend included —
authenticate → create a project → send a Build message → a sandbox is created (or reused) → the
agent writes code into it → the preview URL renders the result. See
[§10 Implementation status](#10-implementation-status) for what's left, which is Polish-tier work
(async Build, codebase persistence, Chat mode, payments, tests), not base wiring.

## 2. Repository map

| Path | Role | Language |
| --- | --- | --- |
| `api/` | Main backend. Auth, users, orchestration. Hexagonal/DDD. | Go 1.24 |
| `ai-service/` | LLM calls, LangChain agents, E2B sandbox control. | Python 3.11 |
| `app/` | Web client. | Next.js 15 / React 19 |
| `proto/` | Contract between `app` and `api`. Source of truth for both. | Protobuf |
| `docker-compose.yaml` | The only supported way to run the stack. | — |
| `Caddyfile` | TLS reverse proxy for the API. | — |
| `makefile` | Entry point for every routine task. | — |

`Taskfile.yml` is unrelated to the application — it manages the author's local dev sandbox VMs.

## 3. Runtime topology ✅

Everything runs under Docker Compose on a single bridge network named `vulx`. Containers address
each other by service name (`sql`, `redis`, `temporal`, `ai-service`, `api`).

```
browser
   │
   ├──► caddy :443 (local.app.vulx.ai, tls internal) ──► app :3000   Next.js dev server
   │                                                        │
   │                                                        │ Connect RPC (JSON over HTTP,
   │                                                        │ credentials: include)
   │                                                        ▼
   └──► caddy :443 (local.api.vulx.ai, tls internal) ──► api :8080   Go — Connect + Vanguard
                                                            │
                                                            ├──► sql       :5432   Postgres 16 (app data + Temporal)
                                                            ├──► redis     :6379   JWT blacklist, OAuth state
                                                            ├──► temporal  :7233   workflow engine
                                                            └──► ai-service:9999   FastAPI
                                                                         │
                                                                         └──► E2B cloud ──► Next.js sandbox :3000
```

| Service | Port | Purpose |
| --- | --- | --- |
| `app` | 3000 | Next.js dev server (`Dockerfile.dev`, source bind-mounted for hot reload) |
| `api` | 8080 | Go API — Connect RPC + transcoded REST |
| `ai-service` | 9999 | FastAPI, uvicorn, `--no-access-log` (it logs via middleware instead) |
| `sql` | 5432 | Postgres 16 — application schema *and* Temporal's own storage |
| `redis` | 6379 | Not published to the host; reachable only inside the network |
| `temporal` | 7233 | `auto-setup` image; health-checked so the API cannot race namespace creation |
| `temporal-ui` | 8081 | Mapped from the container's 8080 (the README's "8080" is wrong) |
| `caddy` | 80/443 | `local.api.vulx.ai` → `api:8080` and `local.app.vulx.ai` → `app:3000`, both `tls internal` |

**Caddy's job** is to terminate HTTPS for both hosts, so the browser will accept `Secure` JWT
cookies during local development and the app/api hop stays same-*site* (`local.app.vulx.ai` →
`local.api.vulx.ai`) even though it's cross-*origin* — required for `credentials: "include"` to
carry the cookie. Caddy also transparently proxies the frontend's HMR WebSocket upgrade. Reaching
either host requires `127.0.0.1 local.api.vulx.ai` and `127.0.0.1 local.app.vulx.ai` in
`/etc/hosts`, and trusting Caddy's local CA (`make trust` / `make trust-rm`).

`docker compose run` (used by the `app` and `api` make targets) does not attach a container's
network alias by default — only `--use-aliases` does — so both targets pass it explicitly. Without
it, `reverse_proxy app:3000` / `api:8080` cannot resolve and Caddy 502s.

**Startup ordering** is enforced through `depends_on` with health conditions: Postgres and Redis
must report healthy, and Temporal must answer `tctl namespace describe`, before `api` starts. The
API panics rather than degrades if any connection fails (`postgres.NewDb`, `cache.NewRedisClient`,
`temporal.New`), so a failed dependency surfaces immediately as a crash loop.

**Secrets live outside the repository.** Compose reads `env_file:
../AI-Website-Builder-Secrets/.api-env` and `.ai-service-env` — a *sibling directory of the repo
root*. Nothing starts without it. Required keys are enumerated in `api/.env.example` and
`ai-service/.env.example`. The Go side falls back to sane in-network defaults for everything except
`JWT_SEED` and the Google OAuth credentials (`api/internal/config/config.go`).

## 4. Communication contracts

| Hop | Protocol | Contract source |
| --- | --- | --- |
| browser → api | Connect RPC, JSON over HTTP/1.1 | `proto/api/v1/*.proto` |
| REST clients → api | Plain REST/JSON, transcoded by Vanguard | `google.api.http` options in the same protos |
| api → ai-service | Hand-rolled REST/JSON over `net/http` | none — untyped, hand-maintained 🟡 |
| ai-service → E2B | E2B Python SDK | — |

The Go↔Python hop is the one seam with **no shared schema**. Request and response shapes are
duplicated by hand in `api/internal/infrastructure/outbound/ai_service/` (Go structs) and
`ai-service/api/routes/models/` (Pydantic models). They have already drifted; see §11.

## 5. Go API ✅

Hexagonal architecture (ports and adapters) with DDD. Business rules sit in the centre and know
nothing about HTTP, SQL, Redis, or Temporal; everything external is reached through an interface.

### 5.1 Layout

```
api/
├── cmd/api/main.go                     load config → application.New() → Start(ctx)
└── internal/
    ├── config/                         env → typed Config; defaults for in-network hosts
    ├── domain/                         ENTITIES + ERROR TAXONOMY. Zero external imports.
    │   ├── user.go                     User, Profile — private fields, accessor methods only
    │   ├── login.go                    LoginProvider enum, UserFromProvider
    │   ├── project.go                  Project entity — sandbox_id/preview_url as "" until set ✅
    │   ├── ai_provider.go              AIProvider enum — String() doubles as the AI-service URL segment
    │   ├── message.go                  Message entity + MessageRole/ChatMode enums ✅
    │   ├── sandbox.go                  Sandbox — value-object DTO for the AI-service's create response
    │   │                                 (NOT an entity; renamed from SandboxInfo after the original
    │   │                                 uuid-keyed Sandbox entity was deleted for having no callers) ✅
    │   ├── code_agent_result.go        CodeAgentResult — value-object DTO for the agent's response ✅
    │   ├── page.go                     Page[T] — generic cursor-paginated result
    │   └── error.go                    Error{type, err}, WrapError, ErrorType constants
    ├── application/
    │   ├── app.go                      ⭐ THE WIRING FILE — every dependency is built here
    │   └── services/                   USE CASES + the ports they depend on
    │       ├── user_service.go         + UserRepository port
    │       ├── project_service.go      + ProjectRepository, ProjectTitler ports ✅
    │       ├── message_service.go      + MessageRepository, AIService ports — the Build flow ✅
    │       ├── auth_service.go         + AuthAdapter port
    │       ├── oauth_service.go        + OauthProvider, OauthProviderRegistry ports
    │       ├── account_service.go      composes oauth + auth + user (no new ports)
    │       └── cahce.go                Cache port [sic — filename is misspelled in the repo]
    ├── infrastructure/
    │   ├── inbound/
    │   │   ├── handlers/               Connect handlers: account.go, user.go, project.go, message.go
    │   │   ├── grpc/gen/               ⚠️ GENERATED by buf — never hand-edit
    │   │   ├── grpc/adapters/auth/     JWT cookie read/write + the auth interceptor
    │   │   ├── grpc/adapters/error/    domain.Error → connect.Code
    │   │   ├── grpc/adapters/logger/   structured per-RPC request logging
    │   │   ├── grpc/adapters/security/ CORS allowlist
    │   │   ├── auth_token/auth.go      Ed25519 JWT mint / validate / parse
    │   │   └── http/handlers/          /healthz and the Swagger UI at /docs/
    │   ├── outbound/
    │   │   ├── oauth/                  Google provider + a name→provider registry
    │   │   ├── temporal/               client, worker registration — currently a no-op; see below
    │   │   └── ai_service/             REST client for the Python service — CreateSandbox,
    │   │                                 RunCodeAgent, GenerateTitle, shared `do` helper,
    │   │                                 15-minute client backstop ✅
    │   ├── persistence/postgres/       repositories (user, project, message), embedded migrations,
    │   │                                 cursor codec
    │   └── cache/redis.go              Cache port implementation
    └── utils/                          slog logger, clamp helper
```

### 5.2 Where ports live — the one convention that surprises people

Ports are **not** in a `ports/` package. Each interface is declared in the file of the service that
consumes it, and implemented far away in `infrastructure/`. `UserRepository` is declared at the top
of `application/services/user_service.go` and implemented by
`infrastructure/persistence/postgres.UserRepository`. This is deliberate — the consumer owns the
contract, so the dependency arrow points inward, toward the domain.

To find an implementation, search for the method set, not the interface name.

### 5.3 Dependency injection

`application/app.go` is the single composition root. It constructs, in order: the AI-service
client, the database and repositories (user, project, message), Redis, Temporal (registering a
worker with nothing on it — see §13), the token service, the auth/oauth/user/account/project/message
services, the Connect handlers, and finally the interceptor chain and mux. Nothing anywhere else
calls a constructor for a long-lived dependency.

Read this file first when tracing any request — it is the map of the whole backend.

`App.Start` serves on `:8080` from a goroutine and blocks on a `select` over the error channel and
`ctx.Done()`, giving a 10-second graceful shutdown that closes workers, Temporal, the DB and Redis.

### 5.4 Error taxonomy — the backbone convention

Every fallible function returns `*domain.Error`, which pairs an `ErrorType` with a wrapped cause.
Layers add context as the error travels up:

```go
return nil, domain.WrapError("user service get by email", err)
```

`WrapError` **preserves the innermost type** through the chain, so a `ErrorTypeNotFound` raised by
the repository is still `NotFound` when the handler sees it. Exactly one place converts the type to
a transport code — `grpc/adapters/error.ToConnectError`. Handlers call it and nothing else:

| Domain type | Connect code |
| --- | --- |
| `ErrorTypeUnauthenticated` | `CodeUnauthenticated` |
| `ErrorTypePermissionDenied` | `CodePermissionDenied` |
| `ErrorTypeInvalid` | `CodeInvalidArgument` |
| `ErrorTypeNotFound` | `CodeNotFound` |
| `ErrorTypeAlreadyExists` | `CodeAlreadyExists` |
| `ErrorTypeInternal` / `Unimplemented` / `Unavailable` / `Timeout` | matching codes |
| anything else | `CodeUnknown` |

Adding a category means editing `domain/error.go` and the switch in the error adapter — never
constructing a `connect.Error` inside a service.

### 5.5 Persistence ✅

`sqlx` over `lib/pq`, hand-written SQL, no ORM. Rows map to package-private structs with `db` tags
and a `ToDomain()` method; the domain type is reconstructed through `domain.RestoreX(...)`
constructors that skip validation (data already in the database is trusted).

**Migrations run automatically at boot.** `postgres.NewDb` embeds `migrations/*.sql` with
`go:embed` and executes `sql-migrate` up before returning; a failure panics the API. Add a new
numbered file — never edit one that has been applied. `make nuke` drops the volumes for a clean
slate.

Current schema: `users` (uuid pk, unique email, `credits INTEGER DEFAULT 10`, `is_active`),
`user_auth_providers` (one row per user, unique on `(provider, provider_user_id)`, cascading
delete), `projects` (owned by `user_id`, cascading delete; `sandbox_id`/`preview_url` are
`NOT NULL DEFAULT ''` rather than nullable — the domain layer already treats `""` as "no sandbox
yet", so `sql.NullString` would be unused complexity), and `messages` (owned by `project_id`,
cascading delete; `role`/`mode`/`provider` are `VARCHAR`, not Postgres enums, so adding a value is a
Go/proto change, not a migration). The `citext` extension is enabled but not yet used by any column.

**Pagination** is keyset-based, not offset, for `users` and `projects`: `cursor.go`
base64-encodes a timestamp (`created_at` for users, `updated_at` for projects — the codec is
column-agnostic) as an opaque token, ordered `DESC`, and results come back as
`domain.Page[T]{Items, Token, HasMore}`. **`messages` is unpaginated on purpose** — a thread reads
oldest-first (`ORDER BY created_at ASC, id ASC`), which doesn't fit the `DESC` keyset codec, and a
thread is small enough that pagination is machinery with no caller yet.

**`message_repository.go`'s `Create` is the one exception to "table X's SQL lives in
X_repository.go".** Inserting a message also has to bump `projects.updated_at` (so the sidebar
reorders by recent activity) — done as a single data-modifying CTE so both writes are atomic with
no transaction helper in this codebase. The `UPDATE projects` living inside
`message_repository.go` is intentional, not a misplaced query.

### 5.6 Project titles: generated asynchronously, always via OpenAI

`ProjectService.Create` returns immediately with `provisionalTitle(firstPrompt)` — a fast, local
whitespace-collapse-and-60-char-clamp — then spawns a detached goroutine (`generateTitleAsync`)
that calls the new `ProjectTitler` port (`AIService.GenerateTitle`, `outbound/ai_service/llm.go`)
and, on success, overwrites the title via `ProjectRepository.UpdateTitle`. This is a deliberate
exception to the provider-selection pattern used everywhere else (`Project.provider` picks which
`{provider}` path segment a Build hits): title generation always calls `POST /openai/query`,
regardless of what provider the project itself uses — a cheap, one-off cosmetic call unrelated to
the code-gen path. One real consequence of that: if `OPENAI_API_KEY` isn't configured, every
project's title silently and permanently falls back to the provisional one, even for an
Anthropic- or Google-only setup. `GenerateTitle`'s own 10s timeout (`generateTitleTimeout`,
`ai_service.go`) sits inside a 15s outer bound (`asyncTitleTimeout`) covering the LLM call plus the
follow-up DB write; any failure — timeout, AI-service outage, empty response — just leaves the
provisional title in place, logged as a warning. `UpdateTitle` (mirrors `UpdateSandbox`'s shape)
does **not** bump `updated_at`, consistent with that existing convention. Because this goroutine
runs outside the request lifecycle, it recovers its own panics — an unrecovered one there would
crash the whole process, not just fail one title.

## 6. Authentication ✅

### 6.1 Token design

Stateless Ed25519 JWTs in an httpOnly cookie named `jwt`. Long-lived and self-renewing, with no
refresh token — a deliberate trade of maximum security for simplicity at this scale.

| Property | Value | Where |
| --- | --- | --- |
| Algorithm | EdDSA (Ed25519) | `inbound/auth_token/auth.go` |
| Key | derived deterministically from `JWT_SEED` (base64, 32 bytes — `openssl rand -base64 32`) | `config.Crypto` |
| Lifetime | 7 days | `CreateJWT` |
| Issuer / audience | `api.vulx.ai` (both, and both verified) | `ValidateJWT` |
| Clock leeway | 5 minutes | `ValidateJWT` |
| Claims | registered only: `iss`, `sub` (user UUID), `aud`, `exp`, `iat`, `jti` | — |
| Renewal threshold | under 42 hours remaining | `ValidateJWT` |

The seed is a *seed*, not the private key: the same `JWT_SEED` always regenerates the same keypair,
so restarting the API does not invalidate live sessions — but changing it invalidates all of them.

### 6.2 Renewal, expressed as a sentinel error

Renewal is not a separate endpoint or a frontend concern. It rides along on ordinary requests via a
sentinel error passed up the stack:

1. `ValidateJWT` finds under 42h left and returns `(userID, services.ErrAuthTokenExpiresSoon)` —
   a valid user *and* a non-nil error.
2. `AuthService.ValidateSession` recognises the sentinel, loads the user, and re-returns the
   sentinel alongside it.
3. `HTTPAuthAdapter.AuthenticateWithJWT` translates it into `(user, refresh=true)`.
4. The interceptor runs the handler, then calls `RefreshJWTCookie` to attach a brand-new 7-day
   cookie to the response.

Callers that treat any non-nil error as failure will break this path. Check for the sentinel with
`errors.Is`.

### 6.3 Logout and revocation

Because the JWT is stateless, logout needs a denylist. `AuthService.Logout` parses the token for
its expiry, SHA-256-hashes the token string, and writes the hash to **Redis** with a TTL equal to
the token's remaining life — so the entry evicts itself exactly when the token would have expired
anyway. `ValidateSession` checks this key on every request before it verifies the signature.

The denylist is in Redis, not Postgres. A Redis flush silently un-revokes every logged-out token.

### 6.4 Enforcement model — read this before adding an RPC

There is **no public-route allowlist and no per-route configuration**. The interceptor
(`grpc/adapters/auth/interceptor.go`) runs on every RPC and is *advisory*: if a valid cookie is
present it injects the user into the context; otherwise it simply calls the handler with an
unmodified context.

A handler makes itself protected by asking for the user and returning the error:

```go
user, err := authAdapter.User(ctx)
if err != nil {
    return nil, err          // already a connect.CodeUnauthenticated error
}
```

Omit those three lines and the RPC is public. `BeginAccountAuth` and `FinishAccountAuth` rely on
this to stay reachable while logged out.

`UserService`'s three RPCs additionally hard-code a superuser check against the literal email
`alitsayyed@gmail.com` (`inbound/handlers/user.go:42,67,94`). They are admin endpoints, not part of
the product surface.

### 6.5 Google OAuth ✅

Passwordless login, one provider today. The provider abstraction (`services.OauthProvider` plus a
name→provider `OauthProviderRegistry`) exists so a second one is a new file rather than a new
branch in the service.

**Begin** — `BeginAccountAuth`, REST `GET /api/v1/account/auth/begin`:

1. Handler maps the proto `LoginProvider` enum to `domain.LoginProvider`; unspecified is rejected.
2. `OauthService.BeginLoginFlow` generates 32 random bytes as CSRF state.
3. State is stored in Redis as `provider:<state>` → provider name, **10-minute TTL**. This doubles
   as the state allowlist — an unknown state simply misses the cache.
4. Google's consent URL is returned to the browser as `login_url`.

**Finish** — `FinishAccountAuth`, REST `POST /api/v1/account/auth/finish`, body `{code, state}`:

1. `provider:<state>` is looked up in Redis; a miss is `ErrorTypeInvalid` (this *is* the CSRF check).
2. Code is exchanged for a Google access token; the profile is fetched from
   `googleapis.com/oauth2/v2/userinfo`. An unverified Google email is rejected.
3. `AccountService.FinishAuth` reconciles identity:
   - email unknown → create the user, then create their `user_auth_providers` row;
   - email known → load the stored provider and reject with `ErrorTypeAlreadyExists` if it differs,
     naming the provider that owns the email. This is what stops account takeover by a second IdP.
4. A 7-day JWT is minted and set as a cookie by the handler via `SetJWTCookie`; the response body
   carries only the `Profile`.

Google's access token is never stored — it is used once to read the profile and discarded.

## 7. AI service ✅ (standalone) / 🟡 (integration)

FastAPI + LangChain + the E2B SDK. It exists as a separate service for one hard reason: **E2B ships
a Python SDK only**, and LangChain's Python implementation is the mature one. Keeping it out of the
Go binary also isolates heavy, fast-moving AI dependencies.

It is stateless and has no database. Everything it needs arrives in the request.

### 7.1 Routes

All routers mount under `/ai-service/v1` (`api/main.py`). Interactive docs at `/docs` are disabled
when `ENVIRONMENT=production`.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/ai-service/v1/healthz` | liveness |
| `POST` | `/ai-service/v1/sandbox/` | create a sandbox → `{id, url}` |
| `GET` | `/ai-service/v1/sandbox/{id}/files?path=` | list a directory |
| `GET` | `/ai-service/v1/sandbox/{id}/file?path=` | read one file |
| `POST` | `/ai-service/v1/sandbox/{id}/files` | write files |
| `POST` | `/ai-service/v1/sandbox/{id}/command` | run a shell command |
| `POST` | `/ai-service/v1/{provider}/query` | plain LLM question → `{content}` |
| `POST` | `/ai-service/v1/{provider}/{sandbox_id}/code` | run the coding agent |

`{provider}` is `openai`, `google`, or `anthropic` — those exact path segments, not `gemini` or
`claude`. Only sandbox *create* is intended for the Go API; the other sandbox routes are
development affordances. CORS is restricted to `http://api:8080`, so the browser cannot call this
service directly.

The three provider route modules are deliberate near-duplicates — same shape, different injected
dependency. Adding a provider means copying one, not generalising them.

### 7.2 Dependency injection

`api/dependencies.py` is the composition root, built from `@lru_cache()` factories exposed as
`Annotated[T, Depends(factory)]` aliases. The cache is what makes them singletons: one
`ChatOpenAI`, one `SandboxService`, one agent executor per provider for the process lifetime.
Routes reference only the alias, never the constructor.

Model choice, temperature, retries, and timeouts live in the client classes (`clients/*.py`) and
`api/config.py` (`pydantic-settings`, `.env`-backed). Logging is `structlog` in JSON.

### 7.3 The coding agent

`CodeAgentService` (`services/ai_services.py`) builds a LangChain tool-calling agent at
construction time: the `NEXTJS_PROMPT` template, a `PydanticOutputParser`, the sandbox tools, and
an `AgentExecutor`. A request is passed in as `"Sandbox ID: {id}\nTask: {message}"` — the sandbox
ID reaches the tools through the prompt text, not through the executor's configuration.

The system prompt (`prompts/nextjs_prompt.py`) carries the sandbox's ground rules: the project is
at `/home/user/`, the dev server is already running with hot reload (the agent must never start
it), styling is Tailwind-only, `.css` files are forbidden, `package.json` is edited only through
`npm install`, and `"use client"` goes at the top of `app/page.tsx`.

### 7.4 Sandbox tools

`SandboxService` wraps the E2B SDK and exposes four tools as nested `BaseTool` classes, each with a
Pydantic `args_schema`:

| Tool | Backing method |
| --- | --- |
| `list_sandbox_files` | `list_files` — refuses `/`, `/root`, `/etc`, `/sys`, `/proc` |
| `read_sandbox_file` | `read_file` |
| `write_sandbox_files` | `write_files` — batched |
| `execute_sandbox_command` | `execute_terminal_command` → `{stdout, stderr}` |

The tools are **synchronous on purpose** — it forces the agent to take one action at a time. Each
one reconnects with `Sandbox.connect(sandbox_id)` per call rather than holding a handle, and each
catches its own exceptions and returns the failure as a *string* so the agent can read and recover
from it instead of the run aborting.

### 7.5 Why a callback intercepts the agent — the trust boundary

An agent's narration of what it did is not evidence. `CodeAgentCallBack`
(`services/agent_callback_service.py`) is constructed fresh per request and hooks the executor:

- `on_tool_start` stages the tool's *inputs* (file paths + contents, or the command string).
- `on_tool_end` promotes staged inputs to the confirmed result only if the tool's output does not
  look like a failure; pending state is cleared either way.

The response the caller receives is therefore split by trust level: `summary` is the model's
prose (parsed out of `CodeAgentResult`), while `files` and `commands` come from the callback's
observed record. `CodeAgentResult` intentionally contains *only* a summary — the model is never
asked to report the file list it wrote, because that is the part it gets wrong.

If the agent returns empty output, the service still answers from the callback with a placeholder
summary rather than failing.

### 7.6 The sandbox template

`ai-service/sandbox-template/nextjs/` builds the E2B image: Node 21, `create-next-app@15.3.3`, and
every shadcn component pre-installed, all moved to `/home/user/` so the agent needs no
subdirectory. `compile_page.sh` runs at sandbox start, launching `next dev --turbopack` and polling
`localhost:3000` until it compiles, so the preview URL is warm on first view.

The built template is registered in the E2B account as `ats-nextjs-template` /
`4tx9flul6x7b6zfn8s47` (`e2b.toml`) and selected at creation through
`E2B_SANDBOX_NEXTJS_TEMPLATE_ID`. **Sandbox lifetime is configurable** via
`E2B_SANDBOX_TIMEOUT_SECONDS` (`api/config.py`, default 300 — E2B's own default, raised for demos);
no keep-alive or reconnection strategy exists yet, and that's deliberately out of scope — a manual
`RefreshSandbox` (Polish) is the intended fix, not background lifecycle management. `Sandbox.create`
is now called with no positional `template_id` — `SandboxService.create()` reads both the template
and the timeout from `settings` internally, since the one caller never varied the template. The
create response's `url` is scheme-qualified (`https://` prepended in `api/routes/sandbox.py`) —
E2B's `get_host()` returns a bare host, which silently fails as an `<iframe src>` without it.

Supporting a different stack (React, Vue) means a new Dockerfile, an `e2b template build` push, a
new template ID, and a matching system prompt — the prompt is stack-specific.

## 8. Proto and code generation ✅

`proto/api/v1/` holds the contract. Each RPC carries a `google.api.http` annotation, which is what
lets one handler serve both protocols.

**Defined today:** `UserService` (`GetUser`, `ListUsers`, `CreateUser`), `AccountService`
(`BeginAccountAuth`, `FinishAccountAuth`, `AccountLogout`, `GetUserProfile`), `ProjectService`
(`ListProjects`, `GetProject`, `CreateProject` — deliberately no `RenameProject` yet), and
`MessageService` (`ListMessages`, `SendMessage`, nested under
`/api/v1/projects/{project_id}/messages`); plus the `LoginProvider`, `AiProvider`, `MessageRole`,
and `ChatMode` enums. `AiProvider` (not `AIProvider`) is spelled that way on purpose — buf's
`protoc-gen-es` strips an enum's name as a value prefix by lowercasing and inserting an underscore
before every capital, so `AIProvider` would produce `AI_PROVIDER_` while the values are
`AI_PROVIDER_*`, breaking the short-name generation; `AiProvider` matches. There are no credit
messages yet, and the Go↔AI-service hop (`SendMessage`'s internals) is still untyped, hand-rolled
REST — see §4 — deliberately, not an oversight.

`make gen` (`buf generate`) fans one proto change out to four trees, **none of which may be edited
by hand**:

| Output | Plugin | Consumer |
| --- | --- | --- |
| `api/internal/infrastructure/inbound/grpc/gen/` | `protocolbuffers/go` + `connectrpc/go` | Go messages and handler interfaces |
| `app/src/gen/` | `protoc-gen-es` (target `ts`) | browser clients |
| `api/internal/infrastructure/inbound/http/handlers/openapi.yaml` | `google-gnostic-openapi` | the Swagger UI at `/docs/` |

Two constraints worth knowing before running it: the TypeScript plugin is resolved from
`app/node_modules/.bin/protoc-gen-es`, so the frontend's npm install must have happened; and
`buf.gen.yaml` overrides `go_package_prefix` for the whole module, which is why generated Go lands
under `.../inbound/grpc/gen` regardless of proto path.

Lint (`STANDARD`) and breaking-change detection (`FILE`) are configured in `buf.yaml`. Format
protos with `make plint` (`clang-format`).

### Vanguard: one handler, two protocols

`vanguard.NewTranscoder` wraps the Connect handlers so the same implementation answers Connect/gRPC
*and* the annotated REST routes. This sidesteps gRPC-web and its proxy requirements while keeping a
single source of truth. The transcoder is mounted at `/`; `/healthz` and `/docs/` are plain
`http.Handler`s registered ahead of it.

Two interceptors wrap every RPC, in this order: request logging, then auth.

## 9. Frontend ✅ (Build path) / 🟡 (Chat, streaming, mobile — out of scope)

The auth surface is ✅ **built** end to end. Past sign-in, `LoggedInScreen` is wired to the real
backend — `ListProjects`, `GetProject`, `CreateProject`, `ListMessages`, `SendMessage` all run
against the live API. `components/session/mock.ts` and `conversation-list.tsx` are gone. See
`.planning/Frontend/logged_in_design.md` for the screen-by-screen structure (still accurate) and
`.planning/tasks/Frontend/responsive-logged-in-screen/` for how the wiring landed.

**Auth is real, not scaffolding.** `src/hooks/services/useServiceClient.ts` builds a memoised
Connect transport from `NEXT_PUBLIC_API_URL` (falling back to `https://local.api.vulx.ai`), sends
JSON (`useBinaryFormat` is derived from that same base-URL check — binary once a non-local API URL
is configured), and every request carries `credentials: "include"` so the `.vulx.ai` session cookie
round-trips across the `app` → `api` origin hop. `useAccountService.ts` mirrors the existing
`useUserService.ts` one-line wrapper pattern.

`src/hooks/useSession.ts` is the single source of truth for "am I logged in": it calls
`GetUserProfile` through React Query (`queryKey: ["profile"]`), treats a `Code.Unauthenticated`
rejection as a normal `null` result rather than an error, and exposes a derived
`"loading" | "authed" | "anon"` status. `src/app/page.tsx` gates on that status alone — splash
(plain black screen, deliberately not a spinner) while loading, `LoggedInScreen` when authed,
`LoggedOutScreen` otherwise. `src/app/auth/callback/page.tsx` completes the round trip:
`FinishAccountAuth` on the `code`/`state` pair Google redirects back with, guarded by a `useRef`
latch against React 19 StrictMode's double-invoked effects, then `router.replace("/")`.

`QueryClientProvider` (`src/app/providers.tsx`, a per-session client via lazy `useState`, not a
module-level singleton — avoids leaking one user's cached profile to another across SSR requests)
and `<Toaster theme="dark" />` are mounted in `layout.tsx`. `@tanstack/react-query` is no longer
just an installed-but-unused dependency.

`src/components/ui/` is a full shadcn install — treat it as vendored, restyle at the call site only.
`src/components/session/` holds all post-login UI, distinct from `src/components/landing/`.

**`LoggedInScreen` (`logged-in-screen.tsx`) is one `SidebarProvider` shell over a three-state `View`
union** — `{ kind: "home" }`, `{ kind: "draft" }` (Workspace open, no project created yet), or
`{ kind: "project"; id }` — collapsed for layout purposes into `isWorkspace = view.kind !== "home"`:

- **Home** (`view.kind === "home"`): the sidebar renders `ProjectList` over `useProjects()` (real
  `Project` rows, `formatRelative` timestamps, an active-row highlight, no client-side sort — the
  backend already returns `updated_at DESC`), and the inset renders `HomeView` (the two-tier
  "Welcome back" headline + a prompt box carrying the same mode/provider `ComposerControls` as the
  Workspace composer, since the provider has to be chosen before `CreateProject` runs). Its
  "New build" button sets the view to `draft` — no project exists until the first send.
- **Workspace** (`draft` or `project`): the *same* sidebar panel becomes the project's chat thread
  (`ChatPanel`, keyed on `projectId ?? "draft"` so switching projects resets the composer draft and
  re-seeds the provider select), reading `useMessages(projectId)` for the thread (`enabled:
  !!projectId`, so a draft simply renders an empty thread) — and the inset becomes `PreviewPane`, a
  real `<iframe>` once `project.previewUrl` is non-empty, with an opaque `bg-surface` overlay and a
  cycling `GeneratingLine` shimmer ("Thinking…" / "Building…" / "Creating…") while a send is in
  flight. `Sidebar`'s `collapsible` prop is `"offcanvas"` in **both** states on purpose: only the
  `--sidebar-width` CSS variable changes (16rem → 26rem) between them, and that variable is what
  animates as a slide. Letting `collapsible` itself differ between Home and Workspace would put the
  two states on different internal render branches of the vendored `Sidebar`, forcing a full
  unmount/remount instead of a transition — this was tried and reverted. Home deliberately renders no
  `SidebarTrigger`/`SidebarRail`, so nothing on that screen can collapse it even though the mechanism
  underneath supports collapsing.
- **`LoggedInScreen.start()` is the one create-then-send flow**, shared by Home's submit, the
  Workspace composer's `onSend`, and a draft's first message: if no project id exists yet, it awaits
  `useCreateProject().mutateAsync({firstPrompt, provider})`, switches the view to the new id
  immediately — not when the build finishes, so the sidebar row and a real thread appear right away
  — then awaits `useSendMessage().mutateAsync(...)` against that id. `generating`
  (`createProject.isPending || sendMessage.isPending`) is the single boolean both `ChatPanel` and
  `PreviewPane` read, so the thread shimmer and the preview overlay can never disagree. A thrown
  error anywhere in the sequence surfaces as a toast; no other error UI is in scope.
- **The preview URL arrives through cache invalidation, not the response.** `SendMessageResponse`
  carries only the user/assistant messages, not the sandbox — `useSendMessage`'s `onSettled`
  invalidates `["messages", id]`, `["project", id]` and `["projects"]` together, and it is the
  `["project", id]` invalidation specifically that makes `previewUrl` show up once a sandbox exists.
- **Project titles fill in a few seconds after creation**, and the frontend has to notice on its
  own — this app runs with `refetchOnWindowFocus: false` (`providers.tsx`), so nothing refetches
  spontaneously. `LoggedInScreen` tracks the id of a just-created project as `pendingTitleId` (set
  only in `start()`'s create branch, cleared after an 18s window matching the backend's bound) and
  passes a `pollForTitle` flag into `useProject`, which turns on a 2s `refetchInterval` only for
  that one query — never for an already-open project. `ChatPanel` snapshots the title it first
  received on mount, shows a typing "Generating title…" (`TypingText`/`TypingTextCursor`, the same
  primitive the logged-out hero's prompt cycle uses) while the polled title still matches that
  snapshot, and cross-fades (`vx-fade`) to the real title the instant it changes, rather than
  waiting for the poll window to close.
- Logging out goes through a confirm `AlertDialog` ("Log out?") before `AccountLogout` fires — sized
  to match the auth dialog's `sm:max-w-sm` for visual consistency between the app's two modals.

`AuthDialog` (`components/auth/auth-dialog.tsx`) is the shared login/signup modal: one
"Continue with Google" button (the default, near-white `Button` variant — no dark outline) that calls
`BeginAccountAuth` the same way for both modes, and a title-only header (no subtext) per mode.

A single non-system accent colour, `--accent-blue` (`#569CD6`), was added on top of the otherwise
near-monochrome palette to mark selected/active state (the mode toggle's active icon, the composer
border, the project-list dot, the generating row's Build icon, the logged-out "Log in" outline). See
`.planning/Frontend/design-system.md` §1 and §3 for the token and the reasoning — this file only
notes that it exists and is intentional, not accidental drift from the one-hue rule.

**What is out of scope, deliberately, not missing by oversight:** Chat mode stays selectable but
returns `Unimplemented` from the backend (`ChatMode`'s toggle is intentionally not gated —
`Polish.md`); there is no routing, so a refresh returns to Home and the open project is lost; no
pagination on the project list or the message thread; no draggable chat/preview split; no mobile
layout. Credits are read (`Profile.credits`, a `bigint`) but never displayed prominently or spent —
the feature doesn't exist yet.

Path aliases: `@/*` → `src/*`, `@apiv1/*` → `src/gen/api/v1/*`.

## 10. Implementation status

| Capability | Status | Notes |
| --- | --- | --- |
| Google OAuth login, JWT sessions, logout revocation | ✅ | end to end |
| User CRUD, keyset pagination | ✅ | admin-gated by a hardcoded email |
| Proto → Go/TS/OpenAPI codegen | ✅ | |
| Connect + REST dual serving | ✅ | |
| Sandbox creation and agent file/command execution | ✅ | via the Python service directly, and via the Go API |
| All three model providers | ✅ | query and code-agent paths |
| Go API → AI service calls | ✅ | `CreateSandbox`, `RunCodeAgent`; 30s timeout ceiling removed — §11 |
| Projects: schema, CRUD, ownership | ✅ | `ListProjects`/`GetProject`/`CreateProject`, keyset pagination on `updated_at`; titles generated asynchronously via OpenAI — §5.6 |
| Messages: schema, synchronous Build flow | ✅ | `ListMessages`/`SendMessage`; Chat mode returns `Unimplemented` — `Polish.md` |
| Temporal workflows | 🟡 | connection + a worker with nothing registered on it; the demo workflow is kept as an unwired reference file only — §11 |
| Frontend auth surface (login, session gate, logout) | ✅ | end to end |
| Frontend beyond auth (dashboard, editor, preview) | ✅ | wired to the real backend — §9; synchronous Build path only, no Chat/streaming/pagination/routing/mobile |
| Credits: schema and display | 🟡 | granted (10 by default) and read; never spent |
| Sandbox persistence / reuse | ✅ | `sandbox_id`/`preview_url` stored on `Project`, reused across messages in the same project |
| Sandbox keep-alive / dead-sandbox recovery | ⛔ | a dead sandbox and a real outage both surface as a plain 500 — `Polish.md` |
| Version history, code export (ZIP), codebase persistence | ⛔ | agent writes are discarded after each run — `Polish.md` |
| Payments | ⛔ | no planning doc exists yet either — `Polish.md` |
| Tests | ⛔ | no test file exists in any service |

## 11. Known gaps and defects

Verified against the code. Fix these deliberately rather than building on top of them.

1. ~~**The Go→AI-service sandbox call cannot succeed.**~~ **Fixed.** `sandbox.go` now issues
   `POST {base}/sandbox/` through a shared `do` helper that checks the transport error *before*
   touching `resp` (the old code closed `resp.Body` first, which nil-derefed on a transport
   failure). The client's `http.Client.Timeout` was also raised from 30s to a 15-minute backstop —
   the 30s ceiling overrode any longer per-call context deadline, which is why a real code-agent run
   could never have succeeded regardless of anything else being correct.
2. ~~**`AIService.CallAI()` is an empty stub.**~~ **Fixed.** Replaced by `RunCodeAgent`
   (`outbound/ai_service/llm.go`), which posts to `/{provider}/{sandboxID}/code` and is now the
   `AIService` port `MessageService` calls for the Build flow. `openai_agent.go` is deleted.
3. ~~**The auth interceptor flattens handler error codes.**~~ **Fixed.** `interceptor.go` now passes
   through the real code for client-caused errors (`Unauthenticated`, `PermissionDenied`,
   `InvalidArgument`, `NotFound`, `AlreadyExists`) and collapses anything else — `Internal`,
   `Unavailable`, `Unknown`, etc. — to a generic `CodeInternal` with a fixed message, logging the
   real error server-side first. This is a deliberate choice, not the literal "always return
   unwrapped" originally proposed: it avoids leaking infrastructure detail (a dropped DB connection,
   a Redis timeout) in the error message reaching the browser, while still giving the frontend
   accurate codes for genuine client-side problems.
4. ~~**Cookie parsing drops the JWT when it is not the last cookie.**~~ **Fixed.** The loop in
   `grpc/adapters/auth/auth.go` now returns on the first `jwt` match instead of continuing and
   resetting `token` to `""` on every non-matching cookie.
5. ~~**OAuth state is replayable within its TTL.**~~ **Fixed.** `CompleteLoginFlow`
   (`oauth_service.go`) now deletes `provider:<state>` from Redis immediately after a successful
   `Exchange` (not before — a transient Google failure must not burn the state). A delete failure is
   logged, not returned, since the login itself already succeeded by that point.
6. **The callback's success test is string matching.** `agent_callback_service.py:42,52` treats any
   tool output containing `"error"` as a failure — including a successful `cat` of a file that
   mentions the word. This silently drops real file writes from the result.
7. ~~**The frontend demo page does not compile.**~~ **Fixed.** `app/src/app/page.tsx` was rewritten
   as the session gate (splash / logged-out / logged-in); the `create-next-app` scaffold and its
   broken `createUser({name: "tony"})` call are gone.
8. **The Temporal workflow is unwired, kept only as a reference.** `user_workflow.go` (a
   `CreateSandbox` activity, an empty `UseLlm`, `StartUserWorkflow` blocking on `workflowRun.Get` —
   making the "async" orchestration synchronous) is deliberately **not deleted**, since it's a
   useful shape to copy from when the real Build workflow is written (`Polish.md`,
   `build-orchestration.md`). But nothing constructs or registers it: `RegisterWorkers()` takes no
   argument and only creates a bare, empty `worker.Worker` (with a comment showing how to register a
   real one), and `StopWorkers()`'s body is fully commented out rather than calling `.Stop()` on a
   worker that was never told to run anything — calling `.Stop()` unconditionally here would be
   harmless today, but the point is there is currently no real worker to stop.
9. **The callback's success test is string matching.** `agent_callback_service.py:42,52` treats any
   tool output containing `"error"` as a failure — including a successful `cat` of a file that
   mentions the word. This silently drops real file writes from the result. `Polish.md`.
10. **A dead sandbox and a real AI-service outage are indistinguishable.** The Python service
    returns a plain 500 for both; `MessageService.Send` reuses a project's stored `sandbox_id`
    blindly once it's non-empty, so a message to a dead sandbox just fails with an opaque
    `Unavailable`. Mitigated, not fixed, by the raised sandbox timeout (§7.6); the real fix is a
    `RefreshSandbox` RPC — `Polish.md`.
11. **`Project.provider` can go stale.** `SendMessage` never updates it even though every send
    carries its own provider — spec'd in `messages-model.md` §8.1/§8.4 as applying to every send,
    but omitted from the task-level Build-flow implementation and found live during MVP testing.
    Deliberately deferred rather than patched in after the fact — `Polish.md`.
12. **Superuser authorisation is a hardcoded email literal** in four places. It needs a role column.
13. **`services/cahce.go` is misspelled**, as is `expirestAt` in the token service. Renaming the
    file is safe; be aware when grepping.

## 12. Extension recipes

**Add an RPC.** Define the message and the `google.api.http` annotation in `proto/api/v1/` →
`make gen` → implement the method on the handler in `inbound/handlers/` (call
`authAdapter.User(ctx)` first if it must be protected, and map errors with
`errorAdapter.ToConnectError`) → if it is a new *service*, register it in the `vanguard.NewService`
list in `application/app.go`.

**Add an outbound dependency.** Declare the interface in the consuming service file under
`application/services/` → implement it in `infrastructure/outbound/` (or `persistence/`, `cache/`)
→ construct and inject it in `application/app.go`. The implementation must return `*domain.Error`
with a meaningful `ErrorType`.

**Add a database table.** New numbered file in `persistence/postgres/migrations/` with `+migrate
Up` and `Down` sections → a row struct with `db` tags and `ToDomain()` → a `RestoreX` constructor in
`domain/` → repository methods. It applies on the next API start; `make nuke` to reset.

**Add an OAuth provider.** New file in `outbound/oauth/` implementing `services.OauthProvider` →
register it in `OauthProviderRegistry.Provider` → extend `domain.LoginProvider` and
`ParseLoginProvider` → extend the `LoginProvider` enum in `enums.proto` and the mapping in
`handlers/account.go`.

**Add a model provider.** New client in `ai-service/clients/` → factories and `Annotated` aliases in
`api/dependencies.py` → copy a route module in `api/routes/` and change the injected dependency →
include the router in `api/main.py` → add the API key and model name to `api/config.py` and
`.env.example`.

**Add a sandbox stack.** New directory under `sandbox-template/` with an `e2b.Dockerfile` and a
start script → build and push to E2B → record the template ID in config → write a matching system
prompt in `prompts/`, since the current one describes Next.js specifics.

## 13. Key decisions and their trade-offs

**Monorepo with separate services.** Shared protos, atomic cross-service commits, one compose file.
The services are split by hard constraint rather than by scaling need: Python is required for E2B
and LangChain, Go is wanted for the API. Independent scaling is a downstream benefit, not the
motivation.

**Hexagonal/DDD in the Go API.** Costs an interface and a mapping layer for every dependency; buys
a domain that can be reasoned about without a database and swapped implementations behind stable
contracts. Given there are no tests yet, the testability argument is currently unrealised.

**Connect + Vanguard over plain gRPC.** Browsers cannot speak gRPC without a proxy. Connect speaks
HTTP natively and Vanguard transcodes REST from the same annotations, so there is one handler, one
schema, and no Envoy.

**Protobuf as the contract.** Type safety across two languages, backward-compatibility rules, and
generated clients. The cost is a codegen step in the loop — and note that the boundary it does
*not* cover, Go↔Python, is exactly the one that has drifted.

**Stateless JWT in an httpOnly cookie.** No session table on the read path and immunity to
JavaScript exfiltration, at the cost of needing a Redis denylist for logout and accepting that a
compromised token is valid until it expires. The 42-hour sliding renewal keeps active users signed
in without a refresh-token dance.

**A separate AI service.** Forced by the E2B Python SDK; also keeps LangChain's dependency weight
and release cadence out of the Go build.

**Temporal.** Chosen for the eventual multi-step generation pipeline — durable retries, crash
recovery, and visibility into long runs. Currently connected but running nothing (§11) — the MVP's
Build flow is synchronous, inline in the `SendMessage` handler, not a workflow. `ai-service-client.md`
set an explicit falsification condition for when `build-orchestration.md` is written: if Build stays
a single activity, remove Temporal rather than keep it as infrastructure ahead of a use case it
never gets.
