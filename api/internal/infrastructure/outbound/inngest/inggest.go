package inngest

import (
	"context"
	"fmt"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	utils "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/util"
	"github.com/inngest/inngestgo"
)

type InngestAdapter struct {
	Client inngestgo.Client
}

func NewInngestAdapter(cfg config.InngestClient) *InngestAdapter {
	client, err := inngestgo.NewClient(inngestgo.ClientOpts{
		AppID:  cfg.AppID,
		Logger: utils.Logger,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create Ingest client: %w", err))
	}

	return &InngestAdapter{
		Client: client,
	}
}

// temp test func to invoke in inngest
func TestInngestFunc(ctx context.Context, client inngestgo.Client) {
	_, err := inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{
			ID:   "hello",
			Name: "say hello",
		},
		// Run on every api/hello.created event.
		inngestgo.EventTrigger("api/hello", nil),
		func(ctx context.Context, input inngestgo.Input[any]) (any, error) {
			utils.Logger.Info("Function started - sleeping for 10 seconds")
			time.Sleep(10 * time.Second)
			utils.Logger.Info("Hello from Inngest function!")
			return map[string]any{"message": "hello"}, nil
		},
	)
	if err != nil {
		panic(err)
	}
}
