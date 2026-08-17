package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const MaxMediaNameLength = 200

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

func (t MediaType) Valid() bool {
	switch t {
	case MediaTypeImage, MediaTypeVideo:
		return true
	default:
		return false
	}
}

func ParseMediaType(s string) (MediaType, error) {
	t := MediaType(s)
	if !t.Valid() {
		return "", UnsupportedMediaTypeError{MediaType: s}
	}
	return t, nil
}

type Media struct {
	ID         uuid.UUID
	Name       string
	StorageKey string
	Type       MediaType
	CreatedAt  time.Time

	Tags []Tag
}

func NewMedia(
	id uuid.UUID,
	name string,
	storageKey string,
	mediaType MediaType,
	now time.Time,
) (Media, error) {
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		return Media{}, ValidationError{Field: "name", Message: "must not be empty"}
	case utf8.RuneCountInString(name) > MaxMediaNameLength:
		return Media{}, ValidationError{Field: "name", Message: "must be at most 200 characters"}
	case strings.TrimSpace(storageKey) == "":
		return Media{}, ValidationError{Field: "storageKey", Message: "must not be empty"}
	case !mediaType.Valid():
		return Media{}, UnsupportedMediaTypeError{MediaType: string(mediaType)}
	}

	return Media{
		ID:         id,
		Name:       name,
		StorageKey: storageKey,
		Type:       mediaType,
		CreatedAt:  now,
		Tags:       nil,
	}, nil
}
