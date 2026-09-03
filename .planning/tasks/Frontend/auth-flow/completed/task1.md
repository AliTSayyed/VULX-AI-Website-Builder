# Task 1 — Infrastructure: two hosts behind Caddy

**Goal.** `https://local.app.vulx.ai` serves the Next dev server and `https://local.api.vulx.ai`
serves the Go API, both with a green padlock, both on the shared `vulx` network.

Most of this is configuration and several steps happen outside the repo. Nothing later works without
it: the cookie needs a real HTTPS origin under `vulx.ai`, and Google will not accept a redirect URI
that is neither HTTPS nor literal `localhost`.

---

## 1. `Caddyfile`

Replace the whole file:

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

Both upstreams resolve by compose service name. Caddy proxies WebSocket upgrades automatically, so
Next's HMR socket works through it unchanged. Caddy needs no `depends_on: app` — `reverse_proxy`
resolves per request, not at boot.

## 2. ⚠️ `make app` breaks service-name DNS — verify this first

`makefile` runs `docker compose run --build --service-ports --rm app`. **`docker compose run` does
not attach the service's network aliases to the container it creates** unless you pass
`--use-aliases` (see `docker compose run --help`). Dependencies started by `depends_on` are created
normally and keep their aliases; only the `run` target loses them.

So with the makefile as written, Caddy's `reverse_proxy app:3000` may fail to resolve `app`, and
`https://local.app.vulx.ai` returns a 502 while `http://localhost:3000` works fine. The same applies
to `make api` and `reverse_proxy api:8080`, which is pre-existing.

**Check first** — if this is already working on your machine, change nothing:

```bash
docker compose exec caddy wget -qO- http://app:3000 >/dev/null && echo "app resolves"
```

If it does not resolve, pick one:

```makefile
# option A — keep `run`, restore the alias
app: clean
	@docker compose run --build --service-ports --use-aliases --rm app
```

```makefile
# option B — use `up`, which always aliases
app: clean
	@docker compose up --build app
```

Option A is the smaller diff and keeps the existing behaviour (`--rm`, interactive TTY). Option B
also gives you the api's logs in the same stream, which is useful while debugging the round trip.

Note this in the commit message either way — it is a non-obvious footgun and the next person will
hit it.

## 3. `/etc/hosts`

```
127.0.0.1 local.api.vulx.ai
127.0.0.1 local.app.vulx.ai
```

Remove the old `127.0.0.1 local.vulx.ai` line once nothing references it.

## 4. Trust Caddy's internal CA

Skipping this is the single most expensive mistake available here: Google's redirect lands on a
certificate interstitial and the round trip dies at the point that is hardest to read.

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

Firefox keeps its own store — import under Settings → Privacy & Security → Certificates. Delete
`caddy-root.crt` afterwards and do not commit it. `make nuke` drops the `caddydata` volume and
regenerates the CA, so this has to be redone after a nuke.

## 5. `../AI-Website-Builder-Secrets/.api-env`

```
APP_URL=https://local.app.vulx.ai
API_URL=https://local.api.vulx.ai
REDIRECT_URL=https://local.app.vulx.ai/auth/callback
```

`APP_URL` and `API_URL` are the entire CORS allowlist in `SecurityAdapterCors`, which already sets
`AllowCredentials: true` against explicit origins. That is correct for credentialed requests and the
reason a wildcard origin must never be introduced here.

`config.go` defaults these to `http://localhost:{8080,3000}` if unset — the app will start and then
fail CORS in a way that reads like a frontend bug, so set them explicitly.

## 6. `app/.env.local` (new file)

```
NEXT_PUBLIC_API_URL=https://local.api.vulx.ai
```

`app/.gitignore` already ignores `.env*`, so this stays local — mention it in the PR description so
the next person knows to create one.

`NEXT_PUBLIC_*` is inlined at build time. The source is bind-mounted for hot reload, so a changed
value needs `docker compose restart app`, not an image rebuild.

## 7. Google Cloud Console

Add to the OAuth client's **Authorized redirect URIs**:

```
https://local.app.vulx.ai/auth/callback
```

Byte-identical to `REDIRECT_URL` — a trailing-slash mismatch produces `redirect_uri_mismatch` at the
consent screen, before any of our code runs. No **Authorized JavaScript origins** entry is needed;
we never load Google's JS SDK, the browser only follows a redirect.

## 8. `app/next.config.ts` — only if Next complains

Next 15 warns when a dev request arrives with a cross-origin `Host`. If it appears in the app logs:

```ts
const nextConfig: NextConfig = {
  allowedDevOrigins: ["local.app.vulx.ai"],
};
```

Do not add this pre-emptively. If the warning never shows, the file stays untouched.

---

## Done when

```bash
# no -k flag. If it needs -k, step 4 is not done.
curl -s  https://local.api.vulx.ai/healthz          # 200
curl -sI https://local.app.vulx.ai | head -1        # HTTP/2 200
```

- Both hosts load in the browser with a padlock and no interstitial.
- Editing `app/src/components/landing/hero.tsx` hot-reloads through
  `https://local.app.vulx.ai` without a manual refresh — this proves the HMR WebSocket survives
  the proxy.
- The old `local.vulx.ai` host is gone from the Caddyfile and from `/etc/hosts`.

## Notes

- The direct ports (`:3000`, `:8080`) stay published and are still useful for `curl`. **The browser
  must always use the Caddy hosts** — a login performed against `http://localhost:8080` cannot set a
  `.vulx.ai` cookie, and it fails silently.
- `docker-compose.yaml` hardcodes `TEMPORAL_CORS_ORIGINS=http://localhost:3000` on `temporal-ui`.
  Cosmetic, unrelated, leave it.
