"use client";

import type { ReactNode } from "react";
import { ExternalLink, RotateCw } from "lucide-react";
import { Button } from "@/components/ui/button";

type PreviewPaneProps = {
  url: string | null;
  /** The sidebar toggle lives in this chrome row rather than a page header, so
   *  the preview panel and the sidebar panel share a top edge. */
  trigger?: ReactNode;
};

/*
 * Where the sandbox iframe goes. POST /ai-service/v1/sandbox/ returns {id, url}
 * and that url is what loads here — but nothing persists it yet, so a
 * conversation reopened tomorrow has no sandbox to point at. Both states are
 * rendered below because both are real: url, and no url.
 *
 * One surface, one hairline. The chrome is separated from the viewport by space
 * and a tone step, not by a divider line — see design-system.md §1 rule 4.
 */
export function PreviewPane({ url, trigger }: PreviewPaneProps) {
  return (
    <div className="border-hairline bg-surface flex h-full flex-col overflow-hidden rounded-2xl border">
      <div className="flex h-12 shrink-0 items-center gap-1.5 px-2.5">
        {trigger}
        <div className="bg-surface-2 text-muted-foreground flex h-7 min-w-0 flex-1 items-center rounded-full px-3 text-[11px]">
          <span className="truncate">{url ?? "No sandbox running"}</span>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full"
          disabled={!url}
          aria-label="Reload preview"
        >
          <RotateCw className="size-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full"
          disabled={!url}
          aria-label="Open preview in a new tab"
        >
          <ExternalLink className="size-3.5" />
        </Button>
      </div>

      <div className="min-h-0 flex-1 p-2.5 pt-0">
        <div className="bg-background flex size-full items-center justify-center rounded-xl">
          <p className="text-muted-foreground max-w-xs px-6 text-center text-xs text-balance">
            {url
              ? "The sandbox iframe renders here once the preview URL is reachable from the browser."
              : "No preview yet. A sandbox is created on the first Build message."}
          </p>
        </div>
      </div>
    </div>
  );
}
