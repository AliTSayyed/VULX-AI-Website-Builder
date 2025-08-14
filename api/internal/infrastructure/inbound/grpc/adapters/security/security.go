package security

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/rs/cors"
)

type SecurityAdapter struct {
	apiUrl string
	appUrl string
}

func NewSecurityAdapter(cfg *config.Config) *SecurityAdapter {
	return &SecurityAdapter{
		apiUrl: cfg.ApiUrl,
		appUrl: cfg.AppUrl,
	}
}

func (s *SecurityAdapter) SecurityAdapterCors(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{s.apiUrl, s.appUrl},
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: true,
		MaxAge:           7200,
	}).Handler(h)
}
