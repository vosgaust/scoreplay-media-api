package application

import "github.com/vosgaust/scoreplay-media-api/internal/domain"

type MediaView struct {
	Media   domain.Media
	FileURL string
}
