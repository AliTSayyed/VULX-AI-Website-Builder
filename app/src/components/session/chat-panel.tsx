"use client";

import { ArrowLeft, ArrowUp, Hammer, MessageSquare } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  PromptInput,
  PromptInputActions,
  PromptInputTextarea,
} from "@/components/ui/prompt-input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { PROMPT_BOX } from "@/components/prompt/prompt-styles";
import { cn } from "@/lib/utils";
import type { Conversation } from "./mock";
import { AiProvider, ChatMode } from "@/gen/api/v1/enums_pb";
import { ComposerControls } from "./composer-controls";
import { DEFAULT_PROVIDER } from "./providers";

type ChatPanelProps = {
  conversation: Conversation;
  onBack: () => void;
};

export function ChatPanel({ conversation, onBack }: ChatPanelProps) {
  const [value, setValue] = useState("");
  const [provider, setProvider] = useState<AiProvider>(DEFAULT_PROVIDER);
  const [mode, setMode] = useState<ChatMode>(ChatMode.BUILD);

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex h-12 shrink-0 items-center gap-0.5 px-2">
        <Button
          variant="ghost"
          size="icon"
          onClick={onBack}
          aria-label="Back"
          className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-8 shrink-0 rounded-full"
        >
          <ArrowLeft className="size-3.5" />
        </Button>
        <span className="text-foreground truncate text-sm font-medium">
          {conversation.title}
        </span>
      </div>

      <ScrollArea className="min-h-0 flex-1">
        <div className="flex flex-col gap-4 p-3">
          {conversation.messages.length === 0 && (
            <p className="text-muted-foreground px-1 py-8 text-center text-xs">
              No messages yet.
            </p>
          )}

          {conversation.messages.map((m) =>
            m.role === "user" ? (
              <div key={m.id} className="flex justify-end">
                <p className="bg-surface-2 text-foreground max-w-[85%] rounded-2xl px-3 py-2 text-[13px] leading-relaxed">
                  {m.body}
                </p>
              </div>
            ) : (
              <div key={m.id} className="flex flex-col gap-1.5">
                <span className="text-muted-foreground flex items-center gap-1.5 px-1 text-[11px]">
                  {m.mode === "build" ? (
                    <Hammer className="size-3 text-accent-blue" />
                  ) : (
                    <MessageSquare className="size-3 text-accent-blue" />
                  )}
                  {m.mode === "build" ? "Build" : "Chat"}
                </span>
                <p className="text-foreground-dim px-1 text-[13px] leading-relaxed">
                  {m.body}
                </p>
              </div>
            ),
          )}
        </div>
      </ScrollArea>

      <div className="shrink-0 p-3">
        <PromptInput
          value={value}
          onValueChange={setValue}
          onSubmit={() => setValue("")}
          className={cn(PROMPT_BOX, "border-accent-blue")}
        >
          <PromptInputTextarea
            placeholder={
              mode === ChatMode.BUILD
                ? "Describe a change..."
                : "Ask a question..."
            }
            className="text-foreground placeholder:text-muted-foreground min-h-16 bg-transparent text-[13px] leading-5 dark:bg-transparent"
          />

          <PromptInputActions className="justify-between gap-2 pt-1">
            {/*
             * Mode is the consequential choice — Build writes to the codebase and
             * spends credits, Chat does neither — so it gets a segmented toggle
             * rather than hiding inside a dropdown. See logged_in_design.md §3.
             */}
            <ComposerControls
              mode={mode}
              onModeChange={setMode}
              provider={provider}
              onProviderChange={setProvider}
            />

            <Button
              size="icon"
              className="size-8 shrink-0 rounded-full"
              disabled={!value.trim()}
              onClick={() => setValue("")}
              aria-label="Send message"
            >
              <ArrowUp className="size-3.5" />
            </Button>
          </PromptInputActions>
        </PromptInput>
      </div>
    </div>
  );
}
