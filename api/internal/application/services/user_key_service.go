package services

import (
	"context"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
)

type UserApiKeyRepository interface {
	CreateOrUpdate(ctx context.Context, userApiKey *domain.UserApiKey) (*domain.UserApiKey, error)
	FindByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*domain.UserApiKey, error)
	FindAllWithUserID(ctx context.Context, userID uuid.UUID) ([]*domain.UserApiKey, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserKeyService struct {
	userKeysRepo UserApiKeyRepository
}

func NewUserKeyService(u UserApiKeyRepository) *UserKeyService {
	return &UserKeyService{
		userKeysRepo: u,
	}
}
