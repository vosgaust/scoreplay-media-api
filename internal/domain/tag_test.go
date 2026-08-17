package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var testTagID = uuid.MustParse("0198b3c4-0000-7000-8000-000000000002")

func TestNewTag(t *testing.T) {
	nameAtLimit := strings.Repeat("é", MaxTagNameLength)
	nameOverLimit := strings.Repeat("é", MaxTagNameLength+1)

	tests := []struct {
		name         string
		tagName      string
		expectedName string
		expectedErr  error
	}{
		{
			name:         "tag is created",
			tagName:      "Messi",
			expectedName: "Messi",
		},
		{
			name:         "name is trimmed",
			tagName:      "  Messi  ",
			expectedName: "Messi",
		},
		{
			name:         "case is preserved",
			tagName:      "FC Barcelona",
			expectedName: "FC Barcelona",
		},
		{
			name:         "name at the rune limit is accepted",
			tagName:      nameAtLimit,
			expectedName: nameAtLimit,
		},
		{
			name:        "empty name is rejected",
			tagName:     "",
			expectedErr: ValidationError{Field: "name", Message: "must not be empty"},
		},
		{
			name:        "whitespace-only name is rejected",
			tagName:     "   \t  ",
			expectedErr: ValidationError{Field: "name", Message: "must not be empty"},
		},
		{
			name:        "name over the rune limit is rejected",
			tagName:     nameOverLimit,
			expectedErr: ValidationError{Field: "name", Message: "must be at most 120 characters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTag(testTagID, tt.tagName, testNow)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("error = %v, want %v", err, tt.expectedErr)
			}
			if tt.expectedErr != nil {
				return
			}

			if got.ID != testTagID {
				t.Errorf("ID = %v, want %v", got.ID, testTagID)
			}
			if got.Name != tt.expectedName {
				t.Errorf("Name = %q, want %q", got.Name, tt.expectedName)
			}
			if !got.CreatedAt.Equal(testNow) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, testNow)
			}
		})
	}
}
