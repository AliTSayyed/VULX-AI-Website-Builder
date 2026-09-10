# Project Service — Proto

> The contract for projects. This document is **authoritative for `proto/api/v1/project_service.proto`
> and the `AiProvider` addition to `enums.proto`** — when it and `.planning/Backend/project-model.md`
> disagree about a field name, number, or annotation, this document wins and the other is corrected.
>
> It also carries the **conventions section (§2) that the other two proto documents reference**
> rather than restate: buf lint rules, field-number discipline, REST annotation mechanics, and what
> `make gen` produces on each side.
>
> Backend behaviour — handlers, services, ownership checks, the title goroutine — is
> `project-model.md`. This document stops at the wire.
>
> Verified against the code on `feature/messages` (2026-09-08).

## ⏸ Phase split — MVP vs Polish

| § | Content | Phase |
| --- | --- | --- |
| §2 | conventions — lint, field numbers, REST annotations, codegen | **MVP** (and it governs all three proto docs) |
| §3 | `AiProvider` in `enums.proto` | **MVP** |
| §4 `ListProjects`, `GetProject`, `CreateProject` | | **MVP** |
| §4 `RenameProject` | | Polish |
| §5.1–§5.3, §5.5 | `Project` fields, timestamps, pagination, why create takes the prompt | **MVP** |
| §5.4 | `PATCH` | Polish — it only exists for `RenameProject` |
| §6 | landing it, the enum-prefix check | **MVP** |

**§2.3 is the one thing to get right on the first `make gen`.** `AiProvider` versus `AIProvider`
decides whether the frontend writes `AiProvider.OPENAI` or `AiProvider.AI_PROVIDER_OPENAI`, and it
fails silently. Everything else here is additive later; this one is annoying to change once the TS
client is wired.

For the MVP, drop the `RenameProject` RPC and its two messages from §4 — the rest of the file is
unchanged.

## 1. What is being added

| File | Change |
| --- | --- |
| `proto/api/v1/enums.proto` | append `AiProvider` |
| `proto/api/v1/project_service.proto` | **new** — `ProjectService`, 4 RPCs, 9 messages |

`RefreshSandbox` is a fifth RPC on this same service, owned by `project-codebase-proto.md`. It is
mentioned there and deliberately absent here so the two can land separately.

## 2. Conventions — read once, applies to all three proto documents

### 2.1 Buf lint is `STANDARD`, breaking detection is `FILE`

From `buf.yaml`. Four consequences that shape every message below:

- **Every RPC needs `FooRequest` / `FooResponse` messages** named after it, even when empty. That is
  why `GetProjectRequest` exists rather than passing a bare string.
- **Enum values must be prefixed with the enum name in UPPER_SNAKE_CASE**, and the zero value must
  end `_UNSPECIFIED`. This is why the constants look verbose — see §2.2 for why that cost is
  invisible to the frontend, and §2.3 for the trap.
- **`FILE`-level breaking detection means field numbers are permanent.** Never renumber, never reuse
  a deleted number — `reserved` it. Renaming a field is safe on the wire in binary but **breaks the
  JSON encoding**, which is what the browser uses today (§2.5), so treat renames as breaking too.
- Format with `make plint` (clang-format) before committing. It reflows the `option` blocks in a
  specific way; hand-formatted protos will churn on the next run.

### 2.2 The verbose enum prefix disappears in TypeScript

`protoc-gen-es` strips the shared prefix from enum values. The existing generated file proves it:

```ts
// app/src/gen/api/v1/enums_pb.ts
export enum LoginProvider {
  UNSPECIFIED = 0,   // from LOGIN_PROVIDER_UNSPECIFIED
  GOOGLE = 1,        // from LOGIN_PROVIDER_GOOGLE
}
```

So the frontend writes `AiProvider.ANTHROPIC`, not `AiProvider.AI_PROVIDER_ANTHROPIC`. The proto
verbosity is a lint requirement, not a burden the UI carries.

### 2.3 `AiProvider`, not `AIProvider` — a verified trap

The stripping in §2.2 is conditional. `@bufbuild/protobuf`'s `findEnumSharedPrefix` computes the
expected prefix with `camelToSnakeCase`, which inserts an underscore before **every** capital:

```js
// node_modules/@bufbuild/protobuf/dist/cjs/registry.js
function camelToSnakeCase(camel) {
    return (camel.substring(0, 1) + camel.substring(1).replace(/[A-Z]/g, (c) => "_" + c)).toLowerCase();
}
```

Run against the candidates:

| Enum name | Prefix protoc-gen-es expects | Values `AI_PROVIDER_*` stripped? |
| --- | --- | --- |
| `AIProvider` | `A_I_PROVIDER_` | ❌ → `AIProvider.AI_PROVIDER_OPENAI` in TS |
| `AiProvider` | `AI_PROVIDER_` | ✅ → `AiProvider.OPENAI` |
| `ChatMode` | `CHAT_MODE_` | ✅ |
| `MessageRole` | `MESSAGE_ROLE_` | ✅ |

**Use `AiProvider`.** Consecutive capitals are the only shape where this bites, and it fails
*silently* — `buf lint` may well accept `AI_PROVIDER_` for `AIProvider` (buf's snake-casing is its
own implementation), so the first sign of trouble would be ugly names appearing in the frontend.

The Go domain type stays `domain.AIProvider`, since Go convention capitalises initialisms. Only the
proto identifier changes; the handler already maps between the two enums (`project-model.md` §8.1).

### 2.4 REST annotations and how Vanguard reads them

Each RPC carries a `google.api.http` option, and Vanguard uses it to serve the same handler over
REST as well as Connect/gRPC.

- **A `{name}` path variable must match a field name on the request message.** `GET /projects/{id}`
  requires `GetProjectRequest.id`. A mismatch is a `make gen` or runtime routing failure, not a
  compile error.
- **`body: "*"` means "every field not already bound to the path".** So `RenameProjectRequest.id`
  comes from the URL and only `title` is read from the JSON body — you do not, and must not, send
  `id` twice.
- **Omit `body` when every field is a path variable.** An RPC whose request is only an id needs no
  body declaration at all.
- **`GET` never takes a body.** Non-path fields become query parameters — `ListProjects`' `limit`
  and `token` arrive as `?limit=25&token=…`.

### 2.5 What the browser actually speaks

The frontend does **not** call the REST paths. `useServiceClient.ts` builds a Connect transport, so
the browser posts to `/api.v1.ProjectService/ListProjects`. The REST annotations exist for `curl`,
the Swagger UI at `/docs/`, and any future non-Connect client. One handler serves both.

Two things follow:

- **JSON is the wire format today**, but `useBinaryFormat` is derived from `NEXT_PUBLIC_API_URL` and
  flips to binary once a non-local API URL is configured. Anything that works only in JSON — a field
  name the client depends on, say — will break in production and not in development.
- The gnostic OpenAPI plugin runs with `enum_type=string`, so the `/docs/` schema and every REST
  request spell enums as their **full proto names**: `"AI_PROVIDER_ANTHROPIC"`, not `"anthropic"`
  and not `2`. That is what the `curl` examples in the backend docs use.

### 2.6 `make gen` mechanics

One command fans out to four trees, **none of which may be hand-edited**:

| Output | Plugin | Consumer |
| --- | --- | --- |
| `api/internal/infrastructure/inbound/grpc/gen/` | `protocolbuffers/go` + `connectrpc/go` | Go messages, handler interfaces |
| `app/src/gen/` | `protoc-gen-es` (target=ts) | browser clients |
| `api/.../http/handlers/openapi.yaml` | `google-gnostic-openapi` | the Swagger UI at `/docs/` |

Two constraints: `protoc-gen-es` is resolved from `app/node_modules/.bin/`, so the frontend's npm
install must have happened first; and `buf.gen.yaml` overrides `go_package_prefix` for the whole
module, which is why generated Go lands under `.../inbound/grpc/gen` regardless of proto path.

**`make gen` rewrites `app/src/gen/`, so the frontend must be rebuilt after any proto change** —
`cd app && npm run lint && npm run build` is part of verifying a proto change, not an afterthought.

## 3. `enums.proto` — modified

Appended beside the existing `LoginProvider`. `messages-proto.md` adds two more enums to this same
file; the three additions are independent and can land in any order.

```proto
enum AiProvider {
  AI_PROVIDER_UNSPECIFIED = 0;
  AI_PROVIDER_OPENAI = 1;
  AI_PROVIDER_GOOGLE = 2;
  AI_PROVIDER_ANTHROPIC = 3;
}
```

**The three names map to AI-service URL path segments and are not free choices.** The FastAPI
routers mount at `/ai-service/v1/openai`, `/google`, `/anthropic` and reject `gemini` and `claude`.
The frontend labels them "OpenAI", "Gemini" and "Claude"; that label lives in the UI only
(`mock.ts`'s `PROVIDERS` already separates value from label for this reason).

`AI_PROVIDER_UNSPECIFIED` exists because lint demands a zero value, **not** because it is a usable
input. Every RPC that accepts a provider rejects it (`project-model.md` §3).

## 4. `project_service.proto` — new

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

## 5. Field-by-field rationale

### 5.1 `Project`

| Field | Note |
| --- | --- |
| `id` | `string`, not `bytes`. It is a UUID; the handler does `uuid.Parse` and returns `InvalidArgument` on failure |
| `title` | Provisional at creation, replaced once by the LLM, renameable after |
| `provider` | The project's **current** selection, restored when the Workspace reopens. Per-message provider is a different field on a different message (`messages-proto.md`) |
| `sandbox_id`, `preview_url` | Empty string when no sandbox has been made. proto3 has no null for scalars, and that is fine — the domain maps them to `""` deliberately rather than `*string` |
| `created_at`, `updated_at` | RFC3339 strings — §5.2 |

**`user_id` is absent, and its absence is load-bearing.** Ownership comes from the JWT via
`authAdapter.User(ctx)`. A `user_id` field on the request would be an authorisation hole with a
field name; on the response it would leak nothing useful but invite the frontend to trust it.

### 5.2 Timestamps are RFC3339 strings, not `google.protobuf.Timestamp`

Postgres generates the values (`DEFAULT NOW()`); the proto only carries them. A plain string keeps
the repo free of well-known-type imports and lets the browser do `new Date(s)` — no
`timestampDate()` helper from `@bufbuild/protobuf/wkt`, no `{seconds, nanos}` unwrapping.

The trade, stated so it is not lost: **the schema no longer enforces that these are instants**, so
the format is a convention the handler must hold up. Always `t.UTC().Format(time.RFC3339)`, never
`t.String()` — Go's default layout (`2006-01-02 15:04:05.999999999 -0700 MST`) is parsed
inconsistently by browsers. RFC3339 carries its own offset, so this survives the binary wire format
(§2.5).

The API sends an absolute instant and nothing else; "2 hours ago" is the browser's job.

### 5.3 `limit` / `token`, not `page_size` / `page_token`

Matches `ListUsersRequest` and the field names on `domain.Page[T]`. Consistency with the repo beats
consistency with Google's AIP when one convention is already generated into the TS client.

### 5.4 `PATCH` is the first non-GET/POST verb in this repo

> **⏸ Polish.** `RenameProject` is not in the MVP — nothing in the demo renames a project, and the
> title is the truncated first prompt. Leave the RPC out of the first `make gen` and add it later;
> adding an RPC to an existing service is additive and breaks nothing.

`google.api.http` and Vanguard both support it, but nothing here has exercised it. If transcoding
misbehaves, the fallback is `post: "/api/v1/projects/{id}/rename"` — the Connect path is unaffected
either way, so the frontend would not notice.

`RenameProject` returns the whole `Project` rather than nothing, so the client can reconcile without
a follow-up `GetProject`.

### 5.5 `CreateProject` takes the prompt but does not persist it

`first_prompt` is used **only** to derive the provisional title and to seed the LLM titling call. The
prompt is not written as a message row here — the client follows with `SendMessage` carrying the
same text (`messages-proto.md`).

Two calls, same string on the wire twice. That is deliberate: the two RPCs have different jobs, and
folding message creation into `ProjectService` would make it depend on `MessageRepository`.

## 6. Landing it

```bash
make plint                                   # clang-format the .proto files
make gen                                     # regenerate all four trees
cd api && go build ./... && go vet ./...     # will fail until the handler exists — expected
cd app && npm run lint && npm run build      # §2.6
```

Check the generated output before writing any Go:

```bash
grep -n "OPENAI\|ANTHROPIC" app/src/gen/api/v1/enums_pb.ts
```

**Expect `OPENAI = 1`, not `AI_PROVIDER_OPENAI = 1`.** If you see the long form, the enum was named
`AIProvider` and §2.3 applies — fix the proto name and regenerate rather than living with it.

Then confirm the four RPCs appear in `app/src/gen/api/v1/project_service_pb.ts` and in
`api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect/`, and that the REST paths landed
in `openapi.yaml`.

## 7. Out of scope

`RefreshSandbox` (`project-codebase-proto.md`) · everything message-shaped (`messages-proto.md`) ·
`DeleteProject` · a file-tree RPC · credits fields on `Project` · streaming · handler and service
implementation (`.planning/Backend/project-model.md`).
