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
	require.Len(t, allImages, 5)
	for _, image := range allImages {
		assert.Regexp(t, regexp.MustCompile(`^sha256:[a-f0-9]{64}$`), image.Digest)
		assert.True(t, strings.HasSuffix(image.Reference, "@"+image.Digest))
	}
	assert.True(t, allImages[0].Enabled)
	assert.False(t, allImages[1].Enabled)
	assert.False(t, allImages[2].Enabled)
	assert.True(t, allImages[3].Enabled)
	assert.True(t, allImages[4].Enabled)

	images := catalog.ImagesFor("codex-cli")
	require.Len(t, images, 1)
	assert.Equal(t, "codex-cli-stable", images[0].Slug)
	assert.True(t, images[0].Enabled)
	videoImages := catalog.ImagesFor("video-studio")
	require.Len(t, videoImages, 1)
	assert.Equal(t, int64(4), videoImages[0].ID)
	assert.Equal(t, "video-studio-stable", videoImages[0].Slug)
	assert.True(t, videoImages[0].Enabled)

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

func TestDefaultCatalogPublishesDoAgentForSeedance(t *testing.T) {
	catalog := DefaultCatalog()

	doAgentImages := catalog.ImagesFor("do-agent")
	seedanceImages := catalog.ImagesFor("seedance-expert")
	require.Len(t, doAgentImages, 1)
	require.Len(t, seedanceImages, 1)
	assert.True(t, doAgentImages[0].Enabled)
	assert.Equal(t, doAgentImages[0], seedanceImages[0])
	assert.Equal(t, []string{"do-agent", "seedance-expert"}, doAgentImages[0].WorkerTypeSlugs)
	assert.Regexp(t, regexp.MustCompile(`^repo\.aiedulab\.cn:8443/agentsmesh/runner-do-agent@sha256:[a-f0-9]{64}$`), doAgentImages[0].Reference)
}

func TestDefaultCatalogDoesNotInventUnknownWorkerImages(t *testing.T) {
	catalog := DefaultCatalog()

	assert.Empty(t, catalog.ImagesFor("unknown-worker"))
}

func TestDefaultCatalogDoesNotUseMutableEnvironmentImageReferences(t *testing.T) {
	expected := DefaultCatalog().ImagesFor("do-agent")
	digest := "sha256:" + strings.Repeat("d", 64)
	t.Setenv(
		"WORKER_RUNTIME_IMAGE_REFERENCES",
		"do-agent=agentsmesh-main-runner-do-agent@"+digest,
	)

	images := DefaultCatalog().ImagesFor("do-agent")

	assert.Equal(t, expected, images)
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
