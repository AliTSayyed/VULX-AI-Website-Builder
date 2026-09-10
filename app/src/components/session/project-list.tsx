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
import type { Project } from "@/gen/api/v1/project_service_pb";
import { formatRelative } from "@/lib/format-time";

type ProjectListProps = {
  projects: Project[];
  activeId: string | null;
  loading: boolean;
  onOpen: (id: string) => void;
  onNewBuild: () => void;
};

export function ProjectList({
  projects,
  activeId,
  loading,
  onOpen,
  onNewBuild,
}: ProjectListProps) {
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
          Projects
        </SidebarGroupLabel>
        {projects.length === 0 && !loading ? (
          <p className="text-muted-foreground px-2 py-6 text-center text-[11px]">
            No projects yet.
          </p>
        ) : (
          <SidebarMenu>
            {projects.map((p) => (
              <SidebarMenuItem key={p.id}>
                <SidebarMenuButton
                  onClick={() => onOpen(p.id)}
                  isActive={p.id === activeId}
                  tooltip={p.title}
                  className="hover:bg-surface-2 h-auto items-start gap-2 py-2"
                >
                  <span className="mt-2 size-1.5 shrink-0 rounded-full bg-accent-blue" />
                  <span className="flex min-w-0 flex-col gap-0.5">
                    <span className="text-foreground truncate text-[13px]">
                      {p.title}
                    </span>
                    <span className="text-muted-foreground text-[11px]">
                      {formatRelative(p.updatedAt)}
                    </span>
                  </span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        )}
      </SidebarGroup>
    </>
  );
}
