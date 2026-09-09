# Task 1 — Migration `00004_add_projects.sql`

**Goal:** the `projects` table.

Source: `.planning/Backend/project-model.md` §2.

## Files

| File | Change |
| --- | --- |
| `api/internal/infrastructure/persistence/postgres/migrations/00004_add_projects.sql` | **new** |

## The migration

```sql
-- +migrate Up
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    sandbox_id VARCHAR(255),
    preview_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_projects_user_updated
    ON projects (user_id, updated_at DESC, id DESC);

-- +migrate Down
DROP TABLE IF EXISTS projects CASCADE;
```

## Notes

**`00004` is correct and there must be no gap.** `messages` takes `00005`;
`project_codebases` takes `00006` if and when it is built. Migrations are `go:embed`ed and run by
`sql-migrate` on API boot — **a failing migration panics the API**, so a syntax error here presents
as a crash loop, not an error message.

**Never edit this file once it has been applied.** `make nuke` drops the volumes for a clean re-run.

**`NOT NULL` on the timestamps diverges from `users` deliberately.** Keyset pagination reads
`updated_at` off the last row of a page; a `NULL` there produces an unusable cursor. Do not "fix"
`users` to match — it is not being changed.

**The index column order is not arbitrary.** `(user_id, updated_at DESC, id DESC)` mirrors the
`WHERE` and `ORDER BY` in task 3's `FindAllByUser` exactly. Changing one without the other silently
drops the index.

**`sandbox_id` and `preview_url` are nullable** because a project has no sandbox until its first
Build message. They are written once, by `SendMessage`.

## Verifying

```bash
make nuke && make
docker compose logs api | grep -i migrat
docker compose exec sql psql -U postgres -d local -c '\d projects'
docker compose exec sql psql -U postgres -d local -c '\di idx_projects_user_updated'
```

Confirm the cascade, which is the only behaviour here not otherwise observable until much later:

```sql
-- get a real user id first: SELECT id FROM users LIMIT 1;
INSERT INTO projects (user_id, title, provider) VALUES ('<uuid>', 'cascade check', 'anthropic');
DELETE FROM users WHERE id = '<uuid>';
SELECT count(*) FROM projects;   -- must be 0
```

(Then `make nuke` again — you just deleted your logged-in user.)

## Done when

- `\d projects` shows all eight columns, the FK, and the index.
- The API starts without a migration panic.
