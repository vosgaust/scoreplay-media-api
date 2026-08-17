//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vosgaust/scoreplay-media-api/internal/domain"
	"github.com/vosgaust/scoreplay-media-api/internal/domain/ports"
)

var _ ports.TagRepository = TagRepository{}

func TestTagRepositoryCreateRoundTrip(t *testing.T) {
	resetDB(t)

	repo := NewTagRepository(testPool)
	want := newTestTag(t, "Messi")

	require.NoError(t, repo.Create(t.Context(), want))

	got, err := repo.List(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, want.ID, got[0].ID)
	assert.Equal(t, want.Name, got[0].Name)
	assert.WithinDuration(t, want.CreatedAt, got[0].CreatedAt, 0)
}

func TestTagRepositoryListIsEmptyNotNil(t *testing.T) {
	resetDB(t)

	got, err := NewTagRepository(testPool).List(t.Context())

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestTagRepositoryCreateRejectsDuplicateName(t *testing.T) {
	resetDB(t)

	repo := NewTagRepository(testPool)
	require.NoError(t, repo.Create(t.Context(), newTestTag(t, "Messi")))

	err := repo.Create(t.Context(), newTestTag(t, "messi"))

	var conflict domain.ConflictError
	require.ErrorAs(t, err, &conflict)
	assert.Equal(t, "tag", conflict.Resource)
	assert.Equal(t, "name", conflict.Field)
}

func TestTagRepositoryListOrdersCaseInsensitively(t *testing.T) {
	resetDB(t)

	repo := NewTagRepository(testPool)

	for _, name := range []string{"Zidane", "mbappé", "Messi"} {
		require.NoError(t, repo.Create(t.Context(), newTestTag(t, name)))
	}

	tags, err := repo.List(t.Context())
	require.NoError(t, err)

	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}

	assert.Equal(t, []string{"mbappé", "Messi", "Zidane"}, names)
}

func newTestTag(t *testing.T, name string) domain.Tag {
	t.Helper()

	tag, err := domain.NewTag(uuid.Must(uuid.NewV7()), name, time.Now().UTC().Truncate(time.Microsecond))
	require.NoError(t, err)

	return tag
}
