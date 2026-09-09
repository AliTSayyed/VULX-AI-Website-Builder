# Task 2 — `ai_service.go`: timeouts and the shared `do` helper

**Goal:** remove the 30-second ceiling and give all methods one correct request path.

Source: `.planning/Backend/ai-service-client.md` §4.1, §5.

## Files

| File | Change |
| --- | --- |
| `.../outbound/ai_service/ai_service.go` | rewrite the client construction; add `do` |

## 1. The timeout — the point of this task

```go
// Backstop only. Real deadlines are per-call, set with context.WithTimeout by each
// method, so a caller (or a Temporal activity) governs how long its own call may run.
// A hard client Timeout would override those and cap every call, which is the bug
// this replaces: 30s is shorter than every code-agent run.
client: &http.Client{Timeout: 15 * time.Minute}
```

`http.Client.Timeout` beats any longer context deadline. The old 30s value is why Build cannot work
today, and no amount of correct code elsewhere fixes it.

The 15-minute backstop is not decoration: `RunCodeAgent` sets no deadline of its own, and Temporal
does not forcibly kill an activity goroutine when `StartToCloseTimeout` expires. Without a backstop a
hung connection leaks a goroutine and a socket for the process lifetime.

## 2. The `do` helper

One private method builds, sends, checks and decodes. Five hand-rolled request blocks is where
forgotten `Body.Close()` calls and inconsistent error mapping come from.

```go
func (a *AIService) do(ctx context.Context, method, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return domain.NewError(domain.ErrorTypeInternal,
				fmt.Errorf("ai service marshal %s: %w", path, err))
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, body)
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal,
			fmt.Errorf("ai service new request %s: %w", path, err))
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return domain.NewError(domain.ErrorTypeTimeout,
				fmt.Errorf("ai service %s timed out: %w", path, err))
		}
		return domain.NewError(domain.ErrorTypeUnavailable,
			fmt.Errorf("ai service %s unreachable: %w", path, err))
	}
	defer resp.Body.Close()          // AFTER the error check — see below

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Detail string `json:"detail"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		utils.Logger.Error("ai service returned an error",
			"path", path, "status", resp.StatusCode, "detail", e.Detail)
		if resp.StatusCode == http.StatusUnprocessableEntity {
			// 422 means WE sent something malformed. A bug in Go, not an outage.
			return domain.NewError(domain.ErrorTypeInternal,
				fmt.Errorf("ai service rejected our request to %s (422)", path))
		}
		return domain.NewError(domain.ErrorTypeUnavailable,
			fmt.Errorf("ai service %s failed with status %d", path, resp.StatusCode))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return domain.NewError(domain.ErrorTypeInternal,
				fmt.Errorf("ai service decode %s: %w", path, err))
		}
	}
	return nil
}
```

## Five things that must be exactly right

1. **`http.NewRequestWithContext`, never `http.NewRequest`.** The context is the only thing that
   makes per-call deadlines and cancellation work.
2. **`defer resp.Body.Close()` goes *after* the error check.** Putting it before is the existing
   defect in `sandbox.go` — on a transport failure `resp` is nil and it panics the API.
3. **Never return a bare `fmt.Errorf`.** `domain.WrapError` assigns `ErrorTypeUnknown` to anything
   that is not already a `*domain.Error`, which surfaces as `CodeUnknown` in the browser.
4. **Never put `detail` into the error returned upward.** It is upstream infrastructure text. Log it
   server-side; the interceptor already collapses non-client errors to a generic message, and copying
   `detail` into the error defeats that.
5. **Check both timeout shapes.** A context deadline gives `context.DeadlineExceeded`; the client
   backstop gives a `net.Error` with `Timeout() == true` that does not necessarily wrap it.

## 3. Timeout constants

Package-level, used by tasks 3 and 4:

```go
const (
	createSandboxTimeout = 90 * time.Second
	// RunCodeAgent deliberately has none — it inherits the caller's context.
)
```

## Verifying

```bash
cd api && go build ./... && go vet ./...
grep -n "Timeout" api/internal/infrastructure/outbound/ai_service/ai_service.go
```

Confirm `30 * time.Second` is gone. Nothing calls `do` yet — that is tasks 3 and 4.

## Done when

- The shared client's `Timeout` is 15 minutes, with the comment explaining why.
- `do` exists, closes the body after the error check, and returns only `*domain.Error`.
