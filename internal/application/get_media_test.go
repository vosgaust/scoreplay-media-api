package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

func TestGetMediaHandlerHandle(t *testing.T) {
	stored := domain.Media{
		ID:         newID(),
		Name:       "Messi free kick",
		StorageKey: "0198b3c4-0000-7000-8000-000000000001",
		Type:       domain.MediaTypeImage,
		Tags:       []domain.Tag{{ID: newID(), Name: "Messi"}},
	}
	media := &fakeMediaRepository{media: stored}

	view, err := NewGetMediaHandler(media, &fakeFileStore{}).Handle(context.Background(), stored.ID)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if view.Media.ID != stored.ID {
		t.Errorf("Media = %v, want %v", view.Media, stored)
	}
	if want := "https://files.test/" + stored.StorageKey; view.FileURL != want {
		t.Errorf("FileURL = %q, want %q", view.FileURL, want)
	}
}

func TestGetMediaHandlerHandleNotFound(t *testing.T) {
	media := &fakeMediaRepository{getErr: domain.ErrNotFound}

	_, err := NewGetMediaHandler(media, &fakeFileStore{}).Handle(context.Background(), uuid.Nil)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Handle() error = %v, want ErrNotFound", err)
	}
}
