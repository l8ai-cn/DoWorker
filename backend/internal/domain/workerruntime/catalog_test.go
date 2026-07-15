package workerruntime

import (
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
	require.Len(t, allImages, 4)
	for _, image := range allImages {
		assert.Regexp(t, regexp.MustCompile(`^sha256:[a-f0-9]{64}$`), image.Digest)
		assert.True(t, strings.HasSuffix(image.Reference, "@"+image.Digest))
		assert.True(t, image.Enabled)
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

func TestDefaultCatalogExposesDoAgentImageToSeedanceExpert(t *testing.T) {
	catalog := DefaultCatalog()

	doAgentImages := catalog.ImagesFor("do-agent")
	require.Len(t, doAgentImages, 1)
	assert.Equal(t, doAgentImages, catalog.ImagesFor("seedance-expert"))
	assert.Empty(t, catalog.ImagesFor("unknown-worker"))
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
