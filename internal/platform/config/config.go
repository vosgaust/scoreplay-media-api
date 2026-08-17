package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	General  General
	Server   Server
	Storage  Storage
	Database Database
}

type General struct {
	LogLevel        slog.Level    `env:"SCOREPLAY_LOG_LEVEL,default=info"`
	ShutdownTimeout time.Duration `env:"SCOREPLAY_SHUTDOWN_TIMEOUT,default=10s"`
}

type Server struct {
	Port    string `env:"SCOREPLAY_SERVER_PORT,default=8080"`
	BaseURL string `env:"SCOREPLAY_SERVER_BASE_URL,default=http://localhost:8080"`
}

type Storage struct {
	Root           string `env:"SCOREPLAY_STORAGE_ROOT,default=./data/media"`
	MaxUploadBytes int64  `env:"SCOREPLAY_STORAGE_MAX_UPLOAD_BYTES,default=5242880"` // 5 MiB
}

type Database struct {
	URL string `env:"SCOREPLAY_DATABASE_URL,required"`
}

func Load(ctx context.Context) (Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return Config{}, fmt.Errorf("read environment: %w", err)
	}

	return cfg, nil
}
