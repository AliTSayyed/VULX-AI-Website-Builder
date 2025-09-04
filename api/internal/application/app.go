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

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	rpclogger "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/logger"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/security"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/handlers"
	httpHandlers "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/http/handlers"
	llm "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/outbound/LLM"
	aiservice "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/outbound/ai_service"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/outbound/temporal"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/persistence/postgres"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"github.com/jmoiron/sqlx"
)

type App struct {
	mux      *http.ServeMux
	security *security.SecurityAdapter
	db       *sqlx.DB
	temporal *temporal.Temporal
}

func New(cfg *config.Config) *App {
	utils.InitilizeLogger()

	// ddd dependency injection
	aiservice := aiservice.NewAIService(cfg.AIServiceUrl)
	llmService := llm.New(cfg.LLM)

	temporalService := temporal.New(cfg.Temporal)
	userWorkflow := temporal.NewUserWorkflow(temporalService, llmService, aiservice)
	temporalService.RegisterWorkers(userWorkflow)

	db := postgres.NewDb(cfg.DB)
	userRepo := postgres.NewUserRepository(db)
	userService := services.NewUserService(userRepo, userWorkflow)
	userServiceHandler := handlers.NewUserServiceHandler(userService)

	// create interceptors (middleware) for connect handlers
	interceptor := connect.WithInterceptors(rpclogger.LoggerInterceptor())

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
	mux.Handle("/healthz", healthz())
	mux.Handle("/", transcoder)
	mux.Handle("/docs/", http.StripPrefix("/docs", httpHandlers.NewHandler()))

	// app stores routes, cors (security), and connections to services
	app := &App{
		mux:      mux,
		security: security.NewSecurityAdapter(cfg),
		db:       db,
		temporal: temporalService,
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
	utils.Logger.Info("API Ready For Requests!")

	// main thread blocks on this statement waiting for channel
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		// close connection to services
		defer a.temporal.StopWorkers()
		defer a.temporal.Client.Close()
		defer a.db.Close()
		defer cancel()
		return server.Shutdown(timeout)
	}
}
