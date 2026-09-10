"use client";

import { useEffect, useState } from "react";
import { TextShimmer } from "@/components/ui/text-shimmer";
import { cn } from "@/lib/utils";

/*
 * Shimmer means "the model is working" — design-system.md §8 reserves it for
 * exactly this flow. The words cycle because a Build takes minutes and a single
 * frozen label starts to read as a hung request after the first thirty seconds.
 *
 * The list is intentionally vague. The backend sends no progress at all
 * (SendMessage is one blocking call — Polish.md, "Async Build"), so these are
 * honest about the phase and dishonest about nothing.
 */
const PHASES = ["Thinking", "Building", "Creating"];

const INTERVAL = 2400;

export function GeneratingLine({ className }: { className?: string }) {
  const [i, setI] = useState(0);

  useEffect(() => {
    const t = setInterval(() => setI((n) => (n + 1) % PHASES.length), INTERVAL);
    return () => clearInterval(t);
  }, []);

  return (
    <TextShimmer className={cn("text-[13px]", className)} duration={3}>
      {`${PHASES[i]}…`}
    </TextShimmer>
  );
}
