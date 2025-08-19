package temporal

import (
	"fmt"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"go.temporal.io/sdk/client"
)

type Temporal struct {
	Client client.Client
}

func New(cfg config.Temporal) *Temporal {
	temporalClient, err := client.Dial(client.Options{
		HostPort: cfg.HostPort,
		Logger:   utils.Logger,
	})
	if err != nil {
		panic(fmt.Errorf("unable to create Temporal Client: %w", err))
	}
	utils.Logger.Info("Connected to Temporal service")
	return &Temporal{
		Client: temporalClient,
	}
}
