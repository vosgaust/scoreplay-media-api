package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/domain/ports"
)

type GetMediaHandler struct {
	media ports.MediaRepository
	files ports.FileStore
}

func NewGetMediaHandler(media ports.MediaRepository, files ports.FileStore) GetMediaHandler {
	return GetMediaHandler{media: media, files: files}
}

func (h GetMediaHandler) Handle(ctx context.Context, id uuid.UUID) (MediaView, error) {
	media, err := h.media.GetByID(ctx, id)
	if err != nil {
		return MediaView{}, err
	}

	return MediaView{Media: media, FileURL: h.files.URL(media.StorageKey)}, nil
}
