package handlers

import (
	_ "embed"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//go:embed openapi.yaml
var spec []byte

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", httpSwagger.Handler(
		httpSwagger.URL("/docs/openapi.yaml"),
	))

	mux.Handle("/openapi.yaml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(spec)
	}))

	return mux
}
