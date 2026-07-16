package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWorkerRuntimeCatalogFile(t *testing.T) {
	setRequiredPreviewOrigin(t)
	t.Setenv("WORKER_RUNTIME_CATALOG_FILE", "/tmp/worker-runtime-catalog.json")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "/tmp/worker-runtime-catalog.json", cfg.WorkerRuntimeCatalogFile)
}
