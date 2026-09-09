package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Project struct {
	ID         uuid.UUID `db:"id"`
	UserID     uuid.UUID `db:"user_id"`
	Title      string    `db:"title"`
	Provider   string    `db:"provider"`
	SandboxID  string    `db:"sandbox_id"`
	PreviewURL string    `db:"preview_url"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (p *Project) ToDomain() *domain.Project {
	return domain.RestoreProject(
		p.ID, p.UserID, p.Title, p.Provider,
		p.SandboxID, p.PreviewURL,
		p.CreatedAt, p.UpdatedAt,
	)
}

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (p *ProjectRepository) Create(ctx context.Context, project *domain.Project) (*domain.Project, error) {
	query := `
		INSERT INTO projects (id, user_id, title, provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, user_id, title, provider, sandbox_id, preview_url, created_at, updated_at
	`
	var dbProject Project
	err := p.db.QueryRowxContext(ctx, query, project.ID(), project.UserID(), project.Title(), project.Provider().String()).StructScan(&dbProject)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to create project %s for user %s, %w", project.Title(), project.UserID(), err))
	}
	return dbProject.ToDomain(), nil
}

func (p *ProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, user_id, title, provider, sandbox_id, preview_url, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	var dbProject Project
	err := p.db.QueryRowxContext(ctx, query, id).StructScan(&dbProject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewError(domain.ErrorTypeNotFound, fmt.Errorf("project %s not found: %w", id, err))
		}
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to find project %s, %w", id, err))
	}
	return dbProject.ToDomain(), nil
}

func (p *ProjectRepository) FindAllByUser(ctx context.Context, userID uuid.UUID, limit int64, token string) (*domain.Page[*domain.Project], error) {
	ts, err := decodeToken(token)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to decode token: %w", err))
	}

	query := `
		SELECT id, user_id, title, provider, sandbox_id, preview_url, created_at, updated_at
		FROM projects
		WHERE user_id = $1
		  AND ($2::timestamptz IS NULL OR updated_at < $2)
		ORDER BY updated_at DESC, id DESC
		LIMIT NULLIF($3, 0)
	`

	var cursor *time.Time
	if !ts.IsZero() {
		cursor = &ts
	}

	var dbProjects []*Project
	err = p.db.SelectContext(ctx, &dbProjects, query, userID, cursor, limit+1)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to find projects for user %s, %w", userID, err))
	}

	var nextToken string
	if len(dbProjects) > int(limit) {
		dbProjects = dbProjects[:limit]
		nextToken = encodeToken(dbProjects[len(dbProjects)-1].UpdatedAt)
	}

	projects := make([]*domain.Project, 0, len(dbProjects))
	for _, dbProject := range dbProjects {
		projects = append(projects, dbProject.ToDomain())
	}

	return &domain.Page[*domain.Project]{
		Items:   projects,
		Token:   nextToken,
		HasMore: nextToken != "",
	}, nil
}

func (p *ProjectRepository) UpdateSandbox(ctx context.Context, id uuid.UUID, sandboxID, previewURL string) error {
	query := `UPDATE projects SET sandbox_id = $2, preview_url = $3 WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, id, sandboxID, previewURL)
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to update sandbox for project %s, %w", id, err))
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to check rows affected updating sandbox for project %s, %w", id, err))
	}
	if rows == 0 {
		return domain.NewError(domain.ErrorTypeNotFound, fmt.Errorf("project %s not found", id))
	}

	return nil
}
