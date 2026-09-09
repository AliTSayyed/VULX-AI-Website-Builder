# Task 4 — `message_service.go` and the synchronous Build flow

**Goal:** the heart of the MVP. A Build message creates a sandbox if needed, runs the agent, and
records the summary.

Source: `.planning/Backend/messages-model.md` §9.0 (the phase split), §7, §8.2, §8.3.

## Files

| File | Change |
| --- | --- |
| `api/internal/application/services/message_service.go` | **new** |

## 1. Ports

Declared here, in the consumer's file. **They use `domain` types and primitives only** — a port
referencing `infrastructure/outbound/ai_service` would point the dependency arrow outward, which is
the one thing the architecture forbids.

```go
type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) (*domain.Message, error)
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Message, error)
}

type SandboxCreator interface {
	CreateSandbox(ctx context.Context) (*domain.SandboxInfo, error)
}

type CodeAgentRunner interface {
	RunCodeAgent(ctx context.Context, provider domain.AIProvider, sandboxID, message string) (*domain.CodeAgentResult, error)
}
```

**`ProjectRepository` is not redeclared** — it already exists in `project_service.go` in this same
`services` package.

`ChatResponder` is **not** declared: Chat mode is Polish.

```go
type MessageService struct {
	messageRepo    MessageRepository
	projectRepo    ProjectRepository
	projectService *ProjectService
	sandbox        SandboxCreator
	codeAgent      CodeAgentRunner
}
```

It composes `*ProjectService` — not `ProjectRepository` — for the ownership check, so that check
lives in exactly one place. It *also* takes `ProjectRepository` directly for the one write
`ProjectService` does not expose: `UpdateSandbox`.

## 2. `List`

```go
func (s *MessageService) List(ctx context.Context, userID, projectID uuid.UUID) ([]*domain.Message, error) {
	if _, err := s.projectService.Get(ctx, userID, projectID); err != nil {
		return nil, domain.WrapError("message service list", err)
	}
	messages, err := s.messageRepo.FindByProject(ctx, projectID)
	if err != nil {
		return nil, domain.WrapError("message service list", err)
	}
	return messages, nil
}
```

The ownership check runs first and its `NotFound` propagates unchanged.

## 3. `Send` — the MVP Build flow

```go
func (s *MessageService) Send(
	ctx context.Context, userID, projectID uuid.UUID,
	body string, mode domain.ChatMode, provider domain.AIProvider,
) (*domain.Message, *domain.Message, error) {

	// 1. ownership
	project, err := s.projectService.Get(ctx, userID, projectID)
	if err != nil {
		return nil, nil, domain.WrapError("message service send", err)
	}

	// 2. validation — §8.3
	body = strings.TrimSpace(body)
	if body == "" || len([]rune(body)) > 10000 { ... ErrorTypeInvalid }
	if mode == domain.ChatModeUnspecified || provider == domain.AIProviderUnspecified { ... }

	// 3. persist the user message (the CTE bumps projects.updated_at)
	userMsg, err := domain.NewMessage(projectID, domain.MessageRoleUser, mode, provider, body)
	// ... s.messageRepo.Create(ctx, userMsg)

	// 4. Chat is not implemented in the MVP
	if mode == domain.ChatModeChat {
		return userMsg, nil, domain.NewError(domain.ErrorTypeUnimplemented,
			errors.New("chat mode is not implemented yet"))
	}

	// 5. ensure a sandbox — ONLY if the project has none
	sandboxID := project.SandboxID()
	if sandboxID == "" {
		info, err := s.sandbox.CreateSandbox(ctx)
		if err != nil {
			return nil, nil, domain.WrapError("message service send: create sandbox", err)
		}
		if err := s.projectRepo.UpdateSandbox(ctx, projectID, info.ID, info.URL); err != nil {
			return nil, nil, domain.WrapError("message service send: store sandbox", err)
		}
		sandboxID = info.ID
	}

	// 6. run the agent — blocks for minutes
	result, err := s.codeAgent.RunCodeAgent(ctx, provider, sandboxID, body)
	if err != nil {
		return nil, nil, domain.WrapError("message service send: code agent", err)
	}

	// 7. persist the assistant message — Summary ONLY
	asstMsg, err := domain.NewMessage(projectID, domain.MessageRoleAssistant, mode, provider, result.Summary)
	// ... s.messageRepo.Create(ctx, asstMsg)

	return userMsg, asstMsg, nil
}
```

## Six things that must be exactly right

**1. The sandbox is created only when `SandboxID()` is empty.** Creating one per message would
abandon the previous sandbox *and all the code in it*, so the second prompt would build on an empty
template. This single `if` is what makes iterative building work.

**2. `UpdateSandbox` is called before the agent runs.** If the agent call fails, the sandbox is still
recorded and the next message reuses it instead of leaking another one.

**3. The user message is committed before the agent call and is never rolled back.** Deleting what
the user typed because a downstream service was unavailable loses their words to protect a symmetry
nobody asked for. The cost is that a thread can hold user messages with no reply — the frontend must
not assume they pair up.

**4. The assistant body is `result.Summary` and nothing else.** Never the agent's own account of
which files it wrote. `Files` and `Commands` are observed fact and belong in `project_codebases`,
which is Polish; discarding them here is correct, and `Summary` comes from the model's parsed output
rather than the callback, which is why this works before `agent-result-capture.md`.

**5. Unspecified enums are rejected, not defaulted.** The composer always has a concrete selection;
unspecified means a client bug and defaulting would hide it.

**6. No timeout is set around the agent call.** It inherits the request context. The client's
15-minute backstop is the only hard limit, by design.

## The known limitation to not accidentally "fix"

If the stored sandbox has **expired**, step 5 reuses its id and step 6 fails with an opaque error —
the AI service returns 500 for a dead sandbox and for an outage alike, so Go cannot tell them apart.
That is accepted for the MVP and mitigated by raising the sandbox timeout
(`tasks/AI-Service/sandbox-service/task1.md`). Do not add a liveness probe or a retry here; that is
`RefreshSandbox`'s job in Polish.

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

Read-check before wiring: is the sandbox creation inside `if sandboxID == ""`? Is `result.Summary`
the only thing written to the assistant message? Does every error path wrap with
`domain.WrapError`?

## Done when

- `go build ./... && go vet ./...` clean.
- The ports reference only `domain` types.
- A second Build on the same project cannot create a second sandbox.
