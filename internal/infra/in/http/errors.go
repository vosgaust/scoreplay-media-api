package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		validation      domain.ValidationError
		unknownTags     domain.UnknownTagsError
		unsupportedType domain.UnsupportedMediaTypeError
		unsupportedMime unsupportedContentTypeError
		conflict        domain.ConflictError
		tooLarge        *http.MaxBytesError
	)

	switch {
	case errors.As(err, &validation):
		writeErrorStatus(w, r, http.StatusBadRequest, errorBody{
			Code:    "validation_error",
			Message: err.Error(),
			Details: map[string]any{"field": validation.Field},
		})

	case errors.As(err, &unknownTags):
		writeErrorStatus(w, r, http.StatusBadRequest, errorBody{
			Code:    "unknown_tags",
			Message: err.Error(),
			Details: map[string]any{"unknown_tag_ids": unknownTags.IDs},
		})

	case errors.As(err, &unsupportedType), errors.As(err, &unsupportedMime):
		writeErrorStatus(w, r, http.StatusUnsupportedMediaType, errorBody{
			Code: "unsupported_media_type", Message: err.Error(),
		})

	case errors.As(err, &conflict):
		writeErrorStatus(w, r, http.StatusConflict, errorBody{
			Code: "conflict", Message: err.Error(),
		})

	case errors.Is(err, domain.ErrNotFound):
		writeErrorStatus(w, r, http.StatusNotFound, errorBody{
			Code: "not_found", Message: "the resource does not exist",
		})

	case errors.As(err, &tooLarge):
		writeErrorStatus(w, r, http.StatusRequestEntityTooLarge, errorBody{
			Code: "payload_too_large", Message: "the upload exceeds the maximum allowed size",
		})

	default:
		slog.ErrorContext(r.Context(), "request failed", "error", err)
		writeErrorStatus(w, r, http.StatusInternalServerError, errorBody{
			Code: "internal_error", Message: "the request could not be processed",
		})
	}
}
