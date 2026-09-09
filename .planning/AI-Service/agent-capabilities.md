# Agent Capabilities — AI Service

> Gaps between what the agent can do in a sandbox and what survives being persisted and replayed.
> Three items: **the agent cannot delete a file**, **paths are not canonical**, and **`npm install`
> mutates a `package.json` nothing captures**.
>
> Lowest priority of the four AI-service documents — **nothing is blocked on it**. Each item is a
> fidelity problem that only appears once `project-codebase-model.md` is replaying real projects.
> Read it then, not before.
>
> Verified against the code on `feature/messages` (2026-09-08).
>
> ## ⏸ Phase: POLISH — none of this is in the MVP
>
> **Every item here is a fidelity problem that only exists once a codebase is persisted and
> replayed**, which the MVP does not do (`project-codebase-model.md`).
>
> - The delete-tool gap matters because replay resurrects deleted files. No replay, no resurrection.
> - Path canonicalisation matters because two keys for one file makes replay non-deterministic. No
>   stored map, no keys.
> - `package.json` read-back matters because replayed installs drift. No replay.
>
> Read this alongside `project-codebase-model.md`, not before it.


## 1. Why these three are one document

They share a cause. The sandbox is no longer the codebase — Postgres is
(`project-codebase-model.md` §1). Once a file map is persisted and replayed into fresh sandboxes,
every gap between "what the agent did" and "what was recorded" becomes a divergence that compounds
over a project's life.

None of them break anything today. All three produce a sandbox that quietly does not match what the
user built.

## 2. There is no delete tool

`SandboxService` exposes four tools: `list_sandbox_files`, `read_sandbox_file`,
`write_sandbox_files`, `execute_sandbox_command`. Nothing deletes.

Consequences, in order of severity:

1. **The persisted map can only grow.** A file the agent removes with
   `execute_sandbox_command("rm app/old-page.tsx")` is still a key in `files`, so **the next refresh
   resurrects it** (`project-codebase-model.md` §10.1).
2. That resurrection is worse than a stale file. A deleted-then-restored component can shadow its
   replacement or break a build that succeeded before the refresh — and the user has no way to see
   why, because the file is not in any diff they saw.
3. The agent works around the gap with `rm`, which the callback stores as a command… except
   `project-codebase-model.md` §8 filters stored commands to `npm install` only. So the deletion is
   discarded — correctly, since replaying arbitrary shell is a hazard, but the deletion is lost
   either way.

### 2.1 The fix

A `SandboxDeleteTool` (`delete_sandbox_files`, batched, taking a list of paths), a
`SandboxService.delete_files` method, and — the part that matters — **callback capture**:

```python
class CodeAgentCallBackResult(BaseModel):
    updated_files: Dict[str, str]
    deleted_files: List[str]        # new
    commands_executed: List[str]
```

`AICodeAgentResponse` gains `deleted: List[str]`, and Go applies it as a JSONB key removal
alongside the merge. `project-codebase-model.md` §5.2's single statement becomes:

```sql
SET files = (project_codebases.files || EXCLUDED.files) - $4::text[]
```

`-` with a `text[]` removes keys from a JSONB object. Still one statement, still atomic — apply the
writes, then subtract the deletions. **Order matters and is not arbitrary:** a file written and then
deleted in the same run must end deleted.

This is the item with real value, and it is the only one of the three that changes the Go side.

Add the tool to `NEXTJS_PROMPT`'s tool list, and add a rule: **delete files with the delete tool,
never with `rm`.** A capability the prompt does not mention is one the agent will not reliably use.

### 2.2 Guard it

`delete_files` must refuse the same forbidden paths `list_files` already refuses (`/`, `/root`,
`/etc`, `/sys`, `/proc`) — and, unlike listing, should refuse anything outside `/home/user/`
entirely. A destructive tool deserves a stricter allowlist than a read-only one.

## 3. Paths are not canonical, and the prompt contradicts itself

`NEXTJS_PROMPT` says both of these:

> File paths for write_sandbox_files can be absolute (e.g., "/home/user/app/page.tsx") or relative
> (e.g., "app/page.tsx") — if relative, /home/user/ will be automatically prepended.

> When calling write_sandbox_files, always use absolute file paths starting with "/home/user/"

Permissive in one place, mandatory in another. In the sandbox both resolve to the same file, so this
never mattered — E2B's working directory is `/home/user`.

**It matters now.** The callback stores paths exactly as the agent typed them, and Go stores them
unchanged (`project-codebase-model.md` §3). So an agent that writes `app/page.tsx` in one run and
`/home/user/app/page.tsx` in the next produces **two JSONB keys for one file**, holding two
different versions.

On replay both are written. Go maps have no iteration order, so **which version wins is
non-deterministic** — a refresh can silently roll a file back to an older revision, and do it
differently next time.

### 3.1 The fix — both halves

**Normalise in `write_files`**, before the callback can see it:

```python
def _canonical(path: str) -> str:
    p = path if path.startswith("/") else f"/home/user/{path}"
    return posixpath.normpath(p)
```

Apply it in `SandboxService.write_files` (and `delete_files`) so the stored key is canonical no
matter what the agent typed. **This is the half that actually fixes it** — it does not depend on the
model complying with anything.

**Then fix the prompt** by deleting the permissive line and keeping the mandatory one. Contradictory
instructions cost tokens and attention even when a downstream guard makes them harmless.

`project-codebase-model.md` §3 says "store paths exactly as received; do not normalise" — that
instruction is aimed at **Go**, and it stays correct. Normalisation belongs at the point of capture,
in Python, where the sandbox's own semantics are known. Go should not be inventing path rules for a
filesystem it cannot see.

### 3.2 While in the prompt: two smaller errors

- It describes `write_sandbox_files` as taking "array of file data with **path/content**". The
  actual field is `data` (`WriteEntry(path, data)`). LangChain hands the model the real
  `args_schema`, so this is stale documentation rather than a live bug — but it is stale
  documentation the model reads.
- It instructs `npm install <package> --yes`. `--yes` is an `npx` flag, not an `npm install` one.
  npm tolerates it, but the string is **persisted and replayed** (`project-codebase-model.md` §8),
  so it is worth not carrying noise into stored state.

## 4. `npm install` mutates a `package.json` nothing captures

The prompt forbids editing `package.json` directly and requires `npm install`. npm then rewrites
`package.json` **inside the sandbox** — but the agent never called the write tool on it, so the
callback never sees it, so it is never persisted.

Today this is handled on the Go side by replaying the stored `npm install` commands
(`project-codebase-model.md` §8). That works, with one weakness named there: **it replays the
request, not the resolution.** `npm install some-lib` recorded in March can resolve to a different
version in June, and the replayed sandbox is not the one the user built.

### 4.1 The optional fix

After a code-agent run that executed any `npm install`, read `package.json` back out of the sandbox
and include it in the returned `files` map as an ordinary entry. It then persists, replays, and
pins versions like any other file.

`SandboxService.read_file` already exists and needs no new route — this happens inside
`process_code_request`, not over HTTP.

**Weigh it before doing it.** It adds a round trip to every Build that installs something, and it is
only worth it if version drift actually bites. The command replay is adequate until it isn't. If you
do take it, `package-lock.json` matters more than `package.json` for real reproducibility — and it
is large, which pushes against the size caps in `project-codebase-model.md` §7.3.

## 5. Priority

| Item | Value | Cost | Do it |
| --- | --- | --- | --- |
| §3 path normalisation | Prevents non-deterministic version rollback | ~10 lines, Python only | **First** — cheapest and fixes a real correctness bug |
| §2 delete tool | Stops resurrecting deleted files | New tool + callback field + response field + Go JSONB subtraction | Second — the only one touching Go |
| §4 `package.json` read-back | Version fidelity across replays | A round trip per Build | Only if drift actually bites |

§3 alone is worth doing as soon as replay is real. §2 becomes worth doing the first time someone
asks the agent to remove a file. §4 may never be worth doing.

## 6. Verifying

```bash
cd ai-service && ruff check .
docker compose restart ai-service
```

**§3 — the canonical-path check.** Two runs against one sandbox, one asking for a relative write and
one absolute, to the same logical file:

```bash
SB=$(curl -s -X POST http://localhost:9999/ai-service/v1/sandbox/ | jq -r .id)
# …two code-agent calls…
# then confirm the returned `files` maps use ONE key, /home/user/app/page.tsx — never both forms
```

**§2 — the delete check.** Ask the agent to create a file, confirm it appears in `files`; then ask
it to delete that file and confirm the path appears in `deleted` and **not** in `files`. Then, on
the Go side once wired, refresh and confirm it does not come back.

Also confirm the guard: a request that would delete outside `/home/user/` is refused (§2.2).

**§4 — the read-back check.** A run that installs a package must return `package.json` among
`files`, with the new dependency in it.

## 7. Out of scope

The callback's success test (`agent-result-capture.md`) · the HTTP contract and sandbox lifetime
(`sandbox-service.md`) · conversation history (`conversation-history.md`) · a general-purpose file
API for the agent · binary assets — `files` is `Dict[str, str]` and the agent has no way to produce
one · supporting a stack other than Next.js, which needs a new template and a new prompt
(`ARCHITECTURE.md` §7.6).
