package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrSandboxUrlEmpty = NewError(ErrorTypeInvalid, errors.New("sandbox url cannot be empty"))

type Sandbox struct {
	id  uuid.UUID
	url string
}

func NewSandbox(url string) (*Sandbox, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, ErrUserNameEmpty
	}

	return &Sandbox{
		id:  uuid.New(),
		url: url,
	}, nil
}

func RestoreSandbox(id uuid.UUID, url string) *Sandbox {
	return &Sandbox{
		id:  id,
		url: url,
	}
}

func (s *Sandbox) ID() uuid.UUID {
	if s == nil {
		return uuid.Nil
	}
	return s.id
}

func (s *Sandbox) Url() string {
	if s == nil {
		return ""
	}
	return s.url
}
