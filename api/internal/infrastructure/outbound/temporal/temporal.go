/*
* This file will create a connection to the Temporal service
* We register workers to the correct worklfow queues
* handle closing all workers when server is shutdown
 */
package temporal

import (
	"fmt"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type Temporal struct {
	Client     client.Client
	UserWorker worker.Worker
}

func New(cfg config.Temporal) *Temporal {
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Logger:    utils.Logger,
		Namespace: "default",
	})
	if err != nil {
		panic(fmt.Errorf("unable to create Temporal Client: %w", err))
	}
	utils.Logger.Info("Connected to Temporal service")
	return &Temporal{
		Client: temporalClient,
	}
}

func (temporal *Temporal) RegisterWorkers(userWorkflowInstance *UserWorkflow) *Temporal {
	userWorker := worker.New(temporal.Client, "user-workflow", worker.Options{})
	userWorker.RegisterWorkflow(userWorkflowInstance.UserWorkflowSteps)
	userWorker.RegisterActivity(userWorkflowInstance.SayHello)
	userWorker.RegisterActivity(userWorkflowInstance.UseLlm)

	utils.Logger.Info("Workers successfully registered")

	go func() {
		err := userWorker.Run(worker.InterruptCh())
		if err != nil {
			utils.Logger.Error("failed to run user worker", "error", err)
		}
	}()

	temporal.UserWorker = userWorker
	return temporal
}

func (temporal *Temporal) StopWorkers() {
	temporal.UserWorker.Stop()
}
