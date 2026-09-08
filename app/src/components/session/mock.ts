/*
 * Static fixtures for the logged-in screen.
 *
 * Nothing here touches the network. This screen exists so the shape can be
 * reviewed and the API surface it implies can be read off before any of it is
 * built — see .planning/Frontend/logged_in_design.md §5 and §7. Every export
 * below names the RPC that will replace it.
 */

export type Provider = "openai" | "google" | "anthropic";
export type ChatMode = "chat" | "build";

/** Labels diverge from values on purpose: the AI-service routes reject
 *  "gemini" and "claude" — the path segments are exactly these three. */
export const PROVIDERS: { value: Provider; label: string }[] = [
  { value: "anthropic", label: "Claude" },
  { value: "openai", label: "OpenAI" },
  { value: "google", label: "Gemini" },
];

export type Message = {
  id: string;
  role: "user" | "assistant";
  mode: ChatMode;
  body: string;
};

export type Conversation = {
  id: string;
  title: string;
  updatedAt: string;
  /** From POST /sandbox/ -> {id, url}. Null until a sandbox exists. */
  previewUrl: string | null;
  messages: Message[];
};

/** Replaces: ListConversations. */
export const CONVERSATIONS: Conversation[] = [
  {
    id: "c1",
    title: "Photography portfolio",
    updatedAt: "2 hours ago",
    previewUrl: "https://3000-ab12cd34.e2b.dev",
    messages: [
      {
        id: "m1",
        role: "user",
        mode: "build",
        body: "Build a portfolio for a photographer. Big full-bleed images, a thin nav, and a contact page.",
      },
      {
        id: "m2",
        role: "assistant",
        mode: "build",
        body: "Created the route group, a full-bleed gallery grid and a contact form. The nav is sticky and collapses on mobile.",
      },
      {
        id: "m3",
        role: "user",
        mode: "chat",
        body: "Why did you use a grid instead of masonry here?",
      },
      {
        id: "m4",
        role: "assistant",
        mode: "chat",
        body: "Masonry needs JS measurement and reflows on image load. A CSS grid with a fixed aspect ratio keeps the first paint stable. Happy to switch if you want varied heights.",
      },
    ],
  },
  {
    id: "c2",
    title: "Pricing page, three tiers",
    updatedAt: "Yesterday",
    previewUrl: null,
    messages: [
      {
        id: "m5",
        role: "user",
        mode: "chat",
        body: "What should go in a three-tier pricing page for a dev tool?",
      },
      {
        id: "m6",
        role: "assistant",
        mode: "chat",
        body: "Anchor on the middle tier, keep feature rows to what changes between tiers, and put the enterprise tier last with a contact CTA rather than a price.",
      },
    ],
  },
  {
    id: "c3",
    title: "Docs site with search",
    updatedAt: "4 days ago",
    previewUrl: null,
    messages: [],
  },
];
