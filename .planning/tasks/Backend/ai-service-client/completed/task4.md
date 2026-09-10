# Task 4 — `llm.go`: `RunCodeAgent`

**Goal:** the call that makes the agent write code. Replaces the `CallAI` stub.

Source: `.planning/Backend/ai-service-client.md` §4. Depends on task 2.

## Files

| File | Change |
| --- | --- |
| `.../outbound/ai_service/llm.go` | **new** |
| `.../outbound/ai_service/openai_agent.go` | **delete** |

`openai_agent.go` contains only `func (a *AIService) CallAI() error { return nil }`. Its name
describes one provider while serving all three; `llm.go` replaces it.

## The code

`CodeAgentResult` goes in **`api/internal/domain/code_agent_result.go`**, for the same reason
`SandboxInfo` does (task 3): the `CodeAgentRunner` port lives in `application/services`, and a port
referencing an infrastructure package would point the dependency arrow outward.

```go
// domain/code_agent_result.go — a value object describing an external result, not an entity.
type CodeAgentResult struct {
	Summary  string            `json:"summary"`
	Commands []string          `json:"commands"`
	Files    map[string]string `json:"files"`
}
```

```go
// outbound/ai_service/llm.go
type codeAgentRequest struct {
	Message string `json:"message"`
}

func (a *AIService) RunCodeAgent(
	ctx context.Context, provider domain.AIProvider, sandboxID, message string,
) (*domain.CodeAgentResult, error) {
	if provider == domain.AIProviderUnspecified {
		return nil, domain.NewError(domain.ErrorTypeInvalid,
			errors.New("ai provider cannot be unspecified"))
	}
	if sandboxID == "" {
		return nil, domain.NewError(domain.ErrorTypeInvalid,
			errors.New("sandbox id cannot be empty"))
	}

	// No timeout of its own: an agent run is minutes and the CALLER owns the deadline.
	// The 15-minute client backstop in ai_service.go is the only hard limit.
	path := fmt.Sprintf("/%s/%s/code", provider.String(), url.PathEscape(sandboxID))

	var out domain.CodeAgentResult
	if err := a.do(ctx, http.MethodPost, path, codeAgentRequest{Message: message}, &out); err != nil {
		return nil, domain.WrapError("ai service run code agent", err)
	}

	utils.Logger.Info("code agent finished",
		"sandbox_id", sandboxID, "provider", provider.String(),
		"files", len(out.Files), "commands", len(out.Commands))

	return &out, nil
}
```

## Five notes

**`provider.String()` is the URL path segment.** `openai` / `google` / `anthropic`, exactly. The
FastAPI routers reject `gemini` and `claude`. This is why `domain.AIProvider.String()` returns those
literals and not display names.

**The unspecified guard is not redundant** with the service-layer validation. `AIProviderUnspecified`
stringifies to `"unspecified"`, which would produce `POST /unspecified/<id>/code` — a 404 from
FastAPI that reads like a routing bug rather than a bad argument.

**`url.PathEscape(sandboxID)`** — the id comes from an external service. Escaping is free; debugging a
malformed path is not.

**`CodeAgentResult` omits `human_message`** from the response — it is the request echoed back and the
caller already has it.

**The MVP reads `Summary` and discards `Files` and `Commands`.** They are decoded anyway so the
struct is complete when `project-codebase-model.md` starts persisting them. `Summary` comes from the
model's parsed output, not from `CodeAgentCallBack`, which is why the MVP works before
`agent-result-capture.md`.

## Not in this task

`Query` — needed only by Chat mode and LLM titles, both Polish.

## Verifying

```bash
rm api/internal/infrastructure/outbound/ai_service/openai_agent.go
cd api && go build ./... && go vet ./...
grep -rn "CallAI" api/ --include=*.go        # expect nothing
```

Exercise the underlying route directly, since nothing calls the client yet:

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .id)
time curl -s -X POST http://localhost:9999/ai-service/v1/anthropic/$SB/code \
  -H 'content-type: application/json' \
  -d '{"message":"build a landing page for a coffee shop"}' | jq '{summary, files: (.files|keys)}'
```

**Note the elapsed time.** If it exceeds 30 seconds — it will — that is the concrete proof that task
2's timeout change was required.

## Done when

- `RunCodeAgent` exists, `openai_agent.go` is gone, no reference to `CallAI` remains.
- The direct `curl` returns a `summary` and renders in the sandbox URL.
- `go build ./... && go vet ./...` clean.
