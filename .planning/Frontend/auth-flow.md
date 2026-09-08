# Auth Flow — Topology, Mechanics, Backend Prerequisites

Companion to `logged_out_screen.md`. Verified against the code on `feature/UI`.

**Task breakdown:** `.planning/tasks/Frontend/auth-flow/completed/` (all nine tasks landed).
**Backend side:** `.planning/Backend/auth-flow.md` — all three tasks also landed.

## 1. The round trip

```
 browser                                api                          google
 https://local.app.vulx.ai       https://local.api.vulx.ai
─────────────────────────       ─────────────────────────         ──────────
 click "Continue with Google"
   │
   ├─ BeginAccountAuth(GOOGLE) ──────────►
   │                              state = 32 random bytes
   │                              redis SET provider:<state>="google" (TTL 10m)
   │  ◄────────── { login_url } ──┘
   │
   ├─ window.location.assign(login_url) ──────────────────────────────►
   │                                                          consent screen
   │  ◄──── 302 to REDIRECT_URL?code=..&state=.. ──────────────────────┘
   │
 /auth/callback mounts
   │
   ├─ FinishAccountAuth({code, state}) ──►
   │                              redis GET provider:<state>  ← this IS the CSRF check
   │                              exchange code → google access token
   │                              GET /oauth2/v2/userinfo     ← rejects unverified emails
   │                              reconcile user + user_auth_providers
   │                              mint 7-day Ed25519 JWT
   │  ◄── { profile } + Set-Cookie: jwt=… ──┘
   │
   ├─ invalidateQueries(["profile"])
   └─ router.replace("/")  →  GetUserProfile succeeds  →  logged-in screen
```

Google's access token is used once to read the profile and is never stored. The JWT is the only
credential that persists.

## 2. Topology

### 2.1 Hosts

| Host | Serves | Via |
| --- | --- | --- |
| `https://local.app.vulx.ai` | Next dev server | Caddy → `app:3000`, `tls internal` |
| `https://local.api.vulx.ai` | Go API | Caddy → `api:8080`, `tls internal` |

Both resolve to `127.0.0.1` through `/etc/hosts`. Caddy reaches both upstreams by compose service
name on the shared `vulx` network. The direct host ports (`:3000`, `:8080`) stay published and
remain useful for `curl`, but **the browser must always use the Caddy hosts** — a login performed
against `http://localhost:8080` cannot set a `.vulx.ai` cookie.

### 2.2 Why the existing cookie code is already correct

`grpc/adapters/auth/auth.go:88` writes:

```
jwt=…; Expires=…; HttpOnly; Secure; SameSite=Lax; Domain=.vulx.ai; Path=/
```

| Attribute | Under this topology | Verdict |
| --- | --- | --- |
| `Domain=.vulx.ai` | Set by `local.api.vulx.ai`, which domain-matches `.vulx.ai` | ✅ accepted, and shared with the app host |
| `Secure` | Caddy terminates real TLS | ✅ |
| `SameSite=Lax` | Both hosts are under registrable domain `vulx.ai` → same-site | ✅ |
| `HttpOnly` | — | ✅ implies JS can never read the session |

**No Go change is needed here.** The `// TODO MUST UPDATE DOMAIN` comment above it applies to
production DNS, not to local development. This is the payoff of the two-host layout: every other
arrangement forces either a code change or a broken Google redirect URI.

For contrast, and so nobody re-derives this later:

- App on `http://localhost:3000` + API on `http://localhost:8080` → `Domain=.vulx.ai` is rejected
  (host `localhost` does not domain-match), so the cookie is silently dropped.
- App on `http://local.app.vulx.ai:3000` (no TLS) → `Secure` is rejected; browsers exempt the
  literal hostname `localhost`, not a hosts-file alias. Google also refuses to register an
  `http://` redirect URI on a non-localhost host.

### 2.3 The cookie parse bug — fixed

`GetJWTCookie` used to reset `token` to `""` on every non-`jwt` cookie in the header, so the session
survived only if `jwt` happened to be parsed last. Less likely to bite on a dedicated subdomain than
on shared `localhost`, but any second cookie scoped to `.vulx.ai` — a theme preference, an analytics
cookie, anything on a sibling subdomain — would have reintroduced it as random logouts. Fixed by
`tasks/Backend/auth-flow/task1.md` (now in `completed/`) — the loop returns on the first match
instead of continuing:

```go
for _, cookie := range cookies {
    if cookie.Name == apiCookieName {
        return cookie.Value
    }
}
return ""
```

## 3. Setup

### 3.1 `Caddyfile`

```caddyfile
{
    admin off
}

local.api.vulx.ai {
    reverse_proxy api:8080
    tls internal
}

local.app.vulx.ai {
    reverse_proxy app:3000
    tls internal
}
```

Caddy proxies WebSocket upgrades automatically, so Next's HMR socket works through it unchanged.
`api` currently declares `depends_on: caddy`, so Caddy is already up first; it does not need a
`depends_on: app` because `reverse_proxy` retries per request rather than resolving at boot.

### 3.2 `/etc/hosts`

```
127.0.0.1 local.api.vulx.ai
127.0.0.1 local.app.vulx.ai
```

Remove the old `127.0.0.1 local.vulx.ai` line once nothing references it.

### 3.3 Trust Caddy's internal CA

Without this, Google's redirect lands on a browser interstitial and the round trip dies where it
is hardest to debug. Pull the root out of the `caddydata` volume:

```bash
docker compose cp caddy:/data/caddy/pki/authorities/local/root.crt ./caddy-root.crt
```

macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain ./caddy-root.crt
```

Linux (Debian/Ubuntu):

```bash
sudo cp caddy-root.crt /usr/local/share/ca-certificates/caddy-root.crt
sudo update-ca-certificates
```

Firefox uses its own store and needs the cert imported under Settings → Privacy & Security →
Certificates. Delete `caddy-root.crt` afterwards; do not commit it. The CA is regenerated by
`make nuke` (it drops the `caddydata` volume), so this has to be redone after a nuke.

### 3.4 `../AI-Website-Builder-Secrets/.api-env`

```
APP_URL=https://local.app.vulx.ai
API_URL=https://local.api.vulx.ai
REDIRECT_URL=https://local.app.vulx.ai/auth/callback
```

`APP_URL` and `API_URL` are the entire CORS allowlist in `SecurityAdapterCors`, which already sets
`AllowCredentials: true` against explicit origins — correct for credentialed requests, and the
reason a wildcard origin must never be introduced.

Optional tidy-up: `docker-compose.yaml` hardcodes `TEMPORAL_CORS_ORIGINS=http://localhost:3000` on
`temporal-ui`. Cosmetic only, unrelated to this feature.

### 3.5 Google Cloud Console

Add to the OAuth client's **Authorized redirect URIs**:

```
https://local.app.vulx.ai/auth/callback
```

Byte-identical to `REDIRECT_URL` — a trailing-slash mismatch produces `redirect_uri_mismatch` at
the consent screen, before any of our code runs. HTTPS on a domain you own satisfies Google's
[URI validation rules](https://developers.google.com/identity/protocols/oauth2/web-server); the
`http://` exemption covers only literal `localhost` and loopback IPs, which is why a hosts-file
alias over plain HTTP is not an option.

No **Authorized JavaScript origins** entry is needed — we never load Google's JS SDK; the browser
only follows a redirect.

### 3.6 `app/.env.local`

```
NEXT_PUBLIC_API_URL=https://local.api.vulx.ai
```

`NEXT_PUBLIC_*` is inlined at build time. The app source is bind-mounted for hot reload, so a new
value needs a dev-server restart (`docker compose restart app`), not an image rebuild.

## 4. The callback route

`app/src/app/auth/callback/page.tsx`.

### 4.1 Inputs to handle

| Query | Meaning | Render |
| --- | --- | --- |
| `code` + `state` | Normal success path | Spinner, then redirect to `/` |
| `error=access_denied` | User pressed Cancel at Google | "Sign-in cancelled" + link home |
| `error=<other>` | Google-side failure | The raw error + link home |
| neither | Route hit directly | "Invalid sign-in link" + link home |

### 4.2 Suspense

`useSearchParams()` opts the route into client rendering and Next 15 requires a `<Suspense>`
boundary above it, or `npm run build` fails with a prerender error. Keep the search-param reader
in a small inner component and wrap it.

### 4.3 The StrictMode double-fire trap

React 19 StrictMode invokes effects twice in development. The `useRef` latch below landed alongside
the backend's Redis-delete fix (`tasks/Backend/auth-flow/task3.md`), in the correct order — the
latch first, so the second StrictMode-fired effect never reaches `FinishAccountAuth` at all and
never has the chance to fail on an already-consumed `state`.

```ts
const fired = useRef(false);
useEffect(() => {
  if (fired.current) return;
  fired.current = true;
  // … FinishAccountAuth
}, []);
```

### 4.4 On success

`invalidateQueries({ queryKey: ["profile"] })`, replay any stashed hero draft, then
`router.replace("/")`.

`FinishAccountAuthResponse` already carries the `Profile`, so `setQueryData` would save a round
trip — but it is a different generated type from `GetUserProfileResponse` and needs a deliberate
unwrap to the inner `Profile`. Invalidate first; optimise only if the flash is visible.

## 5. Backend fixes

All three landed — `.planning/tasks/Backend/auth-flow/` is now `completed/task{1,2,3}.md`. Detail in
`.planning/Backend/auth-flow.md`.

| Task | Fix | Relation to this document | Status |
| --- | --- | --- | --- |
| 1 | `GetJWTCookie` keeps only the last cookie | the defect described in §2.3 above | ✅ landed |
| 2 | Interceptor flattened authenticated error codes to `CodeInvalidArgument` | not blocking; §1.2 of the backend doc explains why logged-out detection still worked | ✅ landed — implemented as a client-caused-vs-server-caused split rather than passing every code through unwrapped, so infrastructure errors (Redis/DB failures) still don't leak their raw message to the browser |
| 3 | OAuth state was never deleted, so it was replayable for 10 minutes | **ordering constraint** — pairs with the `useRef` latch in §4.3 | ✅ landed, latch-first as required |

## 6. Frontend transport — fixed

`useServiceClient.ts` used to hardcode `http://localhost:8080` and pass no credentials, so the
browser would neither store nor send the cookie. Fixed by `tasks/Frontend/auth-flow/task2.md`:

```ts
const baseUrl = process.env.NEXT_PUBLIC_API_URL ?? "https://local.api.vulx.ai";

createConnectTransport({
  baseUrl,
  useBinaryFormat: false,
  fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
});
```

`credentials: "include"` is required: `local.app.vulx.ai` → `local.api.vulx.ai` is same-*site* but
cross-*origin*, and `fetch` omits cookies cross-origin by default. Keep the `useMemo` on the
transport — recreating it per render defeats connection reuse.

## 7. Manual verification

```bash
# TLS and routing — no -k flag; if this needs -k, the CA is not trusted (§3.3)
curl -s https://local.api.vulx.ai/healthz
curl -sI https://local.app.vulx.ai | head -1

# session usable (paste the cookie from DevTools after a browser login)
curl -s -b "jwt=<token>" -H 'Content-Type: application/json' -d '{}' \
  https://local.api.vulx.ai/api/v1/account/profile

# the §2.3 regression — jwt must survive not being last
curl -s -b "other=1; jwt=<token>; zz=2" -H 'Content-Type: application/json' -d '{}' \
  https://local.api.vulx.ai/api/v1/account/profile
```

The last one fails on today's code and passes after the `GetJWTCookie` fix. That route is also
transcoded by Vanguard, so a plain REST `GET` works if you prefer it to the POST form.
