package workercreation

import (
	"context"
	"testing"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeCatalogResolverRejectsUnreleasedSelection(t *testing.T) {
	resolver := newRuntimeCatalogResolver(runtimedomain.DefaultCatalog())

	_, err := resolver.ResolveRuntime(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		specservice.RuntimeSelection{
			RuntimeImageID:    1,
			PlacementPolicy:   specdomain.PlacementPolicyExplicit,
			ComputeTargetID:   1,
			DeploymentMode:    specdomain.DeploymentModePooled,
			ResourceProfileID: 1,
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "selection is disabled")
}

func TestRuntimeCatalogResolverRejectsInvalidOrUnavailableSelections(t *testing.T) {
	resolver := newRuntimeCatalogResolver(enabledCodexRuntimeCatalog())
	valid := specservice.RuntimeSelection{
		RuntimeImageID:    1,
		PlacementPolicy:   specdomain.PlacementPolicyExplicit,
		ComputeTargetID:   1,
		DeploymentMode:    specdomain.DeploymentModePooled,
		ResourceProfileID: 1,
	}

	tests := []struct {
		name       string
		workerType string
		mutate     func(*specservice.RuntimeSelection)
		match      string
	}{
		{
			name:       "image belongs to another worker type",
			workerType: "claude-code",
			match:      "runtime image",
		},
		{
			name:       "managed target is disabled",
			workerType: "codex-cli",
			mutate: func(selection *specservice.RuntimeSelection) {
				selection.ComputeTargetID = 2
				selection.DeploymentMode = specdomain.DeploymentModeDedicated
			},
			match: "Dedicated managed Kubernetes provisioning is not configured",
		},
		{
			name:       "runner pool does not support dedicated mode",
			workerType: "codex-cli",
			mutate: func(selection *specservice.RuntimeSelection) {
				selection.DeploymentMode = specdomain.DeploymentModeDedicated
			},
			match: "deployment mode",
		},
		{
			name:       "resource profile is unknown",
			workerType: "codex-cli",
			mutate: func(selection *specservice.RuntimeSelection) {
				selection.ResourceProfileID = 999
			},
			match: "resource profile",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selection := valid
			if test.mutate != nil {
				test.mutate(&selection)
			}

			_, err := resolver.ResolveRuntime(
				context.Background(),
				specservice.Scope{OrgID: 77, UserID: 7},
				slugkit.MustNewForTest(test.workerType),
				selection,
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
			assert.ErrorContains(t, err, test.match)
		})
	}
}
