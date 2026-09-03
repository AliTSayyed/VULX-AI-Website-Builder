"use client";

import { ArrowUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  PromptInput,
  PromptInputActions,
  PromptInputTextarea,
} from "@/components/ui/prompt-input";

type HeroPromptProps = {
  onAuth: (mode: "login" | "signup") => void;
};

export function HeroPrompt({ onAuth }: HeroPromptProps) {
  const open = () => onAuth("signup");

  return (
    <PromptInput
      value=""
      onValueChange={() => {}}
      onSubmit={open}
      className="bg-card border-border mt-10 w-full max-w-3xl rounded-3xl p-3"
    >
      <PromptInputTextarea
        readOnly
        placeholder="Ask VULX to build..."
        aria-label="Sign up to start building"
        onClick={open}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            open();
          }
        }}
        className="text-foreground placeholder:text-muted-foreground min-h-32 cursor-pointer bg-transparent text-base dark:bg-transparent"
      />

      <PromptInputActions className="justify-end pt-1">
        <Button
          size="icon"
          className="size-9 rounded-full"
          onClick={open}
          aria-label="Sign up to start building"
        >
          <ArrowUp className="size-4" />
        </Button>
      </PromptInputActions>
    </PromptInput>
  );
}