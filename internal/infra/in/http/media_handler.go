package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/application"
	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

const multipartMemory = 10 << 20

type MediaHandler struct {
	create         mediaCreator
	find           mediaFinder
	maxUploadBytes int64
}

func NewMediaHandler(create mediaCreator, find mediaFinder, maxUploadBytes int64) MediaHandler {
	return MediaHandler{create: create, find: find, maxUploadBytes: maxUploadBytes}
}

func (h MediaHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	if err := r.ParseMultipartForm(multipartMemory); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, r, err)

			return
		}

		writeErrorStatus(w, r, http.StatusBadRequest, errorBody{
			Code:    "invalid_body",
			Message: "the request must be multipart/form-data with a name and a file",
		})

		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	file, _, err := r.FormFile("file")
	if err != nil {
		writeErrorStatus(w, r, http.StatusBadRequest, errorBody{
			Code: "missing_file", Message: "a file part is required",
		})

		return
	}
	defer func() { _ = file.Close() }()

	tagIDs, err := parseTagIDs(r.MultipartForm.Value["tags"])
	if err != nil {
		writeError(w, r, err)

		return
	}

	body, mediaType, err := classify(file)
	if err != nil {
		writeError(w, r, err)

		return
	}

	view, err := h.create.Handle(r.Context(), application.CreateMedia{
		Name:   r.FormValue("name"),
		TagIDs: tagIDs,
		Type:   mediaType,
		File:   body,
	})
	if err != nil {
		writeError(w, r, err)

		return
	}

	w.Header().Set("Location", "/media/"+view.Media.ID.String())
	writeJSON(w, r, http.StatusCreated, newMediaResponse(view))
}

func (h MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, domain.ValidationError{Field: "id", Message: "must be a UUID"})

		return
	}

	view, err := h.find.Handle(r.Context(), id)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, newMediaResponse(view))
}

func parseTagIDs(values []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(values))

	for _, value := range values {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, domain.ValidationError{Field: "tags", Message: "must contain only UUIDs"}
		}
		ids = append(ids, id)
	}

	return ids, nil
}
