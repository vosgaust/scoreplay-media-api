package http

import (
	"context"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/application"
	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type tagCreator interface {
	Handle(ctx context.Context, cmd application.CreateTag) (domain.Tag, error)
}

type tagLister interface {
	Handle(ctx context.Context) ([]domain.Tag, error)
}

type mediaCreator interface {
	Handle(ctx context.Context, cmd application.CreateMedia) (application.MediaView, error)
}

type mediaFinder interface {
	Handle(ctx context.Context, id uuid.UUID) (application.MediaView, error)
}
