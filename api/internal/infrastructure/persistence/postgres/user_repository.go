package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (*UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	return user, nil
}

func (*UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := domain.NewUser("bob")
	if err != nil {
		return nil, errors.New("failed to find user")
	}
	return user, nil
}
