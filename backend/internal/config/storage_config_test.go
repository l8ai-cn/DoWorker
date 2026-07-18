package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageAllowedTypesIncludeMP4ByDefault(t *testing.T) {
	setRequiredPreviewOrigin(t)
	t.Setenv("STORAGE_ALLOWED_TYPES", "")

	cfg, err := Load()

	require.NoError(t, err)
	require.Contains(t, cfg.Storage.AllowedTypes, "video/mp4")
}
