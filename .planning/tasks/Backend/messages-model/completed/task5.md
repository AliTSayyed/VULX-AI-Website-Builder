# Task 5 — Handler and wiring

**Goal:** the two message RPCs reachable, and the AI client finally constructed.

Source: `.planning/Backend/messages-model.md` §10.

## Files

| File | Change |
| --- | --- |
| `.../inbound/handlers/message.go` | **new** |
| `.../application/app.go` | wire it, and re-add `aiservice` |

## 1. The handler

```go
type MessageServiceHandler struct {
	apiv1connect.UnimplementedMessageServiceHandler

	messageService *services.MessageService
	authAdapter    *authAdapter.HTTPAuthAdapter
}
```

Same four steps as every handler: `authAdapter.User(ctx)` (**omit and the RPC is public**),
`uuid.Parse` the project id, call the service with `user.ID()`, map errors with
`errorAdapter.ToConnectError`. No superuser check — this is product surface.

- `ListMessages` → `messageService.List(ctx, user.ID(), projectID)` → `[]*apiv1.Message`.
- `SendMessage` → map the proto `mode` and `provider` to domain enums, call
  `messageService.Send(...)`, return both messages.

### `SendMessage`'s error path is not the usual one

`Send` returns `(userMsg, asstMsg, err)`. On a Chat-mode call it returns a **non-nil `userMsg`
alongside a non-nil error** — the message was persisted, then the flow stopped. Do not treat that as
a partial success: return the error. The user message is already in the database and the client will
see it on the next `ListMessages`.

```go
userMsg, asstMsg, err := h.messageService.Send(...)
if err != nil {
	return nil, errorAdapter.ToConnectError(err)
}
```

### Converters

`messageToProto(*domain.Message)` mirroring `projectToProto`, with
`CreatedAt: m.CreatedAt().UTC().Format(time.RFC3339)` — never `t.String()`.

Total mapping helpers both ways for `MessageRole` and `ChatMode`, defaulting to unspecified.
**`providerToProto` / `providerFromProto` already exist** in `handlers/project.go` in this same
`handlers` package — reuse them, do not redeclare.

The domain enums and proto enums stay separate because **the domain must not import generated
code**.

## 2. `app.go`

Re-add what `tasks/Backend/ai-service-client/task1.md` removed, now that something consumes it:

```go
// DI
aiservice := aiservice.NewAIService(cfg.AIServiceUrl)

// persistance
messageRepo := postgres.NewMessageRepository(db)

// business logic
messageService := services.NewMessageService(
	messageRepo, projectRepo, projectService,
	aiservice, // SandboxCreator
	aiservice, // CodeAgentRunner
)

// handlers
messageServiceHandler := handlers.NewMessageServiceHandler(messageService, connectAuthAdapter)
```

and one vanguard entry:

```go
vanguard.NewService(apiv1connect.NewMessageServiceHandler(messageServiceHandler, interceptor)),
```

**Passing `aiservice` twice is correct**, not a copy-paste slip — one concrete type satisfying two
narrow interfaces is ordinary interface segregation. A comment on each line stops the next reader
from "fixing" it.

**The shadowing trap:** the local `services := []*vanguard.Service{...}` shadows the imported
`services` package. Every `services.NewXService(...)` must appear above it.

## Verifying

```bash
cd api && go build ./... && go vet ./...
make
docker compose logs api | tail -20
```

Then a Chat call, which is the cheapest end-to-end proof the wiring is right without spending an
agent run:

```bash
curl -i -b "jwt=$JWT" -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' \
  -d '{"body":"hello","mode":"CHAT_MODE_CHAT","provider":"AI_PROVIDER_ANTHROPIC"}'
```

Expect **501** (`Unimplemented`) — and then confirm the user message was still persisted:

```bash
curl -s -b "jwt=$JWT" $API/api/v1/projects/$PID/messages | jq '.messages | length'   # 1
```

That single check proves ownership, validation, the repository, the CTE and the handler all work,
for free. The Build path is task 6.

## Done when

- `go build ./... && go vet ./...` clean, API starts.
- A Chat send returns 501 **and** leaves one message behind.
- `ListMessages` returns it with `created_at` in RFC3339.
