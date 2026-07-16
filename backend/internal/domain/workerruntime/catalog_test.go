package workerruntime

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCatalogExposesImmutableRuntimeSelections(t *testing.T) {
	catalog := DefaultCatalog()

	allImages := catalog.Images()
	require.Len(t, allImages, 3)
	for _, image := range allImages {
		assert.Regexp(t, regexp.MustCompile(`^sha256:[a-f0-9]{64}$`), image.Digest)
		assert.True(t, strings.HasSuffix(image.Reference, "@"+image.Digest))
		assert.False(t, image.Enabled)
	}

	images := catalog.ImagesFor("codex-cli")
	require.NotEmpty(t, images)

	target := catalog.Target(1)
	require.NotNil(t, target)
	assert.Equal(t, workerspec.ComputeTargetKindRunnerPool, target.Kind)
	assert.True(t, target.SupportsPooled)
	assert.False(t, target.SupportsDedicated)
	assert.True(t, target.Enabled)

	managed := catalog.Target(2)
	require.NotNil(t, managed)
	assert.Equal(t, workerspec.ComputeTargetKindKubernetes, managed.Kind)
	assert.False(t, managed.Enabled)
	assert.NotEmpty(t, managed.DisabledReason)

	profile := catalog.Profile(1)
	require.NotNil(t, profile)
	assert.Positive(t, profile.Resources.CPURequestMilliCPU)
	assert.GreaterOrEqual(
		t,
		profile.Resources.CPULimitMilliCPU,
		profile.Resources.CPURequestMilliCPU,
	)
	assert.Positive(t, profile.Resources.MemoryRequestBytes)
	assert.GreaterOrEqual(
		t,
		profile.Resources.MemoryLimitBytes,
		profile.Resources.MemoryRequestBytes,
	)
}

func TestLoadCatalogReadsExplicitDevelopmentLock(t *testing.T) {
	path := t.TempDir() + "/runtime-catalog.json"
	err := os.WriteFile(path, []byte(`{
		"schema_version": 1,
		"revision": "local-dev-codex",
		"images": [{
			"id": 1,
			"slug": "codex-cli-local",
			"name": "Codex CLI (local development)",
			"reference": "docker-daemon://runner-codex@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"digest": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"worker_type_slugs": ["codex-cli"],
			"enabled": true
		}]
	}`), 0o600)
	require.NoError(t, err)

	catalog, err := LoadCatalog(path)

	require.NoError(t, err)
	assert.Equal(t, "local-dev-codex", catalog.Revision())
	images := catalog.ImagesFor("codex-cli")
	require.Len(t, images, 1)
	assert.True(t, images[0].Enabled)
}

func TestDefaultCatalogDoesNotInventUnavailableWorkerImages(t *testing.T) {
	catalog := DefaultCatalog()

	assert.Empty(t, catalog.ImagesFor("do-agent"))
	assert.Empty(t, catalog.ImagesFor("unknown-worker"))
}

func TestDefaultCatalogDoesNotUseMutableEnvironmentImageReferences(t *testing.T) {
	digest := "sha256:" + strings.Repeat("d", 64)
	t.Setenv(
		"WORKER_RUNTIME_IMAGE_REFERENCES",
		"do-agent=agentsmesh-main-runner-do-agent@"+digest,
	)

	images := DefaultCatalog().ImagesFor("do-agent")

	assert.Empty(t, images)
}

func TestParseRuntimeCatalogLockRejectsMutableReference(t *testing.T) {
	_, err := parseRuntimeCatalogLock([]byte(`{
		"schema_version": 1,
		"revision": "test",
		"images": [{
			"id": 1,
			"slug": "do-agent-stable",
			"name": "Do Agent",
			"reference": "registry.example/do-agent:latest",
			"digest": "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			"worker_type_slugs": ["do-agent"],
			"enabled": true
		}]
	}`))

	require.Error(t, err)
}

func TestCatalogReturnsCopies(t *testing.T) {
	catalog := DefaultCatalog()

	images := catalog.ImagesFor("codex-cli")
	require.NotEmpty(t, images)
	images[0].Digest = "changed"

	assert.NotEqual(t, "changed", catalog.ImagesFor("codex-cli")[0].Digest)

	target := catalog.Target(1)
	require.NotNil(t, target)
	target.Name = "changed"
	assert.NotEqual(t, "changed", catalog.Target(1).Name)
}
