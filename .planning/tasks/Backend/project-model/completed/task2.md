# Task 2 — Domain types

**Goal:** `domain.AIProvider` and `domain.Project`, following `domain/user.go`'s conventions exactly.

Source: `.planning/Backend/project-model.md` §4.

## Files

| File | Change |
| --- | --- |
| `api/internal/domain/ai_provider.go` | **new** |
| `api/internal/domain/project.go` | **new** |

## 1. `domain/ai_provider.go`

The `LoginProvider` pattern from `domain/login.go`, verbatim: `int` enum with `iota`, a `String()`,
and a `ParseX`.

```go
package domain

import "strings"

type AIProvider int

const (
	AIProviderUnspecified AIProvider = iota
	AIProviderOpenAI
	AIProviderGoogle
	AIProviderAnthropic
)

func (a AIProvider) String() string {
	switch a {
	case AIProviderOpenAI:
		return "openai"
	case AIProviderGoogle:
		return "google"
	case AIProviderAnthropic:
		return "anthropic"
	default:
		return "unspecified"
	}
}

func ParseAIProvider(s string) AIProvider {
	switch strings.ToLower(s) {
	case "openai":
		return AIProviderOpenAI
	case "google":
		return AIProviderGoogle
	case "anthropic":
		return AIProviderAnthropic
	default:
		return AIProviderUnspecified
	}
}
```

**The Go type is `AIProvider`; the proto enum is `AiProvider`.** That is not a typo — Go capitalises
initialisms, and the proto spelling is forced by `protoc-gen-es`'s prefix stripping. The handler maps
between them (task 5).

**The three strings are AI-service URL path segments** and are matched exactly by the FastAPI
routers. Never `gemini`, never `claude`.

## 2. `domain/project.go`

```go
type Project struct {
	id         uuid.UUID
	userID     uuid.UUID
	title      string
	provider   AIProvider
	sandboxID  string    // "" until a sandbox has been created
	previewURL string    // "" until a sandbox has been created
	createdAt  time.Time
	updatedAt  time.Time
}
```

Sentinels at the top of the file, as `user.go` does:

```go
var (
	ErrProjectTitleEmpty          = NewError(ErrorTypeInvalid, errors.New("project title cannot be empty"))
	ErrProjectUserEmpty           = NewError(ErrorTypeInvalid, errors.New("project user id cannot be empty"))
	ErrProjectProviderUnspecified = NewError(ErrorTypeInvalid, errors.New("project provider cannot be unspecified"))
)
```

- `NewProject(userID uuid.UUID, title string, provider AIProvider) (*Project, error)` — trims the
  title, rejects `uuid.Nil`, an empty title, and `AIProviderUnspecified`. Sets `id: uuid.New()`.
  **Does not set timestamps** — Postgres owns those.
- `RestoreProject(id, userID uuid.UUID, title, provider, sandboxID, previewURL string, createdAt, updatedAt time.Time) *Project`
  — takes `provider` as a **string** and calls `ParseAIProvider`, mirroring
  `RestoreUserFromProvider`. No validation: data already in the database is trusted.
- Nil-safe accessors: `ID`, `UserID`, `Title`, `Provider`, `SandboxID`, `PreviewURL`, `CreatedAt`,
  `UpdatedAt`, plus `HasSandbox() bool` returning `p != nil && p.sandboxID != ""`.

## Three things to get right

**`UserID()` is not optional.** The ownership check in task 4 is a comparison against it. Omitting
the accessor makes that check impossible to write.

**`RestoreProject` takes the provider as a string, not an `AIProvider`.** This keeps the persistence
row struct (task 3) free of domain enum types.

**Nullable columns map to `""`, not `*string`.** `HasSandbox()` reads better at the call site than a
nil check, and it is the only question anyone asks of those two fields.

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

Nothing calls these yet — that is fine. Unused *functions* are legal in Go.

Sanity-check the enum round-trip mentally: `ParseAIProvider(AIProviderAnthropic.String())` must
return `AIProviderAnthropic`, and `ParseAIProvider("gemini")` must return `AIProviderUnspecified`.

## Done when

- `go build ./... && go vet ./...` clean.
- Both files follow `user.go`: private fields, nil-safe accessors, `NewX` validates, `RestoreX` does
  not.
