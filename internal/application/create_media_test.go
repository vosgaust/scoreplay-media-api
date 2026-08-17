package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

func TestCreateMediaHandlerHandle(t *testing.T) {
	resolved := []domain.Tag{{ID: newID(), Name: "Messi"}}
	media := &fakeMediaRepository{resolved: resolved}
	files := &fakeFileStore{keyPrefix: "bucket/"}
	cmd := newCreateMedia("Messi free kick")

	view, err := NewCreateMediaHandler(media, files).Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if view.Media.ID.Version() != 7 {
		t.Errorf("ID version = %d, want 7", view.Media.ID.Version())
	}
	if files.key != view.Media.ID.String() {
		t.Errorf("stored under key %q, want %q", files.key, view.Media.ID)
	}
	if files.content != "jpeg bytes" {
		t.Errorf("stored content = %q, want the uploaded bytes", files.content)
	}
	if want := "bucket/" + view.Media.ID.String(); view.Media.StorageKey != want {
		t.Errorf("StorageKey = %q, want the key the store returned (%q)", view.Media.StorageKey, want)
	}
	if media.created.StorageKey != view.Media.StorageKey {
		t.Errorf("persisted key %q, want %q", media.created.StorageKey, view.Media.StorageKey)
	}
	if len(view.Media.Tags) != 1 || view.Media.Tags[0] != resolved[0] {
		t.Errorf("Tags = %v, want %v", view.Media.Tags, resolved)
	}
	if want := "https://files.test/" + view.Media.StorageKey; view.FileURL != want {
		t.Errorf("FileURL = %q, want %q", view.FileURL, want)
	}
}

func TestCreateMediaHandlerHandleWithoutTags(t *testing.T) {
	media := &fakeMediaRepository{}
	files := &fakeFileStore{}

	view, err := NewCreateMediaHandler(media, files).Handle(context.Background(), newCreateMedia("untagged"))
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if len(media.tagIDs) != 0 {
		t.Errorf("repository received tag ids %v, want none", media.tagIDs)
	}
	if len(view.Media.Tags) != 0 {
		t.Errorf("Tags = %v, want none", view.Media.Tags)
	}
}

func TestCreateMediaHandlerHandleErrors(t *testing.T) {
	repoFailure := errors.New("insert media: connection lost")
	storeFailure := errors.New("disk full")

	tests := []struct {
		name          string
		cmd           CreateMedia
		putErr        error
		createErr     error
		wantErr       error
		wantPutCalls  int
		wantRepoCalls int
	}{
		{
			name:    "an invalid name costs no bytes on disk",
			cmd:     newCreateMedia("  "),
			wantErr: domain.ValidationError{Field: "name", Message: "must not be empty"},
		},
		{
			name:    "a forged media type costs no bytes on disk",
			cmd:     CreateMedia{Name: "x", Type: domain.MediaType("pdf"), File: strings.NewReader("%PDF")},
			wantErr: domain.UnsupportedMediaTypeError{MediaType: "pdf"},
		},
		{
			name:         "a store failure never reaches the database",
			cmd:          newCreateMedia("Messi free kick"),
			putErr:       storeFailure,
			wantErr:      storeFailure,
			wantPutCalls: 1,
		},
		{
			name:          "a repository failure is passed through",
			cmd:           newCreateMedia("Messi free kick"),
			createErr:     repoFailure,
			wantErr:       repoFailure,
			wantPutCalls:  1,
			wantRepoCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := &fakeMediaRepository{createErr: tt.createErr}
			files := &fakeFileStore{putErr: tt.putErr}

			_, err := NewCreateMediaHandler(media, files).Handle(context.Background(), tt.cmd)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Handle() error = %v, want %v", err, tt.wantErr)
			}
			if files.calls != tt.wantPutCalls {
				t.Errorf("Put calls = %d, want %d", files.calls, tt.wantPutCalls)
			}
			if media.calls != tt.wantRepoCalls {
				t.Errorf("repository calls = %d, want %d", media.calls, tt.wantRepoCalls)
			}
		})
	}
}
