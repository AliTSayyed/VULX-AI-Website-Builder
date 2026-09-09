# Task 1 — Configurable sandbox timeout

**Goal:** stop every sandbox dying after 5 minutes, without hardcoding a new number.

Source: `.planning/AI-Service/sandbox-service.md` §4.

## Files

| File | Change |
| --- | --- |
| `ai-service/api/config.py` | new setting |
| `ai-service/services/sandbox_service.py` | pass it to `Sandbox.create` |
| `ai-service/.env.example` | document the key |

## 1. Add the setting

`api/config.py`, in `Settings`, beside `e2b_sandbox_nextjs_template_id`:

```python
    # e2b sandbox lifetime, seconds. E2B's own default is 300 (5 minutes).
    e2b_sandbox_timeout_seconds: int = 300
```

**Default stays 300 on purpose.** Development means many throwaway sandboxes; a low default keeps
credit burn down. Raise it in the env file for a demo, not in code.

`pydantic-settings` reads `E2B_SANDBOX_TIMEOUT_SECONDS` from the environment automatically — the
field name maps to the upper-case env key. No extra wiring.

## 2. Pass it through

`services/sandbox_service.py`:

```python
from api.config import settings          # check the existing import style in this file first

    def create(self, template_id: str) -> Sandbox:
        return Sandbox.create(
            template=template_id,
            timeout=settings.e2b_sandbox_timeout_seconds,
        )
```

Delete the now-wrong comment `# By default the sandbox is alive for 5 minutes`.

**Verify the SDK's parameter name and unit before trusting this.** `timeout` in seconds is the usual
signature for `e2b_code_interpreter`, but confirm against the installed version rather than
assuming — a silently-ignored keyword argument would look exactly like success here:

```bash
docker compose exec ai-service python -c \
  "from e2b_code_interpreter import Sandbox; help(Sandbox.create)" | head -30
```

If the parameter turns out to have a different name (`timeout_ms`, say), use the real one and note it
in the config comment.

## 3. Document the key

`ai-service/.env.example`:

```
# Sandbox lifetime in seconds. 300 = E2B default (dev). Raise for demos.
E2B_SANDBOX_TIMEOUT_SECONDS=300
```

Then set the real value in `../AI-Website-Builder-Secrets/.ai-service-env` — the repo's env files
live outside the repo, so `.env.example` is documentation only and changing it changes nothing at
runtime.

## Verifying

```bash
cd ai-service && ruff check .
docker compose restart ai-service
```

**The real check takes six minutes and there is no shortcut.** Set
`E2B_SANDBOX_TIMEOUT_SECONDS=1800` in the secrets env file, restart, then:

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .id)
date
# wait 6 minutes
curl -s -X POST "http://localhost:9999/ai-service/v1/sandbox/$SB/command?command=node%20-v"
```

The command must succeed. Before this change it fails — the sandbox is gone at 5 minutes.

A faster smoke test that proves the plumbing but **not** the behaviour: set the value to `60`,
create a sandbox, wait 90 seconds, confirm the command now fails. That proves the parameter is being
read; only the six-minute test proves it is being raised.

## Done when

- `ruff check .` is clean.
- A sandbox created with `E2B_SANDBOX_TIMEOUT_SECONDS=1800` answers a command after six minutes.
- The default in `config.py` is still 300.
