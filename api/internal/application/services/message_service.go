package services

import (
	"context"
	"errors"
	"strings"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) (*domain.Message, error)
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Message, error)
}

type SandboxCreator interface {
	CreateSandbox(ctx context.Context) (*domain.Sandbox, error)
}

type CodeAgentRunner interface {
	RunCodeAgent(ctx context.Context, provider domain.AIProvider, sandboxID, message string) (*domain.CodeAgentResult, error)
}

const maxMessageBodyRunes = 10000

type MessageService struct {
	messageRepo    MessageRepository
	projectRepo    ProjectRepository
	projectService *ProjectService
	sandbox        SandboxCreator
	codeAgent      CodeAgentRunner
}

func NewMessageService(
	messageRepo MessageRepository,
	projectRepo ProjectRepository,
	projectService *ProjectService,
	sandbox SandboxCreator,
	codeAgent CodeAgentRunner,
) *MessageService {
	return &MessageService{
		messageRepo:    messageRepo,
		projectRepo:    projectRepo,
		projectService: projectService,
		sandbox:        sandbox,
		codeAgent:      codeAgent,
	}
}

func (s *MessageService) List(ctx context.Context, userID, projectID uuid.UUID) ([]*domain.Message, error) {
	if _, err := s.projectService.Get(ctx, userID, projectID); err != nil {
		return nil, domain.WrapError("message service list", err)
	}

	messages, err := s.messageRepo.FindByProject(ctx, projectID)
	if err != nil {
		return nil, domain.WrapError("message service list", err)
	}

	return messages, nil
}

func (s *MessageService) Send(
	ctx context.Context, userID, projectID uuid.UUID,
	body string, mode domain.ChatMode, provider domain.AIProvider,
) (*domain.Message, *domain.Message, error) {
	// 1. ownership
	project, err := s.projectService.Get(ctx, userID, projectID)
	if err != nil {
		return nil, nil, domain.WrapError("message service send", err)
	}

	// 2. validation
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, nil, domain.NewError(domain.ErrorTypeInvalid, errors.New("message body cannot be empty"))
	}
	if len([]rune(body)) > maxMessageBodyRunes {
		return nil, nil, domain.NewError(domain.ErrorTypeInvalid, errors.New("message body too long"))
	}
	if mode == domain.ChatModeUnspecified {
		return nil, nil, domain.NewError(domain.ErrorTypeInvalid, errors.New("chat mode cannot be unspecified"))
	}
	if provider == domain.AIProviderUnspecified {
		return nil, nil, domain.NewError(domain.ErrorTypeInvalid, errors.New("ai provider cannot be unspecified"))
	}

	// 3. persist the user message — the CTE bumps projects.updated_at
	userMsg, err := domain.NewMessage(projectID, domain.MessageRoleUser, mode, provider, body)
	if err != nil {
		return nil, nil, domain.WrapError("message service send", err)
	}
	userMsg, err = s.messageRepo.Create(ctx, userMsg)
	if err != nil {
		return nil, nil, domain.WrapError("message service send", err)
	}

	// 4. Chat is not implemented in the MVP
	if mode == domain.ChatModeChat {
		return userMsg, nil, domain.NewError(domain.ErrorTypeUnimplemented,
			errors.New("chat mode is not implemented yet"))
	}

	// 5. ensure a sandbox — ONLY if the project has none
	sandboxID := project.SandboxID()
	if sandboxID == "" {
		info, err := s.sandbox.CreateSandbox(ctx)
		if err != nil {
			return nil, nil, domain.WrapError("message service send: create sandbox", err)
		}
		if err := s.projectRepo.UpdateSandbox(ctx, projectID, info.ID, info.URL); err != nil {
			return nil, nil, domain.WrapError("message service send: store sandbox", err)
		}
		sandboxID = info.ID
	}

	// 6. run the agent — blocks for minutes, inherits the caller's context
	result, err := s.codeAgent.RunCodeAgent(ctx, provider, sandboxID, body)
	if err != nil {
		return nil, nil, domain.WrapError("message service send: code agent", err)
	}

	// 7. persist the assistant message — Summary only
	asstMsg, err := domain.NewMessage(projectID, domain.MessageRoleAssistant, mode, provider, result.Summary)
	if err != nil {
		return nil, nil, domain.WrapError("message service send", err)
	}
	asstMsg, err = s.messageRepo.Create(ctx, asstMsg)
	if err != nil {
		return nil, nil, domain.WrapError("message service send", err)
	}

	return userMsg, asstMsg, nil
}
