package listing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublishRequiresApprovedVersionAndPrimarySpace(t *testing.T) {
	item, err := New(42, 108, "product-listing-optimizer")
	require.NoError(t, err)
	require.NoError(t, item.SetVisibility(VisibilityPublic))
	require.NoError(t, item.Submit(14))
	require.NoError(t, item.Approve(21))

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
	require.NoError(t, item.SetVisibility(VisibilityMembers))
	require.NoError(t, item.Submit(14))
	require.NoError(t, item.Approve(21))

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
		AccessMode:       AccessModeDirect,
		CurrentVersionID: 301,
		SubmittedBy:      14,
		PublishedAt:      &publishedAt,
		HasPrimarySpace:  true,
		Revision:         4,
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
		AccessMode:       AccessModeDirect,
		CurrentVersionID: 301,
		SubmittedBy:      14,
		PublishedAt:      &publishedAt,
		Revision:         4,
	})
	require.ErrorIs(t, err, ErrPrimarySpaceRequired)
}

func TestListingRequiresIndependentReviewer(t *testing.T) {
	item, err := New(42, 108, "product-listing-optimizer")
	require.NoError(t, err)
	require.NoError(t, item.Submit(14))

	require.ErrorIs(t, item.Approve(14), ErrSubmitterCannotReview)
	require.NoError(t, item.Approve(21))
	require.Equal(t, StatusApproved, item.Status())
	require.Equal(t, int64(14), item.SubmittedBy())
}

func TestRestoreSupportsOperationalListingStates(t *testing.T) {
	item, err := Restore(State{
		ID: 5, MarketplaceID: 42, CatalogItemID: 108,
		Slug: "product-listing-optimizer", Status: StatusNeedsChanges,
		Visibility: VisibilityHidden, AccessMode: AccessModeApproval,
		SubmittedBy: 14, Revision: 3,
	})
	require.NoError(t, err)
	require.Equal(t, StatusNeedsChanges, item.Status())
}
