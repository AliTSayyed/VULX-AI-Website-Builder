# Backend Task 3 — OAuth state is replayable

> ⛔ **Do not land this before `tasks/Frontend/auth-flow/task8.md`.** See §Ordering below. This is
> the one task in either tree with a hard sequencing constraint.

**Goal.** A `code`/`state` pair can be exchanged exactly once.

## The defect

`api/internal/application/services/oauth_service.go:125`:

```go
provider, err := o.cache.Get(ctx, fmt.Sprintf("provider:%s", state))
if err != nil {
    return nil, domain.NewError(domain.ErrorTypeInvalid, fmt.Errorf("invalid or expired oauth provider:%w", err))
}
```

That lookup **is** the CSRF check — an unknown state simply misses the cache. But the key is never
deleted, so the same pair stays valid for the remainder of its 10-minute TTL. Anyone who captures a
callback URL (browser history, a referer header, a shared screenshot) can replay it.

## The fix

Delete the key once the exchange succeeds, inside `CompleteLoginFlow`. The `Cache` port already
declares `Delete` (`application/services/cahce.go`) and Redis implements it, so no new port method
is needed.

Place the delete **after** `oauthprovider.Exchange` returns successfully — deleting before the
exchange would burn the state on a transient Google failure and force the user to restart the flow.

```go
// exchange code for token
token, err := oauthprovider.Exchange(ctx, code, options)
if err != nil {
    return nil, domain.WrapError("oauth service complete login flow", err)
}

// state is single-use; consume it now that the code has been redeemed
if err := o.cache.Delete(ctx, fmt.Sprintf("provider:%s", state)); err != nil {
    utils.Logger.Warn("failed to delete oauth state", "error", err)
}
```

A failed delete is logged, not returned — the login itself succeeded, and failing it here would turn
a Redis hiccup into a user-visible auth error.

### Also consider

`BeginLoginFlow` writes a second key, `options:%s`, when extra params are present. Nothing sets extra
params today, so that key is never written in practice. If you clean it up, delete it in the same
place and for the same reason; if not, leave a comment so the asymmetry is not read as an oversight.

## Ordering

React 19 StrictMode fires effects twice in development. Today both `FinishAccountAuth` calls succeed
because the state survives. After this change the second call fails on a consumed state — and the
callback route paints the *second* result, so **a working login renders as an error**.

The frontend latch (`tasks/Frontend/auth-flow/task8.md`, the `useRef` guard) is what prevents the
second call. Land that first, or both in one commit. Never this one alone.

## Done when

- `make api` builds.
- A full browser login still succeeds through `/auth/callback`.
- Replaying the same callback URL a second time now fails with `invalid_argument` /
  "invalid or expired oauth provider" instead of logging in again.
- `docker compose exec redis redis-cli keys 'provider:*'` is empty after a completed login.
