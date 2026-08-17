package application

import (
	"context"
	"time"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
	"github.com/vosgaust/scoreplay-media-api/internal/domain/ports"
)

type CreateTag struct {
	Name string
}

type CreateTagHandler struct {
	tags ports.TagRepository
}

func NewCreateTagHandler(tags ports.TagRepository) CreateTagHandler {
	return CreateTagHandler{tags: tags}
}

func (h CreateTagHandler) Handle(ctx context.Context, cmd CreateTag) (domain.Tag, error) {
	tag, err := domain.NewTag(newID(), cmd.Name, time.Now().UTC())
	if err != nil {
		return domain.Tag{}, err
	}

	if err := h.tags.Create(ctx, tag); err != nil {
		return domain.Tag{}, err
	}

	return tag, nil
}
