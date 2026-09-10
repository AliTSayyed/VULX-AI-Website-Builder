# Task 8 — `PreviewPane`: the real iframe and the generating overlay

**Goal.** The placeholder paragraph becomes a live sandbox, with an overlay while the agent works.
This is the thing the demo is actually about.

## New props

```tsx
type PreviewPaneProps = {
  /** Project.preview_url. "" until a sandbox exists — normalise to null. */
  url: string | null;
  generating: boolean;
  trigger?: ReactNode;
};
```

## Three states, one surface

| State | Chrome pill | Body |
| --- | --- | --- |
| no url, not generating | `No sandbox running` | "No preview yet. A sandbox is created on the first Build message." |
| generating | `Generating…` | the last frame (iframe if one exists, placeholder if not) **under** the overlay |
| url, not generating | the url | the iframe |

The existing chrome row (height, `bg-surface-2` pill, the two ghost buttons, the sidebar `trigger`)
does not change shape.

## The iframe

```tsx
<iframe
  ref={frame}
  src={url}
  title="Sandbox preview"
  className="size-full rounded-xl border-0 bg-white"
  sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
/>
```

- **`bg-white`, not `bg-background`.** The generated Next.js site is a light-mode page; a dark ground
  flashes behind it on load. This is not a palette exception — it is the foreign document's own
  ground showing through an element we do not style, not a VULX surface.
- **Do not re-key or re-`src` the iframe when a build finishes.** The sandbox runs `next dev` with
  hot reload and the HMR websocket lives *inside* the frame — it repaints itself as the agent writes
  files. Tearing it down would reload a page that is already correct and lose the visible
  live-update effect. The only thing that changes at the end of a build is the overlay coming off.
- `sandbox` keeps `allow-same-origin` because the E2B preview needs its own origin's storage and
  websocket; without it HMR breaks. It is a third-party origin either way.
- The URL from the API is already scheme-qualified (`ai-service/api/routes/sandbox.py` prepends
  `https://` — E2B's `get_host()` returns a bare host, which fails silently as an `src`). Do no URL
  building here.

## The overlay

Absolutely positioned over the body area (the body wrapper needs `relative`), covering the iframe
edge to edge inside the same `rounded-xl`:

```tsx
{generating && (
  <div className="bg-surface vx-fade absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 rounded-xl">
    <GeneratingLine className="text-sm" />
    <p className="text-muted-foreground max-w-xs px-6 text-center text-xs text-balance">
      Writing files into your sandbox. This takes a minute or two.
    </p>
  </div>
)}
```

- **Opaque, and no `backdrop-blur`.** `design-system.md` rule 3 and §3 are explicit: no translucent
  surfaces, and `backdrop-blur-*` behind an opaque fill is a no-op that still costs a paint. The
  reasoning applies here harder than anywhere else — what sits behind this overlay is an *arbitrary
  generated website*, which may be white, dark, or animating. A translucent scrim would make the
  shimmer's contrast a moving target with a backdrop we do not control, which is exactly the bug §3
  documents against the silk. `bg-surface` holds the measured 17.05 / 10.12 / 7.64 ladder.
- `bg-surface`, not `bg-background`: the pane is already a `bg-surface` panel, so the overlay reads
  as the panel's own material with the viewport temporarily empty, not as a third layer.
- The overlay must sit *inside* the rounded body, not over the chrome row — the reload and
  open-in-new-tab buttons stay reachable, and the URL pill stays visible.
- `vx-fade` on mount so it does not pop.

## The two chrome buttons

Both are currently rendered `disabled` and wired to nothing.

- **Reload**: `frame.current?.contentWindow?.location.reload()` will throw on a cross-origin frame —
  use `if (frame.current) frame.current.src = frame.current.src;` instead. This is a *user-initiated*
  reload, which is legitimate; the rule above is only about automatic ones.
- **Open in new tab**: render as `<a href={url} target="_blank" rel="noreferrer">` styled with the
  existing button classes (`buttonVariants` or `asChild`), not an `onClick` + `window.open` — a real
  anchor gets middle-click and the browser's own affordances.
- Both stay `disabled` (the anchor: not rendered as a link) when there is no url.

## Done when

- `npm run lint` and `npm run build` pass.
- `LoggedInScreen` passes `generating={false}` and `url={null}` for now; the pane shows the
  no-sandbox state.
- Temporarily hard-coding a known-good E2B URL renders the site in the frame, and hard-coding
  `generating` shows the overlay with the shimmer over it.
