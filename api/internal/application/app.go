/*
* This file will create our application.
* While main will "trigger this", this file will create the server and relevant connections
* uses listen and serve on a different go routine (use channels for graceful shutdown)
 */
package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	connectrpc "connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	logger "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/logger"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/security"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/handlers"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/persistence/postgres"
	utils "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/util"
)

type App struct {
	mux      *http.ServeMux
	security *security.SecurityAdapter
}

func New(cfg *config.Config) *App {
	utils.InitilizeLogger()

	db := postgres.NewDb(cfg.DB)
	userRepo := postgres.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userServiceHandler := handlers.NewUserServiceHandler(userService)

	// use vangaurd to create rest and rpc compatible connect handlers with middleware (interceptors)
	interceptor := connectrpc.WithInterceptors(logger.LoggerInterceptor())
	services := []*vanguard.Service{
		vanguard.NewService(apiv1connect.NewUserServiceHandler(userServiceHandler, interceptor)),
	}

	connectHandlers, err := vanguard.NewTranscoder(services)
	if err != nil {
		panic(fmt.Errorf("failed to mount transcode handlers: %w", err))
	}

	// routing
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())
	mux.Handle("/", connectHandlers)
	// mux.Handle("/api/inngest", inngestAdapter.Client.Serve())

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

	// start sever on new thread
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()
	utils.Logger.Info("API ready for requests")

	// main thread blocks on this statement waiting for err on server or ctx done call.
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
