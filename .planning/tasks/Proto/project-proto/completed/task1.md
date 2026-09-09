# Task 1 — `AiProvider` enum

**Goal:** add the provider enum and prove it generates the right TypeScript before anything depends
on it.

Source: `.planning/Proto/project-proto.md` §2.3, §3.

## Files

| File | Change |
| --- | --- |
| `proto/api/v1/enums.proto` | append one enum |

## The change

Append below the existing `LoginProvider`:

```proto
enum AiProvider {
  AI_PROVIDER_UNSPECIFIED = 0;
  AI_PROVIDER_OPENAI = 1;
  AI_PROVIDER_GOOGLE = 2;
  AI_PROVIDER_ANTHROPIC = 3;
}
```

## Three things that are not free choices

**The enum is `AiProvider`, not `AIProvider`.** `@bufbuild/protobuf`'s `camelToSnakeCase` inserts an
underscore before every capital, so `AIProvider` expects values prefixed `A_I_PROVIDER_`. Values
named `AI_PROVIDER_*` would then **not** be stripped, and the frontend would write
`AIProvider.AI_PROVIDER_OPENAI`. `AiProvider` expects `AI_PROVIDER_` and strips cleanly to
`AiProvider.OPENAI`.

`buf lint` may accept either spelling — its snake-caser is a separate implementation — so lint
passing is **not** evidence this is right. Step 3 below is the evidence.

**The three value names are AI-service URL path segments.** `openai`, `google`, `anthropic` are the
FastAPI router prefixes; the service rejects `gemini` and `claude`. Product labels ("Gemini",
"Claude") belong in the frontend only.

**`AI_PROVIDER_UNSPECIFIED` exists because buf lint requires a zero value**, not because it is a
usable input. Every RPC that takes a provider rejects it.

## Run it

```bash
make plint     # clang-format; run before committing or the next run churns
make gen
```

`make gen` needs `app/node_modules` present — it resolves `protoc-gen-es` from there.

## Verifying — this is the point of the task

```bash
grep -n "OPENAI\|ANTHROPIC\|UNSPECIFIED" app/src/gen/api/v1/enums_pb.ts
```

**Expect:**

```ts
export enum AiProvider {
  UNSPECIFIED = 0,
  OPENAI = 1,
  GOOGLE = 2,
  ANTHROPIC = 3,
}
```

**If you see `AI_PROVIDER_OPENAI = 1`**, the enum was named `AIProvider`. Rename it to `AiProvider`
in the proto and regenerate. Do not work around it in the frontend.

Then confirm the Go side exists:

```bash
grep -rn "AiProvider" api/internal/infrastructure/inbound/grpc/gen/api/v1/enums.pb.go | head
```

And that the frontend still builds, since `make gen` rewrote `app/src/gen/`:

```bash
cd app && npm run lint && npm run build
```

## Done when

- `AiProvider.OPENAI` (short form) appears in `app/src/gen/api/v1/enums_pb.ts`.
- `apiv1.AiProvider` exists in the generated Go.
- `npm run build` passes.
