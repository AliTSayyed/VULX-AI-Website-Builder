package domain

import (
	"errors"
	"strings"
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
