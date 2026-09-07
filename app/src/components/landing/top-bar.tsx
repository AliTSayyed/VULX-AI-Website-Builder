"use client";

import { VxMark } from "@/components/brand/vx-mark";
import { Button } from "@/components/ui/button";

type TopBarProps = {
  onAuth: (mode: "login" | "signup") => void;
};

export function TopBar({ onAuth }: TopBarProps) {
  return (
    <header className="vx-rise h-16 shrink-0">
      <div className="mx-auto flex h-full max-w-6xl items-center justify-between px-5 sm:px-6">
        <div className="flex items-center gap-2.5">
          <VxMark />
          <span className="text-foreground text-[15px] font-medium tracking-tight">
            VULX
          </span>
        </div>

        <div className="flex items-center gap-1.5">
          <Button
            variant="ghost"
            size="sm"
            className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 hidden rounded-full px-4 sm:inline-flex"
            onClick={() => onAuth("login")}
          >
            Log in
          </Button>
          <Button
            size="sm"
            className="rounded-full px-4"
            onClick={() => onAuth("signup")}
          >
            Sign up
          </Button>
        </div>
      </div>
    </header>
  );
}
