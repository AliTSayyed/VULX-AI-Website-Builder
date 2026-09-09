# Task 6 — Verification

**Goal:** prove the things a happy path does not exercise. No code changes.

Source: `.planning/Backend/project-model.md` §11.

There is no test harness in this repo. This is manual, and the checks below are the ones that catch
the two bugs this feature is most likely to ship with.

## Setup

```bash
make nuke && make
# log in through the browser at https://local.app.vulx.ai, then copy the jwt cookie
export JWT=...
export API=https://local.api.vulx.ai
```

## 1. Unauthenticated is rejected — all three

```bash
curl -i -s -o /dev/null -w '%{http_code}\n' $API/api/v1/projects
curl -i -s -o /dev/null -w '%{http_code}\n' $API/api/v1/projects/00000000-0000-0000-0000-000000000000
curl -i -s -o /dev/null -w '%{http_code}\n' -X POST $API/api/v1/projects \
  -H 'content-type: application/json' -d '{"first_prompt":"x","provider":"AI_PROVIDER_ANTHROPIC"}'
```

**All three must be 401.** A 200 means `authAdapter.User(ctx)` is missing from that handler — there
is no allowlist and no interceptor that would have caught it.

## 2. Ownership — the check that leaks data if wrong

Log in as a **second Google account** in a private window and capture its cookie as `$JWT2`.

```bash
PID=$(curl -s -b "jwt=$JWT" -X POST $API/api/v1/projects -H 'content-type: application/json' \
  -d '{"first_prompt":"user one project","provider":"AI_PROVIDER_ANTHROPIC"}' | jq -r .project.id)

curl -s -o /dev/null -w '%{http_code}\n' -b "jwt=$JWT2" $API/api/v1/projects/$PID   # must be 404
curl -s -b "jwt=$JWT2" $API/api/v1/projects | jq '.projects | length'               # must be 0
```

**404, not 403, and not the project.** 403 would confirm the id exists.

## 3. Empty and whitespace prompts are rejected

```bash
for p in '""' '"   "'; do
  curl -s -o /dev/null -w "%{http_code}\n" -b "jwt=$JWT" -X POST $API/api/v1/projects \
    -H 'content-type: application/json' \
    -d "{\"first_prompt\":$p,\"provider\":\"AI_PROVIDER_ANTHROPIC\"}"
done
```

Both **400**. This is the mechanism that stops empty projects existing — if it returns 201, a user
who types nothing will litter the sidebar.

## 4. Unspecified provider is rejected, not defaulted

```bash
curl -s -o /dev/null -w '%{http_code}\n' -b "jwt=$JWT" -X POST $API/api/v1/projects \
  -H 'content-type: application/json' \
  -d '{"first_prompt":"x","provider":"AI_PROVIDER_UNSPECIFIED"}'
```

**400.** Also try omitting `provider` entirely — proto3 defaults it to zero, so it must also be 400.

## 5. Pagination actually pages

Create 25 projects, then:

```bash
curl -s -b "jwt=$JWT" "$API/api/v1/projects?limit=10" | jq '{n: (.projects|length), token, has_more}'
```

Expect `n: 10`, a non-empty `token`, `has_more: true`. Then follow the token:

```bash
curl -s -b "jwt=$JWT" "$API/api/v1/projects?limit=10&token=$TOKEN" | jq '.projects[].title'
```

**Check for overlap and gaps** — collect all ids across pages and confirm 25 unique. A repeat or a
missing row means the cursor column and the `ORDER BY` disagree (task 3, gotcha 1).

Also: `limit=0` must return 10, not 0 (the clamp).

## 6. Timestamp format

```bash
curl -s -b "jwt=$JWT" $API/api/v1/projects | jq -r '.projects[0].created_at'
```

Must look like `2026-09-08T14:03:11Z`. If it looks like `2026-09-08 14:03:11.123 +0000 UTC`, the
converter used `t.String()` and `new Date()` will parse it inconsistently in the browser.

## 7. Nothing generated was hand-edited

```bash
cd /Users/ats/Documents/Code-Projects/AI-Website-Builder && make gen && git status --short
```

`api/.../grpc/gen/`, `app/src/gen/` and `openapi.yaml` must show **no** diff after a regen. A diff
means someone edited generated code.

## Done when

All seven pass. Then move `task1.md`–`task6.md` into `completed/`.
