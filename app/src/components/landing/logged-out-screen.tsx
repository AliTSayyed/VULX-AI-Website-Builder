"use client";

import { useState } from "react";
import { AuthDialog, type AuthMode } from "@/components/auth/auth-dialog";
import { Hero } from "./hero";
import { HeroPrompt } from "./hero-prompt";
import { SuggestionChips } from "./suggestion-chips";
import { TopBar } from "./top-bar";

export function LoggedOutScreen() {
  const [authOpen, setAuthOpen] = useState(false);
  const [authMode, setAuthMode] = useState<AuthMode>("signup");

  const openAuth = (mode: AuthMode) => {
    setAuthMode(mode);
    setAuthOpen(true);
  };

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col">
      <TopBar onAuth={openAuth} />

      <main className="flex flex-1 flex-col items-center justify-center px-6 pb-12">
        <Hero />
        <HeroPrompt onAuth={openAuth} />
        <SuggestionChips onSelect={() => openAuth("signup")} />
      </main>

      <footer className="text-muted-foreground shrink-0 pb-6 text-center text-xs">
        VULX can make mistakes. Check generated code before shipping.
      </footer>

      <AuthDialog open={authOpen} onOpenChange={setAuthOpen} mode={authMode} />
    </div>
  );
}