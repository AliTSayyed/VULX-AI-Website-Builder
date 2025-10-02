package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
)

/*
this service creates auth tokens for users to manage sessions.
a user will have an access tokent that will last 7 days
if token expires in < 42 hours, automaitcally create new accesstoken and override old one
if user logs out, store access token in the black list (db)
if user does not interact with app for > 7 days, access token expires and user must login again
*/

var (
	ErrAuthTokenNotFound    = domain.NewError(domain.ErrorTypeNotFound, errors.New("auth token not found"))
	ErrAuthTokenExpiresSoon = domain.NewError(domain.ErrorTypeInvalid, errors.New("auth token expires soon"))
)

type AuthAdapter interface {
	CreateJWT(uuid.UUID) (string, error)
	ValidateJWT(string) (uuid.UUID, error)
	ParseJWT(string) (uuid.UUID, time.Time, error)
}

// only using access tokens, just black list tokens in cahce if user logs out
type AuthService struct {
	cache       Cache
	authAdapter AuthAdapter
	userRepo    UserRepository
}

func NewAuthService(cache Cache, authAdapter AuthAdapter, userRepo UserRepository) *AuthService {
	return &AuthService{
		cache:       cache,
		authAdapter: authAdapter,
		userRepo:    userRepo,
	}
}

// used when a user logs in
func (a *AuthService) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	jwt, err := a.authAdapter.CreateJWT(userID)
	if err != nil {
		return "", domain.WrapError("auth service create session", err)
	}

	return jwt, nil
}

// used on every request interception
func (a *AuthService) ValidateSession(ctx context.Context, token string) (*domain.User, error) {
	// check black list (if user logs out and wants to immediately log in, must create a new session)
	hashedToken := hash(token)
	exists, err := a.cache.Exists(ctx, hashedToken)
	if err != nil {
		return nil, domain.WrapError("auth service validate session", err)
	}
	if exists {
		return nil, domain.ErrUnauthenticated
	}

	var refresh bool
	userID, err := a.authAdapter.ValidateJWT(token)
	if errors.Is(err, ErrAuthTokenExpiresSoon) {
		refresh = true
		err = nil
	} else if err != nil {
		return nil, domain.WrapError("auth service validate session", err)
	}

	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.WrapError("auth service validate session", err)
	}

	// this will auto create a new jwt for the user
	if refresh {
		return user, ErrAuthTokenExpiresSoon
	}

	return user, nil
}

// used when a user logs out, blacklist token so only one valid token per user is active at any time
func (a *AuthService) Logout(ctx context.Context, token string) error {
	_, expiresAt, err := a.authAdapter.ParseJWT(token)
	if err != nil {
		return domain.WrapError("auth service logout", err)
	}

	// calculate how long to store in cache
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	// black list the hashed token into cache
	hashedToken := hash(token)
	if err = a.cache.Set(ctx, hashedToken, "1", ttl); err != nil {
		return domain.WrapError("auth service logout", err)
	}
	return nil
}

func hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
