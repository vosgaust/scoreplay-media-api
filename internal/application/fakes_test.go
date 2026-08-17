package application

import (
	"context"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type fakeTagRepository struct {
	calls     int
	created   []domain.Tag
	tags      []domain.Tag
	createErr error
	listErr   error
}

func (f *fakeTagRepository) Create(_ context.Context, tag domain.Tag) error {
	f.calls++
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, tag)

	return nil
}

func (f *fakeTagRepository) List(_ context.Context) ([]domain.Tag, error) {
	f.calls++
	if f.listErr != nil {
		return nil, f.listErr
	}

	return f.tags, nil
}

type fakeMediaRepository struct {
	calls     int
	created   domain.Media
	tagIDs    []uuid.UUID
	resolved  []domain.Tag
	media     domain.Media
	createErr error
	getErr    error
}

func (f *fakeMediaRepository) CreateWithTags(
	_ context.Context, media domain.Media, tagIDs []uuid.UUID,
) ([]domain.Tag, error) {
	f.calls++
	f.created = media
	f.tagIDs = tagIDs
	if f.createErr != nil {
		return nil, f.createErr
	}

	return f.resolved, nil
}

func (f *fakeMediaRepository) GetByID(_ context.Context, _ uuid.UUID) (domain.Media, error) {
	f.calls++
	if f.getErr != nil {
		return domain.Media{}, f.getErr
	}

	return f.media, nil
}

type fakeFileStore struct {
	calls     int
	key       string
	content   string
	keyPrefix string
	putErr    error
}

func (f *fakeFileStore) Put(_ context.Context, key string, r io.Reader) (string, error) {
	f.calls++
	f.key = key
	if f.putErr != nil {
		return "", f.putErr
	}

	body, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	f.content = string(body)

	return f.keyPrefix + key, nil
}

func (f *fakeFileStore) URL(key string) string {
	return "https://files.test/" + key
}

func newCreateMedia(name string) CreateMedia {
	return CreateMedia{
		Name: name,
		Type: domain.MediaTypeImage,
		File: strings.NewReader("jpeg bytes"),
	}
}
