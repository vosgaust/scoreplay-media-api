package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
}

type ConflictError struct {
	Resource string
	Field    string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("%s with this %s already exists", e.Resource, e.Field)
}

type UnknownTagsError struct {
	IDs []uuid.UUID
}

func (e UnknownTagsError) Error() string {
	ids := make([]string, len(e.IDs))
	for i, id := range e.IDs {
		ids[i] = id.String()
	}
	return "unknown tag ids: " + strings.Join(ids, ", ")
}

type UnsupportedMediaTypeError struct {
	MediaType string
}

func (e UnsupportedMediaTypeError) Error() string {
	return fmt.Sprintf("unsupported media type %q", e.MediaType)
}
