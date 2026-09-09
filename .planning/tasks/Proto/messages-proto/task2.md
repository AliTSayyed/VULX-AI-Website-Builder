# Task 2 — `message_service.proto`

**Goal:** the two message RPCs on the wire.

Source: `.planning/Proto/messages-proto.md` §3, §4. Depends on this area's task 1 **and** on
`Proto/project-proto/task1.md` (`AiProvider`).

## Files

| File | Change |
| --- | --- |
| `proto/api/v1/message_service.proto` | **new** |

## The file

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

## Five things to get right

**The field is `project_id`, not `id`.** The path variable `{project_id}` must match a request field
name exactly. This differs from `GetProjectRequest.id` on purpose — different message, different
resource.

**`body : "*"` binds only `body`, `mode` and `provider`.** `project_id` comes from the URL and is not
sent twice.

**`SendMessageRequest.body` and the `body : "*"` annotation are unrelated.** One is the message text,
the other is HTTP. Adjacent and confusing on first read; kept aligned with `Message.body` as the
lesser evil.

**`assistant_message` may be unset.** proto3 message fields are nullable. The user message is
committed before the agent call and is not rolled back if it fails, so a thread can contain user
messages with no reply — the frontend must check presence, not compare against an empty object.

**No `updated_at` on the wire, no pagination on `ListMessages`, no `status` field.** All deliberate:
messages are immutable, threads are short, and a `status` shape depends on the poll-vs-stream
decision that has not been made.

## Run it

```bash
make plint && make gen
```

## Verifying

```bash
grep -n "listMessages\|sendMessage" app/src/gen/api/v1/message_service_pb.ts
grep -rn "MessageServiceHandler" api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect/ | head
grep -n "messages" api/internal/infrastructure/inbound/http/handlers/openapi.yaml
cd app && npm run lint && npm run build
```

Confirm the nested REST path landed as `/api/v1/projects/{project_id}/messages` — it is the first
route in this repo with a path variable followed by another segment, so it is worth an eye rather
than an assumption.

**`go build ./...` still fails.** Nothing implements `MessageServiceHandler` yet;
`tasks/Backend/messages-model/task5.md` does that.

## Done when

- Both methods appear in `app/src/gen/api/v1/message_service_pb.ts`.
- Both REST paths appear in `openapi.yaml` with `{project_id}` intact.
- `npm run build` passes.
