package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("invalid config %w", err))
	}
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	app := application.New(config)
	err = app.Start(ctx)
	if err != nil {
		fmt.Printf("failed to start app: %v\n", err)
	}
}
