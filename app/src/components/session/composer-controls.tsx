"use client";

import { Hammer, MessageSquare } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { AiProvider, ChatMode } from "@/gen/api/v1/enums_pb";
import { PROVIDERS } from "./providers";

type ComposerControlsProps = {
  mode: ChatMode;
  onModeChange: (mode: ChatMode) => void;
  provider: AiProvider;
  onProviderChange: (provider: AiProvider) => void;
  disabled?: boolean;
};

export function ComposerControls({
  mode,
  onModeChange,
  provider,
  onProviderChange,
  disabled,
}: ComposerControlsProps) {
  return (
    <div className="flex min-w-0 items-center gap-3">
      <ToggleGroup
        type="single"
        value={String(mode)}
        onValueChange={(v) => v && onModeChange(Number(v) as ChatMode)}
        disabled={disabled}
        className="border-hairline bg-surface-2 h-8 gap-0 rounded-full border p-1"
      >
        <ToggleGroupItem
          value={String(ChatMode.CHAT)}
          aria-label="Chat mode"
          className="group text-foreground-dim data-[state=on]:text-foreground-dim h-full gap-1 rounded-full bg-transparent px-2.5 text-[11px] shadow-none focus-visible:shadow-none focus-visible:ring-0 data-[state=on]:bg-transparent data-[state=on]:shadow-none"
        >
          <MessageSquare className="size-3 group-data-[state=on]:text-accent-blue" />
          Chat
        </ToggleGroupItem>
        <ToggleGroupItem
          value={String(ChatMode.BUILD)}
          aria-label="Build mode"
          className="group text-foreground-dim data-[state=on]:text-foreground-dim h-full gap-1 rounded-full bg-transparent px-2.5 text-[11px] shadow-none focus-visible:shadow-none focus-visible:ring-0 data-[state=on]:bg-transparent data-[state=on]:shadow-none"
        >
          <Hammer className="size-3 group-data-[state=on]:text-accent-blue" />
          Build
        </ToggleGroupItem>
      </ToggleGroup>

      <Select
        value={String(provider)}
        onValueChange={(v) => v && onProviderChange(Number(v) as AiProvider)}
        disabled={disabled}
      >
        <SelectTrigger
          size="sm"
          className="border-hairline bg-surface-2 dark:bg-surface-2 dark:hover:bg-surface-2 h-8 gap-1 rounded-full px-2.5 text-[11px] shadow-none"
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {PROVIDERS.map((p) => (
            <SelectItem key={p.value} value={String(p.value)} className="text-xs">
              {p.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
