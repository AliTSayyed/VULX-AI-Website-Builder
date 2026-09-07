"use client";

import { ArrowUp } from "lucide-react";
import {
  TypingText,
  TypingTextCursor,
} from "@/components/animate-ui/primitives/texts/typing";
import { Button } from "@/components/ui/button";
import { usePrefersReducedMotion } from "@/hooks/use-prefers-reduced-motion";
import {
  PromptInput,
  PromptInputActions,
  PromptInputTextarea,
} from "@/components/ui/prompt-input";

type HeroPromptProps = {
  onAuth: (mode: "login" | "signup") => void;
};

/*
 * Deliberately distinct from SUGGESTIONS in suggestion-chips.tsx — the chips sit
 * directly below this input, and typing the same three strings there reads as a bug.
 */
const PROMPTS = [
  "a landing page for my startup",
  "a portfolio with a photo gallery",
  "a booking page for my studio",
  "a docs site with search",
];

export function HeroPrompt({ onAuth }: HeroPromptProps) {
  const open = () => onAuth("signup");
  const reducedMotion = usePrefersReducedMotion();

  return (
    <PromptInput
      value=""
      onValueChange={() => {}}
      onSubmit={open}
      className="border-hairline bg-surface vx-rise mt-9 w-full max-w-2xl rounded-3xl p-3 shadow-none"
      style={{ animationDelay: "400ms" }}
    >
      {/*
       * The prompt hint types itself, so it cannot be a native placeholder —
       * that would render underneath the animation. The textarea drops its
       * placeholder and keeps aria-label as its accessible name; the overlay
       * is aria-hidden and pointer-events-none so clicks still reach the
       * textarea and open the auth dialog. top-2/left-3 and leading-6 mirror
       * the textarea's own padding and line box so the two align exactly, and
       * right-3 lets a long prompt wrap inside the box instead of overflowing.
       *
       * Only the tail cycles — "Ask VULX to build" stays put so the hint is
       * always readable. TypingText loops forever, which is what
       * prefers-reduced-motion exists for, so that case renders one static
       * prompt instead.
       */}
      <div className="relative">
        <PromptInputTextarea
          readOnly
          aria-label="Sign up to start building"
          onClick={open}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              open();
            }
          }}
          className="text-foreground min-h-21 cursor-pointer bg-transparent text-[15px] leading-6 dark:bg-transparent"
        />

        <span
          aria-hidden
          className="text-muted-foreground pointer-events-none absolute top-2 right-3 left-3 text-[15px] leading-6"
        >
          Ask VULX to build{" "}
          {reducedMotion ? (
            PROMPTS[0]
          ) : (
            <TypingText
              text={PROMPTS}
              loop
              duration={45}
              delay={900}
              holdDelay={1800}
            >
              <TypingTextCursor
                style={{ height: "1em", transform: "translateY(0.2em)" }}
              />
            </TypingText>
          )}
        </span>
      </div>

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
