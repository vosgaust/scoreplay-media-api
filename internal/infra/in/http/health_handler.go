package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type readinessProbe interface {
	Ping(ctx context.Context) error
}

const readinessTimeout = 2 * time.Second

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

func handleReadyz(database readinessProbe) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			slog.ErrorContext(ctx, "readiness check failed", "dependency", "postgres", "error", err)
			writeError(w, r, http.StatusServiceUnavailable, "not_ready", "a dependency is unreachable")

			return
		}

		writeJSON(w, r, http.StatusOK, map[string]any{
			"status":       "ok",
			"dependencies": map[string]string{"postgres": "ok"},
		})
	}
}
