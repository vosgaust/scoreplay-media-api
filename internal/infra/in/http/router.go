package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
	}))

	r.Get("/healthz", handleHealthz)

	return r
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}
