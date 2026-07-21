package orchestrationcontrol

import (
	"context"
	"testing"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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

func TestServiceListResourcesDerivesModelBindingProtocols(t *testing.T) {
	repository := &resourceQueryRepositoryStub{}
	service := &Service{
		repository: repository,
		authorizer: &orchestrationAuthorizerStub{},
		workerDefinitions: modelBindingPolicyStub{
			workerType: "minimax-cli",
			adapters:   []string{"minimax"},
		},
	}

	result, err := service.ListResources(
		context.Background(),
		orchestrationServiceScope(),
		ResourceListFilter{
			Kind:  resource.KindModelBinding,
			Limit: 100,
			ModelBinding: &ModelBindingReferenceFilter{
				WorkerType: slugkit.MustNewForTest("minimax-cli"),
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{"minimax"},
		repository.listFilter.ModelBinding.ProtocolAdapters,
	)
	assert.Equal(
		t,
		[]string{"minimax"},
		result.AppliedFilter.ModelBinding.ProtocolAdapters,
	)
}

type modelBindingPolicyStub struct {
	workerType string
	adapters   []string
}

func (stub modelBindingPolicyStub) EnvironmentBundlePolicy(
	string,
) (workerdefinition.EnvironmentBundlePolicy, bool) {
	return workerdefinition.EnvironmentBundlePolicy{}, false
}

func (stub modelBindingPolicyStub) ModelBindingProtocolAdapters(
	workerType string,
) ([]string, bool) {
	if workerType != stub.workerType {
		return nil, false
	}
	return append([]string{}, stub.adapters...), true
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
