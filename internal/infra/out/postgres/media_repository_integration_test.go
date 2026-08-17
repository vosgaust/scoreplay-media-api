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

var _ ports.MediaRepository = MediaRepository{}

func TestMediaRepositoryCreateWithTags(t *testing.T) {
	resetDB(t)

	tags := NewTagRepository(testPool)
	media := NewMediaRepository(testPool)

	zidane := newTestTag(t, "Zidane")
	mbappe := newTestTag(t, "mbappé")
	require.NoError(t, tags.Create(t.Context(), zidane))
	require.NoError(t, tags.Create(t.Context(), mbappe))

	want := newTestMedia(t, "Final highlights", domain.MediaTypeVideo)

	resolved, err := media.CreateWithTags(t.Context(), want, []uuid.UUID{zidane.ID, mbappe.ID})
	require.NoError(t, err)

	require.Len(t, resolved, 2)
	assert.Equal(t, []string{"mbappé", "Zidane"}, []string{resolved[0].Name, resolved[1].Name})
	assert.Equal(t, mbappe.ID, resolved[0].ID)

	assert.Equal(t, 1, countRows(t, "media"))
	assert.Equal(t, 2, countRows(t, "media_tags"))
}

func TestMediaRepositoryCreateWithoutTags(t *testing.T) {
	resetDB(t)

	want := newTestMedia(t, "Untagged shot", domain.MediaTypeImage)

	resolved, err := NewMediaRepository(testPool).CreateWithTags(t.Context(), want, nil)
	require.NoError(t, err)

	assert.NotNil(t, resolved)
	assert.Empty(t, resolved)
	assert.Equal(t, 1, countRows(t, "media"))
	assert.Equal(t, 0, countRows(t, "media_tags"))
}

func TestMediaRepositoryCreateWithUnknownTagWritesNothing(t *testing.T) {
	resetDB(t)

	tags := NewTagRepository(testPool)
	known := newTestTag(t, "Messi")
	require.NoError(t, tags.Create(t.Context(), known))

	unknown := uuid.Must(uuid.NewV7())

	_, err := NewMediaRepository(testPool).CreateWithTags(
		t.Context(),
		newTestMedia(t, "Free kick", domain.MediaTypeImage),
		[]uuid.UUID{known.ID, unknown},
	)

	var unknownTags domain.UnknownTagsError
	require.ErrorAs(t, err, &unknownTags)
	assert.Equal(t, []uuid.UUID{unknown}, unknownTags.IDs)

	assert.Equal(t, 0, countRows(t, "media"))
	assert.Equal(t, 0, countRows(t, "media_tags"))
}

func TestMediaRepositoryCreateCollapsesDuplicateTagIDs(t *testing.T) {
	resetDB(t)

	tags := NewTagRepository(testPool)
	messi := newTestTag(t, "Messi")
	require.NoError(t, tags.Create(t.Context(), messi))

	resolved, err := NewMediaRepository(testPool).CreateWithTags(
		t.Context(),
		newTestMedia(t, "Free kick", domain.MediaTypeImage),
		[]uuid.UUID{messi.ID, messi.ID},
	)
	require.NoError(t, err)

	assert.Len(t, resolved, 1)
	assert.Equal(t, 1, countRows(t, "media_tags"))
}

func TestMediaRepositoryGetByIDWithTags(t *testing.T) {
	resetDB(t)

	tags := NewTagRepository(testPool)
	media := NewMediaRepository(testPool)

	zidane := newTestTag(t, "Zidane")
	mbappe := newTestTag(t, "mbappé")
	require.NoError(t, tags.Create(t.Context(), zidane))
	require.NoError(t, tags.Create(t.Context(), mbappe))

	want := newTestMedia(t, "Final highlights", domain.MediaTypeVideo)
	_, err := media.CreateWithTags(t.Context(), want, []uuid.UUID{zidane.ID, mbappe.ID})
	require.NoError(t, err)

	got, err := media.GetByID(t.Context(), want.ID)
	require.NoError(t, err)

	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.StorageKey, got.StorageKey)
	assert.Equal(t, domain.MediaTypeVideo, got.Type)
	assert.WithinDuration(t, want.CreatedAt, got.CreatedAt, 0)

	// Whole tags, not name-only fragments: id and created_at come back too, so domain.Tag means
	// the same thing here as it does from CreateWithTags and from TagRepository.List.
	require.Len(t, got.Tags, 2)
	assert.Equal(t, []string{"mbappé", "Zidane"}, []string{got.Tags[0].Name, got.Tags[1].Name})
	assert.Equal(t, mbappe.ID, got.Tags[0].ID)
	assert.Equal(t, zidane.ID, got.Tags[1].ID)
	assert.WithinDuration(t, mbappe.CreatedAt, got.Tags[0].CreatedAt, 0)
}

func TestMediaRepositoryGetByIDWithoutTags(t *testing.T) {
	resetDB(t)

	media := NewMediaRepository(testPool)
	want := newTestMedia(t, "Untagged shot", domain.MediaTypeImage)
	_, err := media.CreateWithTags(t.Context(), want, nil)
	require.NoError(t, err)

	got, err := media.GetByID(t.Context(), want.ID)
	require.NoError(t, err)

	assert.NotNil(t, got.Tags)
	assert.Empty(t, got.Tags)
}

func TestMediaRepositoryGetByIDMissing(t *testing.T) {
	resetDB(t)

	_, err := NewMediaRepository(testPool).GetByID(t.Context(), uuid.Must(uuid.NewV7()))

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMediaRepositoryGetByIDRejectsAnUnreadableType(t *testing.T) {
	resetDB(t)

	media := NewMediaRepository(testPool)
	stored := newTestMedia(t, "Corrupt row", domain.MediaTypeImage)
	_, err := media.CreateWithTags(t.Context(), stored, nil)
	require.NoError(t, err)

	_, err = testPool.Exec(t.Context(), `UPDATE media SET type = 'document' WHERE id = $1`, stored.ID)
	require.NoError(t, err)

	_, err = media.GetByID(t.Context(), stored.ID)

	var unsupported domain.UnsupportedMediaTypeError
	require.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "document", unsupported.MediaType)
}

func newTestMedia(t *testing.T, name string, mediaType domain.MediaType) domain.Media {
	t.Helper()

	id := uuid.Must(uuid.NewV7())

	extension := ".jpg"
	if mediaType == domain.MediaTypeVideo {
		extension = ".mp4"
	}

	media, err := domain.NewMedia(
		id,
		name,
		"media/2026/08/"+id.String()+extension,
		mediaType,
		time.Now().UTC().Truncate(time.Microsecond),
	)
	require.NoError(t, err)

	return media
}

func countRows(t *testing.T, table string) int {
	t.Helper()

	var count int
	err := testPool.QueryRow(t.Context(), "SELECT count(*) FROM "+table).Scan(&count)
	require.NoError(t, err)

	return count
}
