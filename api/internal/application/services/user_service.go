/*
* only
 */
package services

import (
	"context"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
)

// this interface is the "port" to interact with the postgres db
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.WrapError("user service get", err)
	}
	return user, nil
}

func (s *UserService) Add(ctx context.Context, name string) (*domain.User, error) {
	user, err := domain.NewUser(name)
	if err != nil {
		return nil, domain.WrapError("user service add", err)
	}

	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, domain.WrapError("user service add", err)
	}

	return user, nil
}
