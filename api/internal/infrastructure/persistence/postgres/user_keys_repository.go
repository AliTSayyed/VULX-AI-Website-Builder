package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserApiKey struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	ProviderName string    `db:"provider"`
	EncryptedKey string    `db:"encrypted_key"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (u *UserApiKey) ToDomain() *domain.UserApiKey {
	return domain.RestoreKeyFromLlmProvider(u.ID, u.UserID, u.ProviderName, u.EncryptedKey)
}

type UserApiKeyRepository struct {
	db *sqlx.DB
}

func NewUserApiKeyRepository(db *sqlx.DB) *UserApiKeyRepository {
	return &UserApiKeyRepository{
		db: db,
	}
}

func (u *UserApiKeyRepository) CreateOrUpdate(ctx context.Context, userApiKey *domain.UserApiKey) (*domain.UserApiKey, error) {
	query := `
		INSERT INTO user_api_keys (user_id, provider, encrypted_key, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (provider, user_id) 
		DO UPDATE SET 
			encrypted_key = EXCLUDED.encrypted_key,
			updated_at = NOW()
		RETURNING id, user_id, provider, encrypted_key, created_at, updated_at
	`
	var dbUserApiKey UserApiKey
	err := u.db.QueryRowxContext(ctx, query, userApiKey.UserID(), userApiKey.Provider(), userApiKey.Key()).StructScan(&dbUserApiKey)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInternal,
			fmt.Errorf("failed to upsert user id %s, provider %s, encrypted key %s: %w",
				userApiKey.UserID(), userApiKey.Provider(), userApiKey.Key(), err))
	}
	return dbUserApiKey.ToDomain(), nil
}

func (u *UserApiKeyRepository) FindByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*domain.UserApiKey, error) {
	return nil, nil
}

func (u *UserApiKeyRepository) FindAllWithUserID(ctx context.Context, userID uuid.UUID) ([]*domain.UserApiKey, error) {
	return nil, nil
}

func (u *UserApiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
