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

**Where the product actually is:** the AI service can drive a sandbox end to end, and the Go API
can authenticate a user. The seam *between* those two halves is scaffolding — see
[§10 Implementation status](#10-implementation-status). Do not assume a prompt-to-preview path
exists.

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
   ├──────────────► app        :3000   Next.js dev server
   │                  │
   │                  │ Connect RPC (JSON over HTTP)
   │                  ▼
   └──────────────► api        :8080   Go — Connect + Vanguard
       (or via caddy :443,              │
        local.vulx.ai, internal TLS)    │
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
| `caddy` | 80/443 | `local.vulx.ai` → `api:8080` with `tls internal` |

**Caddy's actual job** is narrow: it terminates HTTPS for the API only, so the browser will accept
`Secure` JWT cookies during local development. It does not proxy the frontend and does not serve
documentation. Reaching the API through it requires `127.0.0.1 local.vulx.ai` in `/etc/hosts` and
trusting Caddy's local CA.

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
    │   ├── sandbox.go                  Sandbox entity (defined, not yet persisted or used) 🟡
    │   ├── page.go                     Page[T] — generic cursor-paginated result
    │   └── error.go                    Error{type, err}, WrapError, ErrorType constants
    ├── application/
    │   ├── app.go                      ⭐ THE WIRING FILE — every dependency is built here
    │   └── services/                   USE CASES + the ports they depend on
    │       ├── user_service.go         + UserRepository, UserWorkflowService ports
    │       ├── auth_service.go         + AuthAdapter port
    │       ├── oauth_service.go        + OauthProvider, OauthProviderRegistry ports
    │       ├── account_service.go      composes oauth + auth + user (no new ports)
    │       └── cahce.go                Cache port [sic — filename is misspelled in the repo]
    ├── infrastructure/
    │   ├── inbound/
    │   │   ├── handlers/               Connect handlers: account.go, user.go
    │   │   ├── grpc/gen/               ⚠️ GENERATED by buf — never hand-edit
    │   │   ├── grpc/adapters/auth/     JWT cookie read/write + the auth interceptor
    │   │   ├── grpc/adapters/error/    domain.Error → connect.Code
    │   │   ├── grpc/adapters/logger/   structured per-RPC request logging
    │   │   ├── grpc/adapters/security/ CORS allowlist
    │   │   ├── auth_token/auth.go      Ed25519 JWT mint / validate / parse
    │   │   └── http/handlers/          /healthz and the Swagger UI at /docs/
    │   ├── outbound/
    │   │   ├── oauth/                  Google provider + a name→provider registry
    │   │   ├── temporal/               client, worker registration, the example workflow
    │   │   └── ai_service/             REST client for the Python service 🟡
    │   ├── persistence/postgres/       repositories, embedded migrations, cursor codec
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
client, the database and repositories, Redis, Temporal (registering workers), the token service,
the auth/oauth/user/account services, the Connect handlers, and finally the interceptor chain and
mux. Nothing anywhere else calls a constructor for a long-lived dependency.

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

Current schema: `users` (uuid pk, unique email, `credits INTEGER DEFAULT 10`, `is_active`) and
`user_auth_providers` (one row per user, unique on `(provider, provider_user_id)`, cascading
delete). The `citext` extension is enabled but not yet used by any column.

**Pagination** is keyset-based, not offset: `cursor.go` base64-encodes the `created_at` of the last
row as an opaque token, and results come back as `domain.Page[T]{Items, Token, HasMore}`.

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
`E2B_SANDBOX_NEXTJS_TEMPLATE_ID`. **Sandboxes live 5 minutes by default** — no keep-alive or
reconnection strategy exists yet.

Supporting a different stack (React, Vue) means a new Dockerfile, an `e2b template build` push, a
new template ID, and a matching system prompt — the prompt is stack-specific.

## 8. Proto and code generation ✅

`proto/api/v1/` holds the contract. Each RPC carries a `google.api.http` annotation, which is what
lets one handler serve both protocols.

**Defined today — and only this:** `UserService` (`GetUser`, `ListUsers`, `CreateUser`),
`AccountService` (`BeginAccountAuth`, `FinishAccountAuth`, `AccountLogout`, `GetUserProfile`), and
a `LoginProvider` enum. There are no AI, sandbox, message, or credit messages yet; those hops are
either untyped REST or unbuilt.

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

## 9. Frontend 🟡

Scaffolding with a working RPC transport. Almost none of the product UI exists.

**What is real:** `src/hooks/services/useServiceClient.ts` builds a memoised Connect transport and
returns a typed client for any generated service; `useUserService.ts` is the one-line wrapper
pattern to copy. `src/components/ui/` is a full shadcn install — treat it as vendored.

**What is not:** `src/app/page.tsx` is a stale demo (see §11). There is no `/auth/callback` route,
no login UI, no dashboard, no editor, no preview pane, and no protected-route logic.
`@tanstack/react-query` is installed but not imported anywhere — there is no `QueryClientProvider`
in `layout.tsx`, so descriptions of "React Query cache" are aspirational.

The transport hardcodes `http://localhost:8080` and selects JSON over binary from that same
constant. It sends no credentials option, which will need attention when cookie-authenticated
calls are wired up.

Path aliases: `@/*` → `src/*`, `@apiv1/*` → `src/gen/api/v1/*`.

## 10. Implementation status

| Capability | Status | Notes |
| --- | --- | --- |
| Google OAuth login, JWT sessions, logout revocation | ✅ | end to end |
| User CRUD, keyset pagination | ✅ | admin-gated by a hardcoded email |
| Proto → Go/TS/OpenAPI codegen | ✅ | |
| Connect + REST dual serving | ✅ | |
| Sandbox creation and agent file/command execution | ✅ | via the Python service directly |
| All three model providers | ✅ | query and code-agent paths |
| Go API → AI service calls | 🟡 | one stub method, one broken URL — §11 |
| Temporal workflows | 🟡 | a demo workflow only; no code-generation workflow |
| Frontend beyond a transport | 🟡 | demo page does not compile |
| Credits: schema and display | 🟡 | granted (10 by default) and read; never spent |
| Sandbox persistence / reuse / keep-alive | ⛔ | `domain.Sandbox` is defined but unused; no table |
| Version history, code export (ZIP) | ⛔ | |
| Messages / conversation model | ⛔ | referenced in code comments as the next thing to build |
| Payments | ⛔ | |
| Tests | ⛔ | no test file exists in any service |

## 11. Known gaps and defects

Verified against the code. Fix these deliberately rather than building on top of them.

1. **The Go→AI-service sandbox call cannot succeed.**
   `outbound/ai_service/sandbox.go:16` issues `GET {base}/sandbox/create`, but the route is
   `POST /ai-service/v1/sandbox/`. The same function ignores the returned error (`:18` is an empty
   `if` body with a TODO) and then dereferences `resp`, so a transport failure panics rather than
   returning.
2. **`AIService.CallAI()` is an empty stub** (`outbound/ai_service/openai_agent.go:8`) — it returns
   `nil` without calling anything. No Go code reaches the LLM endpoints.
3. **The auth interceptor flattens handler error codes.** For *authenticated* requests,
   `interceptor.go:26` re-wraps every handler error as `CodeInvalidArgument`, discarding the code
   that `ToConnectError` just derived. A `NotFound` reaches an authenticated caller as
   `InvalidArgument`; unauthenticated callers get the correct code. Return the error unwrapped.
4. **Cookie parsing drops the JWT when it is not the last cookie.** The loop in
   `grpc/adapters/auth/auth.go:70-76` resets `token` to `""` on every non-`jwt` cookie, so a
   session survives only if `jwt` happens to be parsed last. Break on match instead.
5. **OAuth state is replayable within its TTL.** `provider:<state>` is read in `CompleteLoginFlow`
   (`oauth_service.go:125`) but never deleted, so the same `code`/`state` pair can be presented
   repeatedly for up to 10 minutes. Delete the key after a successful exchange.
6. **The callback's success test is string matching.** `agent_callback_service.py:42,52` treats any
   tool output containing `"error"` as a failure — including a successful `cat` of a file that
   mentions the word. This silently drops real file writes from the result.
7. **The frontend demo page does not compile.** `app/src/app/page.tsx` calls
   `createUser({name: "tony"})` and reads `user.name`, but the proto has `first_name`, `last_name`,
   and `email` — `name` does not exist on the generated type. Delete or rewrite the page.
8. **The Temporal workflow is a placeholder.** `user-workflow` runs a `CreateSandbox` activity and
   an empty `UseLlm`; `StartUserWorkflow` blocks on `workflowRun.Get`, making the "async"
   orchestration synchronous. Its own comment says it will be replaced by the messages service.
9. **Superuser authorisation is a hardcoded email literal** in four places. It needs a role column.
10. **`services/cahce.go` is misspelled**, as is `expirestAt` in the token service. Renaming the
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
recovery, and visibility into long runs. Currently carries only a demo workflow, so it is
infrastructure ahead of its use case.
