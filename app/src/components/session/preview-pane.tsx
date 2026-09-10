"use client";

import { type ReactNode, useRef } from "react";
import { ExternalLink, RotateCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { GeneratingLine } from "./generating";

type PreviewPaneProps = {
  /** Project.preview_url. "" until a sandbox exists — normalise to null. */
  url: string | null;
  generating: boolean;
  /** The sidebar toggle lives in this chrome row rather than a page header, so
   *  the preview panel and the sidebar panel share a top edge. */
  trigger?: ReactNode;
};

/*
 * Where the sandbox iframe goes. Three states share one surface: no url yet,
 * generating (an opaque overlay over the last good frame), and a live preview.
 *
 * One surface, one hairline. The chrome is separated from the viewport by space
 * and a tone step, not by a divider line — see design-system.md §1 rule 4.
 */
export function PreviewPane({ url, generating, trigger }: PreviewPaneProps) {
  const frame = useRef<HTMLIFrameElement>(null);

  const reload = () => {
    // contentWindow.location.reload() throws on a cross-origin frame (the E2B
    // sandbox is a different origin). Re-assigning src is the equivalent that
    // works, and this only ever runs on a user click.
    if (frame.current) frame.current.src = frame.current.src;
  };

  return (
    <div className="border-hairline bg-surface flex h-full flex-col overflow-hidden rounded-2xl border">
      <div className="flex h-12 shrink-0 items-center gap-1.5 px-2.5">
        {trigger}
        <div className="bg-surface-2 text-muted-foreground flex h-7 min-w-0 flex-1 items-center rounded-full px-3 text-[11px]">
          <span className="truncate">
            {generating ? "Generating…" : (url ?? "No sandbox running")}
          </span>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full"
          disabled={!url}
          onClick={reload}
          aria-label="Reload preview"
        >
          <RotateCw className="size-3.5" />
        </Button>
        {url ? (
          <Button
            asChild
            variant="ghost"
            size="icon"
            className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full"
            aria-label="Open preview in a new tab"
          >
            <a href={url} target="_blank" rel="noreferrer">
              <ExternalLink className="size-3.5" />
            </a>
          </Button>
        ) : (
          <Button
            variant="ghost"
            size="icon"
            className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full"
            disabled
            aria-label="Open preview in a new tab"
          >
            <ExternalLink className="size-3.5" />
          </Button>
        )}
      </div>

      <div className="min-h-0 flex-1 p-2.5 pt-0">
        <div className="relative size-full">
          {url ? (
            <iframe
              ref={frame}
              src={url}
              title="Sandbox preview"
              className="size-full rounded-xl border-0 bg-white"
              sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
            />
          ) : (
            <div className="bg-background flex size-full items-center justify-center rounded-xl">
              <p className="text-muted-foreground max-w-xs px-6 text-center text-xs text-balance">
                No preview yet. A sandbox is created on the first Build message.
              </p>
            </div>
          )}

          {generating && (
            <div className="bg-surface vx-fade absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 rounded-xl">
              <GeneratingLine className="text-sm" />
              <p className="text-muted-foreground max-w-xs px-6 text-center text-xs text-balance">
                Writing files into your sandbox. This takes a minute or two.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
