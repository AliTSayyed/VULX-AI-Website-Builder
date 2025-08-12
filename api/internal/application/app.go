/*
* This file will create our application.
* While main will "trigger this", this file will create the server
* and relevant connections
* uses listens and serves on a different go routine (uses channels for graceful shutdown)
 */
package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/handlers"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/persistence/postgres"
)

type App struct {
	router http.Handler
}

func New(cfg config.Config) *App {
	db := postgres.NewDb(cfg.DB)
	userRepo := postgres.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userServiceHandler := handlers.NewUserServiceHandler(userService)

	app := &App{
		router: loadRoutes(cfg.Origins, userServiceHandler),
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	// main thread blocks on this statement waiting for channel
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
