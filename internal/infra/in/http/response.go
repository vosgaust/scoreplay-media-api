package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.ErrorContext(r.Context(), "encoding response body failed", "error", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, r, status, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}
