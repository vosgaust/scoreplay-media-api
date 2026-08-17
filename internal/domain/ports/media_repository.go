package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type MediaRepository interface {
	CreateWithTags(ctx context.Context, media domain.Media, tagIDs []uuid.UUID) ([]domain.Tag, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Media, error)
}
