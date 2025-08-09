package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application"
)

func main() {
	app := application.New(application.LoadConfig())
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	err := app.Start(ctx)
	if err != nil {
		fmt.Printf("failed to start app: %v\n", err)
	}
}
