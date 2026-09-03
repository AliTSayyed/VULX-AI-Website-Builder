"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { GoogleIcon } from "./google-icon";

export type AuthMode = "login" | "signup";

type AuthDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: AuthMode;
};

const COPY: Record<AuthMode, { title: string; description: string }> = {
  signup: {
    title: "Create your account",
    description: "Start building in seconds.",
  },
  login: {
    title: "Welcome back",
    description: "Sign in to continue.",
  },
};

export function AuthDialog({ open, onOpenChange, mode }: AuthDialogProps) {
  const copy = COPY[mode];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="text-xl tracking-tight">{copy.title}</DialogTitle>
          <DialogDescription>{copy.description}</DialogDescription>
        </DialogHeader>

        <Button variant="outline" size="lg" className="mt-2 w-full">
          <GoogleIcon className="size-4" />
          Continue with Google
        </Button>

        <p className="text-muted-foreground mt-2 text-center text-xs">
          By continuing you agree to our Terms and Privacy Policy.
        </p>
      </DialogContent>
    </Dialog>
  );
}