package orchestrationcontrol

import (
	"context"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceListResourcesAppliesWorkerDefinitionPolicy(t *testing.T) {
	repository := &resourceQueryRepositoryStub{}
	service := &Service{
		repository: repository,
		authorizer: &orchestrationAuthorizerStub{},
		workerDefinitions: workerDefinitionPolicyStub{
			"cursor-cli": {
				ModelManagedFields:     []string{"CURSOR_MODEL"},
				CredentialBundleFields: []string{"CURSOR_API_KEY"},
			},
		},
	}

	_, err := service.ListResources(
		context.Background(),
		orchestrationServiceScope(),
		environmentBundleListFilter(
			EnvironmentBundlePurposeRuntime,
			"",
		),
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{"CURSOR_MODEL"},
		repository.listFilter.EnvironmentBundle.ModelManagedFields,
	)

	_, err = service.ListResources(
		context.Background(),
		orchestrationServiceScope(),
		environmentBundleListFilter(
			EnvironmentBundlePurposeCredential,
			"CURSOR_API_KEY",
		),
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		"CURSOR_API_KEY",
		repository.listFilter.EnvironmentBundle.TargetName,
	)
}

func TestServiceListResourcesRejectsUndeclaredCredentialTarget(t *testing.T) {
	service := &Service{
		repository: &resourceQueryRepositoryStub{},
		authorizer: &orchestrationAuthorizerStub{},
		workerDefinitions: workerDefinitionPolicyStub{
			"cursor-cli": workerdefinition.EnvironmentBundlePolicy{},
		},
	}

	_, err := service.ListResources(
		context.Background(),
		orchestrationServiceScope(),
		environmentBundleListFilter(
			EnvironmentBundlePurposeCredential,
			"CURSOR_API_KEY",
		),
	)

	require.ErrorIs(t, err, control.ErrInvalid)
}

func environmentBundleListFilter(
	purpose EnvironmentBundlePurpose,
	targetName string,
) ResourceListFilter {
	return ResourceListFilter{
		Kind: resource.KindEnvironmentBundle, Limit: 100,
		EnvironmentBundle: &EnvironmentBundleReferenceFilter{
			Purpose: purpose, WorkerType: slugkit.Slug("cursor-cli"),
			TargetName: targetName,
		},
	}
}
