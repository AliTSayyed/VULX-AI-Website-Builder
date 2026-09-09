package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) (*domain.Project, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	FindAllByUser(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error)
	UpdateSandbox(ctx context.Context, id uuid.UUID, sandboxID, previewURL string) error
}

type ProjectService struct {
	projectRepo ProjectRepository
}

func NewProjectService(projectRepo ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) List(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error) {
	limit = utils.Clamp(limit, 10, 100)

	page, err := s.projectRepo.FindAllByUser(ctx, userID, limit, token)
	if err != nil {
		return nil, domain.WrapError("project service list", err)
	}

	return page, nil
}

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

func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, firstPrompt string, provider domain.AIProvider) (*domain.Project, error) {
	title := provisionalTitle(firstPrompt)

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

// provisionalTitle collapses whitespace and truncates to ~60 characters on a word boundary.
// An empty result is deliberate — domain.NewProject rejects it with ErrProjectTitleEmpty, which
// is the entire mechanism that stops an empty prompt from creating an empty project.
func provisionalTitle(firstPrompt string) string {
	collapsed := strings.Join(strings.Fields(firstPrompt), " ")
	if collapsed == "" {
		return ""
	}

	const maxLen = 60
	runes := []rune(collapsed)
	if len(runes) <= maxLen {
		return collapsed
	}

	truncated := string(runes[:maxLen])
	if idx := strings.LastIndex(truncated, " "); idx > 0 {
		truncated = truncated[:idx]
	}

	return truncated
}
