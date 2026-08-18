package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/application"
	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

var errUnexpected = errors.New("pq: relation \"media\" does not exist")

type fakeProbe struct {
	err error
}

func (f fakeProbe) Ping(context.Context) error {
	return f.err
}

type fakeTagCreator struct {
	cmd  application.CreateTag
	tag  domain.Tag
	err  error
	call int
}

func (f *fakeTagCreator) Handle(_ context.Context, cmd application.CreateTag) (domain.Tag, error) {
	f.call++
	f.cmd = cmd

	return f.tag, f.err
}

type fakeTagLister struct {
	tags []domain.Tag
	err  error
}

func (f *fakeTagLister) Handle(_ context.Context) ([]domain.Tag, error) {
	return f.tags, f.err
}

type fakeMediaCreator struct {
	cmd     application.CreateMedia
	content string
	view    application.MediaView
	err     error
	call    int
}

func (f *fakeMediaCreator) Handle(
	_ context.Context, cmd application.CreateMedia,
) (application.MediaView, error) {
	f.call++
	f.cmd = cmd

	if cmd.File != nil {
		body, err := io.ReadAll(cmd.File)
		if err != nil {
			return application.MediaView{}, err
		}
		f.content = string(body)
	}

	return f.view, f.err
}

type fakeMediaFinder struct {
	id   uuid.UUID
	view application.MediaView
	err  error
}

func (f *fakeMediaFinder) Handle(_ context.Context, id uuid.UUID) (application.MediaView, error) {
	f.id = id

	return f.view, f.err
}

func jpegBytes() []byte {
	return append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x11}, 64)...)
}

func pdfBytes() []byte {
	return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{0x22}, 64)...)
}

type uploadPart struct {
	name    string
	content []byte
}

func buildUpload(t *testing.T, name string, tagIDs []string, file *uploadPart) (string, io.Reader) {
	t.Helper()

	var body bytes.Buffer
	form := multipart.NewWriter(&body)

	if name != "" {
		if err := form.WriteField("name", name); err != nil {
			t.Fatalf("write name field: %v", err)
		}
	}
	for _, id := range tagIDs {
		if err := form.WriteField("tags", id); err != nil {
			t.Fatalf("write tags field: %v", err)
		}
	}
	if file != nil {
		part, err := form.CreateFormFile("file", file.name)
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		if _, err := part.Write(file.content); err != nil {
			t.Fatalf("write file part: %v", err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return form.FormDataContentType(), &body
}
