package catalog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewVersionRequiresSemverAndSHA256(t *testing.T) {
	_, err := NewVersion(1, "release-one", "git-sha", strings.Repeat("a", 64), []byte(`{}`), 14)
	require.ErrorIs(t, err, ErrInvalidVersion)

	_, err = NewVersion(1, "1.0.0", "git-sha", "not-a-digest", []byte(`{}`), 14)
	require.ErrorIs(t, err, ErrInvalidDigest)

	version, err := NewVersion(1, "1.0.0", "git-sha", strings.Repeat("a", 64), []byte(`{}`), 14)
	require.NoError(t, err)
	require.Equal(t, ValidationPending, version.ValidationStatus())
	version.MarkValidationPassed()
	require.Equal(t, ValidationPassed, version.ValidationStatus())
}

func TestNewItemProtectsSlug(t *testing.T) {
	_, err := NewItem(2, "Invalid_Slug", "application", "应用", "开箱即用", "expert", 18, 14)
	require.Error(t, err)

	item, err := NewItem(2, "listing-optimizer", "application", "应用", "开箱即用", "expert", 18, 14)
	require.NoError(t, err)
	require.Equal(t, "listing-optimizer", item.Slug().String())
}

func TestRestorePassedVersion(t *testing.T) {
	version, err := RestoreVersion(VersionState{
		ID:                      3,
		CatalogItemID:           1,
		Version:                 "1.0.0",
		SourceRevision:          "git-sha",
		ContentDigest:           strings.Repeat("a", 64),
		Manifest:                []byte(`{}`),
		ValidationStatus:        ValidationPassed,
		CreatedByPlatformUserID: 14,
	})
	require.NoError(t, err)
	require.Equal(t, ValidationPassed, version.ValidationStatus())
}
