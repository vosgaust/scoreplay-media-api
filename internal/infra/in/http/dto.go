package http

import (
	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/application"
	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

type createTagRequest struct {
	Name string `json:"name"`
}

type tagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type mediaResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Tags    []string  `json:"tags"`
	FileURL string    `json:"fileUrl"`
}

func newTagResponse(tag domain.Tag) tagResponse {
	return tagResponse{ID: tag.ID, Name: tag.Name}
}

func newTagResponses(tags []domain.Tag) []tagResponse {
	responses := make([]tagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, newTagResponse(tag))
	}

	return responses
}

func newMediaResponse(view application.MediaView) mediaResponse {
	names := make([]string, 0, len(view.Media.Tags))
	for _, tag := range view.Media.Tags {
		names = append(names, tag.Name)
	}

	return mediaResponse{
		ID:      view.Media.ID,
		Name:    view.Media.Name,
		Tags:    names,
		FileURL: view.FileURL,
	}
}
