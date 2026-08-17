//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const postgresImage = "postgres:17-alpine"

const (
	testDBName     = "scoreplay"
	testDBUser     = "scoreplay"
	testDBPassword = "scoreplay"
)

const migrationsSource = "file://../../../../migrations"

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPassword),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Printf("start postgres container: %v", err)
		return 1
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("terminate postgres container: %v", err)
		}
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("read container connection string: %v", err)
		return 1
	}

	if err := applyMigrations(dsn); err != nil {
		log.Printf("%v", err)
		return 1
	}

	pool, err := NewPool(ctx, dsn)
	if err != nil {
		log.Printf("open test pool: %v", err)
		return 1
	}
	defer pool.Close()

	testPool = pool

	return m.Run()
}

func applyMigrations(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("parse container dsn: %w", err)
	}
	parsed.Scheme = "pgx5"

	migrator, err := migrate.New(migrationsSource, parsed.String())
	if err != nil {
		return fmt.Errorf("open migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func resetDB(t *testing.T) {
	t.Helper()

	_, err := testPool.Exec(t.Context(), "TRUNCATE media_tags, media, tags")
	require.NoError(t, err)
}
