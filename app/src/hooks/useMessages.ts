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
