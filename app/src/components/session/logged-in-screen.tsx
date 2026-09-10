"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LogOut } from "lucide-react";
import { toast } from "sonner";
import { VxMark } from "@/components/brand/vx-mark";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import type { Profile } from "@/gen/api/v1/account_service_pb";
import type { AiProvider, ChatMode } from "@/gen/api/v1/enums_pb";
import { useAccountService } from "@/hooks/services/useAccountService";
import { useCreateProject, useProject, useProjects } from "@/hooks/useProjects";
import { useSendMessage } from "@/hooks/useMessages";
import { SilkBackground } from "@/components/landing/silk-background";
import { ChatPanel } from "./chat-panel";
import { HomeView } from "./home-view";
import { PreviewPane } from "./preview-pane";
import { ProjectList } from "./project-list";

type LoggedInScreenProps = {
  profile: Profile;
};

type View =
  | { kind: "home" }
  | { kind: "draft" } // Workspace open, no project created yet
  | { kind: "project"; id: string };

const TRIGGER =
  "text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full";

/*
 * Two screens behind one shell — see .planning/Frontend/logged_in_design.md.
 *
 * Home (no project open): the sidebar lists projects, the inset holds the
 * welcome hero and the prompt box.
 * Workspace (draft or an open project): the same sidebar panel becomes the
 * chat thread, and the inset becomes the sandbox preview. A draft is a
 * project with no id, no messages and no preview yet — not a fourth screen.
 *
 * Both panels float on the silk as opaque rounded surfaces, which is the same
 * language as the landing page's prompt box — one hairline per element, no
 * internal dividers, separation carried by space and the surface ladder.
 */
export function LoggedInScreen({ profile }: LoggedInScreenProps) {
  const account = useAccountService();
  const queryClient = useQueryClient();
  const [view, setView] = useState<View>({ kind: "home" });
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const isWorkspace = view.kind !== "home";
  const projectId = view.kind === "project" ? view.id : null;

  const { data: projects = [], isPending: projectsLoading } = useProjects();
  const { data: project } = useProject(projectId);
  const createProject = useCreateProject();
  const sendMessage = useSendMessage();

  // One boolean feeding both ChatPanel and PreviewPane, so the thread
  // shimmer and the preview overlay can never disagree.
  const generating = createProject.isPending || sendMessage.isPending;

  /* Opening a project must expand the rail. Going straight from a collapsed
   * icon rail to a chat thread would switch collapsible icon -> offcanvas on the
   * same frame as the content swap, which reads as a glitch rather than a move. */
  const open = (id: string) => {
    setView({ kind: "project", id });
    setSidebarOpen(true);
  };

  /* One handler for three entry points: Home's submit, the Workspace
   * composer's onSend, and the first send out of a draft. They differ only
   * in whether a project id already exists. */
  const start = async (input: {
    prompt: string;
    mode: ChatMode;
    provider: AiProvider;
  }) => {
    try {
      let id = view.kind === "project" ? view.id : null;

      if (!id) {
        const created = await createProject.mutateAsync({
          firstPrompt: input.prompt,
          provider: input.provider,
        });
        id = created.id;
        // Switch views the moment the project exists, not when the build
        // ends: the sidebar row and the real thread appear immediately, and
        // the send below runs against a Workspace that is already on screen.
        setView({ kind: "project", id });
        setSidebarOpen(true);
      }

      await sendMessage.mutateAsync({
        projectId: id,
        body: input.prompt,
        mode: input.mode,
        provider: input.provider,
      });
    } catch {
      toast.error("Could not build that. Please try again.");
    }
  };

  /* The sidebar's "New build" button skips the Home prompt entirely and drops
   * straight into a blank Workspace — ChatPanel already renders "No messages
   * yet." for an empty thread, so there is nothing else to special-case. The
   * row itself appears once CreateProject resolves. */
  const newBuild = () => {
    setView({ kind: "draft" });
    setSidebarOpen(true);
  };

  // Google does not always return a given name; never render a blank line.
  const name = profile.firstName?.trim() || profile.email.split("@")[0];

  const logout = useMutation({
    mutationFn: async () => {
      await account.accountLogout({});
    },
    onSuccess: () => {
      queryClient.setQueryData(["profile"], null);
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
    onError: () => {
      toast.error("Could not log out. Please try again.");
    },
  });

  return (
    <div className="bg-background text-foreground relative isolate min-h-screen">
      <SilkBackground className="-z-10" />

      <SidebarProvider
        open={sidebarOpen}
        onOpenChange={setSidebarOpen}
        /* The vendored sidebar animates width with duration-200 ease-linear.
         * Retimed to the app's easing so the rail widening and the content
         * fading in land together. */
        className="bg-transparent [&_[data-slot=sidebar-container]]:duration-300 [&_[data-slot=sidebar-container]]:ease-[cubic-bezier(0.16,1,0.3,1)] [&_[data-slot=sidebar-gap]]:duration-300 [&_[data-slot=sidebar-gap]]:ease-[cubic-bezier(0.16,1,0.3,1)]"
        // The rail holds a list on Home and a whole chat thread in Workspace,
        // which needs materially more room.
        style={
          {
            "--sidebar-width": isWorkspace ? "26rem" : "16rem",
          } as React.CSSProperties
        }
      >
        {/*
         * variant="sidebar", not "floating": flush to the viewport on the left,
         * top and bottom with a single right hairline, so the rail reads as part
         * of the ground rather than a card sitting on it. The floating variant's
         * p-2 inset is exactly the gap we do not want. It also collapses to a
         * clean 3rem with no padding arithmetic, which floating does not.
         *
         * Always "offcanvas", even on Home — collapsible must stay the same
         * value across both states so the sidebar is the same DOM branch
         * before and after opening a conversation, which is what lets the
         * `--sidebar-width` change (below) animate as a slide instead of an
         * unmount/remount. Home just never renders a trigger or rail, so
         * nothing can actually collapse it.
         */}
        <Sidebar
          variant="sidebar"
          collapsible="offcanvas"
          className="border-hairline"
        >
          <SidebarHeader className="h-16 shrink-0 justify-center">
            <div className="flex items-center gap-2.5 px-1">
              <VxMark className="size-8" />
              <span className="text-foreground text-xl font-medium tracking-tight">
                VULX
              </span>
            </div>
          </SidebarHeader>

          <SidebarContent className={isWorkspace ? "overflow-hidden" : undefined}>
            <div
              key={isWorkspace ? "chat" : "list"}
              className="vx-fade flex min-h-0 flex-1 flex-col"
            >
              {isWorkspace ? (
                <ChatPanel
                  key={projectId ?? "draft"}
                  projectId={projectId}
                  title={project?.title ?? "New build"}
                  onBack={() => setView({ kind: "home" })}
                  onSend={(m) =>
                    start({ prompt: m.body, mode: m.mode, provider: m.provider })
                  }
                  generating={generating}
                  defaultProvider={project?.provider}
                />
              ) : (
                <ProjectList
                  projects={projects}
                  activeId={projectId}
                  loading={projectsLoading}
                  onOpen={open}
                  onNewBuild={newBuild}
                />
              )}
            </div>
          </SidebarContent>

          <SidebarFooter>
            <div className="flex items-center gap-2 px-1 py-1">
              <div className="flex min-w-0 flex-1 flex-col">
                <span className="text-foreground truncate text-xs">
                  {profile.email}
                </span>
                {/* Credits are granted and read but never spent — display only. */}
                <span className="text-muted-foreground text-[11px]">
                  {profile.credits.toString()} credits
                </span>
              </div>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    disabled={logout.isPending}
                    aria-label="Log out"
                    className="text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-8 shrink-0 rounded-full"
                  >
                    <LogOut className="size-3.5" />
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent className="bg-surface border-hairline text-foreground rounded-2xl shadow-none sm:max-w-sm">
                  <AlertDialogHeader>
                    <AlertDialogTitle>Log out?</AlertDialogTitle>
                    <AlertDialogDescription>
                      You&apos;ll need to sign back in to keep working on your
                      projects.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel className="border-hairline bg-surface-2 hover:bg-surface dark:bg-surface-2 dark:border-hairline dark:hover:bg-surface rounded-full shadow-none">
                      Cancel
                    </AlertDialogCancel>
                    <AlertDialogAction
                      onClick={() => logout.mutate()}
                      className="bg-destructive rounded-full text-white shadow-none hover:bg-destructive/90"
                    >
                      Log out
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </SidebarFooter>

          {/* Edge target: the only way back once the workspace sidebar
              collapses fully offcanvas. Home's sidebar is not collapsible,
              so the rail has nothing to do there. */}
          {isWorkspace && <SidebarRail />}
        </Sidebar>

        <SidebarInset className="relative flex min-h-svh flex-col bg-transparent">
          {isWorkspace ? (
            <div key="preview" className="vx-fade min-h-0 flex-1 p-2">
              <PreviewPane
                url={project?.previewUrl || null}
                generating={generating}
                trigger={<SidebarTrigger className={TRIGGER} />}
              />
            </div>
          ) : (
            <div key="home" className="vx-fade flex flex-1 flex-col">
              <HomeView name={name} onStart={start} disabled={generating} />
            </div>
          )}
        </SidebarInset>
      </SidebarProvider>
    </div>
  );
}
