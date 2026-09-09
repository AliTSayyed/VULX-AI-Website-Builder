# Task 6 — End to end: prompt → sandbox → rendered site

**Goal:** the demo. No code changes.

Source: `.planning/MVP.md` "Done when", `.planning/Backend/messages-model.md` §12.

## Setup

```bash
make nuke && make
# log in at https://local.app.vulx.ai, copy the jwt cookie
export JWT=... API=https://local.api.vulx.ai
```

## 1. The demo path

```bash
PID=$(curl -s -b "jwt=$JWT" -X POST $API/api/v1/projects \
  -H 'content-type: application/json' \
  -d '{"first_prompt":"a landing page for a coffee shop","provider":"AI_PROVIDER_ANTHROPIC"}' \
  | jq -r .project.id)
echo "$PID"

time curl -s -b "jwt=$JWT" -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' \
  -d '{"body":"a landing page for a coffee shop with a hero, menu and contact section","mode":"CHAT_MODE_BUILD","provider":"AI_PROVIDER_ANTHROPIC"}' \
  | jq '{user: .userMessage.body, summary: .assistantMessage.body}'

curl -s -b "jwt=$JWT" $API/api/v1/projects/$PID | jq -r .project.previewUrl
```

**Open that URL.** A coffee shop landing page must render. That is the MVP.

Note the elapsed time from `time` — if it is over 30 seconds and it succeeded, the timeout fix is
confirmed working under real load.

**JSON field names over Connect are camelCase** (`previewUrl`, `assistantMessage`); over REST with
the OpenAPI transcoding they may appear snake_case. Check both rather than assuming.

## 2. A second message reuses the same sandbox — the important one

```bash
SB1=$(curl -s -b "jwt=$JWT" $API/api/v1/projects/$PID | jq -r .project.sandboxId)

curl -s -b "jwt=$JWT" -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' \
  -d '{"body":"make the hero heading dark green","mode":"CHAT_MODE_BUILD","provider":"AI_PROVIDER_ANTHROPIC"}' > /dev/null

SB2=$(curl -s -b "jwt=$JWT" $API/api/v1/projects/$PID | jq -r .project.sandboxId)
[ "$SB1" = "$SB2" ] && echo "SAME sandbox — correct" || echo "BUG: a second sandbox was created"
```

**Different ids mean the `if sandboxID == ""` guard is wrong**, and every message is starting from an
empty template — the first build's work is silently abandoned. Reload the preview URL: the coffee
shop must still be there, now with a green heading.

Expect the heading change to be applied but the *conversation* not to be understood — the agent has
no memory between messages, so phrase follow-ups self-containedly. That is a known MVP limitation,
not a bug.

## 3. The project moved up the sidebar

```bash
curl -s -b "jwt=$JWT" $API/api/v1/projects | jq -r '.projects[0].id'
```

Must be `$PID` — the CTE bumped `projects.updated_at`. If an older project is first, the `bumped`
sub-statement is not running (task 3).

## 4. Unauthenticated is rejected

```bash
curl -s -o /dev/null -w '%{http_code}\n' $API/api/v1/projects/$PID/messages                    # 401
curl -s -o /dev/null -w '%{http_code}\n' -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' -d '{"body":"x","mode":"CHAT_MODE_BUILD","provider":"AI_PROVIDER_ANTHROPIC"}'   # 401
```

## 5. Another user's project

With a second account's cookie in `$JWT2`:

```bash
curl -s -o /dev/null -w '%{http_code}\n' -b "jwt=$JWT2" $API/api/v1/projects/$PID/messages     # 404
```

**404, not 403, and not the thread.**

## 6. Validation

```bash
# empty body
curl -s -o /dev/null -w '%{http_code}\n' -b "jwt=$JWT" -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' -d '{"body":"   ","mode":"CHAT_MODE_BUILD","provider":"AI_PROVIDER_ANTHROPIC"}'   # 400

# unspecified mode
curl -s -o /dev/null -w '%{http_code}\n' -b "jwt=$JWT" -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' -d '{"body":"x","provider":"AI_PROVIDER_ANTHROPIC"}'   # 400
```

## 7. The AI service being down does not lose the user's message

```bash
docker compose stop ai-service
curl -s -o /dev/null -w '%{http_code}\n' -b "jwt=$JWT" -X POST $API/api/v1/projects/$PID/messages \
  -H 'content-type: application/json' -d '{"body":"anything","mode":"CHAT_MODE_BUILD","provider":"AI_PROVIDER_ANTHROPIC"}'
docker compose logs api | tail -20
curl -s -b "jwt=$JWT" $API/api/v1/projects/$PID/messages | jq -r '.messages[-1].body'
docker compose start ai-service
```

The last message must be **"anything"** — the user's message survives. And the API logs must show a
clean `Unavailable`, **not a nil-pointer panic** (`ai-service-client/task2.md`, gotcha 2).

## 8. The whole thread reads back in order

```bash
curl -s -b "jwt=$JWT" $API/api/v1/projects/$PID/messages | jq -r '.messages[] | "\(.role) \(.createdAt) \(.body[0:60])"'
```

Oldest first, alternating user/assistant except where step 7 left an unanswered one.

## Done when

All eight pass, and **the preview URL renders the site you asked for**. Then move `task1.md`–
`task6.md` into `completed/`, and do the same for the other five task directories.

At that point `.planning/MVP.md` is satisfied and the only thing between you and the demo is the
frontend.
