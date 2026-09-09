# Task 4 — `project_service.go`

**Goal:** the use cases, the ports, and the ownership check.

Source: `.planning/Backend/project-model.md` §6.

## Files

| File | Change |
| --- | --- |
| `api/internal/application/services/project_service.go` | **new** |

## 1. The port

Declared **here**, in the consumer's file — not in a `ports/` package.

```go
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) (*domain.Project, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	FindAllByUser(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error)
	UpdateSandbox(ctx context.Context, id uuid.UUID, sandboxID, previewURL string) error
}
```

Four methods — exactly what task 3 implements.

**`ProjectTitler` is not declared in the MVP.** `.planning/Backend/project-model.md` §9 suggests a
stub implementation; not declaring the port at all is the same idea one step further, and avoids an
injected dependency nothing calls. It arrives with LLM titling in Polish, as an additive change to
`NewProjectService`'s signature.

```go
type ProjectService struct {
	projectRepo ProjectRepository
}

func NewProjectService(projectRepo ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}
```

## 2. `List`

```go
func (s *ProjectService) List(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error) {
	limit = utils.Clamp(limit, 10, 100)
	page, err := s.projectRepo.FindAllByUser(ctx, userID, limit, token)
	if err != nil {
		return nil, domain.WrapError("project service list", err)
	}
	return page, nil
}
```

`utils.Clamp(limit, 10, 100)` exactly as `UserService.List` does — a client sending `limit: 0` gets
10, not zero rows.

## 3. `Get` — and the ownership check

```go
func (s *ProjectService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.WrapError("project service get", err)
	}
	if project.UserID() != userID {
		return nil, domain.NewError(domain.ErrorTypeNotFound, fmt.Errorf("project %s not found", id))
	}
	return project, nil
}
```

**`NotFound`, not `PermissionDenied`.** `PermissionDenied` confirms the id exists and belongs to
someone else — an enumeration oracle. The user cannot act on the difference; the server log can,
because the repository already returned the real owner.

**Every method that takes a project id runs this check.** There is no interceptor that will do it.
In the MVP `Get` is the only such method here — but `MessageService` composes this service precisely
so the check exists in one place.

## 4. `Create`

```go
func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, firstPrompt string, provider domain.AIProvider) (*domain.Project, error) {
	title := provisionalTitle(firstPrompt)      // rejects empty — see below
	project, err := domain.NewProject(userID, title, provider)
	if err != nil {
		return nil, domain.WrapError("project service create", err)
	}
	created, err := s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, domain.WrapError("project service create", err)
	}
	return created, nil
}
```

`provisionalTitle` is an unexported helper in this file:

- collapse runs of whitespace to single spaces, trim
- if the result is empty → return `""`, and let `domain.NewProject` reject it with
  `ErrProjectTitleEmpty` (`ErrorTypeInvalid` → `CodeInvalidArgument`)
- truncate to ~60 characters **on a word boundary**, no ellipsis needed

**Rejecting an empty prompt is the entire mechanism that stops empty projects existing.** A user who
clicks "+ New build" and types nothing must leave nothing behind, and the frontend achieves that by
not calling `CreateProject` at all until the first send. No `is_draft` column, no TTL sweep, no
reaper — just this validation.

`domain.NewProject` also rejects `AIProviderUnspecified`, so an unspecified provider is a `400`, not
a silent default. The composer always has a concrete selection; unspecified means a client bug.

## 5. Error wrapping

Every return path uses `domain.WrapError("project service <verb>", err)`. `WrapError` preserves the
**innermost** `ErrorType`, so a repository `NotFound` is still `NotFound` at the handler. Returning a
bare `fmt.Errorf` anywhere makes it `ErrorTypeUnknown` → `CodeUnknown`.

## Not in this task

`Rename` — Polish. No LLM call, no goroutine, no `context.WithoutCancel`.

## Verifying

```bash
cd api && go build ./... && go vet ./...
```

Nothing constructs `ProjectService` yet — task 5 does. Check by reading:

- Does `Get` compare `project.UserID()` against the passed `userID`? If not, the RPC leaks.
- Does `Create` reject `""` and a whitespace-only prompt?
- Is `limit` clamped?

## Done when

- `go build ./... && go vet ./...` clean.
- The ownership check returns `ErrorTypeNotFound`.
- Every error is wrapped with `domain.WrapError`.
