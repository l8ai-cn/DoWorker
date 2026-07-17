package orchestrationworker

import (
	"context"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindingApplyServiceConsumesPlanThroughTrustedTransaction(t *testing.T) {
	state := bindingApplyCreateState(t)
	repository := &bindingApplyRepositoryStub{state: state}
	service, err := NewBindingApplyService(
		workerApplyRegistry(t),
		repository,
	)
	require.NoError(t, err)

	head, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
	)

	require.NoError(t, err)
	assert.Equal(t, state.ResultResourceID, head.ID)
	assert.Zero(t, repository.mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, 1, repository.calls)
}

func TestWorkerTemplateApplyServiceReturnsAtomicSnapshotBinding(t *testing.T) {
	state := workerApplyCreateState(t)
	repository := &workerTemplateApplyRepositoryStub{
		state:      state,
		snapshotID: 901,
	}
	service, err := NewWorkerTemplateApplyService(
		workerApplyRegistry(t),
		repository,
	)
	require.NoError(t, err)

	result, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
	)

	require.NoError(t, err)
	assert.Equal(t, state.ResultResourceID, result.Head.ID)
	assert.Equal(t, int64(901), result.WorkerSpecSnapshotID)
	assert.Equal(t, int64(901), repository.mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, 1, repository.calls)
}

func TestPromptApplyServiceConsumesPlanThroughTrustedTransaction(t *testing.T) {
	state := promptApplyCreateState(t)
	repository := &promptApplyRepositoryStub{state: state}
	service, err := NewPromptApplyService(workerApplyRegistry(t), repository)
	require.NoError(t, err)

	head, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
	)

	require.NoError(t, err)
	assert.Equal(t, state.ResultResourceID, head.ID)
	assert.Equal(t, resource.KindPrompt, head.Identity.Kind)
	assert.Zero(t, repository.mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, 1, repository.calls)
}

func TestExpertApplyServiceBuildsPinnedDomainProjection(t *testing.T) {
	state := expertApplyCreateState(t)
	repository := &expertApplyRepositoryStub{
		state:    state,
		expertID: 701,
	}
	resolver := &definitionResolverStub{
		prompt: resource.PromptSpec{
			Content: "Review {{tone}}",
			Variables: map[string]resource.PromptVariableSpec{
				"tone": {Default: stringPointerForTest("carefully")},
			},
		},
	}
	service, err := NewExpertApplyService(
		workerApplyRegistry(t),
		repository,
		resolver,
	)
	require.NoError(t, err)

	applied, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		"11111111-1111-4111-8111-111111111111",
	)

	require.NoError(t, err)
	assert.Equal(t, int64(701), applied.ExpertID)
	assert.Equal(t, int64(901), applied.WorkerSpecSnapshotID)
	assert.Equal(t, int64(1), applied.ResourceRevision)
	assert.Equal(t, "Review carefully", repository.mutation.Projection.Prompt)
	assert.Equal(t, "engineering", repository.mutation.Projection.Category)
	assert.Equal(t, int64(901), repository.mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, 1, repository.calls)
}

func TestApplyServicesRejectMissingDependencies(t *testing.T) {
	_, err := NewBindingApplyService(nil, nil)
	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
	_, err = NewWorkerTemplateApplyService(nil, nil)
	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
	_, err = NewPromptApplyService(nil, nil)
	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
	_, err = NewExpertApplyService(nil, nil, nil)
	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
}

type bindingApplyRepositoryStub struct {
	state    controlservice.LockedApplyState
	mutation controlservice.ApplyMutation
	calls    int
}

func (stub *bindingApplyRepositoryStub) RunBindingApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build BindingApplyBuilder,
) (control.ResourceHead, error) {
	stub.calls++
	mutation, err := build(stub.state)
	if err != nil {
		return control.ResourceHead{}, err
	}
	stub.mutation = mutation
	return mutation.Head, nil
}

type workerTemplateApplyRepositoryStub struct {
	state      controlservice.LockedApplyState
	snapshotID int64
	mutation   controlservice.ApplyMutation
	calls      int
}

func (stub *workerTemplateApplyRepositoryStub) RunWorkerTemplateApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build WorkerTemplateApplyBuilder,
) (AppliedWorkerTemplate, error) {
	stub.calls++
	mutation, err := build(stub.state, stub.snapshotID)
	if err != nil {
		return AppliedWorkerTemplate{}, err
	}
	stub.mutation = mutation
	return AppliedWorkerTemplate{
		Head:                 mutation.Head,
		WorkerSpecSnapshotID: stub.snapshotID,
	}, nil
}

type promptApplyRepositoryStub struct {
	state    controlservice.LockedApplyState
	mutation controlservice.ApplyMutation
	calls    int
}

func (stub *promptApplyRepositoryStub) RunPromptApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build controlservice.ApplyBuilder,
) (control.ResourceHead, error) {
	stub.calls++
	mutation, err := build(stub.state)
	if err != nil {
		return control.ResourceHead{}, err
	}
	stub.mutation = mutation
	return mutation.Head, nil
}

type expertApplyRepositoryStub struct {
	state    controlservice.LockedApplyState
	expertID int64
	mutation ExpertApplyMutation
	calls    int
}

func (stub *expertApplyRepositoryStub) RunExpertApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build ExpertApplyBuilder,
) (AppliedExpert, error) {
	stub.calls++
	mutation, err := build(stub.state)
	if err != nil {
		return AppliedExpert{}, err
	}
	stub.mutation = mutation
	return AppliedExpert{
		Head:                 mutation.Head,
		ExpertID:             stub.expertID,
		WorkerSpecSnapshotID: mutation.Projection.WorkerSpecSnapshotID,
		ResourceRevision:     mutation.Head.Revision,
	}, nil
}

func stringPointerForTest(value string) *string {
	return &value
}
