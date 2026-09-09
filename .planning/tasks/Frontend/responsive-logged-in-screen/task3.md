# Task 3 — `useMessages.ts`, and the send mutation

**Goal.** The thread query, and the one mutation the whole screen turns on. This is where the
preview URL problem is solved and where the "generating" boolean comes from.

## Create `app/src/hooks/useMessages.ts`

```ts
"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  MessageSchema,
  type Message,
} from "@/gen/api/v1/message_service_pb";
import { AiProvider, ChatMode, MessageRole } from "@/gen/api/v1/enums_pb";
import { useMessageService } from "./services/useMessageService";

export function useMessages(projectId: string | null) {
  const messages = useMessageService();

  return useQuery<Message[]>({
    queryKey: ["messages", projectId],
    enabled: !!projectId,
    // A thread must be current the moment it opens; the 30s default would show
    // a stale thread after a send that happened in another tab.
    staleTime: 0,
    queryFn: async () => {
      const res = await messages.listMessages({ projectId: projectId! });
      return res.messages;
    },
  });
}

type SendInput = {
  projectId: string;
  body: string;
  mode: ChatMode;
  provider: AiProvider;
};

export function useSendMessage() {
  const messages = useMessageService();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: SendInput) => messages.sendMessage(input),

    // The user's own words appear instantly. Waiting minutes to see what you
    // just typed reads as a dropped input.
    onMutate: async (input) => {
      const key = ["messages", input.projectId];
      await queryClient.cancelQueries({ queryKey: key });
      const previous = queryClient.getQueryData<Message[]>(key);

      const optimistic = create(MessageSchema, {
        id: `optimistic-${Date.now()}`,
        projectId: input.projectId,
        role: MessageRole.USER,
        mode: input.mode,
        provider: input.provider,
        body: input.body,
        createdAt: new Date().toISOString(),
      });

      queryClient.setQueryData<Message[]>(key, [...(previous ?? []), optimistic]);
      return { previous };
    },

    onError: (_err, input, context) => {
      queryClient.setQueryData(
        ["messages", input.projectId],
        context?.previous ?? [],
      );
    },

    onSettled: (_data, _err, input) => {
      // 1. the persisted pair replaces the optimistic bubble
      queryClient.invalidateQueries({ queryKey: ["messages", input.projectId] });
      // 2. THIS is how the preview URL arrives — SendMessageResponse has no
      //    sandbox on it; the sandbox lives on the project.
      queryClient.invalidateQueries({ queryKey: ["project", input.projectId] });
      // 3. the insert bumped projects.updated_at, so the sidebar reorders
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}
```

## Points that matter

- **`create(MessageSchema, {...})`, not an object literal.** Generated message types carry a
  `$typeName` field; a bare literal will not typecheck as `Message`.
- **The optimistic id is prefixed.** The thread renders `key={m.id}`; a real UUID collision is
  impossible, but the prefix also lets the thread tell an unconfirmed bubble from a persisted one if
  it ever needs to.
- **Invalidating the project is not optional.** Skip it and the first build finishes, the assistant
  reply lands, and the preview stays empty forever. It is the single most likely way this feature
  ships half-working. Documented in `Polish.md` → *Known runtime trade-offs*.
- **`onSettled`, not `onSuccess`,** for the invalidations: a Build that failed halfway may still have
  persisted the user message and created the sandbox, so the caches are suspect either way.
- **No `timeoutMs` on the call.** `SendMessage(BUILD)` blocks for the whole agent run — minutes.
  The Go client's own backstop is 15 minutes (`outbound/ai_service`); the browser must not undercut
  it.
- **The mutation is not fired from `ChatPanel`.** `LoggedInScreen` owns it (task 10), because a
  draft has to run `CreateProject` first and the two must share one pending flag.
- `mutation.isPending` is the *only* source of "generating" — the thread shimmer and the preview
  overlay both read it, so they cannot disagree.

## On failure

No designed error surface is in scope (doc §8). But `onError` must restore the cache, and task 10
must toast. A rejected send that leaves the shimmer running is worse than a visible failure.

## Done when

- `npm run lint` and `npm run build` pass.
- Nothing imports the hooks yet.
