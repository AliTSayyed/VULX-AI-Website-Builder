# Backend Task 2 — Interceptor flattens authenticated error codes

**Goal.** An authenticated request's error reaches the browser with the code the handler actually
meant.

## The defect

`api/internal/infrastructure/inbound/grpc/adapters/auth/interceptor.go:25-27`:

```go
res, err := next(context.WithValue(ctx, UserContextKey{}, user), req)
if err != nil {
    return nil, connect.NewError(connect.CodeInvalidArgument, err)
}
```

`errorAdapter.ToConnectError` is the one place in the codebase that maps a `domain.ErrorType` to a
transport code — and this line throws that work away for every *authenticated* request. A `NotFound`
arrives as `InvalidArgument`; so does an `Internal`, a `PermissionDenied`, and everything else.

Unauthenticated requests take the `return next(ctx, req)` branch at the bottom and are unaffected,
which is why the logged-out session check works correctly today.

## The fix

Return the error unchanged. It is already a `*connect.Error` with the right code, because every
handler runs it through `errorAdapter.ToConnectError` first.

```go
res, err := next(context.WithValue(ctx, UserContextKey{}, user), req)
if err != nil {
    return nil, err
}
```

## Why this matters for the auth flow

`GetUserProfile` and `AccountLogout` are both authenticated calls. Without this fix, any real
failure in either one — a dropped DB connection, a Redis outage during logout — surfaces in the
browser as `invalid_argument`, which points the reader at a malformed request that isn't malformed.
The frontend's session hook only special-cases `unauthenticated`, so everything else becomes a
generic toast; this is what makes that toast's message trustworthy.

## Done when

- `make api` builds.
- Any authenticated RPC that fails returns its true code. Easiest check: call a protected RPC with a
  valid cookie and a payload the handler rejects, and confirm the code is not universally
  `invalid_argument`.

## Notes

- Do not remove the surrounding `if user, refresh := ...` block. The advisory-interceptor design
  (`ARCHITECTURE.md` §6.4) is deliberate: a handler opts into protection by calling
  `authAdapter.User(ctx)`, and there is no route allowlist to keep in sync.
- The `refresh` path below the error check is untouched.
