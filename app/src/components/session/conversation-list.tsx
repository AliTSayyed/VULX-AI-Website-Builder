"use client";

import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import type { Conversation } from "./mock";

type ConversationListProps = {
  conversations: Conversation[];
  onOpen: (id: string) => void;
  onNewBuild: () => void;
};

export function ConversationList({
  conversations,
  onOpen,
  onNewBuild,
}: ConversationListProps) {
  return (
    <>
      <div className="px-2 pb-1">
        <Button
          variant="outline"
          size="sm"
          onClick={onNewBuild}
          className="border-hairline bg-surface hover:bg-surface-2 hover:border-hairline-strong dark:border-hairline dark:bg-surface dark:hover:bg-surface-2 gap-2 rounded-full has-[>svg]:pr-6 shadow-none"
        >
          <Plus className="size-3.5" />
          <span>New build</span>
        </Button>
      </div>

      <SidebarGroup>
        <SidebarGroupLabel className="text-muted-foreground text-[11px]">
          Conversations
        </SidebarGroupLabel>
        <SidebarMenu>
          {conversations.map((c) => (
            <SidebarMenuItem key={c.id}>
              <SidebarMenuButton
                onClick={() => onOpen(c.id)}
                tooltip={c.title}
                className="hover:bg-surface-2 h-auto items-start gap-2 py-2"
              >
                <span className="mt-2 size-1.5 shrink-0 rounded-full bg-accent-blue" />
                <span className="flex min-w-0 flex-col gap-0.5">
                  <span className="text-foreground truncate text-[13px]">
                    {c.title}
                  </span>
                  <span className="text-muted-foreground text-[11px]">
                    {c.updatedAt}
                  </span>
                </span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroup>
    </>
  );
}
