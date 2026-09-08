This file provides guidance to AI AGENTS when working with code in this repository.

## Overview

VULX — a SaaS that generates frontend code from natural language and previews it live in an E2B
sandbox. Monorepo of four units: `api` (Go, main backend), `ai-service` (Python/FastAPI, LLM +
sandbox), `app` (Next.js), `proto` (the contract between `app` and `api`).

See `ARCHITECTURE.md` for the long-form design rationale and the full OAuth/JWT flow.

## Commands

Everything runs in Docker Compose; there is no supported host-only workflow.

```bash
make            # = make app: build+run the whole stack (compose deps pull in api, ai-service, db, redis, temporal, caddy)
make api        # hot-reload loop for the Go API only (reflex watches api/, rebuilds the container)
make clean      # docker compose down
make nuke       # down -v + prune — use for stale DB/volume/cache problems
make gen        # buf generate: regenerate Go + TS + OpenAPI from proto/
make plint      # clang-format the .proto files
```

Ports: app `3000`, api `8080`, ai-service `9999`, Temporal UI `8081`, Postgres `5432`,
Caddy `80/443` (serves `local.vulx.ai` → api; needs a `127.0.0.1 local.vulx.ai` hosts entry).

Per-service, when you need it outside compose:

```bash
cd api        && go build ./... && go vet ./...
cd app        && npm run lint && npm run build
cd ai-service && ruff check .        # ruff is in .venv/bin
```

There is no test suite in this repo yet — do not claim tests pass; say what you actually ran.

## Environment

`docker-compose.yaml` reads env from a **sibling directory outside the repo**:
`../AI-Website-Builder-Secrets/.api-env` and `.ai-service-env`. Compose fails to start without
these. The required keys are listed in `api/.env.example` and `ai-service/.env.example`.

## Protobuf is the source of truth

`proto/api/v1/*.proto` define both RPC messages and REST routes (via `google.api.http`
annotations). `make gen` fans out to four generated trees — **never hand-edit these**:

| Output | Consumer |
| --- | --- |
| `api/internal/infrastructure/inbound/grpc/gen/` | Go server stubs + Connect handlers |
| `app/src/gen/` | TS clients (`protoc-gen-es`) |
| `api/internal/infrastructure/inbound/http/handlers/` | OpenAPI doc served at `/docs/` |

`make gen` invokes `app/node_modules/.bin/protoc-gen-es`, so `app`'s npm deps must be installed
first. Adding an endpoint means: edit the proto → `make gen` → implement the handler in
`api/internal/infrastructure/inbound/handlers/` → register the service in `application/app.go`.

## Go API (hexagonal / DDD)

Wiring lives in one place: `api/internal/application/app.go` constructs every dependency and
injects it downward. Read it first when tracing anything.

```
cmd/api/main.go              config load → application.New → Start
internal/domain/             pure entities + the error taxonomy; zero external deps
internal/application/services/  use cases. Ports (interfaces) are declared HERE, next to the
                                consumer, and implemented by infrastructure/
internal/infrastructure/
  inbound/handlers/          Connect handlers — thin: unwrap request, call service, map error
  inbound/grpc/adapters/     auth interceptor, logging interceptor, CORS, domain→connect errors
  inbound/auth_token/        JWT mint/verify
  outbound/                  postgres, redis, oauth (google), temporal, ai_service (REST client)
```

Conventions that matter:

- **Errors**: every layer returns `domain.WrapError("context", err)`; the error *type*
  (`domain.ErrorTypeNotFound`, …) propagates through the chain. `grpcerror.ToConnectError` is the
  single place that maps a domain type to a Connect code — add new categories in
  `domain/error.go`, not at the handler.
- **Auth**: the interceptor in `grpc/adapters/auth/interceptor.go` runs on *every* RPC and only
  puts the user in context when a valid `jwt` cookie exists. There is no public-route allowlist —
  a handler makes itself protected by calling `auth.User(ctx)` and returning its error. The
  interceptor also silently re-issues the cookie when the token is near expiry.
- **Migrations**: `internal/infrastructure/persistence/postgres/migrations/*.sql` are `go:embed`ed
  and run automatically on API startup (`sql-migrate` up). Add a new numbered file; never edit an
  applied one. Use `make nuke` to reset the DB.
- **Vanguard** wraps each Connect service so one handler serves both gRPC and REST.

## AI service (FastAPI)

Exists separately because the E2B SDK is Python-only. Called by the Go API over plain REST
(`outbound/ai_service/`); it never talks to the database.

- All routes are mounted under `/ai-service/v1` in `api/main.py`.
- `api/dependencies.py` is the DI container: every client/service is an `@lru_cache()` factory
  exposed as an `Annotated[..., Depends(...)]` alias. Add a new provider by adding a client in
  `clients/`, then a factory + alias here — routes should only reference the alias.
- Three providers (openai / google / anthropic) each expose a `query` endpoint (plain LLM) and a
  `code-agent` endpoint (LangChain agent with sandbox tools). The per-provider route modules are
  near-identical wrappers over the shared `CodeAgentService` / `GeneralAIService`.
- `services/agent_callback_service.py` is the trust boundary for agent output: the LangChain
  callback records actual tool invocations, and responses report **only** the captured file writes,
  never the agent's own narration.
- Config is `pydantic-settings` (`api/config.py`); logging is `structlog` JSON.

## Frontend (Next.js App Router)

Still early — `app/src/app/` is largely scaffold. `src/components/ui/` is generated shadcn; treat
it as vendored. Backend access goes through `src/hooks/services/useServiceClient.ts`, which builds
a Connect transport (currently hardcoded to `http://localhost:8080`) and is wrapped per service by
`useUserService`-style hooks. Import message/service types from `@/gen/api/v1/*`.

