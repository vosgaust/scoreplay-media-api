package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
)

func NewRouter(database readinessProbe, tags TagHandler, media MediaHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
	}))

	r.Get("/healthz", handleHealthz)
	r.Get("/readyz", handleReadyz(database))

	r.Post("/tags", tags.Create)
	r.Get("/tags", tags.List)

	r.Post("/media", media.Create)
	r.Get("/media/{id}", media.Get)

	return r
}
