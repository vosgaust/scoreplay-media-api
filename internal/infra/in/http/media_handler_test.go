package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/application"
	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

const maxUpload = 5 << 20

var testMediaID = uuid.MustParse("0198b3c4-0000-7000-8000-0000000000aa")

func testView() application.MediaView {
	return application.MediaView{
		Media: domain.Media{
			ID:   testMediaID,
			Name: "Messi free kick",
			Tags: []domain.Tag{{Name: "Messi"}},
		},
		FileURL: "http://localhost:8080/files/" + testMediaID.String(),
	}
}

func TestMediaHandlerCreate(t *testing.T) {
	creator := &fakeMediaCreator{view: testView()}
	tagID := "0198b3c4-0000-7000-8000-000000000001"

	contentType, body := buildUpload(t, "Messi free kick", []string{tagID},
		&uploadPart{name: "messi.jpg", content: jpegBytes()})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/media", body)
	request.Header.Set("Content-Type", contentType)

	NewMediaHandler(creator, &fakeMediaFinder{}, maxUpload).Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (%s)", response.Code, http.StatusCreated, response.Body)
	}
	if creator.cmd.Name != "Messi free kick" {
		t.Errorf("use case received name %q, want %q", creator.cmd.Name, "Messi free kick")
	}
	if creator.cmd.Type != domain.MediaTypeImage {
		t.Errorf("use case received type %q, want %q", creator.cmd.Type, domain.MediaTypeImage)
	}
	if len(creator.cmd.TagIDs) != 1 || creator.cmd.TagIDs[0].String() != tagID {
		t.Errorf("use case received tag ids %v, want [%s]", creator.cmd.TagIDs, tagID)
	}
	if creator.content != string(jpegBytes()) {
		t.Errorf("use case read %d bytes, want the whole file (%d)", len(creator.content), len(jpegBytes()))
	}
	if got := response.Header().Get("Location"); got != "/media/"+testMediaID.String() {
		t.Errorf("Location = %q, want /media/%s", got, testMediaID)
	}

	var responseBody mediaResponse
	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(responseBody.Tags) != 1 || responseBody.Tags[0] != "Messi" {
		t.Errorf("tags = %v, want [Messi]", responseBody.Tags)
	}
	if responseBody.FileURL != testView().FileURL {
		t.Errorf("fileUrl = %q, want %q", responseBody.FileURL, testView().FileURL)
	}
}

func TestMediaHandlerCreateResponseFieldNames(t *testing.T) {
	creator := &fakeMediaCreator{view: testView()}

	contentType, body := buildUpload(t, "Messi free kick", nil,
		&uploadPart{name: "messi.jpg", content: jpegBytes()})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/media", body)
	request.Header.Set("Content-Type", contentType)

	NewMediaHandler(creator, &fakeMediaFinder{}, maxUpload).Create(response, request)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	for _, field := range []string{"id", "name", "tags", "fileUrl"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("response has no %q field: %s", field, response.Body)
		}
	}
	if len(raw) != 4 {
		t.Errorf("response has %d fields, want exactly 4: %s", len(raw), response.Body)
	}
}

func TestMediaHandlerCreateErrors(t *testing.T) {
	tests := []struct {
		name       string
		mediaName  string
		tagIDs     []string
		file       *uploadPart
		maxUpload  int64
		useCaseErr error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{
			name:       "no file part",
			mediaName:  "Messi free kick",
			wantStatus: http.StatusBadRequest,
			wantCode:   "missing_file",
		},
		{
			name:       "a pdf is neither a photo nor a video",
			mediaName:  "a document",
			file:       &uploadPart{name: "contract.pdf", content: pdfBytes()},
			wantStatus: http.StatusUnsupportedMediaType,
			wantCode:   "unsupported_media_type",
		},
		{
			name:       "a lying extension does not fool the sniffer",
			mediaName:  "a document",
			file:       &uploadPart{name: "trustme.jpg", content: pdfBytes()},
			wantStatus: http.StatusUnsupportedMediaType,
			wantCode:   "unsupported_media_type",
		},
		{
			name:       "a tag id that is not a uuid",
			mediaName:  "Messi free kick",
			tagIDs:     []string{"messi"},
			file:       &uploadPart{name: "messi.jpg", content: jpegBytes()},
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
		},
		{
			name:       "a body over the limit",
			mediaName:  "Messi free kick",
			file:       &uploadPart{name: "messi.jpg", content: jpegBytes()},
			maxUpload:  8,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantCode:   "payload_too_large",
		},
		{
			name:       "an unknown tag comes back from the use case",
			mediaName:  "Messi free kick",
			tagIDs:     []string{"0198b3c4-0000-7000-8000-000000000009"},
			file:       &uploadPart{name: "messi.jpg", content: jpegBytes()},
			useCaseErr: domain.UnknownTagsError{IDs: []uuid.UUID{uuid.MustParse("0198b3c4-0000-7000-8000-000000000009")}},
			wantStatus: http.StatusBadRequest,
			wantCode:   "unknown_tags",
			wantCalls:  1,
		},
		{
			name:       "an unexpected failure never leaks its message",
			mediaName:  "Messi free kick",
			file:       &uploadPart{name: "messi.jpg", content: jpegBytes()},
			useCaseErr: errUnexpected,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &fakeMediaCreator{err: tt.useCaseErr}
			limit := int64(maxUpload)
			if tt.maxUpload != 0 {
				limit = tt.maxUpload
			}

			contentType, body := buildUpload(t, tt.mediaName, tt.tagIDs, tt.file)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/media", body)
			request.Header.Set("Content-Type", contentType)

			NewMediaHandler(creator, &fakeMediaFinder{}, limit).Create(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", response.Code, tt.wantStatus, response.Body)
			}
			if got := errorCodeOf(t, response); got != tt.wantCode {
				t.Errorf("error code = %q, want %q", got, tt.wantCode)
			}
			if creator.call != tt.wantCalls {
				t.Errorf("use case calls = %d, want %d", creator.call, tt.wantCalls)
			}
			if tt.wantStatus == http.StatusInternalServerError &&
				strings.Contains(response.Body.String(), errUnexpected.Error()) {
				t.Errorf("body leaks the internal error: %s", response.Body)
			}
		})
	}
}

func TestMediaHandlerGet(t *testing.T) {
	finder := &fakeMediaFinder{view: testView()}

	response := serveGetMedia(t, finder, testMediaID.String())

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", response.Code, http.StatusOK, response.Body)
	}
	if finder.id != testMediaID {
		t.Errorf("use case received id %s, want %s", finder.id, testMediaID)
	}

	var body mediaResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.ID != testMediaID || body.FileURL != testView().FileURL {
		t.Errorf("body = %+v, want the view's id and fileUrl", body)
	}
}

func TestMediaHandlerGetErrors(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		useCaseErr error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "not a uuid",
			id:         "messi",
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
		},
		{
			name:       "unknown id",
			id:         testMediaID.String(),
			useCaseErr: domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := serveGetMedia(t, &fakeMediaFinder{err: tt.useCaseErr}, tt.id)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", response.Code, tt.wantStatus, response.Body)
			}
			if got := errorCodeOf(t, response); got != tt.wantCode {
				t.Errorf("error code = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

func serveGetMedia(t *testing.T, finder mediaFinder, id string) *httptest.ResponseRecorder {
	t.Helper()

	router := NewRouter(
		fakeProbe{},
		NewTagHandler(&fakeTagCreator{}, &fakeTagLister{}),
		NewMediaHandler(&fakeMediaCreator{}, finder, maxUpload),
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/media/"+id, nil))

	return response
}
