package ports

import (
	"context"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type TagRepository interface {
	Create(ctx context.Context, tag domain.Tag) error
	List(ctx context.Context) ([]domain.Tag, error)
}
