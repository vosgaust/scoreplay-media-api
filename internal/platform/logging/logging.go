package logging

import (
	"io"
	"log/slog"
)

func Init(w io.Writer, level slog.Level) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})))
}
