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
import { SilkBackground } from "@/components/landing/silk-background";
import { ChatPanel } from "./chat-panel";
import { ConversationList } from "./conversation-list";
import { HomeView } from "./home-view";
import { PreviewPane } from "./preview-pane";
import { CONVERSATIONS, type Conversation } from "./mock";

type LoggedInScreenProps = {
  profile: Profile;
};

const TRIGGER =
  "text-foreground-dim hover:bg-surface-2 hover:text-foreground dark:hover:bg-surface-2 size-7 shrink-0 rounded-full";

/*
 * Two screens behind one shell — see .planning/Frontend/logged_in_design.md.
 *
 * Home (no conversation open): the sidebar lists conversations, the inset holds
 * the welcome hero and the prompt box.
 * Workspace (a conversation open): the same sidebar panel becomes that
 * conversation's chat thread, and the inset becomes the sandbox preview.
 *
 * Everything below is local state over static fixtures. No network call is made
 * except the logout that already existed.
 *
 * Both panels float on the silk as opaque rounded surfaces, which is the same
 * language as the landing page's prompt box — one hairline per element, no
 * internal dividers, separation carried by space and the surface ladder.
 */
export function LoggedInScreen({ profile }: LoggedInScreenProps) {
  const account = useAccountService();
  const queryClient = useQueryClient();
  const [conversations, setConversations] =
    useState<Conversation[]>(CONVERSATIONS);
  const [openId, setOpenId] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);

  /* Opening a conversation must expand the rail. Going straight from a collapsed
   * icon rail to a chat thread would switch collapsible icon -> offcanvas on the
   * same frame as the content swap, which reads as a glitch rather than a move. */
  const openConversation = (id: string) => {
    setOpenId(id);
    setSidebarOpen(true);
  };

  const conversation = conversations.find((c) => c.id === openId) ?? null;

  /* Submitting on Home is the one flow worth demonstrating: it drops you into
   * the Workspace with your prompt as the first message and no sandbox yet —
   * which is the state the preview pane has to handle and currently cannot. */
  const start = (input: { prompt: string; mode: ChatMode; provider: AiProvider }) => {
    const id = `draft-${Date.now()}`;
    setConversations((prev) => [
      {
        id,
        title:
          input.prompt.length > 48
            ? `${input.prompt.slice(0, 48)}…`
            : input.prompt,
        updatedAt: "Just now",
        previewUrl: null,
        messages: [
          { id: `${id}-m1`, role: "user", mode: "build", body: input.prompt },
        ],
      },
      ...prev,
    ]);
    openConversation(id);
  };

  /* The sidebar's "New build" button skips the Home prompt entirely and drops
   * straight into a blank Workspace — ChatPanel already renders "No messages
   * yet." for an empty thread, so there is nothing else to special-case. */
  const newBuild = () => {
    const id = `draft-${Date.now()}`;
    setConversations((prev) => [
      {
        id,
        title: "New build",
        updatedAt: "Just now",
        previewUrl: null,
        messages: [],
      },
      ...prev,
    ]);
    openConversation(id);
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
            "--sidebar-width": conversation ? "26rem" : "16rem",
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

          <SidebarContent className={conversation ? "overflow-hidden" : undefined}>
            <div
              key={conversation ? "chat" : "list"}
              className="vx-fade flex min-h-0 flex-1 flex-col"
            >
              {conversation ? (
                <ChatPanel
                  projectId={conversation.id}
                  title={conversation.title}
                  onBack={() => setOpenId(null)}
                  onSend={() => {}}
                  generating={false}
                />
              ) : (
                <ConversationList
                  conversations={conversations}
                  onOpen={openConversation}
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
          {conversation && <SidebarRail />}
        </Sidebar>

        <SidebarInset className="relative flex min-h-svh flex-col bg-transparent">
          {conversation ? (
            <div key="preview" className="vx-fade min-h-0 flex-1 p-2">
              <PreviewPane
                url={conversation.previewUrl}
                trigger={<SidebarTrigger className={TRIGGER} />}
              />
            </div>
          ) : (
            <div key="home" className="vx-fade flex flex-1 flex-col">
              <HomeView name={name} onStart={start} />
            </div>
          )}
        </SidebarInset>
      </SidebarProvider>
    </div>
  );
}
