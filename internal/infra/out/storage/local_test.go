package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

const testKey = "0198b3c4-0000-7000-8000-000000000001.jpg"

func newTestStore(t *testing.T) (LocalFileStore, string) {
	t.Helper()

	parent := t.TempDir()

	return NewLocalFileStore(filepath.Join(parent, "media"), "http://localhost:8080"), parent
}

func TestLocalFileStorePut(t *testing.T) {
	store, parent := newTestStore(t)

	key, err := store.Put(context.Background(), testKey, strings.NewReader("an image"))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if key != testKey {
		t.Errorf("Put() key = %q, want %q", key, testKey)
	}

	content, err := os.ReadFile(filepath.Join(parent, "media", testKey))
	if err != nil {
		t.Fatalf("read stored object: %v", err)
	}
	if string(content) != "an image" {
		t.Errorf("stored content = %q, want %q", content, "an image")
	}
}

func TestLocalFileStorePutInvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "empty", key: ""},
		{name: "nested path", key: "media/a.jpg"},
		{name: "absolute", key: "/etc/passwd"},
		{name: "traversal", key: "../escaped.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, parent := newTestStore(t)

			key, err := store.Put(context.Background(), tt.key, strings.NewReader("payload"))

			if !errors.Is(err, ErrInvalidKey) {
				t.Fatalf("Put(%q) error = %v, want ErrInvalidKey", tt.key, err)
			}
			if key != "" {
				t.Errorf("Put(%q) key = %q, want empty", tt.key, key)
			}
			if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
				t.Errorf("parent contains %v (err %v), want empty", entries, err)
			}
		})
	}
}

func TestLocalFileStorePutInterruptedUploadLeavesPartialFile(t *testing.T) {
	store, parent := newTestStore(t)
	wantErr := errors.New("connection reset")
	body := io.MultiReader(strings.NewReader("half an image"), iotest.ErrReader(wantErr))

	if _, err := store.Put(context.Background(), testKey, body); !errors.Is(err, wantErr) {
		t.Fatalf("Put() error = %v, want %v", err, wantErr)
	}

	content, err := os.ReadFile(filepath.Join(parent, "media", testKey))
	if err != nil {
		t.Fatalf("read partial object: %v", err)
	}
	if string(content) != "half an image" {
		t.Errorf("partial content = %q, want %q", content, "half an image")
	}
}

func TestLocalFileStoreURL(t *testing.T) {
	want := "http://localhost:8080/files/" + testKey

	if got := NewLocalFileStore("", "http://localhost:8080").URL(testKey); got != want {
		t.Errorf("URL() = %q, want %q", got, want)
	}

	if got := NewLocalFileStore("", "http://localhost:8080/").URL(testKey); got != want {
		t.Errorf("URL() with a trailing slash on the base = %q, want %q", got, want)
	}
}
