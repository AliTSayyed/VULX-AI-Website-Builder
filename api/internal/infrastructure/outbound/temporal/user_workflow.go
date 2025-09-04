package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	llm "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/outbound/LLM"
	aiservice "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/outbound/ai_service"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type UserWorkflow struct {
	client    client.Client
	llm       *llm.LLM
	aiservice *aiservice.AIService
}

func NewUserWorkflow(temporal *Temporal, llm *llm.LLM, aiservice *aiservice.AIService) *UserWorkflow {
	return &UserWorkflow{
		client:    temporal.Client,
		llm:       llm,
		aiservice: aiservice,
	}
}

type UserWorkflowData struct {
	Greeting string
}

// This will queue a workflow
func (u *UserWorkflow) StartUserWorkflow(ctx context.Context) error {
	options := client.StartWorkflowOptions{
		TaskQueue: "user-workflow",
	}
	input := UserWorkflowData{
		Greeting: "Hello there user from temporal worfklow execution",
	}
	workflowRun, err := u.client.ExecuteWorkflow(
		context.Background(),
		options,
		"UserWorkflowSteps",
		input,
	)
	utils.Logger.Info("Starting user workflow by placing it in queue")
	if err != nil {
		return err
	}
	// Optional: wait for result or just return will block if we wait
	return workflowRun.Get(context.Background(), nil)
}

// Activities need context.Context, not workflow.Context
func (u *UserWorkflow) CreateSandbox(ctx context.Context, data UserWorkflowData) error {
	SandboxResponse, err := u.aiservice.CreateSandbox(ctx)
	utils.Logger.Info(SandboxResponse.ID)
	return err
}

func (u *UserWorkflow) UseLlm(ctx context.Context, data UserWorkflowData) error {
	// prompt := "Write a simple react button component function"

	// response, err := u.llm.OpenaiClient.GenerateContent(ctx, []llms.MessageContent{
	// 	llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	// })
	// if err != nil {
	// 	return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed at UseLlm, %w", err))
	// }

	// fmt.Println("LLM Response:", response.Choices[0].Content)
	return nil
}

// Workflows need to return error and properly handle activity execution
func (u *UserWorkflow) UserWorkflowSteps(ctx workflow.Context, input UserWorkflowData) error {
	// Need activity options and proper error handling
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	err := workflow.ExecuteActivity(ctx, "CreateSandbox", input).Get(ctx, nil)
	utils.Logger.Info("Running workflow step 1 inside user workflow")
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed at UserWorkflowSteps, %w", err))
	}

	err = workflow.ExecuteActivity(ctx, "UseLlm", input).Get(ctx, nil)
	utils.Logger.Info("Running workflow step 2 inside user workflow")
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed at UserWorkflowSteps, %w", err))
	}
	return nil
}
