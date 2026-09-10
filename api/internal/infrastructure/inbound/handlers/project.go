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

type ProjectServiceHandler struct {
	apiv1connect.UnimplementedProjectServiceHandler

	projectService *services.ProjectService
	authAdapter    *authAdapter.HTTPAuthAdapter
}

func NewProjectServiceHandler(projectService *services.ProjectService, authAdapter *authAdapter.HTTPAuthAdapter) *ProjectServiceHandler {
	return &ProjectServiceHandler{
		projectService: projectService,
		authAdapter:    authAdapter,
	}
}

func (h *ProjectServiceHandler) ListProjects(ctx context.Context, req *connect.Request[apiv1.ListProjectsRequest]) (*connect.Response[apiv1.ListProjectsResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	page, err := h.projectService.List(ctx, user.ID(), req.Msg.GetLimit(), req.Msg.GetToken())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	projects := make([]*apiv1.Project, len(page.Items))
	for i, project := range page.Items {
		projects[i] = projectToProto(project)
	}

	return connect.NewResponse(&apiv1.ListProjectsResponse{
		Projects: projects,
		Token:    page.Token,
		HasMore:  page.HasMore,
	}), nil
}

func (h *ProjectServiceHandler) GetProject(ctx context.Context, req *connect.Request[apiv1.GetProjectRequest]) (*connect.Response[apiv1.GetProjectResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	project, err := h.projectService.Get(ctx, user.ID(), id)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.GetProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ProjectServiceHandler) CreateProject(ctx context.Context, req *connect.Request[apiv1.CreateProjectRequest]) (*connect.Response[apiv1.CreateProjectResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	provider := providerFromProto(req.Msg.GetProvider())

	project, err := h.projectService.Create(ctx, user.ID(), req.Msg.GetFirstPrompt(), provider)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.CreateProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func projectToProto(p *domain.Project) *apiv1.Project {
	if p == nil {
		return nil
	}
	return &apiv1.Project{
		Id:         p.ID().String(),
		Title:      p.Title(),
		Provider:   providerToProto(p.Provider()),
		SandboxId:  p.SandboxID(),
		PreviewUrl: p.PreviewURL(),
		CreatedAt:  p.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt:  p.UpdatedAt().UTC().Format(time.RFC3339),
	}
}

func providerToProto(p domain.AIProvider) apiv1.AiProvider {
	switch p {
	case domain.AIProviderOpenAI:
		return apiv1.AiProvider_AI_PROVIDER_OPENAI
	case domain.AIProviderGoogle:
		return apiv1.AiProvider_AI_PROVIDER_GOOGLE
	case domain.AIProviderAnthropic:
		return apiv1.AiProvider_AI_PROVIDER_ANTHROPIC
	default:
		return apiv1.AiProvider_AI_PROVIDER_UNSPECIFIED
	}
}

func providerFromProto(p apiv1.AiProvider) domain.AIProvider {
	switch p {
	case apiv1.AiProvider_AI_PROVIDER_OPENAI:
		return domain.AIProviderOpenAI
	case apiv1.AiProvider_AI_PROVIDER_GOOGLE:
		return domain.AIProviderGoogle
	case apiv1.AiProvider_AI_PROVIDER_ANTHROPIC:
		return domain.AIProviderAnthropic
	default:
		return domain.AIProviderUnspecified
	}
}
