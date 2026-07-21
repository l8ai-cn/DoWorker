package orchestrationworker

import (
	"context"
	"errors"
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerApplyServiceMaterializesDurableDispatch(t *testing.T) {
	state := workerInvocationApplyCreateState(t)
	repository := &workerApplyRepositoryStub{
		state: state,
		claim: WorkerLaunchClaim{
			LaunchID: 801, PlanID: state.Plan.ID,
			OrganizationID: 42, ActorID: 7,
			ResourceID: 301, ResourceRevision: 1,
			WorkerSpecSnapshotID: 901,
			Prompt:               stringPointerForTest("Review authorization carefully"),
			Alias:                "reviewer-42",
		},
	}
	launcher := &workerPodLauncherStub{result: WorkerPodLaunch{
		PodID: 601, PodKey: "7-standalone-12345678", RunnerID: 11,
		CommandPayload: []byte{1, 2, 3},
	}}
	notifier := &workerDispatchNotifierStub{}
	service, err := NewWorkerApplyService(
		workerApplyRegistry(t),
		repository,
		workerApplyServiceResolver(),
		launcher,
		notifier,
	)
	require.NoError(t, err)

	result, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		state.Plan.ID,
	)

	require.NoError(t, err)
	assert.Equal(t, launcher.result.PodKey, result.PodKey)
	assert.Equal(t, int64(801), result.LaunchID)
	assert.Equal(t, repository.claim, launcher.claim)
	assert.Equal(t, launcher.result, repository.completedLaunch)
	assert.Equal(t, int64(11), notifier.runnerID)
	assert.Equal(t, 24*time.Hour, repository.dispatchTTL)
}

func TestWorkerApplyServiceReplayDoesNotMaterializeAgain(t *testing.T) {
	state := workerInvocationApplyCreateState(t)
	repository := &workerApplyRepositoryStub{
		state: state,
		applied: AppliedWorker{
			Head: control.ResourceHead{
				ID: 301, Revision: 1,
			},
			LaunchID: 801, WorkerSpecSnapshotID: 901,
			ResourceRevision: 1, PodID: 601,
			PodKey: "7-standalone-12345678", RunnerID: 11,
		},
	}
	launcher := &workerPodLauncherStub{}
	notifier := &workerDispatchNotifierStub{}
	service, err := NewWorkerApplyService(
		workerApplyRegistry(t),
		repository,
		workerApplyServiceResolver(),
		launcher,
		notifier,
	)
	require.NoError(t, err)

	result, err := service.Apply(
		context.Background(),
		state.Plan.Scope,
		state.Plan.ID,
	)

	require.NoError(t, err)
	assert.Equal(t, repository.applied, result)
	assert.Zero(t, launcher.calls)
	assert.Zero(t, repository.claimCalls)
	assert.Equal(t, int64(11), notifier.runnerID)
}

func TestWorkerApplyServiceReleasesClaimAfterMaterializationFailure(t *testing.T) {
	state := workerInvocationApplyCreateState(t)
	repository := &workerApplyRepositoryStub{
		state: state,
		claim: WorkerLaunchClaim{
			LaunchID: 801, PlanID: state.Plan.ID,
			OrganizationID: 42, ActorID: 7,
			ResourceID: 301, ResourceRevision: 1,
			WorkerSpecSnapshotID: 901,
		},
	}
	launchErr := errors.New("runner selection failed")
	launcher := &workerPodLauncherStub{err: launchErr}
	service, err := NewWorkerApplyService(
		workerApplyRegistry(t),
		repository,
		workerApplyServiceResolver(),
		launcher,
		&workerDispatchNotifierStub{},
	)
	require.NoError(t, err)

	_, err = service.Apply(
		context.Background(),
		state.Plan.Scope,
		state.Plan.ID,
	)

	assert.ErrorIs(t, err, launchErr)
	assert.Equal(t, 1, repository.releaseCalls)
	assert.Equal(t, "materialization failed", repository.releaseReason)
}

func TestWorkerApplyServiceReleasesClaimAfterInvalidMaterialization(t *testing.T) {
	state := workerInvocationApplyCreateState(t)
	repository := &workerApplyRepositoryStub{
		state: state,
		claim: WorkerLaunchClaim{
			LaunchID: 801, PlanID: state.Plan.ID,
			OrganizationID: 42, ActorID: 7,
			ResourceID: 301, ResourceRevision: 1,
			WorkerSpecSnapshotID: 901,
		},
	}
	service, err := NewWorkerApplyService(
		workerApplyRegistry(t),
		repository,
		workerApplyServiceResolver(),
		&workerPodLauncherStub{result: WorkerPodLaunch{
			PodID: 601, PodKey: "7-standalone-12345678", RunnerID: 11,
		}},
		&workerDispatchNotifierStub{},
	)
	require.NoError(t, err)

	_, err = service.Apply(
		context.Background(),
		state.Plan.Scope,
		state.Plan.ID,
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
	assert.Equal(t, 1, repository.releaseCalls)
	assert.Equal(t, "materialization result invalid", repository.releaseReason)
}

type workerApplyRepositoryStub struct {
	state           controlservice.LockedApplyState
	applied         AppliedWorker
	claim           WorkerLaunchClaim
	completedLaunch WorkerPodLaunch
	dispatchTTL     time.Duration
	claimCalls      int
	releaseCalls    int
	releaseReason   string
}

func (stub *workerApplyRepositoryStub) RunWorkerApplyTransaction(
	_ context.Context,
	_ control.Scope,
	_ string,
	build WorkerApplyBuilder,
) (AppliedWorker, error) {
	if stub.applied.PodKey != "" {
		return stub.applied, nil
	}
	mutation, err := build(stub.state)
	if err != nil {
		return AppliedWorker{}, err
	}
	stub.applied = AppliedWorker{
		Head: mutation.Head, LaunchID: stub.claim.LaunchID,
		WorkerSpecSnapshotID: mutation.Launch.WorkerSpecSnapshotID,
		ResourceRevision:     mutation.Head.Revision,
	}
	return stub.applied, nil
}

func (stub *workerApplyRepositoryStub) ClaimWorkerLaunch(
	_ context.Context,
	_ control.Scope,
	_ int64,
	leaseDuration time.Duration,
	claimToken string,
) (WorkerLaunchClaim, error) {
	stub.claimCalls++
	stub.claim.ClaimToken = claimToken
	stub.claim.LeaseExpiresAt = time.Now().Add(leaseDuration)
	return stub.claim, nil
}

func (stub *workerApplyRepositoryStub) ReleaseWorkerLaunch(
	_ context.Context,
	_ control.Scope,
	_ WorkerLaunchClaim,
	reason string,
) error {
	stub.releaseCalls++
	stub.releaseReason = reason
	return nil
}

func (stub *workerApplyRepositoryStub) CompleteWorkerLaunch(
	_ context.Context,
	_ control.Scope,
	_ WorkerLaunchClaim,
	launch WorkerPodLaunch,
	dispatchTTL time.Duration,
) (AppliedWorker, error) {
	stub.completedLaunch = launch
	stub.dispatchTTL = dispatchTTL
	return AppliedWorker{
		Head: stub.applied.Head, LaunchID: stub.claim.LaunchID,
		WorkerSpecSnapshotID: stub.claim.WorkerSpecSnapshotID,
		ResourceRevision:     stub.claim.ResourceRevision,
		PodID:                launch.PodID, PodKey: launch.PodKey, RunnerID: launch.RunnerID,
	}, nil
}

type workerPodLauncherStub struct {
	claim  WorkerLaunchClaim
	result WorkerPodLaunch
	err    error
	calls  int
}

func (stub *workerPodLauncherStub) MaterializeWorkerPod(
	_ context.Context,
	claim WorkerLaunchClaim,
) (WorkerPodLaunch, error) {
	stub.calls++
	stub.claim = claim
	return stub.result, stub.err
}

type workerDispatchNotifierStub struct {
	runnerID int64
}

func (stub *workerDispatchNotifierStub) TriggerWorkerDispatch(
	runnerID int64,
) {
	stub.runnerID = runnerID
}

func workerApplyServiceResolver() *definitionResolverStub {
	return &definitionResolverStub{
		prompt: resource.PromptSpec{
			Content: "Review {{scope}} {{tone}}",
			Variables: map[string]resource.PromptVariableSpec{
				"scope": {Required: true},
				"tone":  {Default: stringPointerForTest("carefully")},
			},
		},
	}
}
