# Task 3 — `message_repository.go`

**Goal:** persist a message and move its project up the sidebar, atomically.

Source: `.planning/Backend/messages-model.md` §4.

## Files

| File | Change |
| --- | --- |
| `.../persistence/postgres/message_repository.go` | **new** |

## Row struct

```go
type Message struct {
	ID        uuid.UUID `db:"id"`
	ProjectID uuid.UUID `db:"project_id"`
	Role      string    `db:"role"`
	Mode      string    `db:"mode"`
	Provider  string    `db:"provider"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (m *Message) ToDomain() *domain.Message {
	return domain.RestoreMessage(m.ID, m.ProjectID, m.Role, m.Mode, m.Provider,
		m.Body, m.CreatedAt, m.UpdatedAt)
}
```

No `sql.NullString` — every column is `NOT NULL`.

## `Create` — the CTE

```sql
WITH inserted AS (
    INSERT INTO messages (id, project_id, role, mode, provider, body, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    RETURNING id, project_id, role, mode, provider, body, created_at, updated_at
), bumped AS (
    UPDATE projects SET updated_at = NOW() WHERE id = $2
)
SELECT id, project_id, role, mode, provider, body, created_at, updated_at FROM inserted
```

`QueryRowxContext(...).StructScan(&dbMessage)`, then `dbMessage.ToDomain()`.

Write the enums with `.String()`: `message.Role().String()`, `message.Mode().String()`,
`message.Provider().String()`.

### Why a CTE and not two statements

The sidebar orders projects by `updated_at`, and inserting a *message* does not move the *project*
row. Two separate statements are non-atomic and this repo has no transaction helper; a failure
between them leaves a message whose project never moved.

**A data-modifying CTE executes exactly once even when nothing references it**, so `bumped` runs
despite being unreferenced. Both sub-statements see the same snapshot and commit together.

**This is decided — do not substitute `messageRepo.Create` followed by `projectRepo.Touch`.**
`ProjectRepository` has no `Touch` method, deliberately.

**The accepted cost:** there is now one `UPDATE projects` living in `message_repository.go`, which
contradicts "all `projects` SQL lives in `project_repository.go`". Both planning documents name this
as the exception from their own side, so a `grep` for `UPDATE projects` finding it here is expected,
not a mistake.

### Three ways to get the CTE wrong

1. **Omitting the explicit column list** in the final `SELECT`. `SELECT *` over a CTE works but makes
   the column order implicit, and `StructScan` maps by `db` tag — be explicit anyway so a future
   column addition cannot reorder anything.
2. **Referencing `bumped`** in the final `SELECT`. It has no `RETURNING`; referencing it is a syntax
   error. Leave it unreferenced.
3. **Using a different parameter for the project id.** `$2` appears in both sub-statements on
   purpose — the message's project and the bumped project must be the same row.

### Error mapping

A foreign-key violation means the project vanished between the ownership check and the insert. That
is a client-visible fact, not an internal failure:

```go
var pqErr *pq.Error
if errors.As(err, &pqErr) && pqErr.Code == "23503" {
	return nil, domain.NewError(domain.ErrorTypeNotFound,
		fmt.Errorf("project %s not found: %w", message.ProjectID(), err))
}
return nil, domain.NewError(domain.ErrorTypeInternal, ...)
```

`lib/pq` is already a dependency.

## `FindByProject`

```sql
SELECT id, project_id, role, mode, provider, body, created_at, updated_at
FROM messages
WHERE project_id = $1
ORDER BY created_at ASC, id ASC
```

`SelectContext` into `[]*Message`, map each with `ToDomain()`.

**Returns an empty slice, not `NotFound`, for a project with no messages.** A project created before
its first send legitimately has an empty thread — that is a normal state the Workspace renders as an
empty composer, not an error.

**Unpaginated on purpose.** A thread is a handful of messages and the Workspace renders all of them.
`ORDER BY ... ASC` also means `cursor.go`'s `DESC` keyset codec would not fit without new machinery.

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

Then prove the CTE in psql, since nothing calls it yet — **this is the check that matters**:

```sql
SELECT id, updated_at FROM projects WHERE title = '<your project>';
-- note updated_at, then run the CTE by hand with real ids
-- then:
SELECT id, updated_at FROM projects WHERE title = '<your project>';   -- must have moved
SELECT count(*) FROM messages;                                        -- must be 1
```

Both must be true from a **single** statement execution. If the message exists but `updated_at` did
not change, the `bumped` CTE is not running — check it is a data-modifying CTE and not accidentally
commented out or wrapped in a `SELECT`.

Then the FK path:

```sql
INSERT INTO messages (project_id, role, mode, provider, body)
  VALUES (gen_random_uuid(),'user','build','anthropic','x');   -- must fail 23503
```

## Done when

- One statement inserts the message **and** bumps the project.
- A bad `project_id` maps to `ErrorTypeNotFound`, not `Internal`.
- `go build ./... && go vet ./...` clean.
