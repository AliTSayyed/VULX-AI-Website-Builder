# Backend — Project Model Tasks

Six tasks. Context: `.planning/Backend/project-model.md` — **MVP subset**. Authoritative wire
contract: `.planning/Proto/project-proto.md`.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | Migration `00004_add_projects.sql` | Isolated; `make nuke` to redo |
| 2 | `task2.md` | `domain/ai_provider.go` + `domain/project.go` | Pure Go, no deps |
| 3 | `task3.md` | `cursor.go` rename + `project_repository.go` | The pagination is the subtle part |
| 4 | `task4.md` | `application/services/project_service.go` | Ownership check lives here |
| 5 | `task5.md` | `handlers/project.go` + `app.go` wiring | First point where it compiles and runs |
| 6 | `task6.md` | End-to-end verification | Includes the checks a happy path misses |

## Order

Strictly 1 → 6. Each task leaves the tree compiling **except** tasks 1–4, which add code nothing
calls yet; `go build ./...` stays green throughout because unused *functions* are legal in Go (unused
*imports* and *local variables* are not — that is the only thing that will bite you mid-task).

## Depends on

`tasks/Proto/project-proto/` (both tasks) — the generated `apiv1.ProjectServiceHandler` interface
must exist before task 5.

**Not** on `tasks/Backend/ai-service-client/`. The MVP does not generate titles with an LLM, so this
area has no AI-service dependency at all and can be built in parallel with it.

## Not in these tasks

`RenameProject` · `UpdateTitle` · `UpdateProvider` · LLM title generation and its goroutine ·
`RefreshSandbox` · anything in `project-codebase-model.md`.

## The two bugs most likely to slip through

1. **A missing `authAdapter.User(ctx)` makes an RPC public.** There is no allowlist and no
   interceptor that will catch it. Task 5.
2. **A missing ownership comparison leaks other users' projects.** The repository deliberately does
   not filter by user. Task 4.

Task 6 tests both explicitly, because neither shows up on a happy path.
