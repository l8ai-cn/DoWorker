package runner

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/stretchr/testify/require"
)

type workerDefinitionCatalog struct {
	WorkerTypes []struct {
		Slug           string `json:"slug"`
		DefinitionPath string `json:"definition_path"`
	} `json:"worker_types"`
}

type workerDefinitionAdapter struct {
	AdapterID        string   `json:"adapter_id"`
	InteractionModes []string `json:"interaction_modes"`
}

func TestDefinitionACPAdaptersAreRegisteredByRunner(t *testing.T) {
	catalog := loadWorkerDefinitionCatalog(t)
	for _, workerType := range catalog.WorkerTypes {
		definition := loadWorkerDefinitionAdapter(t, workerType.DefinitionPath)
		if !supportsACP(definition.InteractionModes) {
			continue
		}
		transport, err := acp.NewTransport(
			definition.AdapterID,
			acp.EventCallbacks{},
			slog.Default(),
		)
		require.NoErrorf(t, err, "%s adapter %q", workerType.Slug, definition.AdapterID)
		require.NotNilf(t, transport, "%s adapter %q", workerType.Slug, definition.AdapterID)
	}
}

func loadWorkerDefinitionCatalog(t *testing.T) workerDefinitionCatalog {
	t.Helper()
	var catalog workerDefinitionCatalog
	content, err := os.ReadFile(filepath.Join(workerDefinitionRoot(t), "catalog.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(content, &catalog))
	require.NotEmpty(t, catalog.WorkerTypes)
	return catalog
}

func loadWorkerDefinitionAdapter(t *testing.T, relativePath string) workerDefinitionAdapter {
	t.Helper()
	var definition workerDefinitionAdapter
	content, err := os.ReadFile(filepath.Join(workerRepositoryRoot(t), relativePath))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(content, &definition))
	require.NotEmpty(t, definition.AdapterID)
	return definition
}

func workerDefinitionRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(workerRepositoryRoot(t), "config", "worker-types")
}

func workerRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

func supportsACP(modes []string) bool {
	for _, mode := range modes {
		if mode == "acp" {
			return true
		}
	}
	return false
}
