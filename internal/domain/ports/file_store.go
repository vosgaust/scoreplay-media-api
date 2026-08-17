package ports

import (
	"context"
	"io"
)

type FileStore interface {
	Put(ctx context.Context, key string, r io.Reader) (string, error)
	URL(key string) string
}
