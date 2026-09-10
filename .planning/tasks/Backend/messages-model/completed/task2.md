# Task 2 — `domain/message.go`

**Goal:** the `Message` entity and the two enums it carries.

Source: `.planning/Backend/messages-model.md` §3.

## Files

| File | Change |
| --- | --- |
| `api/internal/domain/message.go` | **new** |

## The entity

```go
type Message struct {
	id        uuid.UUID
	projectID uuid.UUID
	role      MessageRole
	mode      ChatMode
	provider  AIProvider
	body      string
	createdAt time.Time
	updatedAt time.Time
}
```

Sentinels at the top of the file, as `user.go` does: `ErrMessageBodyEmpty`,
`ErrMessageProjectEmpty`, `ErrMessageRoleUnspecified`, `ErrMessageModeUnspecified` — all
`ErrorTypeInvalid`.

- `NewMessage(projectID uuid.UUID, role MessageRole, mode ChatMode, provider AIProvider, body string) (*Message, error)`
  — rejects `uuid.Nil`, an empty body after trimming, and any unspecified enum. Sets
  `id: uuid.New()`, no timestamps (Postgres owns those).
- `RestoreMessage(id, projectID uuid.UUID, role, mode, provider, body string, createdAt, updatedAt time.Time) *Message`
  — takes the **three enums as strings** and parses them, mirroring `RestoreUserFromProvider`. That
  keeps the row struct in task 3 free of domain enum types.
- Nil-safe accessors: `ID`, `ProjectID`, `Role`, `Mode`, `Provider`, `Body`, `CreatedAt`,
  `UpdatedAt`.

## The two enums, same file

The `LoginProvider` pattern verbatim — `int` with `iota`, a `String()`, and a `ParseX`, including an
unspecified zero value whose `String()` is `"unspecified"`.

```go
type MessageRole int

const (
	MessageRoleUnspecified MessageRole = iota
	MessageRoleUser
	MessageRoleAssistant
)
// String(): "user", "assistant", default "unspecified"
// ParseMessageRole(s string) MessageRole

type ChatMode int

const (
	ChatModeUnspecified ChatMode = iota
	ChatModeChat
	ChatModeBuild
)
// String(): "chat", "build", default "unspecified"
// ParseChatMode(s string) ChatMode
```

Those strings are what land in the `VARCHAR` columns, so they must round-trip:
`ParseChatMode(ChatModeBuild.String()) == ChatModeBuild`.

**`AIProvider` is not redeclared** — `tasks/Backend/project-model/task2.md` already created
`domain/ai_provider.go`, and its `String()` doubles as the AI-service URL path segment.

**`ChatModeChat` exists even though the MVP never uses it.** The enum is a complete type; leaving a
hole would mean a migration and a proto renumber when Chat mode lands.

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

Nothing calls these yet. Read-check the round-trips: every `String()` value must be parseable back to
the same constant, and every unknown string must map to the unspecified zero value rather than
panicking or defaulting to a real value.

## Done when

- `go build ./... && go vet ./...` clean.
- Private fields, nil-safe accessors, `NewMessage` validates, `RestoreMessage` does not.
