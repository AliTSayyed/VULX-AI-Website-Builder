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

func (temporal *Temporal) RegisterWorkers() *Temporal {
	// No real workflow registered yet. To register one, build a worker.Worker and store
	// it on Temporal.UserWorker, then return temporal. See NewUserWorkflow in
	// user_workflow.go for the workflow/activity shape this would register:
	//
	//   userWorker := worker.New(temporal.Client, "user-workflow", worker.Options{})
	//   userWorker.RegisterWorkflow(userWorkflowInstance.UserWorkflowSteps)
	//   userWorker.RegisterActivity(userWorkflowInstance.CreateSandbox)
	//   userWorker.RegisterActivity(userWorkflowInstance.UseLlm)
	//   go func() {
	//       err := userWorker.Run(worker.InterruptCh())
	//       if err != nil {
	//           utils.Logger.Error("failed to run user worker", "error", err)
	//       }
	//   }()
	//   temporal.UserWorker = userWorker
	return temporal
}

func (temporal *Temporal) StopWorkers() {
	// No real worker running yet — RegisterWorkers doesn't create one. When it does,
	// stop it here:
	//   temporal.UserWorker.Stop()
}
