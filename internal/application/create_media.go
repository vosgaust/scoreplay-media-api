package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
	"github.com/vosgaust/scoreplay-media-api/internal/domain/ports"
)

type CreateMedia struct {
	Name   string
	TagIDs []uuid.UUID
	Type   domain.MediaType
	File   io.Reader
}

type CreateMediaHandler struct {
	media ports.MediaRepository
	files ports.FileStore
}

func NewCreateMediaHandler(media ports.MediaRepository, files ports.FileStore) CreateMediaHandler {
	return CreateMediaHandler{media: media, files: files}
}

func (h CreateMediaHandler) Handle(ctx context.Context, cmd CreateMedia) (MediaView, error) {
	id := newID()

	// Built before any I/O: any validation error must not cost bytes on disk
	media, err := domain.NewMedia(id, cmd.Name, id.String(), cmd.Type, time.Now().UTC())
	if err != nil {
		return MediaView{}, err
	}

	key, err := h.files.Put(ctx, media.StorageKey, cmd.File)
	if err != nil {
		return MediaView{}, fmt.Errorf("store media file: %w", err)
	}
	media.StorageKey = key

	tags, err := h.media.CreateWithTags(ctx, media, cmd.TagIDs)
	if err != nil {
		// The file is already stored and cannot be removed through the port, we log the orphan
		slog.WarnContext(ctx, "media file orphaned: stored but not persisted in db",
			"storage_key", media.StorageKey, "error", err)

		return MediaView{}, err
	}
	media.Tags = tags

	return MediaView{Media: media, FileURL: h.files.URL(media.StorageKey)}, nil
}
