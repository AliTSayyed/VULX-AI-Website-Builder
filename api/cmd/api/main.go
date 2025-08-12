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
	app := application.New(config)
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	err = app.Start(ctx)
	if err != nil {
		fmt.Printf("failed to start app: %v\n", err)
	}
}
