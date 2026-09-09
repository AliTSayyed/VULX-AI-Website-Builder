# MVP — the demo path

> One page. Every document in `.planning/` now carries a **⏸ Phase split** section; this is the
> index that puts the MVP halves in build order.
>
> **The demo:** type a prompt → a project is created → a sandbox starts → the agent writes code into
> it → you watch the site appear in the iframe.
>
> Delete this file when the MVP is done — it is a sequencing aid, not a spec. The documents are the
> spec.

## What is in, and what is not

| In | Out |
| --- | --- |
| `projects` + `messages` tables | `project_codebases` — nothing is stored or replayed |
| `CreateProject`, `GetProject`, `ListProjects` | `RenameProject`, LLM-generated titles |
| `ListMessages`, `SendMessage` with `mode = BUILD` | Chat mode, conversation history |
| `CreateSandbox`, `RunCodeAgent` | `Query`, `WriteFiles`, `RunCommand`, `RefreshSandbox` |
| Sandbox timeout raised so it outlives a demo | Temporal, `status`, polling, credits |

Titles are the first prompt, truncated. `resp.Files` and `resp.Commands` come back from the agent
and are **discarded**.

## Build order

### 0 — Prove the agent works standalone

No document, no code. Do this before anything else, so a later failure is unambiguous.

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | tee /dev/stderr | jq -r .id)
curl -s -X POST http://localhost:9999/ai-service/v1/anthropic/$SB/code \
  -H 'content-type: application/json' \
  -d '{"message":"build a landing page for a coffee shop"}' | jq -r .summary
```

Open the sandbox URL. **If that renders, everything below is plumbing between two working ends.** If
it does not, fix it here rather than debugging through three layers later.

### 1 — Sandbox lifetime · `AI-Service/sandbox-service.md` §4

One config key, one keyword argument, one `.env.example` line. No dependencies — do it whenever.

Keep the dev default low (300s) so repeated test runs do not burn credits; set 1800 for the demo.
This is what makes "no refresh needed" true rather than hopeful.

### 2 — Proto · `Proto/project-proto.md` + `Proto/messages-proto.md`

Both at once, then one `make plint && make gen`.

- `enums.proto`: `AiProvider`, `MessageRole`, `ChatMode`
- `project_service.proto`: `ListProjects`, `GetProject`, `CreateProject` — **omit `RenameProject`**
- `message_service.proto`: `ListMessages`, `SendMessage` — complete as written

**Check before writing any Go:** `grep -n "OPENAI" app/src/gen/api/v1/enums_pb.ts` must show
`OPENAI = 1`, not `AI_PROVIDER_OPENAI = 1`. If it shows the long form the enum was named `AIProvider`
— see `project-proto.md` §2.3. Cheap now, annoying once the TS client is wired.

### 3 — Projects · `Backend/project-model.md`

Migration `00004`, `domain/project.go`, `domain/ai_provider.go`, `project_repository.go`
(`Create`, `FindByID`, `FindAllByUser`, `UpdateSandbox`), `project_service.go`, the handler, `app.go`.

**Stub `ProjectTitler`** with an in-package implementation returning the provisional title. No LLM
call, no goroutine, no `context.WithoutCancel` — so this step has **no dependency on step 4**.

Verifiable on its own: create a project over `curl`, list it, confirm a second account gets `404`.

### 4 — AI client · `Backend/ai-service-client.md`

Two methods — `CreateSandbox`, `RunCodeAgent` — plus the shared `do` helper, and delete
`user_workflow.go` (which touches `temporal.go`, `user_service.go` and `app.go`).

**The 30-second `http.Client` ceiling is the point of this step.** It overrides any longer context
deadline, so until it is gone a code-agent run cannot succeed no matter what else is right.

Independent of step 3 — these two can run in either order or in parallel.

### 5 — Messages and Build · `Backend/messages-model.md`

Migration `00005`, domain, repository (the CTE), service, handler, `app.go`. Then **§9.0**, the
synchronous Build flow. Needs steps 3 and 4.

### 6 — Frontend

Deferred by decision. Verify 0–5 with `curl` and paste the returned `preview_url` into a browser;
that proves the whole pipe without touching React.

## Done when

```bash
PID=$(curl -s -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects \
  -H 'content-type: application/json' \
  -d '{"first_prompt":"a landing page for a coffee shop","provider":"AI_PROVIDER_ANTHROPIC"}' | jq -r .project.id)

curl -s -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' \
  -d '{"body":"a landing page for a coffee shop","mode":"CHAT_MODE_BUILD","provider":"AI_PROVIDER_ANTHROPIC"}'

curl -s -b "jwt=$JWT" https://local.api.vulx.ai/api/v1/projects/$PID | jq -r .project.preview_url
```

Open that URL and see the coffee shop.

## Known MVP limitations — accepted, not overlooked

1. **A dead sandbox is unrecoverable.** `SendMessage` reuses `sandbox_id` blindly and cannot tell a
   live sandbox from an expired one, because the AI service returns 500 for both. Step 1 is the
   mitigation; `RefreshSandbox` is the fix.
2. **`SendMessage` does not return the preview URL** — refetch `GetProject` after a send.
3. **The call blocks for minutes** with no progress feedback beyond a spinner. That is what will make
   you want `build-orchestration.md`, and that wanting is the right trigger.
4. **The agent has no memory between messages.** A second Build message is answered as if it were the
   first — though it still edits the same sandbox, so incremental changes partly work by accident.

## After the MVP, in rough order

`agent-result-capture.md` → `project-codebase-model.md` (+ `sandbox-service.md` §2–§3,
`project-codebase-proto.md`) → Chat mode (`messages-model.md` §8 + `Query`) → LLM titles →
`conversation-history.md` → `build-orchestration.md` → `agent-capabilities.md` → `credits.md`.

`agent-result-capture.md` comes first among these because `project-codebase-model.md` stores what
that callback reports, and today it reports unreliably.
