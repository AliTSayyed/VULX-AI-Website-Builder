package inngest

import (
	"context"
	"fmt"
	"net/url"

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

func InitilizeIngest(ctx context.Context, client inngestgo.Client) {
	conn, err := inngestgo.Connect(ctx, inngestgo.ConnectOpts{
		InstanceID: inngestgo.Ptr("vulx-worker"),
		Apps:       []inngestgo.Client{client},
		RewriteGatewayEndpoint: func(endpoint url.URL) (url.URL, error) {
			utils.Logger.Info("Original endpoint", "url", endpoint.String())
			endpoint.Host = "inngest:8289"
			utils.Logger.Info("Rewritten endpoint", "url", endpoint.String())
			return endpoint, nil
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed connection with Inngest: %w", err))
	}
	utils.Logger.Info("Connection to Inngest established")

	// block this thread since this function will be called in a go rotuine
	<-ctx.Done()
	err = conn.Close()
	if err != nil {
		utils.Logger.Error("could not close connection to Inngest", "error", err)
	}
}
