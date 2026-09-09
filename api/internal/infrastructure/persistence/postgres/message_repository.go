package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

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

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	query := `
		WITH inserted AS (
			INSERT INTO messages (id, project_id, role, mode, provider, body, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			RETURNING id, project_id, role, mode, provider, body, created_at, updated_at
		), bumped AS (
			UPDATE projects SET updated_at = NOW() WHERE id = $2
		)
		SELECT id, project_id, role, mode, provider, body, created_at, updated_at FROM inserted
	`

	var dbMessage Message
	err := r.db.QueryRowxContext(ctx, query,
		message.ID(), message.ProjectID(), message.Role().String(), message.Mode().String(),
		message.Provider().String(), message.Body(),
	).StructScan(&dbMessage)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return nil, domain.NewError(domain.ErrorTypeNotFound,
				fmt.Errorf("project %s not found: %w", message.ProjectID(), err))
		}
		return nil, domain.NewError(domain.ErrorTypeInternal,
			fmt.Errorf("failed to create message for project %s, %w", message.ProjectID(), err))
	}

	return dbMessage.ToDomain(), nil
}

func (r *MessageRepository) FindByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Message, error) {
	query := `
		SELECT id, project_id, role, mode, provider, body, created_at, updated_at
		FROM messages
		WHERE project_id = $1
		ORDER BY created_at ASC, id ASC
	`

	var dbMessages []*Message
	if err := r.db.SelectContext(ctx, &dbMessages, query, projectID); err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal,
			fmt.Errorf("failed to find messages for project %s, %w", projectID, err))
	}

	messages := make([]*domain.Message, 0, len(dbMessages))
	for _, dbMessage := range dbMessages {
		messages = append(messages, dbMessage.ToDomain())
	}

	return messages, nil
}
