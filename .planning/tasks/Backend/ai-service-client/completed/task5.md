# Task 5 — Verification

**Goal:** prove the client behaves under failure, not just on the happy path. No code changes.

Source: `.planning/Backend/ai-service-client.md` §10.

The client has no Go caller until `tasks/Backend/messages-model/task5.md`. Some checks here are
direct against the Python service; the rest run once that task lands. **Come back and finish this
file then** — the failure paths are the ones that will bite in a demo.

## 1. The 30-second ceiling is really gone — the single most important check

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .id)
time curl -s -X POST "http://localhost:9999/ai-service/v1/sandbox/$SB/command?command=sleep%2090" 
```

That proves the Python side survives 90 seconds. The Go side is proved by a real Build taking longer
than 30 seconds and succeeding (task 4's timing, and later a full `SendMessage`).

If a Build fails at almost exactly 30 seconds, `Timeout: 30 * time.Second` is still in
`ai_service.go`.

## 2. A transport failure returns; it does not panic

```bash
docker compose stop ai-service
# then trigger any Go path that calls CreateSandbox (after messages-model/task5)
docker compose logs api | tail -30
docker compose start ai-service
```

Expect a clean `Unavailable`. **A nil-pointer stack trace means `defer resp.Body.Close()` is above
the error check** — the original defect.

## 3. The trailing slash

```bash
curl -s -o /dev/null -w '%{http_code}\n' -X POST http://localhost:9999/ai-service/v1/sandbox/   # 200
curl -s -o /dev/null -w '%{http_code}\n' -X POST http://localhost:9999/ai-service/v1/sandbox    # 307
```

Confirm the Go code uses the first form — a 307 works but doubles every sandbox creation.

## 4. `preview_url` is usable in an iframe

```bash
curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .url
```

Must be scheme-qualified (`https://…`). A bare host renders as a relative path in an `<iframe>` and
fails silently in the browser — the demo's whole payoff. See task 3.

Then actually open it. The Next.js template should render before you write a single file, because
`compile_page.sh` starts the dev server at sandbox boot.

## 5. Temporal survived the deletion

```bash
docker compose logs api | grep -i "temporal\|workers"
docker compose stop api && docker compose logs api | tail -20
```

Connection established, workers registered, and **shutdown with no panic** — the nil-worker case from
task 1.

## 6. Nothing references the deleted code

```bash
grep -rn "CallAI\|UserWorkflow\|SandboxResponse" api/ --include=*.go | grep -v "/gen/"
```

Expect nothing.

## Done when

Checks 1, 3, 4, 5, 6 pass now. Check 2 passes after `tasks/Backend/messages-model/task5.md`.
Then move `task1.md`–`task5.md` into `completed/`.
