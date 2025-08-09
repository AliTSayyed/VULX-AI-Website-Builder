package application

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type App struct {
	router http.Handler
	config Config
}

func New(cfg Config) *App {
	app := &App{
		router: loadRoutes(cfg),
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
