"use client";

import { Suspense, useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { ConnectError } from "@connectrpc/connect";
import { useAccountService } from "@/hooks/services/useAccountService";

function CallbackHandler() {
  const params = useSearchParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const account = useAccountService();
  const fired = useRef(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (fired.current) return;
    fired.current = true;

    const oauthError = params.get("error");
    if (oauthError) {
      setError(
        oauthError === "access_denied"
          ? "Sign-in was cancelled."
          : `Google returned an error: ${oauthError}`
      );
      return;
    }

    const code = params.get("code");
    const state = params.get("state");
    if (!code || !state) {
      setError("Invalid sign-in link.");
      return;
    }

    account
      .finishAccountAuth({ code, state })
      .then((res) => {
        queryClient.setQueryData(["profile"], res.profile ?? null);
        queryClient.invalidateQueries({ queryKey: ["profile"] });
        router.replace("/");
      })
      .catch((err) => {
        setError(
          err instanceof ConnectError
            ? err.rawMessage
            : "Could not complete sign-in."
        );
      });
  }, [account, params, queryClient, router]);

  if (error) {
    return (
      <div className="bg-background text-foreground flex min-h-screen flex-col items-center justify-center gap-4 px-6 text-center">
        <p className="text-muted-foreground text-sm text-balance">{error}</p>
        <a href="/" className="underline underline-offset-4">
          Back to VULX
        </a>
      </div>
    );
  }

  return <div className="bg-background min-h-screen" />;
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={<div className="bg-background min-h-screen" />}>
      <CallbackHandler />
    </Suspense>
  );
}
