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
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindProvider(ctx context.Context, userID uuid.UUID) (*domain.UserFromProvider, error)
	CreateProvider(ctx context.Context, userID uuid.UUID, provider domain.LoginProvider, providerID string) (*domain.UserFromProvider, error)
}

type UserWorkflowService interface {
	StartUserWorkflow(ctx context.Context) error
}

type UserService struct {
	userRepo     UserRepository
	userWorkflow UserWorkflowService
}

func NewUserService(userRepo UserRepository, userWorkflow UserWorkflowService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		userWorkflow: userWorkflow,
	}
}

func (u *UserService) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.WrapError("user service get", err)
	}
	return user, nil
}

func (u *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.WrapError("user service get by email", err)
	}
	return user, nil
}

func (u *UserService) Add(ctx context.Context, firstName string, lastName string, email string) (*domain.User, error) {
	user, err := domain.NewUser(firstName, lastName, email)
	if err != nil {
		return nil, domain.WrapError("user service add", err)
	}

	user, err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, domain.WrapError("user service add", err)
	}

	// TODO get rid of this example call to workflow orchestration
	// err = u.userWorkflow.StartUserWorkflow(ctx)
	// if err != nil {
	// 	return nil, domain.WrapError("user service add", err)
	// }

	return user, nil
}

func (u *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO update user to have is active false
	return nil
}

func (u *UserService) GetProvider(ctx context.Context, userID uuid.UUID) (*domain.UserFromProvider, error) {
	return nil, nil
}

func (u *UserService) AddProvider(ctx context.Context, userID uuid.UUID, provider domain.LoginProvider, providerID string) (*domain.UserFromProvider, error) {
	// TODO, after a user is created, then insert their provider into the db

	return nil, nil
}
