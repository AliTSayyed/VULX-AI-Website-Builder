# Task 1 — Migration `00005_add_messages.sql`

**Goal:** the `messages` table.

Source: `.planning/Backend/messages-model.md` §2.

## Files

| File | Change |
| --- | --- |
| `.../persistence/postgres/migrations/00005_add_messages.sql` | **new** |

## The migration

```sql
-- +migrate Up
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL,
    mode VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_project_created
    ON messages (project_id, created_at ASC, id ASC);

-- +migrate Down
DROP TABLE IF EXISTS messages CASCADE;
```

## Notes

**`00005`, not `00006`.** `project_codebases` is Polish and takes `00006` if it is ever built. Do not
leave a gap; `sql-migrate` runs these in filename order on API boot and a failure **panics the API**.

**`ASC` in the index, unlike `projects`.** A thread reads oldest-first; the project list reads
newest-activity-first. The index must mirror its query's `ORDER BY` or it is not used.

**`body` is `NOT NULL`** — an agent returning nothing still yields a placeholder summary from the AI
service, so there is no legitimate empty case. A failed run is a body that says so, not a null.

**`role`, `mode`, `provider` are `VARCHAR`, not Postgres `ENUM`s**, following
`user_auth_providers.provider`. A Postgres enum needs a migration to add a value; adding a fourth
model provider should be a Go and proto change.

**No `status` column.** It is needed only for asynchronous Build, which the MVP does not do. Its
shape depends on a poll-vs-stream decision that has not been made — adding it now would pre-commit
the answer.

## Verifying

```bash
make nuke && make
docker compose logs api | grep -i migrat
docker compose exec sql psql -U postgres -d local -c '\d messages'
```

Confirm the cascade — the only behaviour here not observable until much later:

```sql
INSERT INTO projects (user_id, title, provider) VALUES ('<user-uuid>','cascade','anthropic') RETURNING id;
INSERT INTO messages (project_id, role, mode, provider, body)
  VALUES ('<project-uuid>','user','build','anthropic','hello');
DELETE FROM projects WHERE title = 'cascade';
SELECT count(*) FROM messages;   -- must be 0
```

## Done when

- `\d messages` shows eight columns, the FK, and the `ASC` index.
- The API starts without a migration panic.
