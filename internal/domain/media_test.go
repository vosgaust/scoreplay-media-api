package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	testMediaID = uuid.MustParse("0198b3c4-0000-7000-8000-000000000001")
	testNow     = time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
)

func TestNewMedia(t *testing.T) {
	nameAtLimit := strings.Repeat("é", MaxMediaNameLength)
	nameOverLimit := strings.Repeat("é", MaxMediaNameLength+1)

	tests := []struct {
		name         string
		mediaName    string
		storageKey   string
		mediaType    MediaType
		expectedName string
		expectedErr  error
	}{
		{
			name:         "image is created",
			mediaName:    "Messi free kick",
			storageKey:   "media/2026/08/abc.jpg",
			mediaType:    MediaTypeImage,
			expectedName: "Messi free kick",
		},
		{
			name:         "video is created",
			mediaName:    "Second half highlights",
			storageKey:   "media/2026/08/abc.mp4",
			mediaType:    MediaTypeVideo,
			expectedName: "Second half highlights",
		},
		{
			name:         "name is trimmed",
			mediaName:    "  Messi free kick  ",
			storageKey:   "media/2026/08/abc.jpg",
			mediaType:    MediaTypeImage,
			expectedName: "Messi free kick",
		},
		{
			name:         "name at the rune limit is accepted",
			mediaName:    nameAtLimit,
			storageKey:   "media/2026/08/abc.jpg",
			mediaType:    MediaTypeImage,
			expectedName: nameAtLimit,
		},
		{
			name:        "empty name is rejected",
			storageKey:  "media/2026/08/abc.jpg",
			mediaType:   MediaTypeImage,
			expectedErr: ValidationError{Field: "name", Message: "must not be empty"},
		},
		{
			name:        "whitespace-only name is rejected",
			mediaName:   "   \t  ",
			storageKey:  "media/2026/08/abc.jpg",
			mediaType:   MediaTypeImage,
			expectedErr: ValidationError{Field: "name", Message: "must not be empty"},
		},
		{
			name:        "name over the rune limit is rejected",
			mediaName:   nameOverLimit,
			storageKey:  "media/2026/08/abc.jpg",
			mediaType:   MediaTypeImage,
			expectedErr: ValidationError{Field: "name", Message: "must be at most 200 characters"},
		},
		{
			name:        "empty storage key is rejected",
			mediaName:   "Messi free kick",
			mediaType:   MediaTypeImage,
			expectedErr: ValidationError{Field: "storageKey", Message: "must not be empty"},
		},
		{
			name:        "incorrect media type is rejected",
			mediaName:   "Messi free kick",
			storageKey:  "media/2026/08/abc.pdf",
			mediaType:   MediaType("pdf"),
			expectedErr: UnsupportedMediaTypeError{MediaType: "pdf"},
		},
		{
			name:        "zero media type is rejected",
			mediaName:   "Messi free kick",
			storageKey:  "media/2026/08/abc.jpg",
			expectedErr: UnsupportedMediaTypeError{MediaType: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMedia(testMediaID, tt.mediaName, tt.storageKey, tt.mediaType, testNow)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("error = %v, want %v", err, tt.expectedErr)
			}
			if tt.expectedErr != nil {
				return
			}

			if got.ID != testMediaID {
				t.Errorf("ID = %v, want %v", got.ID, testMediaID)
			}
			if got.Name != tt.expectedName {
				t.Errorf("Name = %q, want %q", got.Name, tt.expectedName)
			}
			if got.StorageKey != tt.storageKey {
				t.Errorf("StorageKey = %q, want %q", got.StorageKey, tt.storageKey)
			}
			if got.Type != tt.mediaType {
				t.Errorf("Type = %q, want %q", got.Type, tt.mediaType)
			}
			if !got.CreatedAt.Equal(testNow) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, testNow)
			}
			if got.Tags != nil {
				t.Errorf("Tags = %v, want nil (resolved after persistence)", got.Tags)
			}
		})
	}
}

func TestParseMediaType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MediaType
		wantErr error
	}{
		{name: "image", input: "image", want: MediaTypeImage},
		{name: "video", input: "video", want: MediaTypeVideo},
		{
			name:    "a mime type is not the domain vocabulary",
			input:   "image/jpeg",
			wantErr: UnsupportedMediaTypeError{MediaType: "image/jpeg"},
		},
		{
			name:    "case matters",
			input:   "Image",
			wantErr: UnsupportedMediaTypeError{MediaType: "Image"},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: UnsupportedMediaTypeError{MediaType: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMediaType(tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseMediaType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
