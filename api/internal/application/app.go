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

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/persistence/postgres"
	"github.com/jmoiron/sqlx"
)

type App struct {
	router http.Handler
	db     *sqlx.DB
	config config.Config
}

func New(cfg config.Config) *App {
	app := &App{
		router: loadRoutes(cfg.Origins),
		db:     postgres.NewDb(cfg.DB),
		config: cfg,
	}
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", a.config.ServerPort),
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

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
