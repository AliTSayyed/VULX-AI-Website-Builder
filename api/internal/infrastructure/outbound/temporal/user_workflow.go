package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type UserWorkflow struct {
	client client.Client
}

func NewUserWorkflow(temporal *Temporal) *UserWorkflow {
	return &UserWorkflow{
		client: temporal.Client,
	}
}

type UserWorkflowDetails struct {
	Greeting string
}

// This will queue a workflow
func (u *UserWorkflow) StartUserWorkflow(ctx context.Context) error {
	options := client.StartWorkflowOptions{
		TaskQueue: "user-workflow",
	}
	input := UserWorkflowDetails{
		Greeting: "Hello there user from temporal worfklow execution",
	}
	workflowRun, err := u.client.ExecuteWorkflow(
		context.Background(),
		options,
		UserWorkFlow,
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
func SayHello(ctx context.Context, data UserWorkflowDetails) error {
	time.Sleep(5 * time.Second)
	fmt.Println(data.Greeting)
	return nil
}

// Workflows need to return error and properly handle activity execution
func UserWorkFlow(ctx workflow.Context, input UserWorkflowDetails) error {
	// Need activity options and proper error handling
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	err := workflow.ExecuteActivity(ctx, SayHello, input).Get(ctx, nil)
	utils.Logger.Info("Running workflow steps inside user workflow")
	if err != nil {
		return err
	}
	return nil
}
