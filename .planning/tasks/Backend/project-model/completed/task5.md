# Task 5 — Handler and wiring

**Goal:** the three RPCs reachable over Connect and REST. First task where the feature actually runs.

Source: `.planning/Backend/project-model.md` §8.

## Files

| File | Change |
| --- | --- |
| `api/internal/infrastructure/inbound/handlers/project.go` | **new** |
| `api/internal/application/app.go` | wire it up |

## 1. The handler

Model on `handlers/user.go`, **minus the superuser check** — projects are product surface, so the
hardcoded `alitsayyed@gmail.com` literal must not be copied.

```go
type ProjectServiceHandler struct {
	apiv1connect.UnimplementedProjectServiceHandler

	projectService *services.ProjectService
	authAdapter    *authAdapter.HTTPAuthAdapter
}

func NewProjectServiceHandler(projectService *services.ProjectService, authAdapter *authAdapter.HTTPAuthAdapter) *ProjectServiceHandler
```

Embedding `UnimplementedProjectServiceHandler` means an RPC you have not written returns
`Unimplemented` rather than failing to compile — convenient, and the reason to double-check all three
methods are actually present.

Every method is the same four steps:

```go
user, err := authAdapter.User(ctx)   // omit these 3 lines and the RPC is PUBLIC
if err != nil {
	return nil, err                  // already a connect.CodeUnauthenticated error
}
```

then `uuid.Parse` the id where there is one (mapping failures through
`errorAdapter.ToConnectError`), call the service with `user.ID()`, and map the result.

**Never construct a `connect.Error` for a domain failure.** `errorAdapter.ToConnectError` is the
single translation point.

### The three methods

- `ListProjects` → `projectService.List(ctx, user.ID(), req.Msg.GetLimit(), req.Msg.GetToken())`,
  then build `[]*apiv1.Project` from `page.Items` and set `Token` / `HasMore`.
- `GetProject` → `uuid.Parse(req.Msg.GetId())`, then `projectService.Get(ctx, user.ID(), id)`.
- `CreateProject` → map the proto provider to `domain.AIProvider`, then
  `projectService.Create(ctx, user.ID(), req.Msg.GetFirstPrompt(), provider)`.

### Converters, package-level in this file

```go
func projectToProto(p *domain.Project) *apiv1.Project {
	if p == nil {
		return nil
	}
	return &apiv1.Project{
		Id:         p.ID().String(),
		Title:      p.Title(),
		Provider:   providerToProto(p.Provider()),
		SandboxId:  p.SandboxID(),
		PreviewUrl: p.PreviewURL(),
		CreatedAt:  p.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt:  p.UpdatedAt().UTC().Format(time.RFC3339),
	}
}
```

Plus `providerToProto(domain.AIProvider) apiv1.AiProvider` and
`providerFromProto(apiv1.AiProvider) domain.AIProvider` — both total, both defaulting to
unspecified. The domain enum and the proto enum stay separate because **the domain must not import
generated code**.

**Timestamps: `t.UTC().Format(time.RFC3339)`, never `t.String()`.** Go's default layout
(`2006-01-02 15:04:05.999999999 -0700 MST`) is parsed inconsistently by browsers. This converter is
the only place timestamps are formatted.

**Generated Go field names are `SandboxId` and `PreviewUrl`** — protoc-gen-go's camel-casing, not a
typo.

## 2. `app.go`

Three additions under the existing comment banners:

```go
// persistance
projectRepo := postgres.NewProjectRepository(db)

// business logic
projectService := services.NewProjectService(projectRepo)

// handlers
projectServiceHandler := handlers.NewProjectServiceHandler(projectService, connectAuthAdapter)
```

and one entry in the vanguard slice:

```go
vanguard.NewService(apiv1connect.NewProjectServiceHandler(projectServiceHandler, interceptor)),
```

**The one trap in this file:** the local `services := []*vanguard.Service{...}` **shadows the
imported `services` package**. Every `services.NewXService(...)` call must appear *above* that line.
It already does — keep it that way, or you get a confusing "services is not a package" error.

## Verifying

```bash
cd api && go build ./... && go vet ./...
make api        # hot-reload loop, or `make` for the whole stack
```

Log in through the browser once to get a `jwt` cookie, then:

```bash
curl -i -b "jwt=$JWT" -X POST https://local.api.vulx.ai/api/v1/projects \
  -H 'content-type: application/json' \
  -d '{"first_prompt":"a landing page for a coffee shop","provider":"AI_PROVIDER_ANTHROPIC"}'

curl -s -b "jwt=$JWT" https://local.api.vulx.ai/api/v1/projects | jq
```

Note enums are spelled as **full proto names** over REST — the OpenAPI plugin runs with
`enum_type=string`.

Check `/docs/` renders the three new paths.

## Done when

- All three RPCs return 200 with a cookie.
- `preview_url` and `sandbox_id` come back as `""`, and `created_at` looks like
  `2026-09-08T14:03:11Z` — not `2026-09-08 14:03:11.123 +0000 UTC`.
- Task 6's negative checks pass.
