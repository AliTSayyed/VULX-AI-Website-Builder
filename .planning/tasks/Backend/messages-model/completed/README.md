# Backend — Messages Model Tasks

Six tasks. Context: `.planning/Backend/messages-model.md` — **MVP subset**, and note that document's
Chat/Build split is **inverted** by the MVP: Build ships, Chat does not.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | Migration `00005_add_messages.sql` | Isolated |
| 2 | `task2.md` | `domain/message.go` — entity + two enums | Pure Go |
| 3 | `task3.md` | `message_repository.go` — the CTE | The CTE is the subtle part |
| 4 | `task4.md` | `message_service.go` — **the synchronous Build flow** | The heart of the MVP |
| 5 | `task5.md` | `handlers/message.go` + `app.go` | Where everything finally connects |
| 6 | `task6.md` | End-to-end: prompt → sandbox → rendered site | The demo |

## Order

Strictly 1 → 6.

## Depends on

- `tasks/Proto/messages-proto/` — both tasks, for the generated handler interface.
- `tasks/Backend/project-model/` — **all six**. `MessageService` composes `*ProjectService` for the
  ownership check and `ProjectRepository` for `UpdateSandbox`.
- `tasks/Backend/ai-service-client/` — **all five**. Task 5 here re-adds the `aiservice` construction
  that `ai-service-client/task1.md` removed.

This is the last area. Nothing else depends on it.

## What is different from the source document

`.planning/Backend/messages-model.md` §8 (Chat) is **Polish** — it needs `Query`, which the MVP
client does not implement. §9 (Build) is **MVP**, implemented synchronously per §9.0 of that
document's phase split. §9.4's "return `Unimplemented` and disable the Build toggle" does not apply.

## The three bugs most likely to slip through

1. **The CTE not bumping `projects.updated_at`** — the sidebar silently stops reordering. Task 3.
2. **A second sandbox being created on the second message** — losing the first one's work. Task 4.
3. **A missing ownership check or `authAdapter.User(ctx)`** — the same two leaks as project-model.
   Tasks 4 and 5.

Task 6 tests all three.
