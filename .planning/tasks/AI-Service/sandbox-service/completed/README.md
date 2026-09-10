# AI Service — Sandbox Service Tasks

One task. Context: `.planning/AI-Service/sandbox-service.md` — **§4 only**; §2 and §3 are Polish.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | Make the E2B sandbox timeout configurable and raise it | ~6 lines, isolated |

## Why this is on the MVP path

The MVP has **no `RefreshSandbox`**. When a sandbox expires, the project is stuck — `SendMessage`
reuses `sandbox_id` blindly and cannot tell a live sandbox from a dead one. The mitigation is to
make the sandbox outlive the session.

E2B's default is 5 minutes, which is shorter than one Build run plus reading the result.

## Order

**None.** No dependency on any Go work, any proto, or any other task in this repo. Do it first, last,
or while waiting on something else.

## Verifying

`docker compose restart ai-service`, then create a sandbox and run a command against it **six
minutes later**. It must still answer. There is no faster way to prove this one.
