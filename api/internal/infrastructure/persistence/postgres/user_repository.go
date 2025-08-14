package postgres

import (
	"context"
	"fmt"
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

func (u *User) ToDomain() *domain.User {
	return domain.RestoreUser(u.ID, u.Name)
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := ` 
		INSERT INTO users (id, name, created_at, updated_at)
		VALUES($1, $2, NOW(), NOW())
		RETURNING id, name, created_at, updated_at
	`
	var dbUser User
	err := u.db.QueryRowxContext(ctx, query, user.ID(), user.Name()).StructScan(&dbUser)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to create user %s, %w", user.Name(), err))
	}
	return dbUser.ToDomain(), nil
}

func (u *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var dbUser User
	err := u.db.QueryRowxContext(ctx, query, id).StructScan(&dbUser)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to find user %s, %w", id, err))
	}
	return dbUser.ToDomain(), nil
}
