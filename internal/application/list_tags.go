package application

import (
	"context"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
	"github.com/vosgaust/scoreplay-media-api/internal/domain/ports"
)

type ListTagsHandler struct {
	tags ports.TagRepository
}

func NewListTagsHandler(tags ports.TagRepository) ListTagsHandler {
	return ListTagsHandler{tags: tags}
}

func (h ListTagsHandler) Handle(ctx context.Context) ([]domain.Tag, error) {
	return h.tags.List(ctx)
}
