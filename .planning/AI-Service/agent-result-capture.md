# Agent Result Capture — AI Service

> Making the callback's record of what the agent did **true**. Today it infers success by
> substring-matching the tool's prose output, which silently drops real file writes and real
> `npm install`s.
>
> **This blocks `.planning/Backend/project-codebase-model.md`.** Everything that document persists
> comes from here — if the capture is wrong, the stored codebase is wrong, and a replayed sandbox is
> missing files nobody can account for.
>
> Verified against the code on `feature/messages` (2026-09-08).
>
> ## ⏸ Phase: POLISH — not needed for the MVP
>
> **The MVP discards `files` and `commands`, so the callback's accuracy does not matter yet.**
>
> The assistant message body is `resp.Summary`, which is parsed from the model's own output by
> `PydanticOutputParser` — it does not pass through `CodeAgentCallBack` at all. So a broken callback
> produces a correct demo.
>
> This becomes a **blocker** the moment `project-codebase-model.md` starts persisting `files`. Do it
> then, and before that document, not before the MVP.


## 1. Why the callback exists at all

`ARCHITECTURE.md` §7.5: an agent's narration of what it did is not evidence. `CodeAgentCallBack` is
constructed fresh per request and hooks the executor so the response can be split by trust level —
`summary` is model prose, while `files` and `commands` are **observed fact**.

That split is the whole design. This document is about the observation being reliable enough to
deserve the name.

## 2. The defect

`agent_callback_service.py:47`, the only place pending inputs are promoted to the result:

```python
def on_tool_end(self, output: str, **kwargs: Any) -> None:
    tool_name = kwargs.get("name", "")
    success = "failed to" not in output and "error" not in output
    ...
    if tool_name == "write_sandbox_files" or tool_name == "execute_sandbox_command":
        if "failed to" not in output and "error" not in output:
            self.updated_files.update(self.pending_files)
            self.commands_executed.extend(self.pending_commands)
    self.pending_files.clear()
    self.pending_commands.clear()
```

`SandboxCommandTool._run` builds its return string from the command's **full stdout and stderr**:

```python
output_parts.append(f"\nStdout:\n{result.stdout}")
output_parts.append(f"\nStderr:\n{result.stderr}")
```

So the success test is a substring scan over arbitrary program output. `npm install` writes
deprecation notices to stderr as a matter of course, and any package whose name or output contains
the substring `error` causes the command to be discarded from `commands_executed`.

**The asymmetry makes it worse, not better.** The check is case-sensitive and lowercase, so
`npm ERR!` — npm's actual failure prefix — does **not** match. Genuine failures can be recorded as
successes while harmless installs are thrown away. It fails in both directions, and silently.

`ARCHITECTURE.md` §11 #6 already lists this. It is being promoted from a known defect to a blocker
because a feature is now being built on top of it.

### 2.1 What is *not* wrong — verified, so it is not re-litigated

**`kwargs.get("name")` works.** `langchain-core` 0.3.79 passes the tool name explicitly:

```python
# langchain_core/tools/base.py:897
run_manager.on_tool_end(output, color=color, name=self.name, **kwargs)
```

and `handle_event` forwards it to the handler. `tool_name` is populated.

**`output` really is a `str`.** `_format_output` returns the raw `_run` return value unless a
`tool_call_id` was supplied, and the classic `AgentExecutor` does not supply one. So the substring
test is genuinely operating on prose — which is precisely the problem, not a misreading of it.

## 3. The fix

The callback only ever sees the tool's **returned string**. It has no access to a return value from
`SandboxService.execute_terminal_command`. So adding an exit code to `TerminalInfo` is necessary but
not sufficient — the tool has to surface it in what it returns.

**Give the tool and the callback a contract: a fixed status marker on the first line.**

```
TOOL_STATUS: ok
Command: npm install clsx
Stdout:
…
```

```
TOOL_STATUS: failed exit=1
Command: npm install nonexistent-pkg-xyz
Stderr:
…
```

The callback matches that line exactly:

```python
_OK = "TOOL_STATUS: ok"

def on_tool_end(self, output: str, **kwargs: Any) -> None:
    tool_name = kwargs.get("name", "")
    if tool_name in ("write_sandbox_files", "execute_sandbox_command"):
        if str(output).startswith(_OK):
            self.updated_files.update(self.pending_files)
            self.commands_executed.extend(self.pending_commands)
        else:
            logger.warning("agent_tool_failed", tool_name=tool_name,
                           output_preview=str(output)[:200])
    self.pending_files.clear()
    self.pending_commands.clear()
```

**Why a marker rather than an exception.** Raising on a non-zero exit would route through
`on_tool_error`, which never promotes — clean, but it would break a deliberate design decision:
`sandbox_service.py`'s tools each catch their own exceptions and return the failure *as a string* so
the agent can read it and recover instead of the run aborting. The marker preserves that; the agent
still reads everything after the first line.

**Why the marker is on line one.** `startswith` cannot be fooled by program output. A marker
anywhere else is a substring scan again, wearing better clothes.

## 4. Changes, file by file

### 4.1 `services/models/sandbox_models.py`

```python
class TerminalInfo(BaseModel):
    stdout: str
    stderr: str
    exit_code: int          # new
```

The existing field docstrings are misleading and should be corrected while here: `stderr` is
described as "output of failed command", but a successful command routinely writes to stderr — that
misconception is the origin of this whole defect.

### 4.2 `services/sandbox_service.py`

`execute_terminal_command` already receives a `CommandResult` from the E2B SDK and discards
everything but `stdout`/`stderr`. Keep the exit code:

```python
def execute_terminal_command(self, sandbox_id: str, command: str) -> TerminalInfo:
    sbx = Sandbox.connect(sandbox_id)
    result: CommandResult = sbx.commands.run(cmd=command)
    return TerminalInfo(stdout=result.stdout, stderr=result.stderr,
                        exit_code=result.exit_code)
```

Check the SDK's behaviour on a non-zero exit before relying on this: `commands.run` may **raise**
rather than return a result with a non-zero code. If it raises, the `except` branch is the failure
path and must emit the `TOOL_STATUS: failed` marker — the exit code is then a nicety, not the
mechanism.

Then `SandboxCommandTool._run` and `SandboxWriteTool._run` both lead with the marker, and their
`except` branches lead with the failure marker.

**`SandboxListTool` and `SandboxReadTool` should get the marker too**, for consistency — the
callback ignores them today, but a reader should not have to know which two tools are special.

### 4.3 `services/agent_callback_service.py`

§3. Also drop the now-duplicated `success = …` computation at the top of the method — it is
calculated, used only for logging, and then recomputed inline.

## 5. What this does not fix

**Write failures are still inferred, not observed.** `write_sandbox_files` reports success if it did
not raise. E2B's `files.write_files` returns a list of `WriteInfo` for what it actually wrote; a
partial write — three of four files — would still be recorded as all four.

Comparing `len(result)` against `len(write_data)` and emitting the failure marker on a mismatch is a
small addition and worth doing here if it is cheap. If it turns out E2B guarantees all-or-nothing,
say so in a comment rather than leaving the reader to wonder.

## 6. Verifying

No test suite exists in this repo. Manual, against the service directly — do this **before** any Go
work, so a downstream failure is unambiguous.

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .id)

curl -s -X POST http://localhost:9999/ai-service/v1/anthropic/$SB/code \
  -H 'content-type: application/json' \
  -d '{"message":"create app/page.tsx with a red heading, then install the clsx package"}' | jq '{files: (.files|keys), commands}'
```

Four checks:

1. **The happy path captures both.** `files` has the path, `commands` has the `npm install`. If
   either is empty, nothing downstream can work.
2. **The regression this fixes.** Ask the agent to install a package whose output mentions "error"
   — `npm install serialize-error` is the direct case. Before the fix that command is dropped; after
   it, it is recorded. This is *the* check.
3. **A real failure is recorded as a failure.** Ask it to install a package that does not exist. The
   command must be **absent** from `commands`, and the log must carry `agent_tool_failed`.
4. **The agent still recovers.** In case 3, confirm the agent read the failure and responded to it
   rather than the run aborting — that is the behaviour the marker approach exists to preserve.

## 7. Out of scope

The `command` route's query-parameter binding and HTTP status codes (`sandbox-service.md`) ·
conversation history (`conversation-history.md`) · a delete tool and path normalisation
(`agent-capabilities.md`) · anything in Go.
