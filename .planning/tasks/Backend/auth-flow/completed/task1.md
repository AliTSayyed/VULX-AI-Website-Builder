# Backend Task 1 — `GetJWTCookie` keeps only the last cookie

**Goal.** A `jwt` cookie is found regardless of its position in the `Cookie` header.

## The defect

`api/internal/infrastructure/inbound/grpc/adapters/auth/auth.go:69-77`:

```go
var token string
for _, cookie := range cookies {
    if cookie.Name == apiCookieName {
        token = cookie.Value
    } else {
        token = ""          // ← wipes the token we just found
    }
}
return token
```

The `else` branch clears the match, so the loop returns a token **only when `jwt` happens to be the
last cookie in the header**. Any cookie sorted after it silently logs the user out.

It does not bite today because `.vulx.ai` carries exactly one cookie. It starts biting the moment
anything else is set on that domain — a theme preference, an analytics cookie, anything on a sibling
subdomain. The failure mode is "random logouts", which is expensive to diagnose and cheap to prevent.

## The fix

```go
func (h *HTTPAuthAdapter) GetJWTCookie(ctx context.Context, req connect.AnyRequest) string {
	cookies, err := http.ParseCookie(req.Header().Get("Cookie"))
	if err != nil {
		return ""
	}

	for _, cookie := range cookies {
		if cookie.Name == apiCookieName {
			return cookie.Value
		}
	}
	return ""
}
```

## Done when

`make api` builds, and with a real `jwt` value pasted from DevTools:

```bash
# passes before and after
curl -s -b "jwt=<token>" https://local.api.vulx.ai/api/v1/account/profile

# fails before, passes after — this is the regression test
curl -s -b "other=1; jwt=<token>; zz=2" https://local.api.vulx.ai/api/v1/account/profile
```

The second command returns the profile JSON instead of an `unauthenticated` error.

## Notes

- `GetJWTCookie` is called from two places — the interceptor via `AuthenticateWithJWT`, and
  `AccountLogout` in `handlers/account.go:57`. Both benefit; neither needs changing.
- Leave the `ctx` parameter alone even though it is unused. It matches the shape of the other
  methods on this adapter.
