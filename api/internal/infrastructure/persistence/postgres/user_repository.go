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

type User struct {
	ID        uuid.UUID `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Credits   int       `db:"credits"`
	Is_active bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserFromProvider struct {
	UserID       uuid.UUID
	ProviderName string
	ProviderID   string
}

func (u *User) ToDomain() *domain.User {
	return domain.RestoreUser(u.ID, u.FirstName, u.LastName, u.Email, u.Credits, u.Is_active)
}

func (u *UserFromProvider) ToDomain() *domain.UserFromProvider {
	return domain.RestoreUserFromProvider(u.UserID, u.ProviderName, u.ProviderID)
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
		INSERT INTO users (id, first_name, last_name, email, created_at, updated_at)
		VALUES($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, first_name, last_name, email, created_at, updated_at
	`
	var dbUser User
	err := u.db.QueryRowxContext(ctx, query, user.ID(), user.FirstName(), user.LastName(), user.Email()).StructScan(&dbUser)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to create user %s %s (%s), %w", user.FirstName(), user.LastName(), user.Email(), err))
	}
	return dbUser.ToDomain(), nil
}

func (u *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, email, credits, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var dbUser User
	err := u.db.QueryRowxContext(ctx, query, id).StructScan(&dbUser)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to find user id %s, %w", id, err))
	}
	return dbUser.ToDomain(), nil
}

func (u *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, email, credits, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var dbUser User
	err := u.db.QueryRowxContext(ctx, query, email).StructScan(&dbUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewError(domain.ErrorTypeNotFound, fmt.Errorf("failed to find user email %s, %w", email, err))
		}
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to find user email %s, %w", email, err))
	}
	return dbUser.ToDomain(), nil
}

func (u *UserRepository) FindProvider(ctx context.Context, userID uuid.UUID) (*domain.UserFromProvider, error) {
	query := ` 
		SELECT user_id, provider, provider_user_id
		FROM user_auth_providers
		WHERE user_id = $1
	`

	var dbUserFromProvider UserFromProvider
	err := u.db.QueryRowxContext(ctx, query, userID).StructScan(&dbUserFromProvider)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to find provider of user id %s, %w", userID, err))
	}

	return dbUserFromProvider.ToDomain(), nil
}

func (u *UserRepository) CreateProvider(ctx context.Context, userID uuid.UUID, providerName string, providerID string) (*domain.UserFromProvider, error) {
	query := `
		INSERT INTO user_auth_providers (user_id, provider, provider_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING user_id, provider, provider_user_id
	`
	var dbUserFromProvider UserFromProvider
	err := u.db.QueryRowxContext(ctx, query, userID, providerName, providerID).StructScan(&dbUserFromProvider)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal,
			fmt.Errorf("failed to insert provider information for user id %s, provider %s, and provider user id %s, error: %w",
				userID, providerName, providerID, err))
	}
	return dbUserFromProvider.ToDomain(), nil
}
