package listing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewVersionRequiresPresentationAndJSONArrays(t *testing.T) {
	_, err := NewVersion(12, 31, 1, "", "一句话价值", "完整介绍",
		json.RawMessage(`[]`), json.RawMessage(`[]`), json.RawMessage(`[]`),
		json.RawMessage(`[]`), nil, "首次发布")
	require.ErrorIs(t, err, ErrInvalidPresentation)

	version, err := NewVersion(12, 31, 1, "商品优化应用", "一句话价值", "完整介绍",
		json.RawMessage(`["提升转化"]`), json.RawMessage(`["批量优化"]`),
		json.RawMessage(`["跨境运营"]`), json.RawMessage(`[]`),
		[]string{"跨境电商"}, "首次发布")
	require.NoError(t, err)
	require.Equal(t, ReviewDraft, version.ReviewStatus())
	require.Equal(t, 1, version.Revision())
}

func TestVersionReviewLifecycleLocksSubmittedRevision(t *testing.T) {
	version, err := NewVersion(12, 31, 1, "商品优化应用", "一句话价值", "完整介绍",
		json.RawMessage(`[]`), json.RawMessage(`[]`), json.RawMessage(`[]`),
		json.RawMessage(`[]`), nil, "首次发布")
	require.NoError(t, err)

	require.NoError(t, version.Submit())
	require.ErrorIs(t, version.Submit(), ErrInvalidReviewTransition)
	require.NoError(t, version.Approve())
	require.Equal(t, ReviewApproved, version.ReviewStatus())
}

func TestRestoreApprovedVersion(t *testing.T) {
	version, err := RestoreVersion(VersionState{
		ID:                   71,
		ListingID:            61,
		CatalogItemVersionID: 51,
		Revision:             1,
		DisplayName:          "商品优化应用",
		Tagline:              "一句话价值",
		Description:          "完整介绍",
		Outcomes:             json.RawMessage(`[]`),
		UseCases:             json.RawMessage(`[]`),
		TargetAudience:       json.RawMessage(`[]`),
		Requirements:         json.RawMessage(`[]`),
		ReviewStatus:         ReviewApproved,
	})
	require.NoError(t, err)
	require.Equal(t, int64(71), version.ID())
	require.Equal(t, ReviewApproved, version.ReviewStatus())
}

func TestRestoreRejectedVersion(t *testing.T) {
	version, err := RestoreVersion(VersionState{
		ID: 71, ListingID: 61, CatalogItemVersionID: 51, Revision: 1,
		DisplayName: "商品优化应用", Tagline: "一句话价值", Description: "完整介绍",
		Outcomes: json.RawMessage(`[]`), UseCases: json.RawMessage(`[]`),
		TargetAudience: json.RawMessage(`[]`), Requirements: json.RawMessage(`[]`),
		ReviewStatus: ReviewRejected,
	})
	require.NoError(t, err)
	require.Equal(t, ReviewRejected, version.ReviewStatus())
}
