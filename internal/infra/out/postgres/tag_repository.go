package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) TagRepository {
	return TagRepository{pool: pool}
}

func (r TagRepository) Create(ctx context.Context, tag domain.Tag) error {
	const query = `INSERT INTO tags (id, name, created_at) VALUES ($1, $2, $3)`

	if _, err := r.pool.Exec(ctx, query, tag.ID, tag.Name, tag.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ConflictError{Resource: "tag", Field: "name"}
		}

		return fmt.Errorf("insert tag: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505" // unique_violation
}

func (r TagRepository) List(ctx context.Context) ([]domain.Tag, error) {
	const query = `SELECT id, name, created_at FROM tags ORDER BY lower(name)`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}

	tags, err := collectTags(rows)
	if err != nil {
		return nil, fmt.Errorf("collect tags: %w", err)
	}

	return tags, nil
}

// collectTags rehydrates tag rows for every query in this package that selects id, name and
// created_at. Field assignment rather than domain.NewTag: a constructor guards creation, while
// this is reconstituting rows that already passed it — and reading through it would make every
// stored row unreadable the day a validation rule tightens.
//
// CollectRows closes rows and returns a non-nil empty slice when there are none, which is what
// keeps an empty tag list serialising as [] rather than null.
func collectTags(rows pgx.Rows) ([]domain.Tag, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Tag, error) {
		var tag domain.Tag
		err := row.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)

		return tag, err
	})
}
