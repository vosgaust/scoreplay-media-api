package application

import (
	"context"
	"errors"
	"testing"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

func TestCreateTagHandlerHandle(t *testing.T) {
	tags := &fakeTagRepository{}

	tag, err := NewCreateTagHandler(tags).Handle(context.Background(), CreateTag{Name: "Messi"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if tag.ID.Version() != 7 {
		t.Errorf("ID version = %d, want 7", tag.ID.Version())
	}
	if tag.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, want the creation instant")
	}
	if len(tags.created) != 1 || tags.created[0] != tag {
		t.Errorf("repository received %v, want exactly [%v]", tags.created, tag)
	}
}

func TestCreateTagHandlerHandleErrors(t *testing.T) {
	conflict := domain.ConflictError{Resource: "tag", Field: "name"}

	tests := []struct {
		name      string
		tagName   string
		createErr error
		wantErr   error
		wantCalls int
	}{
		{
			name:    "an invalid name never reaches the repository",
			tagName: "  ",
			wantErr: domain.ValidationError{Field: "name", Message: "must not be empty"},
		},
		{
			name:      "a conflict is passed through untranslated",
			tagName:   "Messi",
			createErr: conflict,
			wantErr:   conflict,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := &fakeTagRepository{createErr: tt.createErr}

			_, err := NewCreateTagHandler(tags).Handle(context.Background(), CreateTag{Name: tt.tagName})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Handle() error = %v, want %v", err, tt.wantErr)
			}
			if tags.calls != tt.wantCalls {
				t.Errorf("repository calls = %d, want %d", tags.calls, tt.wantCalls)
			}
		})
	}
}
