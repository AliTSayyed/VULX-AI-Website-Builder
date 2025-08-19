package temporal

import (
	"context"
	"fmt"
	"time"

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

// TODO create workers, register workflows and activities to workers.
// This will queue a workflow
func (u *UserWorkflow) StartUserWorkflow(ctx context.Context) error {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "user-work",
	}
	workflowRun, err := u.client.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		UserWorkFlow,
	)
	if err != nil {
		return err
	}
	// Optional: wait for result or just return
	return workflowRun.Get(context.Background(), nil)
}

// Fix: Activities need context.Context, not workflow.Context
func SayHello(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	fmt.Println("Hello from the activity function")
	return nil
}

// Fix: Workflows need to return error and properly handle activity execution
func UserWorkFlow(ctx workflow.Context) error {
	// Fix: Need activity options and proper error handling
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	err := workflow.ExecuteActivity(ctx, SayHello).Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
