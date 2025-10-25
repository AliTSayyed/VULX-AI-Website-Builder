package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
)

/*
account service will combine the oauth flow, jwt, and creating / logging back in a user.
If there is an existing email in the db, do not create a new user with the same email, give an error with which provider has used that email
This service is what is directly used by the account handler.
JWT will handle user login, not storing oauth access tokens
*/

type AuthResult struct {
	Token   string
	Profile *domain.Profile
}

type AccountService struct {
	oauthService *OauthService
	authService  *AuthService
	userService  *UserService
}

func NewAccountService(oauthService *OauthService, authService *AuthService, userService *UserService) *AccountService {
	return &AccountService{
		oauthService: oauthService,
		authService:  authService,
		userService:  userService,
	}
}

func (a *AccountService) BeginAuth(ctx context.Context, provider domain.LoginProvider) (string, error) {
	if provider == domain.LoginProviderUnspecified {
		return "", domain.ErrLoginProviderUnspecified
	}

	// TODO better place holder for options param other than nil?
	url, err := a.oauthService.BeginLoginFlow(ctx, provider, nil)
	if err != nil {
		return "", domain.WrapError("account service begin auth", err)
	}

	return url, nil
}

// create or get user and give them a jwt
func (a *AccountService) FinishAuth(ctx context.Context, code string, state string) (*AuthResult, error) {
	loginResult, err := a.oauthService.CompleteLoginFlow(ctx, code, state, nil)
	if err != nil {
		return nil, domain.WrapError("account service finish auth", err)
	}

	// get user from email, if email does not exist then create a new user
	var user *domain.User
	user, err = a.userService.GetByEmail(ctx, loginResult.Email)

	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) && domainErr.Type() == domain.ErrorTypeNotFound {
			// if email does not exist, then create the new user and then create their provider
			user, err = a.userService.Add(ctx, loginResult.FirstName, loginResult.LastName, loginResult.Email)
			if err != nil {
				return nil, domain.WrapError("account service finish auth", nil)
			}
			// now create the login provider
			_, err = a.userService.AddProvider(ctx, user.ID(), loginResult.Credentials.ProviderName, loginResult.Credentials.ProviderUserID)
			if err != nil {
				return nil, domain.WrapError("account service finish auth", err)
			}
		} else {
			return nil, domain.WrapError("account service finish auth", err)
		}
	} else {
		// if the email exists, cehck that the loginResult.credentials.ProviderName == Provider for user id
		// if not then throw error this email already exists
		userWithProvider, err := a.userService.GetProvider(ctx, user.ID())
		if err != nil {
			return nil, domain.WrapError("account service finish auth", err)
		}
		if loginResult.Credentials.ProviderName != userWithProvider.Provider() {
			return nil, domain.NewError(domain.ErrorTypeAlreadyExists, fmt.Errorf("this email is already associated with the following provider: %s", userWithProvider.Provider()))
		}
	}

	// create a session for them using the auth service jwt
	token, err := a.authService.CreateSession(ctx, user.ID())
	if err != nil {
		return nil, domain.WrapError("account service finish auth", err)
	}

	profile, err := domain.NewProfile(user.FirstName(), user.LastName(), user.Email(), user.Credits())
	if err != nil {
		return nil, domain.WrapError("account service finish auth", err)
	}

	return &AuthResult{
		Token:   token,
		Profile: profile,
	}, nil
}

func (a *AccountService) Logout(ctx context.Context, token string) error {
	err := a.authService.Logout(ctx, token)
	if err != nil {
		return domain.WrapError("account service logout", err)
	}

	return nil
}

func (a *AccountService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	user, err := a.userService.Get(ctx, userID)
	if err != nil {
		return nil, domain.WrapError("account service get profile", err)
	}

	profile, err := domain.NewProfile(user.FirstName(), user.LastName(), user.Email(), user.Credits())
	if err != nil {
		return nil, domain.WrapError("account service finish auth", err)
	}

	return profile, nil
}
