package market

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRejectsInvalidSlug(t *testing.T) {
	_, err := New("Higher_Education", "高校 AI 市场", "面向高校的应用市场", 9, 14)
	require.Error(t, err)
}

func TestTransitionRequiresConfiguredLifecycle(t *testing.T) {
	item, err := New("higher-education", "高校 AI 市场", "面向高校的应用市场", 9, 14)
	require.NoError(t, err)

	require.ErrorIs(t, item.Transition(StatusPublished), ErrInvalidTransition)
	require.NoError(t, item.Transition(StatusConfiguring))
	require.NoError(t, item.Transition(StatusReview))
	require.NoError(t, item.Transition(StatusPublished))
	require.Equal(t, StatusPublished, item.Status())
}

func TestRestorePublishedMarket(t *testing.T) {
	item, err := Restore(State{
		ID:                      42,
		Slug:                    "higher-education",
		Name:                    "高校 AI 市场",
		Summary:                 "面向高校的应用市场",
		Status:                  StatusPublished,
		Visibility:              "public",
		OwnerPlatformOrgID:      9,
		CreatedByPlatformUserID: 14,
	})
	require.NoError(t, err)
	require.Equal(t, "higher-education", item.Slug().String())
	require.Equal(t, StatusPublished, item.Status())
}
