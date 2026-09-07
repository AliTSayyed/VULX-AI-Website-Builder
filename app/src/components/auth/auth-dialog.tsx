"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { GoogleIcon } from "./google-icon";
import { LoginProvider } from "@/gen/api/v1/enums_pb";
import { useAccountService } from "@/hooks/services/useAccountService";

export type AuthMode = "login" | "signup";

type AuthDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: AuthMode;
};

const COPY: Record<AuthMode, { title: string; description: string }> = {
  signup: {
    title: "Create your account",
    description: "Create an account with Google to start building.",
  },
  login: {
    title: "Sign in to continue",
    description: "Sign in to your account with Google.",
  },
};

export function AuthDialog({ open, onOpenChange, mode }: AuthDialogProps) {
  const copy = COPY[mode];
  const account = useAccountService();
  const [failed, setFailed] = useState(false);

  const begin = useMutation({
    mutationFn: async () => {
      const res = await account.beginAccountAuth({
        loginProvider: LoginProvider.GOOGLE,
      });
      if (!res.loginUrl) {
        throw new Error("no login url returned");
      }
      window.location.assign(res.loginUrl);
    },
    onMutate: () => setFailed(false),
    onError: () => {
      setFailed(true);
      toast.error("Could not start sign-in. Please try again.");
    },
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="border-hairline bg-popover rounded-2xl sm:max-w-sm">
        <DialogHeader className="sm:text-center">
          <DialogTitle className="text-xl font-medium tracking-tight">{copy.title}</DialogTitle>
          <DialogDescription className="sr-only">{copy.description}</DialogDescription>
        </DialogHeader>

        <Button
          className="mt-2 justify-self-center rounded-full shadow-none"
          onClick={() => begin.mutate()}
          disabled={begin.isPending}
        >
          <GoogleIcon className="size-4" />
          {begin.isPending ? "Redirecting…" : "Continue with Google"}
        </Button>

        {failed && (
          <p className="text-destructive mt-2 text-center text-sm">
            Something went wrong. Please try again.
          </p>
        )}

        <p className="text-muted-foreground mt-2 text-center text-xs">
          By continuing you agree to our Terms and Privacy Policy.
        </p>
      </DialogContent>
    </Dialog>
  );
}
