package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/vosgaust/scoreplay-media-api/internal/infra/in/http"
	"github.com/vosgaust/scoreplay-media-api/internal/platform/config"
	"github.com/vosgaust/scoreplay-media-api/internal/platform/logging"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	logging.Init(os.Stdout, cfg.General.LogLevel)

	srv := &http.Server{
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	listener, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		return err
	}

	slog.Info("server listening", "addr", listener.Addr().String())

	go func() {
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("serving stopped unexpectedly", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.General.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("shutdown complete")

	return nil
}
