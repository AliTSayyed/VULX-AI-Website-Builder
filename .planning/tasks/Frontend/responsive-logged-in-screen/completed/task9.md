# Task 9 — `project-list.tsx`

**Goal.** The sidebar list over real `Project`s, with an active row, relative timestamps and an
empty state. Built as a new file; nothing imports it until task 10.

## Create `app/src/components/session/project-list.tsx`

Start from `conversation-list.tsx` — the `New build` button, the `SidebarGroup` / `SidebarMenu`
structure, the accent dot, the two-line row and every class string stay. What changes:

```tsx
type ProjectListProps = {
  projects: Project[];
  activeId: string | null;
  loading: boolean;
  onOpen: (id: string) => void;
  onNewBuild: () => void;
};
```

1. **Rows come from `Project`**: `p.title` on line one, `formatRelative(p.updatedAt)` on line two.
   The group label becomes `Projects`, not `Conversations` — the backend's noun is a project, and
   two names for one thing in the same product is how a demo gets confusing.
2. **Do not sort.** `ListProjects` already returns `updated_at DESC` (keyset-paginated on that
   column). Re-sorting in the browser would fight the server on ties.
3. **Active row**: `isActive={p.id === activeId}` on `SidebarMenuButton` — the vendored component
   already styles `data-active`. Do not invent a highlight.
4. **Empty state**, in place of `SidebarMenu` when `projects.length === 0 && !loading`:

```tsx
<p className="text-muted-foreground px-2 py-6 text-center text-[11px]">
  No projects yet.
</p>
```

5. **Loading**: while `loading` and the list is empty, render nothing where the rows would be — no
   skeletons. `ListProjects` against a local API returns in milliseconds; a skeleton that flashes for
   80ms is noisier than a beat of empty space.

## Points that matter

- **`New build` does not create anything.** It sets the view to a draft (task 10). Keep the button
  exactly where and how it is; only the handler's meaning changes, and that lives in the parent.
- `formatRelative` comes from task 1. Do not format dates inline.
- The `text-[11px]` above matches the sibling rows, not `design-system.md` §2's `text-[11.5px]` meta
  size. The rail already sits at 11px throughout (rows, footer); matching the neighbour is right and
  re-typing the whole rail is not this task.
- Titles are server-generated (the truncated first prompt). Do not derive or truncate again — the
  row already truncates with `truncate`.
- Leave `conversation-list.tsx` in place; task 10 deletes it. Two files briefly coexist so this task
  compiles on its own.

## Done when

- `npm run lint` and `npm run build` pass.
- Nothing imports the file yet; the app is unchanged.
