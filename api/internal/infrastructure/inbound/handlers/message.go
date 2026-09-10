package handlers

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	authAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/auth"
	errorAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/error"
	apiv1 "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	"github.com/google/uuid"
)

type MessageServiceHandler struct {
	apiv1connect.UnimplementedMessageServiceHandler

	messageService *services.MessageService
	authAdapter    *authAdapter.HTTPAuthAdapter
}

func NewMessageServiceHandler(messageService *services.MessageService, authAdapter *authAdapter.HTTPAuthAdapter) *MessageServiceHandler {
	return &MessageServiceHandler{
		messageService: messageService,
		authAdapter:    authAdapter,
	}
}

func (h *MessageServiceHandler) ListMessages(ctx context.Context, req *connect.Request[apiv1.ListMessagesRequest]) (*connect.Response[apiv1.ListMessagesResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := uuid.Parse(req.Msg.GetProjectId())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	messages, err := h.messageService.List(ctx, user.ID(), projectID)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	protoMessages := make([]*apiv1.Message, len(messages))
	for i, m := range messages {
		protoMessages[i] = messageToProto(m)
	}

	return connect.NewResponse(&apiv1.ListMessagesResponse{
		Messages: protoMessages,
	}), nil
}

func (h *MessageServiceHandler) SendMessage(ctx context.Context, req *connect.Request[apiv1.SendMessageRequest]) (*connect.Response[apiv1.SendMessageResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := uuid.Parse(req.Msg.GetProjectId())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	mode := modeFromProto(req.Msg.GetMode())
	provider := providerFromProto(req.Msg.GetProvider())

	userMsg, asstMsg, err := h.messageService.Send(ctx, user.ID(), projectID, req.Msg.GetBody(), mode, provider)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.SendMessageResponse{
		UserMessage:      messageToProto(userMsg),
		AssistantMessage: messageToProto(asstMsg),
	}), nil
}

func messageToProto(m *domain.Message) *apiv1.Message {
	if m == nil {
		return nil
	}
	return &apiv1.Message{
		Id:        m.ID().String(),
		ProjectId: m.ProjectID().String(),
		Role:      roleToProto(m.Role()),
		Mode:      modeToProto(m.Mode()),
		Provider:  providerToProto(m.Provider()),
		Body:      m.Body(),
		CreatedAt: m.CreatedAt().UTC().Format(time.RFC3339),
	}
}

func roleToProto(r domain.MessageRole) apiv1.MessageRole {
	switch r {
	case domain.MessageRoleUser:
		return apiv1.MessageRole_MESSAGE_ROLE_USER
	case domain.MessageRoleAssistant:
		return apiv1.MessageRole_MESSAGE_ROLE_ASSISTANT
	default:
		return apiv1.MessageRole_MESSAGE_ROLE_UNSPECIFIED
	}
}

func modeToProto(m domain.ChatMode) apiv1.ChatMode {
	switch m {
	case domain.ChatModeChat:
		return apiv1.ChatMode_CHAT_MODE_CHAT
	case domain.ChatModeBuild:
		return apiv1.ChatMode_CHAT_MODE_BUILD
	default:
		return apiv1.ChatMode_CHAT_MODE_UNSPECIFIED
	}
}

func modeFromProto(m apiv1.ChatMode) domain.ChatMode {
	switch m {
	case apiv1.ChatMode_CHAT_MODE_CHAT:
		return domain.ChatModeChat
	case apiv1.ChatMode_CHAT_MODE_BUILD:
		return domain.ChatModeBuild
	default:
		return domain.ChatModeUnspecified
	}
}
