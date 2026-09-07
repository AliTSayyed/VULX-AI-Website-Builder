"use client";

import { ArrowUp } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  PromptInput,
  PromptInputActions,
  PromptInputTextarea,
} from "@/components/ui/prompt-input";
import { PROMPT_BOX, PROMPT_TEXTAREA } from "@/components/prompt/prompt-styles";
import { cn } from "@/lib/utils";

type HomeViewProps = {
  name: string;
  onStart: (prompt: string) => void;
};

/*
 * The logged-out hero's prompt box, minus the devices that only make sense to a
 * first-time visitor: no TypingText cycle, a real placeholder, and the textarea
 * is writable. A returning user does not need to be told what the box is for.
 */
export function HomeView({ name, onStart }: HomeViewProps) {
  const [value, setValue] = useState("");

  const submit = () => {
    const trimmed = value.trim();
    if (trimmed) onStart(trimmed);
  };

  return (
    <div className="flex flex-1 flex-col items-center justify-center px-5 pb-12 sm:px-6">
      <h1 className="vx-rise text-center text-balance">
        <span className="text-foreground-dim block text-4xl font-normal tracking-tight sm:text-5xl">
          Welcome back,
        </span>
        <strong className="text-foreground mt-1 block text-5xl font-medium tracking-tight sm:text-6xl">
          {name}
        </strong>
      </h1>

      <PromptInput
        value={value}
        onValueChange={setValue}
        onSubmit={submit}
        className={cn("vx-rise mt-9 max-w-2xl", PROMPT_BOX, "border-accent-blue")}
        style={{ animationDelay: "200ms" }}
      >
        <PromptInputTextarea
          placeholder="Describe the site you want to build..."
          className={`placeholder:text-muted-foreground ${PROMPT_TEXTAREA}`}
        />

        <PromptInputActions className="justify-end pt-1">
          <Button
            size="icon"
            className="size-9 rounded-full"
            onClick={submit}
            disabled={!value.trim()}
            aria-label="Start building"
          >
            <ArrowUp className="size-4" />
          </Button>
        </PromptInputActions>
      </PromptInput>
    </div>
  );
}
