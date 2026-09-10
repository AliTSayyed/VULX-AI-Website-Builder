# Project Codebase — Proto

> One RPC. This document is **authoritative for the `RefreshSandbox` addition to
> `proto/api/v1/project_service.proto`** — when it and
> `.planning/Backend/project-codebase-model.md` disagree about the wire, this wins.
>
> Conventions — buf lint rules, field-number discipline, REST annotation mechanics, `make gen`
> behaviour — are in `project-proto.md` §2 and are not restated here.
>
> Verified against the code on `feature/messages` (2026-09-08).
>
> ## ⏸ Phase: POLISH — not in the MVP
>
> `RefreshSandbox` exists to rebuild a sandbox from stored files. The MVP stores nothing and
> refreshes nothing (`project-codebase-model.md`), so this RPC has no purpose yet.
>
> **The MVP creates its sandbox inside `SendMessage`** on the project's first Build message — no
> client-visible RPC, no button.


## 1. What is being added

| File | Change |
| --- | --- |
| `proto/api/v1/project_service.proto` | append one RPC and two messages |

**No new proto file. No new service. No new enums.** The codebase itself — the JSONB file map and
the command log — never crosses the wire; it is server-side state that only the replay path reads.
The one thing the browser needs is a button that says "give me a working preview again", and that is
this RPC.

## 2. Why it lives on `ProjectService`

It operates on a project, it is a single RPC, and a separate `CodebaseService` would mean a second
`vanguard.NewService` entry, a second handler type, and a second generated TS client — for one
method.

The handler gains a dependency (`*services.CodebaseService`) rather than the service gaining a
sibling. That is `project-codebase-model.md` §12.1's job; the proto is unaffected by it.

## 3. The addition

```proto
  rpc RefreshSandbox(RefreshSandboxRequest) returns (RefreshSandboxResponse) {
    option (google.api.http) = {
      post : "/api/v1/projects/{id}/sandbox/refresh"
    };
  }
```

```proto
message RefreshSandboxRequest { string id = 1; }

message RefreshSandboxResponse {
  string sandbox_id = 1;
  string preview_url = 2;
}
```

## 4. Four decisions

### 4.1 `POST`, not `GET`

It creates a sandbox and costs money. `GET` would be wrong on semantics and actively harmful in
practice — browsers, proxies and prefetchers treat `GET` as safe to repeat, and every repeat here
provisions another E2B sandbox that nothing reaps.

### 4.2 No `body` declaration

`project-proto.md` §2.4: omit `body` when every request field is bound to a path variable. `id` is
the only field and it comes from the URL, so there is nothing left to put in a body.

The practical difference is in the `curl`: `-X POST` with **no** `-d`. Note that
`project-codebase-model.md` §14 currently shows `-d '{}'`, which is harmless but unnecessary —
this document is authoritative.

If a field is ever added — a template selector, say — add `body : "*"` at that time. One line.

### 4.3 The response is two strings, not a `Project`

Only `sandbox_id` and `preview_url` change. Returning the whole `Project` would hand the client a
`title` snapshot that may be **older than one the user just typed** — a rename racing a refresh
would silently revert the sidebar.

Narrow responses are the general rule for RPCs that touch one part of an entity.

### 4.4 No status or progress field

The call is synchronous and takes 10–30 seconds; it returns when the sandbox is genuinely ready. So
there is nothing to report progress *about* — the frontend shows a pending state on the button and
waits for the response.

This is the second-best candidate for asynchronous treatment after Build
(`project-codebase-model.md` §9.3). When that happens the shape changes to a job handle plus a poll
or stream, and this message changes with it. Do not pre-build a `status` field for a flow that is
still synchronous.

## 5. What deliberately does not exist

| Not added | Why |
| --- | --- |
| `GetProjectFiles` / a file tree | Deferred (`logged_in_design.md` §7.4). Nothing in the MVP screen reads the file map — only the server replays it |
| A `Codebase` message | The map never crosses the wire. Adding a message for it would invite a client to fetch tens of KB it has no use for |
| `commands` on the response | The `npm install` log is an implementation detail of replay |
| `DeleteSandbox` | Sandboxes expire on their own in five minutes; nothing reaps them and nothing needs to |

## 6. Landing it

```bash
make plint && make gen
cd app && npm run lint && npm run build
```

```bash
curl -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects/$PID/sandbox/refresh
```

Note the absence of `-d` (§4.2). Three checks:

1. **The path transcoded.** `/api/v1/projects/{id}/sandbox/refresh` is the first two-segment suffix
   after a path variable in this repo — confirm it appears in `openapi.yaml` and that the REST call
   actually routes, rather than falling through to a 404.
2. **`RefreshSandbox` is on the existing client**, not a new one — check
   `app/src/gen/api/v1/project_service_pb.ts` and that no new `*_connect` service appeared in Go.
3. **Unauthenticated returns 401.** Adding an RPC to an existing service does not make it protected;
   the handler has to call `authAdapter.User(ctx)` itself.

## 7. Out of scope

The `project_codebases` table, the JSONB merge, the `npm install` filter, and the replay flow —
all `.planning/Backend/project-codebase-model.md` · asynchronous refresh (§4.4) · the AI-service
routes it calls (`.planning/Backend/ai-service-client.md`).
