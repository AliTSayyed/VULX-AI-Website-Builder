# Sandbox Service — AI Service

> The HTTP surface the Go API calls, and how long a sandbox lives. Three changes: `command` moves
> from a query parameter to a request body, routes return **meaningful status codes** instead of a
> blanket 500, and sandbox lifetime becomes configurable instead of E2B's 5-minute default.
>
> **Do this before writing the Go client.** `.planning/Backend/ai-service-client.md` is currently
> written against today's shape and says so explicitly; changing the contract afterwards means
> writing the Go side twice. §6.
>
> Verified against the code on `feature/messages` (2026-09-08).

## ⏸ Phase split — MVP vs Polish

**One section is MVP: §4, sandbox lifetime.** Everything else here serves `RunCommand` and the
replay path, which the MVP does not have.

| § | Content | Phase |
| --- | --- | --- |
| §1 | who calls these routes | context for both |
| §2 | `command` query param → body | Polish |
| §3 | meaningful status codes | Polish |
| §4 | **`Sandbox.create(timeout=…)`** | **MVP** |
| §6 | ordering vs the Go client | Polish — it is about §2 |

### Why §4 is MVP

The MVP has **no `RefreshSandbox`** — when a sandbox expires, the project is stuck
(`messages-model.md` §9.0). The mitigation is simply to make the sandbox outlive the session.

E2B's default is 5 minutes, which is shorter than a single Build run plus reading the result. Make
it a config key, keep the default low while developing so repeated test runs do not burn credits,
and raise it for a demo:

```
E2B_SANDBOX_TIMEOUT_SECONDS=300     # dev — cheap, dies fast
E2B_SANDBOX_TIMEOUT_SECONDS=1800    # demo — outlives the session
```

That is the whole MVP change here: one config key, one keyword argument, and the `.env.example`
entry. It is also what makes "refreshes are not needed" true rather than hopeful.

## 1. Who calls these routes

Worth stating first, because it decides what is safe to change.

| Surface | Caller | Mechanism |
| --- | --- | --- |
| `SandboxCommandTool`, `SandboxWriteTool`, … | **the AI**, via LangChain | in-process Python, `args_schema` |
| `POST /sandbox/`, `/sandbox/{id}/files`, `/sandbox/{id}/command` | **the Go API** | HTTP |

Both reach the same `SandboxService` methods. The agent never makes HTTP calls to its own service.

**So nothing in this document touches the agent's interface.** `CommandToolInput` and
`WriteToolInput` are unchanged; the AI keeps passing `command` as a plain string. Only the HTTP
parameter binding changes, and its only consumer is Go.

The Go consumer is `RefreshSandbox` (`project-codebase-model.md` §9.1):

```
RefreshSandbox
  ├─ CreateSandbox → POST /sandbox/
  ├─ WriteFiles    → POST /sandbox/{id}/files       ← files first…
  └─ RunCommand    → POST /sandbox/{id}/command     ← …then npm install
```

That order matters: a stored `package.json` written *after* an install would clobber what npm just
resolved. Build's "ensure a live sandbox" step reuses the same three calls, so it is the same flow,
not a second consumer.

**`ARCHITECTURE.md` §7.1 currently says only sandbox *create* is intended for the Go API and the
rest are development affordances. That sentence is now false** — update it in this change.

## 2. `command` becomes a body

> **⏸ Polish.** The only Go caller of this route is `RunCommand`, which exists to replay
> `npm install` during a codebase restore — and the MVP neither stores nor replays a codebase. Leave
> the query-param binding alone until `project-codebase-model.md`.

```python
# api/routes/sandbox.py — today
@router.post("/{sandbox_id}/command")
async def execute_sandbox_command(
    sandbox_id: str, command: str, sandbox_service: sandbox_service_dependency
) -> ExecuteSandboxResponse:
```

`command` is a bare `str` that is not a path parameter, so **FastAPI binds it to the query string**.
The request is `POST /sandbox/{id}/command?command=npm+install+clsx` with an empty body.

Three reasons to change it:

1. **It is inconsistent with its neighbour.** `POST /sandbox/{id}/files` takes a Pydantic body. Two
   adjacent mutation routes with opposite conventions is a trap for the Go client, and the kind of
   thing that costs an afternoon.
2. **URLs have length limits; commands do not.** A collapsed `npm install a b c d …` from a project
   with many dependencies (`project-codebase-model.md` §8) is exactly the case that grows.
3. **Commands land in access logs and proxy logs** when they are in the URL. Less of a concern
   inside the compose network, but it is free to avoid.

```python
# services/models/... or routes/models/sandbox_models.py
class CommandRequest(BaseModel):
    command: str = Field(..., description="terminal command to execute")
```

```python
@router.post("/{sandbox_id}/command")
async def execute_sandbox_command(
    sandbox_id: str, request: CommandRequest, sandbox_service: sandbox_service_dependency
) -> ExecuteSandboxResponse:
```

This is a **breaking change to a route that has no working caller today** — the Go client is a stub
and the browser cannot reach this service (CORS allows only `http://api:8080`). So it is free now
and expensive later. §6.

## 3. Status codes that mean something

> **⏸ Polish.** Same reason as §2 — plus the MVP has no recovery path that would act on a 404. It
> reuses `sandbox_id` blindly and fails opaquely when the sandbox is gone
> (`messages-model.md` §9.0), which is why §4 matters instead.

Every route currently ends the same way:

```python
except Exception as e:
    raise HTTPException(status_code=500, detail=f"Failed to …: {str(e)}")
```

So a dead sandbox, E2B being down, a bad template id, and a genuine bug are **all 500 with prose**.
`ai-service-client.md` §3.3 had to tell the Go side not to branch on a distinction that does not
exist.

The one that matters in practice: **`Sandbox.connect(sandbox_id)` on an expired sandbox.** That is a
routine, expected condition — sandboxes are supposed to die — and it should not look like an
outage.

| Condition | Status | What Go does with it |
| --- | --- | --- |
| `Sandbox.connect` fails: unknown or expired id | **404** | `ErrorTypeNotFound` → refresh and retry, don't alarm |
| E2B unreachable, quota, auth failure | **502** | `ErrorTypeUnavailable` → surface as "try again" |
| Forbidden path in `list_files` | **400** | `ErrorTypeInvalid` → a bug in the caller |
| Anything else | 500 | `ErrorTypeInternal` |

The precise E2B exception types need checking against the SDK — that is implementation work, not a
decision. If they cannot be told apart cleanly, **say so in a comment and keep 500**; a wrong 404 is
worse than an honest 500, because it tells Go to retry a sandbox that will never exist.

Keep the `detail` string. It is genuinely useful in the API's logs and Go never shows it to a
browser (`ai-service-client.md` §3.3).

## 4. Sandbox lifetime

> **▶ MVP — the only part of this document the demo needs.** With no `RefreshSandbox`, the sandbox
> must simply outlive the demo. Five minutes does not; it is shorter than one Build run plus reading
> the result.

```python
# services/sandbox_service.py
def create(self, template_id: str) -> Sandbox:
    sbx = Sandbox.create(
        template=template_id
    )  # By default the sandbox is alive for 5 minutes
    return sbx
```

**The cheapest real improvement in the whole planning set.** The E2B SDK takes a `timeout`
parameter and nothing passes one, so every sandbox gets the 5-minute default.

```python
# api/config.py
e2b_sandbox_timeout_seconds: int = 1800   # 30 minutes
```

```python
def create(self, template_id: str) -> Sandbox:
    return Sandbox.create(
        template=template_id,
        timeout=settings.e2b_sandbox_timeout_seconds,
    )
```

Add the key to `.env.example`. Confirm the parameter's name and unit against the installed
`e2b_code_interpreter` version — `timeout` in seconds is the usual signature, but verify rather than
assume.

**Why it matters more than it looks.** Five minutes is shorter than a single Build run plus reading
the result. It means a user who builds something, reads the summary, and looks away has a dead
preview before they look back. At 30 minutes, refresh becomes an occasional action rather than a
constant one — and every avoided refresh is a sandbox creation plus a full file replay plus an
`npm install` not paid for.

**Costs**, honestly: E2B bills for sandbox uptime, and abandoned sandboxes stay alive six times
longer. Nothing reaps them (`project-codebase-model.md` §10.6). 30 minutes is a starting point, not
a researched number — check it against E2B's pricing and pick deliberately.

**Not in scope: keep-alive.** E2B exposes `set_timeout` on a live sandbox, so extending one while a
user is active is possible. That is the lazy-lifecycle management explicitly deferred in favour of
the manual refresh button (`project-codebase-model.md` §1). Raising the initial timeout is a
constant; keeping a sandbox alive on activity is a policy, and policy needs the frontend to have an
opinion first.

## 5. Files

| File | Change |
| --- | --- |
| `api/routes/models/sandbox_models.py` | add `CommandRequest` |
| `api/routes/sandbox.py` | `command` from body; per-condition status codes |
| `services/sandbox_service.py` | `timeout=` on `Sandbox.create` |
| `api/config.py` | `e2b_sandbox_timeout_seconds` |
| `ai-service/.env.example` | the new key |
| `ARCHITECTURE.md` | §7.1's "development affordances" sentence (§1) |
| `.planning/Backend/ai-service-client.md` | §3.2 and §3.3 — §6 below |

## 6. Ordering — the one coordination point

> **⏸ Polish only.** This constraint exists because of §2, which the MVP skips. **§4 has no ordering
> constraint at all** — it is a config key and a keyword argument, independent of every other
> document. Do it whenever.

`ai-service-client.md` §3.2 documents `command` as a query parameter and states plainly that "the Go
client matches what exists today". §3.3 tells Go not to distinguish failure classes because it
cannot.

**Both become wrong when this lands.** So:

1. Do this document first.
2. Update `ai-service-client.md` §3.2 (body, not query) and §3.3 (404 / 502 / 400 are now
   meaningful, and `SandboxRunner` can map them).
3. Then write the Go client.

Doing it the other way round means writing the Go request-building and error-mapping twice.

**`agent-result-capture.md` has no such constraint** — it changes what the callback *reports*, not
any wire shape, so it can land before, after, or alongside this without touching Go.

## 7. Verifying

```bash
cd ai-service && ruff check .
docker compose restart ai-service
```

```bash
# command now takes a body, not a query param
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .id)
curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/$SB/command \
  -H 'content-type: application/json' -d '{"command":"node -v"}'

# the old form must now fail with 422 — proves the binding actually moved
curl -s -o /dev/null -w '%{http_code}\n' -X POST \
  "http://localhost:9999/ai-service/v1/sandbox/$SB/command?command=node%20-v"

# an expired or unknown sandbox is 404, not 500
curl -s -o /dev/null -w '%{http_code}\n' -X POST \
  http://localhost:9999/ai-service/v1/sandbox/does-not-exist/command \
  -H 'content-type: application/json' -d '{"command":"node -v"}'
```

Then the one that takes patience but is the whole point of §4: **create a sandbox, wait six minutes,
and run a command against it.** It must still work. Before this change it is dead at five.

Also confirm `/docs` still renders — `CommandRequest` should now appear as a request body schema
rather than a query parameter, which is the visible proof the contract changed.

## 8. Out of scope

The callback's success test (`agent-result-capture.md`) · conversation history
(`conversation-history.md`) · a delete tool (`agent-capabilities.md`) · keep-alive and sandbox
reaping (§4) · a `GET /providers` endpoint — deliberately skipped, the `AiProvider` enum is the list
· `ReadFile` / `ListFiles` in the Go client — no caller needs them.
