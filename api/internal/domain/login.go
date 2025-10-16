package domain

/*
This is the oauth providers for the applicaitons
*/
import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrLoginProviderUnspecified = NewError(ErrorTypeInvalid, errors.New("login provider cannot be unspecified"))

type LoginProvider int

const (
	LoginProviderUnspecified LoginProvider = iota
	LoginProviderGoogle
)

type UserFromProvider struct {
	userID     uuid.UUID
	provider   LoginProvider
	providerID string
}

func RestoreUserFromProvider(userID uuid.UUID, providerName string, providerID string) *UserFromProvider {
	provider := ParseLoginProvider(providerName)
	return &UserFromProvider{
		userID:     userID,
		provider:   provider,
		providerID: providerID,
	}
}

func (u *UserFromProvider) Provider() string {
	return u.provider.String()
}

func (l LoginProvider) String() string {
	switch l {
	case LoginProviderGoogle:
		return "google"
	case LoginProviderUnspecified:
		fallthrough
	default:
		return "unspecified"
	}
}

func ParseLoginProvider(s string) LoginProvider {
	switch strings.ToLower(s) {
	case "google":
		return LoginProviderGoogle
	default:
		return LoginProviderUnspecified
	}
}
