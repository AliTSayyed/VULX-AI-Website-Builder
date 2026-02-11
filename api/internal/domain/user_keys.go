package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrAPIKeyEmpty = NewError(ErrorTypeInvalid, errors.New("api key can not be empty"))

type LlmProvider int

const (
	LlmProviderUnspecified LlmProvider = iota
	LlmProviderGemini
	LlmProviderOpenAI
	LlmProviderClaude
)

type UserApiKey struct {
	id          uuid.UUID
	userID      uuid.UUID
	provider    LlmProvider
	providerKey string
}

func RestoreKeyFromLlmProvider(id uuid.UUID, userID uuid.UUID, providerName string, key string) *UserApiKey {
	provider := ParseLlmProvider(providerName)
	return &UserApiKey{
		id:          id,
		userID:      userID,
		provider:    provider,
		providerKey: key,
	}
}

func (u *UserApiKey) ID() uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return u.id
}

func (u *UserApiKey) UserID() uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return u.userID
}

func (u *UserApiKey) Provider() string {
	if u == nil {
		return ""
	}
	return u.provider.String()
}

func (u *UserApiKey) Key() string {
	if u == nil {
		return ""
	}
	return u.providerKey
}

func (l LlmProvider) String() string {
	switch l {
	case LlmProviderGemini:
		return "gemini"
	case LlmProviderOpenAI:
		return "openai"
	case LlmProviderClaude:
		return "claude"
	case LlmProviderUnspecified:
		fallthrough
	default:
		return "unspecified"
	}
}

func ParseLlmProvider(s string) LlmProvider {
	switch strings.ToLower(s) {
	case "gemini":
		return LlmProviderGemini
	case "openai":
		return LlmProviderOpenAI
	case "claude":
		return LlmProviderClaude
	default:
		return LlmProviderUnspecified
	}
}
