# Message Service — Proto

> The contract for the chat thread. This document is **authoritative for
> `proto/api/v1/message_service.proto` and the `MessageRole` / `ChatMode` additions to
> `enums.proto`** — when it and `.planning/Backend/messages-model.md` disagree about the wire, this
> wins.
>
> Conventions — buf lint rules, field-number discipline, REST annotation mechanics, `make gen`
> behaviour — are in `project-proto.md` §2 and are not restated here. `AiProvider` is defined there
> (§3) and reused, not redeclared.
>
> Verified against the code on `feature/messages` (2026-09-08).

## ⏸ Phase split — MVP vs Polish

**Effectively all MVP.** The wire contract is the same whether the server implements Chat, Build, or
both — the phase difference is entirely in the Go handler, not the proto.

| § | Content | Phase |
| --- | --- | --- |
| §2 | `MessageRole`, `ChatMode` | **MVP** — both enums, both values each |
| §3 | `message_service.proto`, both RPCs, all five messages | **MVP** |
| §4 | the decisions behind them | **MVP** |
| §5 | Build-is-unimplemented | **Superseded** — see the note in that section |

`CHAT_MODE_CHAT` ships in the enum even though the MVP never sends it. Enum values are permanent
once numbered (`project-proto.md` §2.1), so defining both now costs nothing and avoids renumbering
later.

## 1. What is being added

| File | Change |
| --- | --- |
| `proto/api/v1/enums.proto` | append `MessageRole` and `ChatMode` |
| `proto/api/v1/message_service.proto` | **new** — `MessageService`, 2 RPCs, 5 messages |

## 2. `enums.proto` — modified

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

Both names snake-case cleanly (`MESSAGE_ROLE_`, `CHAT_MODE_`), so `protoc-gen-es` strips the prefix
and the frontend writes `MessageRole.ASSISTANT` and `ChatMode.BUILD` — `project-proto.md` §2.2. The
consecutive-capitals trap in §2.3 does not apply to either.

`CHAT_MODE_CHAT` reads as a stutter and is unavoidable: the prefix rule applies to every value,
including the one that shares the enum's name.

**Both zero values are rejected inputs, not defaults.** `SendMessage` returns `InvalidArgument` for
either — the composer always has a concrete mode and provider selected, so an unspecified value
means a client bug and defaulting would hide it.

## 3. `message_service.proto` — new

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

## 4. Decisions

### 4.1 A separate service, not more RPCs on `ProjectService`

Messages are their own resource with their own lifetime, and the split has a concrete payoff on the
client: the generated `MessageService` client gets its own React Query cache keyed on the thread, so
sending a message does not invalidate the project list, and listing projects does not touch threads.

Cost: a second `vanguard.NewService` entry and a second handler. Worth it here in a way it was not
for `RefreshSandbox` (`project-codebase-proto.md` §2), which is a single method operating on a
project.

### 4.2 `project_id` is a path variable on both RPCs

`/api/v1/projects/{project_id}/messages` — messages are nested under their project in the URL
because they cannot exist without one, and the ownership check runs against the project either way.

Note the field is `project_id`, not `id`: `project-proto.md` §2.4 requires the path variable to
match a request field name exactly. On `SendMessage`, `body : "*"` then binds only `body`, `mode`
and `provider` from the JSON — `project_id` is not sent twice.

There is an unfortunate collision worth knowing: `SendMessageRequest.body` is the message text, and
`body : "*"` is the HTTP annotation. They are unrelated. Renaming the field to `text` would avoid
the confusion but diverge from `Message.body`; keeping them aligned is the lesser evil.

### 4.3 `SendMessage` returns both messages

The client gets server-assigned ids and timestamps for the **user's own** message rather than
inventing them, so an optimistic append can be reconciled instead of duplicated.

`assistant_message` is unset when the reply failed — proto3 message fields are nullable, so the
frontend checks presence rather than comparing to an empty object. That case is real: the user
message is committed *before* the AI call and is not rolled back if it fails
(`messages-model.md` §8.2). **A thread can contain user messages with no reply, and the frontend
must render it without assuming messages pair up.**

### 4.4 `ListMessages` is unpaginated

No `limit`, no `token`. A thread is a handful of messages and the Workspace renders all of them;
`cursor.go`'s keyset codec is built for `DESC` listing, not `ASC` threads, so a cursor here would be
new machinery with no caller.

The consequence is stated rather than discovered later: a long thread is one large response, and
Build summaries can be verbose. The index (`project_id, created_at ASC, id ASC`) already supports a
cursor when it is wanted; adding `limit`/`token` later is a backward-compatible field addition.

### 4.5 No `updated_at` on the wire

A message is immutable — nothing edits, deletes or regenerates one. Exposing a field that never
differs from `created_at` would invite the frontend to render it. The column exists in Postgres for
consistency with the other tables; it just does not cross.

### 4.6 `mode` and `provider` are on the message, not the project

Both are re-read from the composer on every send: a user asks a Chat question mid-Build thread, or
switches models between turns. `Message.provider` is the **historical record** of what produced that
message; `Project.provider` (`project-proto.md` §5.1) is the **current selection** restored when the
Workspace reopens. Two fields, two questions.

### 4.7 What is deliberately absent from `Message`

| Not a field | Why |
| --- | --- |
| `files_written` | Paths and contents live in `project_codebases`. The thread renders the agent's prose summary; a paths array would be a second, weaker copy |
| `status` | Build writes the user message before the agent finishes, so a pending/complete/failed state will be needed — but its shape depends on the poll-vs-stream decision in `build-orchestration.md`. Adding it now would pre-commit that |
| `token_count` / `cost` | `credits.md` decides the unit |
| `parent_id` | A project has one linear thread |

Each is a backward-compatible addition later: new field numbers, nothing renumbered.

## 5. Build mode is on the wire before it works

> **⏸ Superseded by the MVP.** The MVP implements Build synchronously
> (`messages-model.md` §9.0), so it is **Chat** that returns nothing useful, not Build. The Build
> toggle is the one the demo uses; the Chat toggle is the one to disable.
>
> This section stays correct for any intermediate state where the AI client is not yet wired.

`CHAT_MODE_BUILD` is a valid enum value and `SendMessage` accepts it — then persists the user
message and returns `CodeUnimplemented` (`messages-model.md` §9.4). The contract, the enum and the
validation are all real and testable; only execution is missing.

**The frontend's Build toggle should be visibly disabled until `build-orchestration.md` lands.** An
enabled control that always errors is worse than an absent one.

## 6. Landing it

```bash
make plint && make gen
cd api && go build ./... && go vet ./...     # fails until the handler exists — expected
cd app && npm run lint && npm run build
```

```bash
curl -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' \
  -d '{"body":"what goes in a three-tier pricing page?","mode":"CHAT_MODE_CHAT","provider":"AI_PROVIDER_ANTHROPIC"}'

curl -b "jwt=$JWT" https://local.api.vulx.ai/api/v1/projects/$PID/messages
```

Enums are spelled as full proto names in REST because the OpenAPI plugin runs with
`enum_type=string` (`project-proto.md` §2.5) — `"CHAT_MODE_CHAT"`, not `"chat"` and not `1`.

Four checks:

1. **Prefix stripping.** `grep -n "ASSISTANT\|BUILD" app/src/gen/api/v1/enums_pb.ts` — expect
   `ASSISTANT = 2` and `BUILD = 2`, not the long forms.
2. **`project_id` is not required in the body.** Send the `curl` above without it; the path variable
   must supply it (§4.2).
3. **`assistant_message` is genuinely optional.** Stop the `ai-service` container, send, and confirm
   the client sees the field unset rather than an empty `Message` — this is the case the thread UI
   has to handle.
4. **`CHAT_MODE_BUILD` returns 501**, and the user message is still persisted (§5).

## 7. Out of scope

The `messages` table, repository, service and handler — all `.planning/Backend/messages-model.md` ·
Build execution (`build-orchestration.md`) · a `status` field (§4.7) · conversation history in the
AI-service request (`ai-service-client.md`) · streaming · editing or deleting a message · thread
pagination (§4.4).
