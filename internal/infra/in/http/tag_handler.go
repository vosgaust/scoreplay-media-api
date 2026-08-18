package http

import (
	"encoding/json"
	"net/http"

	"github.com/vosgaust/scoreplay-media-api/internal/application"
)

const maxJSONBody = 1 << 20

type TagHandler struct {
	create tagCreator
	list   tagLister
}

func NewTagHandler(create tagCreator, list tagLister) TagHandler {
	return TagHandler{create: create, list: list}
}

func (h TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody))
	decoder.DisallowUnknownFields()

	var body createTagRequest
	if err := decoder.Decode(&body); err != nil {
		writeErrorStatus(w, r, http.StatusBadRequest, errorBody{
			Code:    "invalid_body",
			Message: "the request body must be a JSON object with a name",
		})

		return
	}

	tag, err := h.create.Handle(r.Context(), application.CreateTag{Name: body.Name})
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusCreated, newTagResponse(tag))
}

func (h TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.list.Handle(r.Context())
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, newTagResponses(tags))
}
