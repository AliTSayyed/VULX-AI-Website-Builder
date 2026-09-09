# Task 1 — Delete the demo Temporal workflow

**Goal:** remove `user_workflow.go` without breaking the build, and keep the Temporal connection for
later.

Source: `.planning/Backend/ai-service-client.md` §7.1.

## Why delete rather than update

It is non-functional (`UseLlm` is empty), `StartUserWorkflow` blocks on `workflowRun.Get` — making
the "async" orchestration synchronous — its `CreateSandbox` activity nil-derefs, and its
`StartToCloseTimeout: 10 * time.Second` is shorter than sandbox creation takes. Every one of those is
a pattern worth not copying. It also has to change regardless, because task 3 changes the signature
it calls.

**`StartUserWorkflow` has no caller.** It is stored on `UserService` and never invoked — verified.
Nothing behavioural is lost.

## Files

| File | Change |
| --- | --- |
| `.../outbound/temporal/user_workflow.go` | **delete** |
| `.../outbound/temporal/temporal.go` | `RegisterWorkers` loses its parameter |
| `.../application/services/user_service.go` | drop the `UserWorkflowService` port |
| `.../application/app.go` | drop four lines |

## 1. Delete the file

```bash
rm api/internal/infrastructure/outbound/temporal/user_workflow.go
```

## 2. `temporal.go`

```go
func (temporal *Temporal) RegisterWorkers() *Temporal {
	// No workflows or activities yet. build-orchestration.md registers the real
	// Build pipeline here. The worker is still created so StopWorkers has something
	// to stop on shutdown.
	userWorker := worker.New(temporal.Client, "user-workflow", worker.Options{})
	...unchanged: go userWorker.Run(...), temporal.UserWorker = userWorker, return temporal
}
```

**Do not skip creating the worker.** `StopWorkers()` calls `temporal.UserWorker.Stop()`, and a nil
worker panics during graceful shutdown — a failure that only shows up on Ctrl-C, which is exactly
when nobody is looking.

Leave the task-queue name as-is; renaming it belongs with the real workflow.

## 3. `user_service.go`

Delete:

```go
type UserWorkflowService interface {
	StartUserWorkflow(ctx context.Context) error
}
```

Remove the `userWorkflow` struct field and the second parameter of `NewUserService`, so it becomes
`NewUserService(userRepo UserRepository) *UserService`.

Watch for a now-unused `context` import if nothing else in the file uses it — Go rejects unused
imports.

## 4. `app.go` — including the trap

Delete:

```go
userWorkflow := temporal.NewUserWorkflow(temporalService, aiservice)
temporalService.RegisterWorkers(userWorkflow)
```

Replace with:

```go
temporalService.RegisterWorkers()
```

and change `services.NewUserService(userRepo, userWorkflow)` → `services.NewUserService(userRepo)`.

**Now the trap.** `aiservice` was only used by `NewUserWorkflow`. Left in place it is an **unused
local variable — a compile error, not a warning**. So also delete:

```go
aiservice := aiservice.NewAIService(cfg.AIServiceUrl)
```

and its import. `tasks/Backend/messages-model/task5.md` re-adds both when `MessageService` finally
consumes the client.

Do **not** paper over it with `_ = aiservice`; that is a placeholder someone will forget.

## Verifying

```bash
cd api && go build ./... && go vet ./...
grep -rn "UserWorkflow\|StartUserWorkflow" api/ --include=*.go | grep -v "/gen/"   # expect nothing
make nuke && make
docker compose logs api | grep -i "temporal\|workers"
```

Expect "Connected to Temporal service" and "Workers successfully registered". The API must start
clean.

**Then test shutdown**, since that is where the nil-worker bug hides:

```bash
docker compose stop api
docker compose logs api | tail -20     # "API Shutting Down", no panic
```

## Done when

- `go build ./... && go vet ./...` clean.
- No reference to `UserWorkflow` remains outside generated code.
- The API starts *and* stops cleanly.
