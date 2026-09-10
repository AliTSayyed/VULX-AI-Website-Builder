# Task 1 — `MessageRole` and `ChatMode` enums

**Goal:** the two message enums, generating short TypeScript names.

Source: `.planning/Proto/messages-proto.md` §2.

## Files

| File | Change |
| --- | --- |
| `proto/api/v1/enums.proto` | append two enums |

## The change

Append below `LoginProvider` and `AiProvider`:

```proto
enum MessageRole {
  MESSAGE_ROLE_UNSPECIFIED = 0;
  MESSAGE_ROLE_USER = 1;
  MESSAGE_ROLE_ASSISTANT = 2;
}

enum ChatMode {
  CHAT_MODE_UNSPECIFIED = 0;
  CHAT_MODE_CHAT = 1;
  CHAT_MODE_BUILD = 2;
}
```

## Notes

**Neither name hits the consecutive-capitals trap.** `MessageRole` → `MESSAGE_ROLE_` and `ChatMode`
→ `CHAT_MODE_`, both of which match their values, so `protoc-gen-es` strips the prefix. Unlike
`AiProvider`, there was no naming decision to make here.

**`CHAT_MODE_CHAT` is a stutter and is unavoidable** — buf lint's prefix rule applies to every value,
including the one sharing the enum's name.

**`CHAT_MODE_CHAT` ships even though the MVP never sends it.** Enum numbers are permanent under
`FILE` breaking detection, so defining both values now costs nothing and avoids renumbering when Chat
mode lands.

**Both zero values are rejected inputs, not defaults.** `SendMessage` returns `InvalidArgument` for
either — the composer always has a concrete selection, so unspecified means a client bug and
defaulting would hide it. That enforcement is in the Go service, not here.

## Run it

```bash
make plint && make gen
```

## Verifying

```bash
grep -n "ASSISTANT\|BUILD" app/src/gen/api/v1/enums_pb.ts
```

Expect `ASSISTANT = 2` and `BUILD = 2` — short forms. If you see `MESSAGE_ROLE_ASSISTANT`, the enum
name and its value prefix disagree; fix the proto rather than the frontend.

```bash
cd app && npm run lint && npm run build
```

## Done when

- `MessageRole.ASSISTANT` and `ChatMode.BUILD` exist in the generated TS, short form.
- `npm run build` passes.
