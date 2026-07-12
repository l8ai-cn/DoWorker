package market

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSpaceRequiresValidSlugAndPublishesOnce(t *testing.T) {
	_, err := NewSpace(42, "Invalid_Space", "运营", "运营应用", 14)
	require.Error(t, err)

	space, err := NewSpace(42, "operations", "运营", "运营应用", 14)
	require.NoError(t, err)
	publishedAt := time.Now().UTC()
	require.NoError(t, space.Publish(publishedAt))
	require.Equal(t, SpaceStatusPublished, space.Status())
	require.Equal(t, publishedAt, *space.PublishedAt())
	require.ErrorIs(t, space.Publish(publishedAt), ErrInvalidSpaceTransition)
}
