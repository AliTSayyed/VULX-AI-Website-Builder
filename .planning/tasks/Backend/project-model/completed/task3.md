# Task 3 — Cursor rename + `project_repository.go`

**Goal:** persistence for projects, with keyset pagination that actually matches its `ORDER BY`.

Source: `.planning/Backend/project-model.md` §5.

## Files

| File | Change |
| --- | --- |
| `.../persistence/postgres/cursor.go` | rename one parameter |
| `.../persistence/postgres/project_repository.go` | **new** |

## 1. `cursor.go` — rename the parameter

```go
func encodeToken(ts time.Time) string          // was: createdAt
func decodeToken(token string) (time.Time, error)
```

Update the doc comment too. **Zero behaviour change** — the token is an opaque base64 RFC3339Nano
timestamp and nothing binds it to any column. `user_repository.go` passes positionally, so no call
site changes.

The rename exists so the next reader does not conclude the project list is sorted wrongly: projects
page on `updated_at`, users page on `created_at`, one codec serves both.

## 2. Row struct

```go
type Project struct {
	ID         uuid.UUID      `db:"id"`
	UserID     uuid.UUID      `db:"user_id"`
	Title      string         `db:"title"`
	Provider   string         `db:"provider"`
	SandboxID  sql.NullString `db:"sandbox_id"`
	PreviewURL sql.NullString `db:"preview_url"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

func (p *Project) ToDomain() *domain.Project {
	return domain.RestoreProject(
		p.ID, p.UserID, p.Title, p.Provider,
		p.SandboxID.String, p.PreviewURL.String,   // "" when invalid — that is the point
		p.CreatedAt, p.UpdatedAt,
	)
}
```

`sql.NullString` is confined to this struct. `.String` is `""` when the column is NULL, which is
exactly what the domain wants — the nullability stops at the persistence boundary.

Constructor `NewProjectRepository(db *sqlx.DB) *ProjectRepository`, same as `NewUserRepository`.

## 3. Methods

### `Create`

```sql
INSERT INTO projects (id, user_id, title, provider, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, user_id, title, provider, sandbox_id, preview_url, created_at, updated_at
```

`QueryRowxContext(...).StructScan(&dbProject)`, then `dbProject.ToDomain()`. **Return the
re-hydrated row, not the input** — the DB assigns the timestamps.

Write the provider with `project.Provider().String()`.

### `FindByID`

```sql
SELECT id, user_id, title, provider, sandbox_id, preview_url, created_at, updated_at
FROM projects WHERE id = $1
```

**Not user-scoped.** The service compares owners (task 4). Map `sql.ErrNoRows` to
`ErrorTypeNotFound`:

```go
if errors.Is(err, sql.ErrNoRows) {
	return nil, domain.NewError(domain.ErrorTypeNotFound, fmt.Errorf("project %s not found: %w", id, err))
}
return nil, domain.NewError(domain.ErrorTypeInternal, ...)
```

Follow `FindByEmail`, **not** `FindByID` in `user_repository.go` — that one maps `ErrNoRows` to
`Internal`, which is a latent bug. Do not copy it.

### `FindAllByUser` — the subtle one

```sql
SELECT id, user_id, title, provider, sandbox_id, preview_url, created_at, updated_at
FROM projects
WHERE user_id = $1
  AND ($2::timestamptz IS NULL OR updated_at < $2)
ORDER BY updated_at DESC, id DESC
LIMIT NULLIF($3, 0)
```

```go
var cursor *time.Time
if !ts.IsZero() {
	cursor = &ts
}
err = p.db.SelectContext(ctx, &dbProjects, query, userID, cursor, limit+1)
```

Then the house convention: query `limit+1`, and only if more came back do you trim and mint a token.

```go
var nextToken string
if len(dbProjects) > int(limit) {
	dbProjects = dbProjects[:limit]
	nextToken = encodeToken(dbProjects[len(dbProjects)-1].UpdatedAt)
}
projects := make([]*domain.Project, 0, len(dbProjects))
for _, row := range dbProjects {
	projects = append(projects, row.ToDomain())
}
return &domain.Page[*domain.Project]{Items: projects, Token: nextToken, HasMore: nextToken != ""}, nil
```

**Four ways to get this wrong:**

1. **Paginating on `created_at`.** The `WHERE` and the `ORDER BY` must reference the same column, or
   pages overlap and skip. It is `updated_at` here — the sidebar sorts by last activity.
2. **`encodeToken` on the wrong row.** It is the *last kept* row, after trimming.
3. **`$2::timestamptz`, not `$2::timestamp`.** The column is `TIMESTAMP WITH TIME ZONE`.
   `user_repository.go` uses `::timestamp`, which forces an implicit session-timezone conversion —
   it happens to work there and is not worth replicating.
4. **Forgetting `limit+1`.** Then `HasMore` is always false and the sidebar silently stops paging.

### `UpdateSandbox`

```sql
UPDATE projects SET sandbox_id = $2, preview_url = $3 WHERE id = $1
```

`ExecContext`; check `RowsAffected() == 0` → `ErrorTypeNotFound`. Returns `error` only.

**Deliberately does not touch `updated_at`.** Attaching a sandbox is not authorship and must not
reorder the sidebar. In the MVP its caller is `SendMessage`, which already bumps the row via its own
statement.

## Not in this task

`UpdateTitle`, `UpdateProvider` (Polish), and `Touch` — which does not exist at all, because the
parent bump happens inside `messageRepo.Create` (`messages-model.md` §4.2).

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

Then exercise the SQL directly, since nothing calls it yet:

```sql
INSERT INTO projects (user_id, title, provider) VALUES ('<user-uuid>', 'a', 'anthropic');
INSERT INTO projects (user_id, title, provider) VALUES ('<user-uuid>', 'b', 'openai');
SELECT title, updated_at FROM projects
 WHERE user_id = '<user-uuid>' ORDER BY updated_at DESC, id DESC;
```

Confirm the index is used rather than a sequential scan:

```sql
EXPLAIN SELECT id FROM projects WHERE user_id = '<uuid>' ORDER BY updated_at DESC, id DESC LIMIT 11;
```

## Done when

- `go build ./... && go vet ./...` clean.
- `cursor.go`'s parameter is `ts` and `user_repository.go` still compiles untouched.
- Every error return is a `*domain.Error` with a deliberate type — a bare `fmt.Errorf` becomes
  `ErrorTypeUnknown` at the handler, which surfaces as `CodeUnknown` in the browser.
