/*
* This file will create our application.
* While main will "trigger this", this file will create the server and relevant connections
* uses listen and serve on a different go routine (use channels for graceful shutdown)
 */
package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	logger "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/logger"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/security"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/handlers"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/persistence/postgres"
)

type App struct {
	mux      *http.ServeMux
	security *security.SecurityAdapter
}

func New(cfg *config.Config) *App {
	// ddd dependency injection
	db := postgres.NewDb(cfg.DB)
	userRepo := postgres.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userServiceHandler := handlers.NewUserServiceHandler(userService)

	// create interceptors (middleware) for connect handlers
	interceptor := connect.WithInterceptors(logger.LoggerInterceptor())

	// use vangaurd to create rest and rpc compatible connect handlers
	services := []*vanguard.Service{
		vanguard.NewService(apiv1connect.NewUserServiceHandler(userServiceHandler, interceptor)),
	}

	transcoder, err := vanguard.NewTranscoder(services)
	if err != nil {
		panic(fmt.Errorf("failed to mount transcode handlers: %w", err))
	}

	// create routes
	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}))
	mux.Handle("/", transcoder)

	// store routes and cors config in the app
	app := &App{
		mux:      mux,
		security: security.NewSecurityAdapter(cfg),
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.security.SecurityAdapterCors(a.mux),
	}

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()
	slog.Info("API Ready For Requests!")

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
