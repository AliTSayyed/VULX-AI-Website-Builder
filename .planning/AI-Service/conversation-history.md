# Conversation History — AI Service

> Both LLM endpoints take a single string and have no memory. The UI renders a conversation; the
> model is not having one. This document makes the thread part of the request.
>
> Fixes the limitation recorded in `.planning/Backend/messages-model.md` §5, which ships one-shot for
> MVP and points here for the real fix.
>
> Verified against the code on `feature/messages` (2026-09-08).
>
> ## ⏸ Phase: POLISH — not needed for the MVP
>
> **The MVP is one Build message producing one site.** A follow-up that depends on the previous turn
> ("now make the header blue") is exactly what this document enables, and exactly what the demo does
> not need to do.
>
> Worth knowing while demoing: the agent has no memory between messages, so a second Build message is
> answered as if it were the first. It still edits the same sandbox, so incremental changes partly
> work by accident — the files are there even though the conversation is not.


## 1. The problem

```python
# api/routes/models/ai_models.py
class AIRequest(BaseModel):
    message: str

class AICodeAgentRequest(BaseModel):
    message: str
```

One string. No history parameter, no message list, no conversation id. So "make it blue" after
"build a hero section" does not know what "it" is — in **both** Chat and Build.

This will read as a bug to every user who hits it, because it looks exactly like one.

## 2. Why it cannot be fixed inside the Python service

The obvious move — attach LangChain memory to the agent — does not work here, and the reason is
structural.

```python
# api/dependencies.py
@lru_cache()
def get_openai_code_agent_service(
    openai: openai_dependency, sandbox: sandbox_service_dependency
) -> CodeAgentService:
    ...
```

`@lru_cache()` is what makes these singletons: **one `CodeAgentService` per provider for the entire
process lifetime**, shared by every request from every user (`ARCHITECTURE.md` §7.2). The
`AgentExecutor` is built once in `__init__`.

Memory attached there would be **shared across all users** — one person's project context leaking
into another's. The isolation is correct and must stay; what has to change is that continuity
arrives *in the request* instead.

That also fits the service's defining property: it is stateless and has no database
(`ARCHITECTURE.md` §7). Everything it needs comes in the call. History is no different.

Note the one piece of per-request state that already exists and shows the pattern:
`CodeAgentCallBack` is constructed fresh per request and passed through
`config={"callbacks": [callback]}` rather than held on the service. History should arrive the same
way — as an argument, never as instance state.

## 3. The shape

```python
class ChatMessage(BaseModel):
    role: Literal["user", "assistant"]
    content: str

class AIRequest(BaseModel):
    message: str
    history: List[ChatMessage] = Field(default_factory=list)

class AICodeAgentRequest(BaseModel):
    message: str
    history: List[ChatMessage] = Field(default_factory=list)
```

Three decisions:

**`history` is separate from `message`, not a single `messages` list.** Both service methods already
take `user_message` and interpolate it into a prompt template at a specific position — `NEXTJS_PROMPT`
ends with `{input}` then `{agent_scratchpad}`. Folding the current turn into a list would mean
restructuring both prompts. Keeping the new turn distinct is a smaller change and keeps "what is
being asked right now" unambiguous to the model.

**It defaults to empty**, so this is a **backward-compatible change**. Existing callers — including
the Go client, whenever it lands — keep working with no history and get today's behaviour. That
matters because it decouples this document from the Go side entirely (§6).

**`role` is `Literal["user", "assistant"]`**, matching `domain.MessageRole`'s two non-unspecified
values (`messages-model.md` §3.1). No `system` role: the system prompt belongs to the service, not
the caller. Accepting a caller-supplied system message would let the Go API — and therefore
anything upstream of it — override the sandbox ground rules in `NEXTJS_PROMPT`.

## 4. Threading it through

### 4.1 `GeneralAIService.process_query_request`

Currently builds a single `HumanMessagePromptTemplate` from `QUERY_PROMPT` and invokes. Prepend the
history as real LangChain messages (`HumanMessage` / `AIMessage`) before the current turn, so the
model sees a genuine conversation rather than a transcript pasted into one string.

The `PydanticOutputParser` and `LLMQueryResult` shape do not change.

### 4.2 `CodeAgentService.process_code_request`

Harder, and the reason this document is not trivial. The prompt is a single template string ending:

```
{format_instructions}
{input}
{agent_scratchpad}
```

`{input}` receives `f"Sandbox ID: {sandbox_id}\nTask: {user_message}"` — the sandbox id reaches the
tools through **prompt text**, not through executor configuration (`ARCHITECTURE.md` §7.3).

Two options:

| Option | Change | Verdict |
| --- | --- | --- |
| **(a)** Render history into the `{input}` string above the task | No prompt restructuring. But history becomes indistinguishable from the current instruction, and the agent may re-execute a previous turn's task | Rejected — it invites duplicate builds |
| **(b)** Add a `MessagesPlaceholder("history")` and move to `ChatPromptTemplate.from_messages` | Correct: history is real message objects the model can tell apart from the instruction. Requires restructuring `NEXTJS_PROMPT` from one template string into system + placeholder + input | **Recommended** |

Take (b). (a) has a specific, likely failure: an agent that reads a prior "build a pricing page" as
part of its current instruction will build it again, write the files again, and the callback will
faithfully record a duplicate run.

### 4.3 Token budget

A long thread plus `NEXTJS_PROMPT` — which is itself very large — will hit context limits. Truncate
before sending: keep the most recent N turns (start with 10) or a character budget, dropping from
the oldest. Do it **in the service**, not the routes, so all three providers inherit it.

Log when truncation happens. A model that has silently forgotten the start of a thread behaves
strangely in ways nobody will connect to a token budget.

## 5. Go's side, when this lands

For reference — owned by `messages-model.md`, not this document.

`SendMessage` gains a `FindByProject` call before the AI call and passes the prior turns.
`MessageRepository.FindByProject` already returns the whole thread ordered `created_at ASC`
(`messages-model.md` §4.1), so the data is there. The `ChatResponder` and `CodeAgentRunner` ports
grow a history parameter, and `ai-service-client.md`'s DTOs gain the field.

## 6. Ordering

**None.** `history` defaults to empty, so this is backward compatible in both directions:

- Land it before the Go client → the client sends no history and gets today's behaviour.
- Land it after → the client adds a field to a request struct.

Unlike `sandbox-service.md`, there is no coordination point. This is the reason for the default —
not politeness, but decoupling.

## 7. Verifying

```bash
cd ai-service && ruff check .
docker compose restart ai-service
```

**The check that matters is the referential one** — a follow-up that only makes sense given the
previous turn:

```bash
curl -s -X POST http://localhost:9999/ai-service/v1/anthropic/query \
  -H 'content-type: application/json' -d '{
    "message": "why did you pick that over masonry?",
    "history": [
      {"role":"user","content":"how should I lay out a photo gallery?"},
      {"role":"assistant","content":"Use a CSS grid with a fixed aspect ratio."}
    ]
  }' | jq -r .content
```

The answer must be about **grid versus masonry**. Without history it will ask what you are referring
to, or invent a subject.

Four more:

1. **Backward compatibility.** The same call with `history` omitted must still succeed — this is
   what makes §6 true.
2. **No duplicate builds.** Send a code-agent request whose history contains a completed build
   instruction. The agent must not rebuild it. This is the §4.2 failure mode and the reason for
   option (b).
3. **No cross-request leakage.** Two requests in a row with *different* histories; the second must
   not see the first's. If it does, state ended up on the `@lru_cache()` singleton — §2.
4. **Truncation fires and is logged.** Send 30 turns; confirm the log records it and the call
   succeeds rather than erroring on context length.

## 8. Out of scope

The callback's success test (`agent-result-capture.md`) · the sandbox HTTP contract and lifetime
(`sandbox-service.md`) · a delete tool (`agent-capabilities.md`) · streaming responses · summarising
long threads rather than truncating them · persisting anything in this service — it stays stateless
· the Go changes in §5 (`messages-model.md`).
