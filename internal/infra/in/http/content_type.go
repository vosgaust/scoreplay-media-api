package http

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

const sniffLen = 512

type unsupportedContentTypeError struct {
	contentType string
}

func (e unsupportedContentTypeError) Error() string {
	return fmt.Sprintf("unsupported content type %q: only images and videos are accepted", e.contentType)
}

func classify(file io.Reader) (io.Reader, domain.MediaType, error) {
	buffered := bufio.NewReaderSize(file, sniffLen)

	head, err := buffered.Peek(sniffLen)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, "", fmt.Errorf("read upload: %w", err)
	}

	contentType := http.DetectContentType(head)

	switch {
	case strings.HasPrefix(contentType, "image/"):
		return buffered, domain.MediaTypeImage, nil
	case strings.HasPrefix(contentType, "video/"):
		return buffered, domain.MediaTypeVideo, nil
	default:
		return nil, "", unsupportedContentTypeError{contentType: contentType}
	}
}
