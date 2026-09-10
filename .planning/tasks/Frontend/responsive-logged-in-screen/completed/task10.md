# Task 10 — `LoggedInScreen`: the cutover

**Goal.** Replace the fixture state with the hooks, orchestrate create-then-send, and delete
`mock.ts`. Everything built in tasks 1–9 connects here.

## 1. The view union

Replace `openId: string | null` and the `conversations` array:

```ts
type View =
  | { kind: "home" }
  | { kind: "draft" }                 // Workspace open, no project created yet
  | { kind: "project"; id: string };
```

- `home` → `HomeView` in the inset, `ProjectList` in the rail.
- `draft` and `project` → **the same** Workspace: `ChatPanel` in the rail, `PreviewPane` in the
  inset. A draft is a project with no id, no messages and no preview — not a fourth screen.
- `const projectId = view.kind === "project" ? view.id : null;` is the only thing the children need
  to tell them apart.
- Everything that keys off "is a conversation open" (`--sidebar-width` 16rem→26rem, `SidebarRail`,
  `vx-fade` keys) now keys off `view.kind !== "home"`. Keep `collapsible="offcanvas"` in both states —
  changing it between states puts the vendored `Sidebar` on a different render branch and forces an
  unmount instead of a slide (this was tried and reverted; see `logged_in_design.md` §2).

## 2. The data

```ts
const { data: projects = [], isPending: projectsLoading } = useProjects();
const { data: project } = useProject(projectId);
const createProject = useCreateProject();
const sendMessage = useSendMessage();

const generating = createProject.isPending || sendMessage.isPending;
```

`generating` is one boolean feeding both `ChatPanel` and `PreviewPane`, so the thread shimmer and the
preview overlay can never disagree.

## 3. The one flow

```ts
const start = async (input: {
  prompt: string;
  mode: ChatMode;
  provider: AiProvider;
}) => {
  try {
    let id = view.kind === "project" ? view.id : null;

    if (!id) {
      const project = await createProject.mutateAsync({
        firstPrompt: input.prompt,
        provider: input.provider,
      });
      id = project.id;
      // Switch views the moment the project exists, not when the build ends:
      // the sidebar row and the real thread appear immediately, and the send
      // below runs against a Workspace that is already on screen.
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
```

- **One handler for three entry points**: Home submit, the Workspace composer's `onSend`, and the
  first send out of a draft. They differ only in whether a project id already exists.
- **`mutateAsync` + `try/catch`.** The sequence is genuinely sequential (the send needs the id), and
  the `catch` is what stops a failed build shimmering forever. No designed error surface is in scope
  (doc §8) — a toast is the whole of it.
- **`New build`** becomes `setView({ kind: "draft" }); setSidebarOpen(true);` — no fabricated row.
  The row appears when `CreateProject` resolves.
- **Back** becomes `setView({ kind: "home" })`.
- Opening a row: `setView({ kind: "project", id }); setSidebarOpen(true);` — the forced open is the
  existing rule and still applies.

## 4. Wiring the children

```tsx
<ProjectList
  projects={projects}
  activeId={projectId}
  loading={projectsLoading}
  onOpen={open}
  onNewBuild={newBuild}
/>

<ChatPanel
  key={projectId ?? "draft"}
  projectId={projectId}
  title={project?.title ?? "New build"}
  onBack={() => setView({ kind: "home" })}
  onSend={(m) => start({ prompt: m.body, mode: m.mode, provider: m.provider })}
  generating={generating}
  defaultProvider={project?.provider}
/>

<PreviewPane
  url={project?.previewUrl || null}
  generating={generating}
  trigger={<SidebarTrigger className={TRIGGER} />}
/>
```

- **`key={projectId ?? "draft"}`** remounts the panel per project, which is what resets the composer
  draft text and re-seeds the provider select. Cheaper and less bug-prone than syncing state in an
  effect.
- **`project?.previewUrl || null`** — `preview_url` is `""`, never null (`NOT NULL DEFAULT ''`), so
  `?? null` would pass the empty string straight through into `src=""`.
- `HomeView` gets `onStart={start}` and `disabled={generating}`.

## 5. Delete

- `app/src/components/session/mock.ts`
- `app/src/components/session/conversation-list.tsx`

`grep -r "mock" app/src` must come back empty. Nothing else may import either file.

## Points that matter

- **Do not add a route.** The open project is local state; a refresh returns to Home. Accepted for
  the MVP (doc §8).
- Leave the logout `AlertDialog`, the footer (email + credits), the silk background and the sidebar
  timing overrides exactly as they are.
- Credits are still display-only. Nothing debits them.

## Done when

- `npm run lint` and `npm run build` pass with no reference to `mock.ts`.
- The full demo path in `responsive-logged-in-screen.md` §10 runs against the live stack. Task 11
  walks it.
