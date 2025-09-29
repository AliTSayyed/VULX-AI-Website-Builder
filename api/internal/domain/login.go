package domain

/*
This is the oauth providers for the applicaitons
*/
import (
	"errors"
	"strings"
)

var ErrLoginProviderUnspecified = NewError(ErrorTypeInvalid, errors.New("login provider cannot be unspecified"))

type ProviderLoginCredentials struct {
	provider   string
	providerID string
}

type LoginProvider int

const (
	LoginProviderUnspecified LoginProvider = iota
	LoginProviderGoogle
)

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
