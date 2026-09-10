# Task 3 — `sandbox.go`: rewrite `CreateSandbox`

**Goal:** the one sandbox call the MVP needs, with the right verb, the right path, and no panic.

Source: `.planning/Backend/ai-service-client.md` §1 defects 1–2, §4. Depends on tasks 1 and 2.

## The defects being fixed

```go
resp, err := a.client.Get(a.baseURL + "/sandbox/create")   // wrong method AND wrong path
if err != nil {
	// TODO domain wrap this error                          // EMPTY BODY
}
defer resp.Body.Close()                                    // nil-derefs on transport failure
```

The route is `POST /sandbox/`. `AI_SERVICE_URL` already ends in `/ai-service/v1`, so only the tail is
wrong.

## Files

| File | Change |
| --- | --- |
| `.../outbound/ai_service/sandbox.go` | rewrite |

## The code

`SandboxInfo` goes in **`api/internal/domain/sandbox_info.go`**, not in this package — see the note
below.

```go
// domain/sandbox_info.go — a plain value object describing an external result, not an
// entity. Public fields with json tags on purpose: application ports reference this type,
// and it crosses a Temporal activity boundary later (Temporal's JSON converter would
// serialise unexported fields to {}).
type SandboxInfo struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
```

```go
// outbound/ai_service/sandbox.go
func (a *AIService) CreateSandbox(ctx context.Context) (*domain.SandboxInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, createSandboxTimeout)
	defer cancel()

	var info domain.SandboxInfo
	if err := a.do(ctx, http.MethodPost, "/sandbox/", nil, &info); err != nil {
		return nil, domain.WrapError("ai service create sandbox", err)
	}
	if info.ID == "" || info.URL == "" {
		return nil, domain.NewError(domain.ErrorTypeInternal,
			errors.New("ai service returned an empty sandbox id or url"))
	}

	utils.Logger.Info("sandbox created", "sandbox_id", info.ID, "url", info.URL)
	return &info, nil
}
```

Delete the old `SandboxResponse` type and its unchecked logging.

**Why the type lives in `domain` and not here.** `MessageService` declares the `SandboxCreator` port
(`tasks/Backend/messages-model/task4.md`), and a port in `application/services` that referenced
`infrastructure/outbound/ai_service` would point the dependency arrow **outward** — the one thing the
hexagonal layout forbids. `domain` imports nothing and is imported by everything, so putting the DTO
there keeps every arrow pointing inward.

It is a value object, not an entity: public fields, no constructor, no accessors. Do not give it the
`domain/user.go` treatment.

**The trailing slash on `/sandbox/` is required.** FastAPI mounts it as `@router.post("/")` under
`prefix="/sandbox"`. Without it you get a 307 redirect; Go follows it and 307 preserves the method,
so it appears to work while costing an extra round trip on every call.

**The empty-response guard matters.** A 200 with `{}` would otherwise produce a project with an empty
`sandbox_id`, and `SendMessage` would then create a *second* sandbox on the next message, silently
losing the first one's work.

## The thing to actually verify before trusting `url`

The Python side builds the response as:

```python
CreateSandboxResponse(id=sbx.sandbox_id, url=sbx.get_host(3000))
```

**E2B's `get_host(port)` returns a bare host** (`3000-abc123.e2b.dev`) — not a URL with a scheme. If
that is what comes back, `preview_url` will be unusable as an `<iframe src>` and the frontend will
silently resolve it as a relative path.

Check it explicitly:

```bash
curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .url
```

- If it starts with `https://` — nothing to do.
- If it is a bare host — **fix it in Python**, where the sandbox's own semantics are known:
  `url=f"https://{sbx.get_host(3000)}"` in `api/routes/sandbox.py`. Do not prepend the scheme in Go;
  the AI service owns what a usable preview URL looks like, and a second fix site is a second thing
  to forget.

Note this in `.planning/AI-Service/sandbox-service.md` if you change it.

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

The client has no caller yet, so exercise the route directly and compare:

```bash
curl -i -X POST http://localhost:9999/ai-service/v1/sandbox/          # 200, {id, url}
curl -i -X POST http://localhost:9999/ai-service/v1/sandbox           # note: 307
curl -i -X GET  http://localhost:9999/ai-service/v1/sandbox/create    # 404 — the old path
```

The third confirms the bug was real.

## Done when

- `CreateSandbox` issues `POST {base}/sandbox/` and handles the error before touching `resp`.
- `url` is verified to be a scheme-qualified URL, in Python if it was not.
- `go build ./... && go vet ./...` clean.
