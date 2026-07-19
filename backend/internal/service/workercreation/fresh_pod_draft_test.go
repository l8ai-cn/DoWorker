package workercreation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFreshPodDraftRejectsInvalidScopeBeforeCatalogUse(t *testing.T) {
	service := NewService(newWorkerCreationServiceFixture().deps())

	_, err := service.NewFreshPodDraft(
		context.Background(),
		specservice.Scope{OrgID: 0, UserID: 7},
		validFreshPodDraftInput(),
	)

	assert.ErrorIs(t, err, specservice.ErrInvalidScope)
}

func TestNewFreshPodDraftRejectsStaleOptionsRevision(t *testing.T) {
	service := NewService(newWorkerCreationServiceFixture().deps())
	input := validFreshPodDraftInput()
	input.OptionsRevision = "stale"

	_, err := service.NewFreshPodDraft(context.Background(), validFreshScope(), input)

	assert.ErrorIs(t, err, ErrStaleOptions)
}

func TestNewFreshPodDraftRejectsMissingPlacement(t *testing.T) {
	service := NewService(newWorkerCreationServiceFixture().deps())
	input := validFreshPodDraftInput()
	input.Runtime.ComputeTargetID = 0

	_, err := service.NewFreshPodDraft(context.Background(), validFreshScope(), input)

	require.Error(t, err)
	assert.Equal(t, "compute_target_id", specservice.InvalidDraftField(err))
}

func TestNewFreshPodDraftRejectsDisabledRuntimeImage(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	deps := fixture.deps()
	deps.Catalog = disabledCodexRuntimeCatalog(t)
	service := NewService(deps)

	_, err := service.NewFreshPodDraft(
		context.Background(),
		validFreshScope(),
		validFreshPodDraftInput(),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "runtime image")
	assert.ErrorContains(t, err, "disabled")
}

func TestNewFreshPodDraftRejectsIncompatibleRuntimeCombination(t *testing.T) {
	service := NewService(newWorkerCreationServiceFixture().deps())
	input := validFreshPodDraftInput()
	input.Runtime.DeploymentMode = specdomain.DeploymentModeDedicated

	_, err := service.NewFreshPodDraft(context.Background(), validFreshScope(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "deployment mode")
}

func TestNewFreshPodDraftRejectsInvalidAutomationLevel(t *testing.T) {
	service := NewService(newWorkerCreationServiceFixture().deps())
	input := validFreshPodDraftInput()
	input.AutomationLevel = "unknown"

	_, err := service.NewFreshPodDraft(context.Background(), validFreshScope(), input)

	require.Error(t, err)
	assert.Equal(t, "automation_level", specservice.InvalidDraftField(err))
}

func validFreshScope() specservice.Scope {
	return specservice.Scope{OrgID: 77, UserID: 7}
}

func validFreshPodDraftInput() FreshPodDraftInput {
	return FreshPodDraftInput{
		OptionsRevision:  runtimedomain.DefaultCatalogRevision(),
		OrganizationSlug: "dev-org",
		WorkerTypeSlug:   "codex-cli",
		Runtime: specservice.RuntimeSelection{
			RuntimeImageID:    1,
			PlacementPolicy:   specdomain.PlacementPolicyExplicit,
			ComputeTargetID:   1,
			DeploymentMode:    specdomain.DeploymentModePooled,
			ResourceProfileID: 1,
		},
		AutomationLevel: specdomain.AutomationLevelAutonomous,
	}
}

func disabledCodexRuntimeCatalog(t *testing.T) runtimedomain.Catalog {
	t.Helper()
	path := filepath.Join(t.TempDir(), "disabled-runtime-catalog.json")
	content := strings.Replace(
		`{
  "schema_version": 1,
  "revision": "`+runtimedomain.DefaultCatalogRevision()+`",
  "images": [{
    "id": 1,
    "slug": "codex-cli-disabled",
    "name": "Codex CLI (disabled)",
    "reference": "registry.example.com/runner-codex@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
    "digest": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
    "worker_type_slugs": ["codex-cli"],
    "enabled": true
  }]
}`,
		`"enabled": true`,
		`"enabled": false`,
		1,
	)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	catalog, err := runtimedomain.LoadCatalog(path)
	require.NoError(t, err)
	return catalog
}
