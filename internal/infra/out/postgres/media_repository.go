package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type MediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) MediaRepository {
	return MediaRepository{pool: pool}
}

func (r MediaRepository) CreateWithTags(
	ctx context.Context,
	media domain.Media,
	tagIDs []uuid.UUID,
) ([]domain.Tag, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	tags, err := resolveTags(ctx, tx, tagIDs)
	if err != nil {
		return nil, err
	}

	if err := insertMedia(ctx, tx, media); err != nil {
		return nil, err
	}

	if err := linkTags(ctx, tx, media.ID, tagIDs); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return tags, nil
}

func resolveTags(ctx context.Context, tx pgx.Tx, tagIDs []uuid.UUID) ([]domain.Tag, error) {
	const query = `SELECT id, name, created_at FROM tags WHERE id = ANY($1) ORDER BY lower(name)`

	if len(tagIDs) == 0 {
		return []domain.Tag{}, nil
	}

	rows, err := tx.Query(ctx, query, tagIDs)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}

	tags, err := collectTags(rows)
	if err != nil {
		return nil, fmt.Errorf("collect tags: %w", err)
	}

	if missing := missingTagIDs(tagIDs, tags); len(missing) > 0 {
		return nil, domain.UnknownTagsError{IDs: missing}
	}

	return tags, nil
}

func missingTagIDs(requested []uuid.UUID, found []domain.Tag) []uuid.UUID {
	existing := make(map[uuid.UUID]struct{}, len(found))
	for _, tag := range found {
		existing[tag.ID] = struct{}{}
	}

	var missing []uuid.UUID

	for _, id := range requested {
		if _, ok := existing[id]; ok {
			continue
		}
		existing[id] = struct{}{}
		missing = append(missing, id)
	}

	return missing
}

func insertMedia(ctx context.Context, tx pgx.Tx, media domain.Media) error {
	const query = `INSERT INTO media (id, name, storage_key, type, created_at)
	               VALUES ($1, $2, $3, $4, $5)`

	_, err := tx.Exec(ctx, query,
		media.ID, media.Name, media.StorageKey, string(media.Type), media.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert media: %w", err)
	}

	return nil
}

func linkTags(ctx context.Context, tx pgx.Tx, mediaID uuid.UUID, tagIDs []uuid.UUID) error {
	const query = `INSERT INTO media_tags (media_id, tag_id)
	               SELECT DISTINCT $1::uuid, tag_id FROM unnest($2::uuid[]) AS tag_id`

	if len(tagIDs) == 0 {
		return nil
	}

	if _, err := tx.Exec(ctx, query, mediaID, tagIDs); err != nil {
		if isForeignKeyViolation(err) {
			return domain.UnknownTagsError{IDs: tagIDs}
		}

		return fmt.Errorf("link tags: %w", err)
	}

	return nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23503" // foreign_key_violation
}

func (r MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Media, error) {
	media, err := r.selectMedia(ctx, id)
	if err != nil {
		return domain.Media{}, err
	}

	tags, err := r.selectTagsOf(ctx, id)
	if err != nil {
		return domain.Media{}, err
	}

	media.Tags = tags

	return media, nil
}

func (r MediaRepository) selectMedia(ctx context.Context, id uuid.UUID) (domain.Media, error) {
	const query = `SELECT id, name, storage_key, type, created_at FROM media WHERE id = $1`

	var (
		media   domain.Media
		rawType string
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&media.ID, &media.Name, &media.StorageKey, &rawType, &media.CreatedAt,
	)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Media{}, domain.ErrNotFound
	case err != nil:
		return domain.Media{}, fmt.Errorf("query media: %w", err)
	}

	media.Type, err = domain.ParseMediaType(rawType)
	if err != nil {
		return domain.Media{}, fmt.Errorf("media %s has an unreadable type: %w", id, err)
	}

	return media, nil
}

func (r MediaRepository) selectTagsOf(ctx context.Context, mediaID uuid.UUID) ([]domain.Tag, error) {
	const query = `SELECT t.id, t.name, t.created_at
	               FROM tags t
	               JOIN media_tags mt ON mt.tag_id = t.id
	               WHERE mt.media_id = $1
	               ORDER BY lower(t.name)`

	rows, err := r.pool.Query(ctx, query, mediaID)
	if err != nil {
		return nil, fmt.Errorf("query media tags: %w", err)
	}

	tags, err := collectTags(rows)
	if err != nil {
		return nil, fmt.Errorf("collect media tags: %w", err)
	}

	return tags, nil
}
