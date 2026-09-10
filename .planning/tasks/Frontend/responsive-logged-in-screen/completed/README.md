# Frontend — Responsive Logged-In Screen

Eleven tasks. Context: `.planning/Frontend/responsive-logged-in-screen.md`. Structure and tokens are
already settled in `logged_in_design.md` and `design-system.md` — nothing here redesigns the screen,
it only replaces `components/session/mock.ts` with the RPCs that now exist.

| Task | File | Change | Risk |
| --- | --- | --- | --- |
| 1 | `task1.md` | `@keyframes shimmer` + `lib/format-time.ts` | Isolated |
| 2 | `task2.md` | Service clients + `useProjects.ts` | Isolated |
| 3 | `task3.md` | `useMessages.ts` — the send mutation | The cache rules are the subtle part |
| 4 | `task4.md` | `providers.ts` + `ComposerControls` | Refactor, no behaviour change |
| 5 | `task5.md` | `generating.tsx` — the cycling shimmer | Isolated |
| 6 | `task6.md` | `ChatPanel` on real messages | The thread's pending state |
| 7 | `task7.md` | `HomeView` gains the two selectors | Small |
| 8 | `task8.md` | `PreviewPane` — real iframe + overlay | The demo's centrepiece |
| 9 | `task9.md` | `project-list.tsx` (from `conversation-list.tsx`) | Isolated, not yet wired |
| 10 | `task10.md` | `LoggedInScreen` cutover, `mock.ts` deleted | **Where everything connects** |
| 11 | `task11.md` | End-to-end demo pass + doc updates | The demo |

## Order

Strictly 1 → 11. The order is chosen so **every task leaves the app compiling and running**; only
task 10 changes behaviour on more than one screen at a time.

Two deliberate transients to expect:

- After task 6, `ChatPanel` queries `ListMessages` with fixture ids (`"c1"`), which the backend does
  not have. The thread renders empty until task 10. That is correct, not a bug.
- Tasks 4 and 9 create files that nothing imports yet. They are wired in tasks 6/7 and 10.

## Depends on

Nothing frontend-side beyond what is already merged. Backend-side this needs **all** of
`tasks/Backend/project-model/`, `tasks/Backend/messages-model/`, `tasks/Backend/ai-service-client/`
and `tasks/AI-Service/sandbox-service/` — all complete on `feature/messages`. No proto change is
needed; `make gen` does not have to run.

## Design system

`.planning/Frontend/design-system.md` governs every visual decision here and none of them are
re-opened. The three of its rules this feature is most likely to break, because the generating state
is new surface area:

- **Rule 3 — no translucency, and §3's "do not add `backdrop-blur-*`".** The preview's generating
  overlay is opaque `bg-surface`. Task 8 explains why the arbitrary generated page behind it makes
  this stricter here, not looser.
- **Rule 4 — no shadows, no internal dividers.** The thread, the list and the overlay separate by
  surface tone and space. `shadow-none` when overriding a vendored component.
- **§6 — the tailwind-merge trap.** Every `dark:` on a vendored base needs a `dark:` on the override.
  Task 4 moves existing overrides between files, which is exactly where one gets dropped.

Shimmer is reserved for the generation flow (§8) — it means "the model is working" and appears
nowhere else. `--accent-blue` is the one non-system hue and rule 1 requires its call sites to be
listed in `design-system.md`; task 11 adds the generating row's icon to that list.

## The five things most likely to slip through

1. **The shimmer renders static.** `TextShimmer` has always been broken — no `@keyframes shimmer`
   exists and Tailwind v4 will not generate one. It fails silently. Task 1.
2. **A request timeout capping the agent run.** A Build send legitimately blocks for minutes. Any
   `timeoutMs` anywhere turns a working build into a failure. Tasks 2 and 3.
3. **The preview URL never appearing.** `SendMessageResponse` does not carry it — the project must
   be invalidated after the send. Task 3.
4. **A second send while one is in flight.** The composer must be fully locked, not just the button.
   Task 6.
5. **Refreshing the iframe when the reply lands.** The sandbox hot-reloads itself; a `key` bump
   tears down the HMR socket and looks worse. Task 8.

## Out of scope

See `responsive-logged-in-screen.md` §8. Short version: no dead-sandbox handling, no Chat gating, no
streaming, no pagination, no routing, no mobile layout. Each has an owner in `Polish.md`.
