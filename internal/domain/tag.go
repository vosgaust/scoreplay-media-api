package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const MaxTagNameLength = 120

type Tag struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

func NewTag(id uuid.UUID, name string, now time.Time) (Tag, error) {
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		return Tag{}, ValidationError{Field: "name", Message: "must not be empty"}
	case utf8.RuneCountInString(name) > MaxTagNameLength:
		return Tag{}, ValidationError{Field: "name", Message: "must be at most 120 characters"}
	}

	return Tag{
		ID:        id,
		Name:      name,
		CreatedAt: now,
	}, nil
}
