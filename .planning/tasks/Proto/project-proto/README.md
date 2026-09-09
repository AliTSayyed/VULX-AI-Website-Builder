# Proto — Project Service Tasks

Two tasks. Context: `.planning/Proto/project-proto.md` — **MVP subset**: `RenameProject` is Polish
and is not written yet.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | `AiProvider` in `enums.proto`, regenerate, verify prefix stripping | Small, but §2.3's trap is here |
| 2 | `task2.md` | `project_service.proto` — 3 RPCs, 9 messages | Larger, mechanical |

## Order

**Task 1 before task 2.** Task 1 exists as a separate step for one reason: it proves the enum name
generates correctly in TypeScript *before* a whole service file depends on it. `AiProvider` versus
`AIProvider` fails silently and is annoying to change once the TS client is wired.

`Proto/messages-proto/task1.md` also edits `enums.proto`. The two are additive and independent —
either order, and one `make gen` can cover both.

## Not in these tasks

`RenameProject` and its two messages, `PATCH`, and everything in `project-codebase-proto.md`
(`RefreshSandbox`). All Polish.

## Verifying

Both tasks end with `make gen` and an inspection of generated output. Nothing compiles on the Go side
yet — `go build` will fail until `tasks/Backend/project-model/` lands the handler, and that is
expected, not a problem to debug.
