# Task 11 — End-to-end demo pass and doc updates

**Goal.** Prove the whole path in a browser against the running stack, then make the documents match
the code.

## 1. Before you start

```bash
make            # whole stack
```

- `../AI-Website-Builder-Secrets/.api-env` and `.ai-service-env` must exist, with a real key for
  whichever provider you demo with.
- `/etc/hosts` needs `127.0.0.1 local.api.vulx.ai` and `127.0.0.1 local.app.vulx.ai`; Caddy's local
  CA must be trusted (`make trust`).
- **Raise `E2B_SANDBOX_TIMEOUT_SECONDS`** before demoing. The default is short, there is no
  keep-alive and no `RefreshSandbox` (`Polish.md`), so a sandbox that dies mid-demo cannot be
  recovered — only outlasted.
- Open the app at `https://local.app.vulx.ai`, not `localhost:3000`, or the session cookie will not
  ride along.

## 2. The walk

Run it in one sitting, in one browser, in this order. Each step is a claim the feature makes.

| # | Do | Expect |
| --- | --- | --- |
| 1 | Log in with an account that has no projects | Sidebar: `No projects yet.` Home: `Welcome back, <name>` and a prompt box with both selectors |
| 2 | Type a prompt, pick Build + a provider, submit | Workspace opens; your message is in the thread; shimmer cycles Thinking/Building/Creating; preview shows the overlay and `Generating…`; composer fully locked; a new row with a real title appears in the sidebar within a second |
| 3 | Wait for the agent (minutes) | Assistant summary lands in the thread; overlay lifts; the generated site renders in the iframe; composer unlocks |
| 4 | Send a second prompt in the same project | Overlay returns **over the existing live preview**; when it resolves the site has changed with no full reload |
| 5 | `← Back` | Home. The project is listed with a relative timestamp, at the top of the list |
| 6 | Reopen it | Full thread and the same live preview come back |
| 7 | `New build` | Empty Workspace, no new sidebar row; the first send creates the project |
| 8 | Reload button, open-in-new-tab | The frame reloads; the tab opens the sandbox |

### What to watch in the Network tab

- One `ListProjects` on load; one `GetProject` + one `ListMessages` per project opened.
- `SendMessage` stays pending for the whole agent run with **no client timeout**.
- After it resolves: `ListMessages`, `GetProject` and `ListProjects` all refetch. **If `GetProject`
  does not refetch, the preview will never appear** — that is task 3's `onSettled`.

## 3. Known things that are not bugs

- A dead sandbox and a real outage both surface as one opaque failure and a toast (`Polish.md`).
- Chat mode fails with `Unimplemented` — expected, not gated (doc §8).
- Switching provider mid-thread does not stick on reopen (`Polish.md` → *`Project.provider` goes
  stale*).
- A browser refresh returns you to Home; the open project is not a route.
- The layout assumes a desktop viewport.

## 4. Documents to update in the same change

- **`ARCHITECTURE.md` §9** — the "static UI shell over fixtures" paragraph is no longer true. Rewrite
  it to describe the wired screen: the hooks, the query keys, the create→send flow, the generating
  state, and `mock.ts` being gone. Keep the parts that are still accurate (the `SidebarProvider`
  shell, `collapsible="offcanvas"` in both states, the accent-blue rationale).
- **`ARCHITECTURE.md` §10** — flip *Frontend beyond auth (dashboard, editor, preview)* from 🟡 to ✅
  with a note that it is the synchronous Build path only.
- **`ARCHITECTURE.md` §11.7** — nothing to change; leave the fixed items alone.
- **`.planning/Frontend/logged_in_design.md`** — §5 ("What this needs that does not exist") and §6
  ("Endpoints this screen implies") describe RPCs by names that were never built
  (`ListConversations`, `GetPreview`). Add a short note at the top of each pointing at
  `responsive-logged-in-screen.md` as the built version, rather than rewriting the design doc.
- **`.planning/Polish.md`** — the Frontend section's "beyond wiring the now-real RPCs to the existing
  fixture-backed shell" phrasing is stale once this ships; the open questions it lists (sandbox
  lifecycle, poll-vs-stream, file tree, draggable split) all stand.
- **`design-system.md` §8** — **both** bullets are stale. Replace the `TextShimmer` one with a line
  recording that the keyframe now exists and that shimmer is in use in the generation flow, noting
  `ThinkingBar` stays unused (it assumes a cancellable stream). Delete the "static shell over
  fixtures" bullet outright — that is what this feature ends.
- **`design-system.md` §1 rule 1 and §3** — `--accent-blue`'s call-site list gains the generating
  row's Build icon, and the surface ladder's `--surface` row now also backs the preview's generating
  overlay. Add both; rule 1 requires the accent's uses to be written down.
- Move the whole task folder into `completed/` and commit, per the convention in the other task
  directories.

## Done when

All eight rows of §2 pass in one sitting, and the four documents in §4 match the code.
