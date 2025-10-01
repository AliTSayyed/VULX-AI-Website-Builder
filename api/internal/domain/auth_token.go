package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

/*
This is the auth token a user has to maintian their session (JWT).
Ready for other token types but unlikely for now
*/

var (
	ErrAuthTokenNameEmpty        = NewError(ErrorTypeInvalid, errors.New("auth token name cannot be empty"))
	ErrAuthTokenExpiresAtInvalid = NewError(ErrorTypeInvalid, errors.New("auth token expires at is invalid"))
	ErrAuthTokenTypeUnspecified  = NewError(ErrorTypeInvalid, errors.New("auth token type cannot be unspecified"))
)

type TokenType int

const (
	TokenTypeUnspecified TokenType = iota
	TokenTypeSession
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeSession:
		return "session"
	case TokenTypeUnspecified:
		fallthrough
	default:
		return "unspecified"
	}
}

func ParseTokenType(s string) TokenType {
	switch strings.ToLower(s) {
	case "session":
		return TokenTypeSession
	default:
		return TokenTypeUnspecified
	}
}

type AuthToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	name      string
	tokenType TokenType
	createdAt time.Time
	expiresAt time.Time
}

func NewAuthToken(userID uuid.UUID, name string, tokenType TokenType, expirestAt time.Time) (*AuthToken, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrAuthTokenNameEmpty
	}

	if tokenType == TokenTypeUnspecified {
		return nil, ErrAuthTokenTypeUnspecified
	}

	now := time.Now()
	if expirestAt.Before(now) || expirestAt.Equal(now) {
		return nil, ErrAuthTokenExpiresAtInvalid
	}

	return &AuthToken{
		id:        uuid.New(),
		userID:    userID,
		name:      name,
		tokenType: tokenType,
		createdAt: time.Now(),
		expiresAt: expirestAt,
	}, nil
}

func RestoreAuthToken(id uuid.UUID, userID uuid.UUID, name string, tokenType TokenType, createdAt, expirestAt time.Time) *AuthToken {
	return &AuthToken{
		id:        id,
		userID:    userID,
		name:      name,
		tokenType: tokenType,
		createdAt: createdAt,
		expiresAt: expirestAt,
	}
}

func (a *AuthToken) ID() uuid.UUID {
	return a.id
}

func (a *AuthToken) UserID() uuid.UUID {
	return a.userID
}

func (a *AuthToken) Name() string {
	return a.name
}

func (a *AuthToken) TokenType() TokenType {
	return a.tokenType
}

func (a *AuthToken) CreatedAt() time.Time {
	return a.createdAt
}

func (a *AuthToken) ExpiresAt() time.Time {
	return a.expiresAt
}

func (a *AuthToken) IsExpired() bool {
	return time.Now().After(a.expiresAt)
}
