package listing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublishRequiresApprovedVersionAndPrimarySpace(t *testing.T) {
	item, err := New(42, 108, "product-listing-optimizer")
	require.NoError(t, err)
	require.NoError(t, item.Approve())
	require.NoError(t, item.SetVisibility(VisibilityPublic))

	err = item.Publish(0, true, time.Now())
	require.ErrorIs(t, err, ErrVersionRequired)

	err = item.Publish(301, false, time.Now())
	require.ErrorIs(t, err, ErrPrimarySpaceRequired)

	publishedAt := time.Now().UTC()
	require.NoError(t, item.Publish(301, true, publishedAt))
	require.True(t, item.IsPublic())
	require.Equal(t, int64(301), item.CurrentVersionID())
	require.Equal(t, publishedAt, *item.PublishedAt())
}

func TestMembersListingIsNotPublic(t *testing.T) {
	item, err := New(42, 108, "internal-teaching-assistant")
	require.NoError(t, err)
	require.NoError(t, item.Approve())
	require.NoError(t, item.SetVisibility(VisibilityMembers))

	require.NoError(t, item.Publish(302, true, time.Now().UTC()))
	require.False(t, item.IsPublic())
}

func TestRestorePublishedListing(t *testing.T) {
	publishedAt := time.Now().UTC()
	item, err := Restore(State{
		ID:               5,
		MarketplaceID:    42,
		CatalogItemID:    108,
		Slug:             "product-listing-optimizer",
		Status:           StatusPublished,
		Visibility:       VisibilityPublic,
		CurrentVersionID: 301,
		PublishedAt:      &publishedAt,
		HasPrimarySpace:  true,
	})
	require.NoError(t, err)
	require.True(t, item.IsPublic())
	require.Equal(t, "product-listing-optimizer", item.Slug().String())
}

func TestRestoreRejectsPublishedListingWithoutPrimarySpace(t *testing.T) {
	publishedAt := time.Now().UTC()
	_, err := Restore(State{
		ID:               5,
		MarketplaceID:    42,
		CatalogItemID:    108,
		Slug:             "product-listing-optimizer",
		Status:           StatusPublished,
		Visibility:       VisibilityPublic,
		CurrentVersionID: 301,
		PublishedAt:      &publishedAt,
	})
	require.ErrorIs(t, err, ErrPrimarySpaceRequired)
}
