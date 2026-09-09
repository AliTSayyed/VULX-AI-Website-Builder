# Backend — AI Service Client Tasks

Five tasks. Context: `.planning/Backend/ai-service-client.md` — **MVP subset**: two methods, not
five.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | Delete the demo Temporal workflow | Touches 4 files; **do it first** |
| 2 | `task2.md` | `ai_service.go` — kill the 30s ceiling, add the shared `do` helper | The single most important change in the MVP |
| 3 | `task3.md` | `sandbox.go` — rewrite `CreateSandbox` | Fixes a live panic |
| 4 | `task4.md` | `llm.go` — `RunCodeAgent`; delete `openai_agent.go` | Mechanical once 2 is done |
| 5 | `task5.md` | Verification against the running AI service | — |

## Order

**Task 1 before task 3**, non-negotiable: `user_workflow.go` calls `CreateSandbox`, so changing that
signature while the workflow exists breaks the build in a confusing place.

Tasks 2 → 3 → 4 in order; 3 and 4 both depend on the `do` helper.

## Depends on

Nothing. Independent of `tasks/Proto/` and `tasks/Backend/project-model/` — build them in parallel
if you like.

## Why this area is not optional

`http.Client{Timeout: 30 * time.Second}` is a hard ceiling that **overrides any longer context
deadline**. A code-agent run takes minutes. Until task 2 lands, Build cannot succeed no matter what
else is correct.

## Not in these tasks

`Query` (Chat + LLM titles) · `WriteFiles` / `RunCommand` (codebase replay) · generated DTOs ·
meaningful HTTP status codes from the Python side. All Polish.

## The trap in task 1

Deleting the workflow leaves `aiservice` as an **unused local variable in `app.go`**, which is a Go
compile error, not a warning. Task 1 removes that line too; `tasks/Backend/messages-model/task5.md`
re-adds it when something finally consumes the client.
