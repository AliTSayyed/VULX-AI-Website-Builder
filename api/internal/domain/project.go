package domain

/*
* A project is the durable thing a user builds: an owned, named record with a codebase, a chat
* thread, and a (usually dead) sandbox pointer. Postgres is the source of truth; the sandbox is a
* disposable materialization of it — sandboxID and previewURL are a cache pointer that is usually
* stale, not state.
 */

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProjectTitleEmpty          = NewError(ErrorTypeInvalid, errors.New("project title cannot be empty"))
	ErrProjectUserEmpty           = NewError(ErrorTypeInvalid, errors.New("project user id cannot be empty"))
	ErrProjectProviderUnspecified = NewError(ErrorTypeInvalid, errors.New("project provider cannot be unspecified"))
)

type Project struct {
	id         uuid.UUID
	userID     uuid.UUID
	title      string
	provider   AIProvider
	sandboxID  string // "" until a sandbox has been created
	previewURL string // "" until a sandbox has been created
	createdAt  time.Time
	updatedAt  time.Time
}

func NewProject(userID uuid.UUID, title string, provider AIProvider) (*Project, error) {
	if userID == uuid.Nil {
		return nil, ErrProjectUserEmpty
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrProjectTitleEmpty
	}

	if provider == AIProviderUnspecified {
		return nil, ErrProjectProviderUnspecified
	}

	return &Project{
		id:       uuid.New(),
		userID:   userID,
		title:    title,
		provider: provider,
	}, nil
}

func RestoreProject(id, userID uuid.UUID, title, provider, sandboxID, previewURL string, createdAt, updatedAt time.Time) *Project {
	return &Project{
		id:         id,
		userID:     userID,
		title:      title,
		provider:   ParseAIProvider(provider),
		sandboxID:  sandboxID,
		previewURL: previewURL,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (p *Project) ID() uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return p.id
}

func (p *Project) UserID() uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return p.userID
}

func (p *Project) Title() string {
	if p == nil {
		return ""
	}
	return p.title
}

func (p *Project) Provider() AIProvider {
	if p == nil {
		return AIProviderUnspecified
	}
	return p.provider
}

func (p *Project) SandboxID() string {
	if p == nil {
		return ""
	}
	return p.sandboxID
}

func (p *Project) PreviewURL() string {
	if p == nil {
		return ""
	}
	return p.previewURL
}

func (p *Project) CreatedAt() time.Time {
	if p == nil {
		return time.Time{}
	}
	return p.createdAt
}

func (p *Project) UpdatedAt() time.Time {
	if p == nil {
		return time.Time{}
	}
	return p.updatedAt
}

func (p *Project) HasSandbox() bool {
	return p != nil && p.sandboxID != ""
}
