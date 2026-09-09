# Proto — Message Service Tasks

Two tasks. Context: `.planning/Proto/messages-proto.md` — effectively **all MVP**. The wire contract
is identical whether the server implements Chat, Build, or both; the phase difference lives in the Go
handler, not the proto.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | `MessageRole` + `ChatMode` in `enums.proto` | Small |
| 2 | `task2.md` | `message_service.proto` — 2 RPCs, 5 messages | Mechanical |

## Order

Task 1 before task 2. Both are additive to `enums.proto`, which
`Proto/project-proto/task1.md` also edits — the two areas are independent and one `make gen` can
cover everything.

Neither depends on `AiProvider` existing *first*, but `message_service.proto` **references** it, so
`Proto/project-proto/task1.md` must be done before this area's task 2 compiles.

## A note on Build vs Chat

`.planning/Proto/messages-proto.md` §5 says Build returns `Unimplemented` and the Build toggle should
be disabled. **That is superseded by the MVP** — Build is the mode the demo uses, and Chat is the one
that does not work. The proto is unaffected either way; both enum values ship.

## Verifying

`go build` on the API will fail until `tasks/Backend/messages-model/` lands the handler. Expected.
