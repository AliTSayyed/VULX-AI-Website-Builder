package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) (*domain.Project, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	FindAllByUser(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error)
	UpdateSandbox(ctx context.Context, id uuid.UUID, sandboxID, previewURL string) error
	UpdateTitle(ctx context.Context, id uuid.UUID, title string) error
}

// ProjectTitler generates a short project title from the first prompt. Create falls back to
// provisionalTitle on any error or empty result — title generation is a cosmetic nicety, never
// load-bearing for whether a project can be created.
type ProjectTitler interface {
	GenerateTitle(ctx context.Context, firstPrompt string) (string, error)
}

type ProjectService struct {
	projectRepo ProjectRepository
	titler      ProjectTitler
}

func NewProjectService(projectRepo ProjectRepository, titler ProjectTitler) *ProjectService {
	return &ProjectService{projectRepo: projectRepo, titler: titler}
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

	go s.generateTitleAsync(created.ID(), firstPrompt)

	return created, nil
}

const asyncTitleTimeout = 15 * time.Second

// generateTitleAsync runs after Create has already returned the provisional title to the
// caller — the request's context is gone by the time this finishes, so it owns a fresh
// background one. Any failure just leaves the provisional title in place; there is no retry.
// recover() is required here: this goroutine runs outside any request lifecycle, so an
// unrecovered panic would crash the whole process instead of one HTTP request.
func (s *ProjectService) generateTitleAsync(id uuid.UUID, firstPrompt string) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Error("panic in async project title generation",
				"project_id", id, "panic", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), asyncTitleTimeout)
	defer cancel()

	generated, err := s.titler.GenerateTitle(ctx, firstPrompt)
	if err != nil {
		utils.Logger.Warn("project title generation failed, keeping provisional title",
			"project_id", id, "error", err)
		return
	}

	title := provisionalTitle(generated)
	if title == "" {
		return
	}

	if err := s.projectRepo.UpdateTitle(ctx, id, title); err != nil {
		utils.Logger.Warn("failed to persist generated project title",
			"project_id", id, "error", err)
	}
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
