"use client";

import { Button } from "@/components/ui/button";

type TopBarProps = {
  onAuth: (mode: "login" | "signup") => void;
};

export function TopBar({ onAuth }: TopBarProps) {
  return (
    <header className="border-border h-14 shrink-0 border-b">
      <div className="mx-auto flex h-full max-w-6xl items-center justify-between px-6">
        <div className="flex items-center gap-2">
          <div className="bg-foreground size-5 rounded-md" aria-hidden />
          <span className="font-medium tracking-tight">VULX</span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            className="hidden sm:inline-flex"
            onClick={() => onAuth("login")}
          >
            Log in
          </Button>
          <Button size="sm" onClick={() => onAuth("signup")}>
            Sign up
          </Button>
        </div>
      </div>
    </header>
  );
}