package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
)

// NewRouter wires the transport layer. It reads slog.Default(), so logging.Init must run first.
func NewRouter(database readinessProbe) http.Handler {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
	}))

	r.Get("/healthz", handleHealthz)
	r.Get("/readyz", handleReadyz(database))

	return r
}
