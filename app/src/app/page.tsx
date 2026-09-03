"use client";

import { useEffect } from "react";
import { toast } from "sonner";
import { LoggedOutScreen } from "@/components/landing/logged-out-screen";
import { LoggedInScreen } from "@/components/session/logged-in-screen";
import { useSession } from "@/hooks/useSession";

const Page = () => {
  const { profile, status, error } = useSession();

  useEffect(() => {
    if (error) {
      toast.error("Could not reach VULX. Some features may not work.");
    }
  }, [error]);

  if (status === "loading") {
    return <div className="bg-background min-h-screen" />;
  }

  if (status === "authed" && profile) {
    return <LoggedInScreen profile={profile} />;
  }

  return <LoggedOutScreen />;
};

export default Page;
