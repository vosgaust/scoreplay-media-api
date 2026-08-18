package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
)

func TestTagHandlerCreate(t *testing.T) {
	tag := domain.Tag{ID: uuid.MustParse("0198b3c4-0000-7000-8000-000000000001"), Name: "Messi"}
	creator := &fakeTagCreator{tag: tag}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(`{"name":"Messi"}`))

	NewTagHandler(creator, &fakeTagLister{}).Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (%s)", response.Code, http.StatusCreated, response.Body)
	}
	if creator.cmd.Name != "Messi" {
		t.Errorf("use case received name %q, want %q", creator.cmd.Name, "Messi")
	}

	var body tagResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body != (tagResponse{ID: tag.ID, Name: tag.Name}) {
		t.Errorf("body = %+v, want %+v", body, tag)
	}
}

func TestTagHandlerCreateErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		useCaseErr error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{
			name:       "not json",
			body:       `not json`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_body",
		},
		{
			name:       "an unknown field is rejected rather than ignored",
			body:       `{"name":"Messi","colour":"blue"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_body",
		},
		{
			name:       "an empty name is the domain's verdict, not the transport's",
			body:       `{"name":""}`,
			useCaseErr: domain.ValidationError{Field: "name", Message: "must not be empty"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
			wantCalls:  1,
		},
		{
			name:       "a duplicate name is a conflict",
			body:       `{"name":"messi"}`,
			useCaseErr: domain.ConflictError{Resource: "tag", Field: "name"},
			wantStatus: http.StatusConflict,
			wantCode:   "conflict",
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &fakeTagCreator{err: tt.useCaseErr}

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(tt.body))

			NewTagHandler(creator, &fakeTagLister{}).Create(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", response.Code, tt.wantStatus, response.Body)
			}
			if got := errorCodeOf(t, response); got != tt.wantCode {
				t.Errorf("error code = %q, want %q", got, tt.wantCode)
			}
			if creator.call != tt.wantCalls {
				t.Errorf("use case calls = %d, want %d", creator.call, tt.wantCalls)
			}
		})
	}
}

func TestTagHandlerList(t *testing.T) {
	tags := []domain.Tag{
		{ID: uuid.MustParse("0198b3c4-0000-7000-8000-000000000001"), Name: "Messi"},
		{ID: uuid.MustParse("0198b3c4-0000-7000-8000-000000000002"), Name: "Mbappé"},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/tags", nil)

	NewTagHandler(&fakeTagCreator{}, &fakeTagLister{tags: tags}).List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body []tagResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 2 || body[0].Name != "Messi" || body[1].Name != "Mbappé" {
		t.Errorf("body = %+v, want the repository's order preserved", body)
	}
}

func TestTagHandlerListEmpty(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/tags", nil)

	NewTagHandler(&fakeTagCreator{}, &fakeTagLister{}).List(response, request)

	if got := strings.TrimSpace(response.Body.String()); got != "[]" {
		t.Errorf("body = %s, want []", got)
	}
}

func errorCodeOf(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()

	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}

	return envelope.Error.Code
}
