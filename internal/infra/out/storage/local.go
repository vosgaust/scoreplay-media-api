// Package storage implements the FileStore port on the local filesystem.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirPerm  = 0o750
	filePerm = 0o640
)

var ErrInvalidKey = errors.New("invalid storage key")

type LocalFileStore struct {
	root    string
	baseURL string
}

func NewLocalFileStore(root, baseURL string) LocalFileStore {
	return LocalFileStore{
		root:    root,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (s LocalFileStore) Put(_ context.Context, key string, r io.Reader) (string, error) {
	if key == "" || key != filepath.Base(key) || strings.Contains(key, "..") {
		return "", fmt.Errorf("%w: %q", ErrInvalidKey, key)
	}

	if err := os.MkdirAll(s.root, dirPerm); err != nil {
		return "", fmt.Errorf("create storage root: %w", err)
	}

	file, err := os.OpenFile(filepath.Join(s.root, key), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerm)
	if err != nil {
		return "", fmt.Errorf("create %q: %w", key, err)
	}
	defer func() { _ = file.Close() }()

	if _, err := io.Copy(file, r); err != nil {
		return "", fmt.Errorf("write %q: %w", key, err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close %q: %w", key, err)
	}

	return key, nil
}

func (s LocalFileStore) URL(key string) string {
	return s.baseURL + "/files/" + key
}
