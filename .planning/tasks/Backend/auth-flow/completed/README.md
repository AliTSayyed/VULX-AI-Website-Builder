# Backend — Auth Flow Tasks

Three defect fixes in the Go API. Context and rationale: `.planning/Backend/auth-flow.md`.

None of these are features. The OAuth login path already works; these stop it from breaking in
ways that are hard to diagnose later.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | `GetJWTCookie` — return on match instead of resetting | 3 lines, isolated |
| 2 | `task2.md` | Interceptor — stop re-wrapping handler errors | 1 line, isolated |
| 3 | `task3.md` | Delete OAuth state after a successful exchange | ~6 lines, **ordering constraint** |

## Order

Tasks 1 and 2 are independent — do them in any order, they can share one commit.

**Task 3 must not land before `tasks/Frontend/auth-flow/task8.md`** (the callback route's `useRef`
latch). See `.planning/Backend/auth-flow.md` §3. Doing it early makes every local login *appear* to
fail while actually succeeding.

## Verifying

`make api` rebuilds and hot-reloads the Go service. There is no test harness in this repo, so each
task carries its own manual `curl`. The API must be reachable at `https://local.api.vulx.ai`, which
means `tasks/Frontend/auth-flow/task1.md` (infrastructure) is done first.
