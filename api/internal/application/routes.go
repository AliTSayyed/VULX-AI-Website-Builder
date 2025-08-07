/*
* This file will load all endpoint routes for the server
* Uses connectrpc + vangaurd to make gRPC calls compatible with rest urls
 */

package application

import (
	"net/http"

	"connectrpc.com/vanguard"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	handlers "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func loadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	services := []*vanguard.Service{
		vanguard.NewService(apiv1connect.NewUserServiceHandler(&handlers.UserService{})),
	}

	transcoder, err := vanguard.NewTranscoder(services)
	if err != nil {
		panic(err)
	}

	router.Mount("/", transcoder)
	return router
}
