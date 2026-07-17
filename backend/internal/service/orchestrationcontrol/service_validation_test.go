package orchestrationcontrol

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceRejectsDuplicateMissingAndUnregisteredPlanners(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	duplicate := fixture.deps
	duplicate.Planners = append(duplicate.Planners, duplicate.Planners[0])
	_, err := NewService(duplicate)
	require.Error(t, err)

	missing := fixture.deps
	missing.Planners = nil
	_, err = NewService(missing)
	require.Error(t, err)

	unregistered := fixture.deps
	unregistered.RequiredTypes = append(
		[]orchestrationresource.TypeMeta{},
		unregistered.RequiredTypes...,
	)
	unregistered.RequiredTypes = append(
		unregistered.RequiredTypes,
		orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "Expert",
		},
	)
	unregistered.Planners = append(
		unregistered.Planners,
		&orchestrationPlannerStub{meta: unregistered.RequiredTypes[1]},
	)
	_, err = NewService(unregistered)
	require.Error(t, err)
}

func TestValidateUsesOneCanonicalDraftForJSONAndYAML(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	service := fixture.service(t)
	jsonResult, err := service.Validate(context.Background(), ValidateRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	require.NoError(t, err)
	yamlResult, err := service.Validate(context.Background(), ValidateRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatYAML, Content: []byte(testResourceYAML)},
	})
	require.NoError(t, err)

	assert.Equal(t, jsonResult.CanonicalManifest, yamlResult.CanonicalManifest)
	assert.Equal(t, orchestrationcontrol.PlanOperationCreate, jsonResult.Operation)
	assert.Empty(t, jsonResult.Issues)
	assert.Equal(t, 2, fixture.authorizer.createCalls)
	assert.Zero(t, fixture.references.calls)
	assert.Zero(t, fixture.planner.planCalls)
	assert.Zero(t, fixture.repository.createPlanCalls)
}

func TestValidateReturnsSafeDeterministicIssueForInvalidSource(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	service := fixture.service(t)
	result, err := service.Validate(context.Background(), ValidateRequest{
		Scope: fixture.scope,
		Source: ResourceSource{
			Format:  SourceFormatJSON,
			Content: []byte(`{"apiVersion":"agentsmesh.io/v1alpha1","kind":"WorkerTemplate","metadata":{"name":"worker-one","namespace":"team-alpha"},"spec":{"unknownSecret":"sk-do-not-echo"}}`),
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "/", result.Issues[0].Path)
	assert.Equal(t, "invalid-draft", result.Issues[0].Code)
	assert.NotContains(t, result.Issues[0].Message, "sk-do-not-echo")
	assert.Zero(t, fixture.authorizer.createCalls)
	assert.Zero(t, fixture.references.calls)
}

func TestValidateAuthorizesTargetBeforeReferenceResolution(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.authorizer.createErr = ErrForbidden
	service := fixture.service(t)

	_, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Equal(t, 1, fixture.authorizer.createCalls)
	assert.Zero(t, fixture.planner.referenceCalls)
	assert.Zero(t, fixture.references.calls)
	assert.Zero(t, fixture.repository.createPlanCalls)
}
