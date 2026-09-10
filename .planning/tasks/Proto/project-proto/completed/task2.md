# Task 2 — `project_service.proto`

**Goal:** the three project RPCs the MVP needs, on the wire.

Source: `.planning/Proto/project-proto.md` §4, §5. Depends on task 1.

## Files

| File | Change |
| --- | --- |
| `proto/api/v1/project_service.proto` | **new** |

## The file

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
```

## What is deliberately absent

**`RenameProject`, `RenameProjectRequest`, `RenameProjectResponse`** — Polish. Field numbers 1–7 on
`Project` are already claimed; adding an RPC later is additive and breaks nothing.

**`user_id` — on no message, request or response.** Ownership comes from the JWT via
`authAdapter.User(ctx)`. Accepting it from the client would be an authorisation hole wearing a field
name.

## Four things to get right

**Path variables must match request field names exactly.** `get: "/api/v1/projects/{id}"` requires
`GetProjectRequest.id`. A mismatch is a routing failure at runtime, not a compile error.

**`body : "*"` on `CreateProject` means "every field not bound to the path"** — here, both of them,
since nothing is in the URL.

**Timestamps are RFC3339 `string`s, not `google.protobuf.Timestamp`.** Postgres generates the values;
the proto only carries them. The handler will format with `t.UTC().Format(time.RFC3339)` — never
`t.String()`, whose Go layout browsers parse inconsistently.

**`limit`/`token`, not `page_size`/`page_token`** — matches `ListUsersRequest` and `domain.Page[T]`.

## Run it

```bash
make plint && make gen
```

## Verifying

```bash
# TS client has all three methods
grep -n "listProjects\|getProject\|createProject" app/src/gen/api/v1/project_service_pb.ts

# Go Connect handler interface exists
grep -rn "ProjectServiceHandler" api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect/ | head

# REST paths transcoded into the OpenAPI doc
grep -n "/api/v1/projects" api/internal/infrastructure/inbound/http/handlers/openapi.yaml

cd app && npm run lint && npm run build
```

**`cd api && go build ./...` will still fail** — nothing implements `ProjectServiceHandler` yet.
That is expected and is `tasks/Backend/project-model/task5.md`'s job. Do not try to fix it here.

## Done when

- `app/src/gen/api/v1/project_service_pb.ts` exists with three methods.
- The three REST paths appear in `openapi.yaml`.
- `npm run build` passes.
- Generated files are **not** hand-edited — if something looks wrong, fix the `.proto` and rerun.
