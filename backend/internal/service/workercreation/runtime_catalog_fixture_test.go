package workercreation

import (
	"os"
	"path/filepath"
	"sync"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
)

var (
	enabledCodexCatalog     runtimedomain.Catalog
	enabledCodexCatalogOnce sync.Once
)

func enabledCodexRuntimeCatalog() runtimedomain.Catalog {
	enabledCodexCatalogOnce.Do(func() {
		path := filepath.Join(os.TempDir(), "workercreation-enabled-codex-catalog.json")
		content := `{
  "schema_version": 1,
  "revision": "` + runtimedomain.DefaultCatalogRevision() + `",
  "images": [{
    "id": 1,
    "slug": "codex-cli-test",
    "name": "Codex CLI (test)",
    "reference": "docker-daemon://runner-codex@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "digest": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "worker_type_slugs": ["codex-cli"],
    "enabled": true
  }]
}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			panic(err)
		}
		catalog, err := runtimedomain.LoadCatalog(path)
		if err != nil {
			panic(err)
		}
		enabledCodexCatalog = catalog
	})
	return enabledCodexCatalog
}

func runtimeCatalogWithE2EEcho() runtimedomain.Catalog {
	path := filepath.Join(os.TempDir(), "workercreation-enabled-e2e-echo-catalog.json")
	content := `{
  "schema_version": 1,
  "revision": "` + runtimedomain.DefaultCatalogRevision() + `",
  "images": [{
    "id": 6,
    "slug": "e2e-echo-test",
    "name": "E2E Echo (test)",
    "reference": "docker-daemon://runner-e2e-echo@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
    "digest": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
    "worker_type_slugs": ["e2e-echo"],
    "enabled": true
  }]
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		panic(err)
	}
	catalog, err := runtimedomain.LoadCatalog(path)
	if err != nil {
		panic(err)
	}
	return catalog
}
