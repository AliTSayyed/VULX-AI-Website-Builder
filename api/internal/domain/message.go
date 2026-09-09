package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMessageBodyEmpty       = NewError(ErrorTypeInvalid, errors.New("message body cannot be empty"))
	ErrMessageProjectEmpty    = NewError(ErrorTypeInvalid, errors.New("message project id cannot be empty"))
	ErrMessageRoleUnspecified = NewError(ErrorTypeInvalid, errors.New("message role cannot be unspecified"))
	ErrMessageModeUnspecified = NewError(ErrorTypeInvalid, errors.New("message mode cannot be unspecified"))
)

type MessageRole int

const (
	MessageRoleUnspecified MessageRole = iota
	MessageRoleUser
	MessageRoleAssistant
)

func (r MessageRole) String() string {
	switch r {
	case MessageRoleUser:
		return "user"
	case MessageRoleAssistant:
		return "assistant"
	default:
		return "unspecified"
	}
}

func ParseMessageRole(s string) MessageRole {
	switch strings.ToLower(s) {
	case "user":
		return MessageRoleUser
	case "assistant":
		return MessageRoleAssistant
	default:
		return MessageRoleUnspecified
	}
}

type ChatMode int

const (
	ChatModeUnspecified ChatMode = iota
	ChatModeChat
	ChatModeBuild
)

func (m ChatMode) String() string {
	switch m {
	case ChatModeChat:
		return "chat"
	case ChatModeBuild:
		return "build"
	default:
		return "unspecified"
	}
}

func ParseChatMode(s string) ChatMode {
	switch strings.ToLower(s) {
	case "chat":
		return ChatModeChat
	case "build":
		return ChatModeBuild
	default:
		return ChatModeUnspecified
	}
}

type Message struct {
	id        uuid.UUID
	projectID uuid.UUID
	role      MessageRole
	mode      ChatMode
	provider  AIProvider
	body      string
	createdAt time.Time
	updatedAt time.Time
}

func NewMessage(projectID uuid.UUID, role MessageRole, mode ChatMode, provider AIProvider, body string) (*Message, error) {
	if projectID == uuid.Nil {
		return nil, ErrMessageProjectEmpty
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, ErrMessageBodyEmpty
	}

	if role == MessageRoleUnspecified {
		return nil, ErrMessageRoleUnspecified
	}

	if mode == ChatModeUnspecified {
		return nil, ErrMessageModeUnspecified
	}

	return &Message{
		id:        uuid.New(),
		projectID: projectID,
		role:      role,
		mode:      mode,
		provider:  provider,
		body:      body,
	}, nil
}

func RestoreMessage(id, projectID uuid.UUID, role, mode, provider, body string, createdAt, updatedAt time.Time) *Message {
	return &Message{
		id:        id,
		projectID: projectID,
		role:      ParseMessageRole(role),
		mode:      ParseChatMode(mode),
		provider:  ParseAIProvider(provider),
		body:      body,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (m *Message) ID() uuid.UUID {
	if m == nil {
		return uuid.Nil
	}
	return m.id
}

func (m *Message) ProjectID() uuid.UUID {
	if m == nil {
		return uuid.Nil
	}
	return m.projectID
}

func (m *Message) Role() MessageRole {
	if m == nil {
		return MessageRoleUnspecified
	}
	return m.role
}

func (m *Message) Mode() ChatMode {
	if m == nil {
		return ChatModeUnspecified
	}
	return m.mode
}

func (m *Message) Provider() AIProvider {
	if m == nil {
		return AIProviderUnspecified
	}
	return m.provider
}

func (m *Message) Body() string {
	if m == nil {
		return ""
	}
	return m.body
}

func (m *Message) CreatedAt() time.Time {
	if m == nil {
		return time.Time{}
	}
	return m.createdAt
}

func (m *Message) UpdatedAt() time.Time {
	if m == nil {
		return time.Time{}
	}
	return m.updatedAt
}
