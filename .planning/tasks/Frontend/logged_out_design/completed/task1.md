# Task 1 — Tokens and shell

**Goal.** Black page, white text, and an app that compiles. Right now it does not.

## 1. `app/src/app/globals.css`

Replace the entire `.dark { … }` block with the palette from `logged_out_design.md` §2. Leave
`:root` and `@theme inline` alone.

## 2. `app/src/app/layout.tsx`

Add `dark` to the `<html>` tag and replace the `create-next-app` metadata:

```tsx
export const metadata: Metadata = {
  title: "VULX — Your personal AI website creator",
  description: "Describe what you want. Watch it build itself.",
};

// …

return (
  <html lang="en" className="dark">
    <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
      {children}
    </body>
  </html>
);
```

## 3. `app/src/app/page.tsx`

**Delete everything in it.** It calls `userService.createUser({ name: "tony" })` and reads
`user.name`; the generated type has `first_name`, `last_name`, `email`. There is no `name` field,
so the build fails today. Temporary placeholder — task 6 replaces it again:

```tsx
const Page = () => {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <span className="text-muted-foreground text-sm">VULX</span>
    </div>
  );
};

export default Page;
```

No `"use client"` — nothing here needs it yet.

## Verify

```bash
docker compose exec app npm run build
```

It should pass. It does not pass before this task.

## Done when

- `https://local.app.vulx.ai` is a pure black page with small grey centred text.
- DevTools → Elements: `<html class="dark">`, and `body` computed `background-color` is `rgb(0, 0, 0)`.
- `npm run build` exits 0.
