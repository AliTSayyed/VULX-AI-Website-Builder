package main

import (
	"context"
	"fmt"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application"
)

func main() {
	app := application.New()
	err := app.Start(context.TODO())
	if err != nil {
		fmt.Errorf("failed to start app: %w", err)
	}
}
