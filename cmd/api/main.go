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

	"github.com/vosgaust/scoreplay-media-api/internal/application"
	api "github.com/vosgaust/scoreplay-media-api/internal/infra/in/http"
	"github.com/vosgaust/scoreplay-media-api/internal/infra/out/postgres"
	"github.com/vosgaust/scoreplay-media-api/internal/infra/out/storage"
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

	pool, err := postgres.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	slog.Info("database connected")

	var (
		tagRepository   = postgres.NewTagRepository(pool)
		mediaRepository = postgres.NewMediaRepository(pool)
		fileStore       = storage.NewLocalFileStore(cfg.Storage.Root, cfg.Server.BaseURL)

		tags = api.NewTagHandler(
			application.NewCreateTagHandler(tagRepository),
			application.NewListTagsHandler(tagRepository),
		)
		media = api.NewMediaHandler(
			application.NewCreateMediaHandler(mediaRepository, fileStore),
			application.NewGetMediaHandler(mediaRepository, fileStore),
			cfg.Storage.MaxUploadBytes,
		)
	)

	srv := &http.Server{
		Handler:           api.NewRouter(pool, tags, media),
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
