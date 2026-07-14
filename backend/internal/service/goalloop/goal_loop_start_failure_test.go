package goalloop

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentpodsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestStartFailsClosedWhenPodKeyPersistenceFails(t *testing.T) {
	loop := startableGoalLoop()
	repo := &failingGoalLoopUpdateRepo{
		goalLoopTestRepo: newGoalLoopTestRepo(loop),
		failOnCall:       2,
	}
	terminator := &goalLoopTerminator{}
	service := startFailureService(repo, terminator)

	started, err := service.Start(context.Background(), 1, 2, loop.Slug)

	require.Nil(t, started)
	require.ErrorContains(t, err, "persist pod key")
	require.Equal(t, domain.StatusFailed, loop.Status)
	require.Equal(t, []string{"loop-pod"}, terminator.keys)
	require.Contains(t, stringValue(loop.VerificationError), "persist pod key")
}

func TestStartLeavesAutopilotControllerUnset(t *testing.T) {
	loop := startableGoalLoop()
	repo := newGoalLoopTestRepo(loop)
	terminator := &goalLoopTerminator{}
	service := startFailureService(repo, terminator)

	started, err := service.Start(context.Background(), 1, 2, loop.Slug)

	require.NoError(t, err)
	require.NotNil(t, started)
	require.Equal(t, domain.StatusActive, loop.Status)
	require.Equal(t, "loop-pod", stringValue(loop.PodKey))
	require.Nil(t, loop.AutopilotControllerKey)
	require.Empty(t, terminator.keys)
}

func TestStartDoesNotCreatePodWhenStatusClaimIsLost(t *testing.T) {
	loop := startableGoalLoop()
	repo := &lostStartClaimRepo{goalLoopTestRepo: newGoalLoopTestRepo(loop)}
	creator := &countingLoopPodCreator{}
	service := startFailureService(repo, &goalLoopTerminator{})
	service.podCreator = creator

	started, err := service.Start(context.Background(), 1, 2, loop.Slug)

	require.Nil(t, started)
	require.ErrorIs(t, err, ErrInvalidState)
	require.Zero(t, creator.calls)
	require.Equal(t, domain.StatusDraft, loop.Status)
}

func TestStartTerminatesCreatedPodWhenCancelWins(t *testing.T) {
	loop := startableGoalLoop()
	repo := &cancelDuringStartRepo{goalLoopTestRepo: newGoalLoopTestRepo(loop)}
	creator := &countingLoopPodCreator{}
	terminator := &goalLoopTerminator{}
	service := startFailureService(repo, terminator)
	service.podCreator = creator

	started, err := service.Start(context.Background(), 1, 2, loop.Slug)

	require.Nil(t, started)
	require.ErrorIs(t, err, ErrInvalidState)
	require.Equal(t, 1, creator.calls)
	require.Equal(t, []string{"loop-pod"}, terminator.keys)
	require.Equal(t, domain.StatusCancelled, loop.Status)
}

func TestStartPersistsCleanupWhenCancelWinsAndPodStopFails(t *testing.T) {
	loop := startableGoalLoop()
	repo := &cancelDuringStartRepo{goalLoopTestRepo: newGoalLoopTestRepo(loop)}
	terminator := &goalLoopTerminator{err: errors.New("runner unavailable")}
	service := startFailureService(repo, terminator)

	started, err := service.Start(context.Background(), 1, 2, loop.Slug)

	require.Nil(t, started)
	require.ErrorIs(t, err, ErrInvalidState)
	require.Equal(t, domain.StatusCancelled, loop.Status)
	require.Equal(t, "loop-pod", stringValue(loop.PodKey))
	pending, ok := loop.PendingPodCleanup()
	require.True(t, ok)
	require.Equal(t, domain.StatusCancelled, pending.TargetStatus)

	terminator.err = nil
	updated, err := service.Cancel(context.Background(), 1, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusCancelled, updated.Status)
	_, stillPending := updated.PendingPodCleanup()
	require.False(t, stillPending)
}

func TestStartFailureRetainsPodUntilCleanupCanBeRetried(t *testing.T) {
	loop := startableGoalLoop()
	repo := &failingGoalLoopUpdateRepo{
		goalLoopTestRepo: newGoalLoopTestRepo(loop),
		failOnCall:       2,
	}
	terminator := &goalLoopTerminator{err: errors.New("runner unavailable")}
	service := startFailureService(repo, terminator)

	started, err := service.Start(context.Background(), 1, 2, loop.Slug)

	require.Nil(t, started)
	require.ErrorContains(t, err, "persist pod key")
	require.Equal(t, domain.StatusFailed, loop.Status)
	require.Equal(t, "loop-pod", stringValue(loop.PodKey))
	pending, ok := loop.PendingPodCleanup()
	require.True(t, ok)
	require.Equal(t, domain.StatusFailed, pending.TargetStatus)

	terminator.err = nil
	updated, err := service.Cancel(context.Background(), 1, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusFailed, updated.Status)
	require.Contains(t, stringValue(updated.VerificationError), "persist pod key")
	_, stillPending := updated.PendingPodCleanup()
	require.False(t, stillPending)
}

func startableGoalLoop() *domain.GoalLoop {
	return &domain.GoalLoop{
		ID: 1, OrganizationID: 1, CreatedByID: 2,
		Slug: "checkout-fix", Name: "checkout-fix", Status: domain.StatusDraft,
		WorkerSpecSnapshotID: 42, Objective: "fix checkout", AcceptanceCriteria: []byte(`["tests pass"]`),
		VerificationCommand: "pnpm test", MaxIterations: 5, TimeoutMinutes: 60,
		NoProgressLimit: 3, SameErrorLimit: 2, EscalationPolicy: domain.EscalationPause,
	}
}

func startFailureService(
	repo domain.Repository,
	terminator *goalLoopTerminator,
) *Service {
	service := NewService(repo)
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshot: workerspecdomain.Snapshot{ID: 42, OrganizationID: 1},
	})
	service.SetWorkerTypeSnapshotValidator(&goalLoopWorkerTypeValidator{})
	service.podCreator = loopPodCreator{}
	service.podLookup = &goalLoopPodStore{pod: &agentpod.Pod{
		OrganizationID: 1, PodKey: "loop-pod", RunnerID: 7, Status: agentpod.StatusRunning,
	}}
	service.podTerminator = terminator
	service.verificationSender = loopVerificationDispatcher{}
	service.promptSender = &recordingPromptDispatcher{}
	return service
}

type failingGoalLoopUpdateRepo struct {
	*goalLoopTestRepo
	updateCalls int
	failOnCall  int
}

func (r *failingGoalLoopUpdateRepo) TransitionStatus(
	ctx context.Context,
	id int64,
	from []string,
	updates map[string]any,
) (bool, error) {
	r.updateCalls++
	if r.updateCalls == r.failOnCall {
		return false, errors.New("database write failed")
	}
	return r.goalLoopTestRepo.TransitionStatus(ctx, id, from, updates)
}

func (r *failingGoalLoopUpdateRepo) Update(ctx context.Context, id int64, updates map[string]any) error {
	r.updateCalls++
	if r.updateCalls == r.failOnCall {
		return errors.New("database write failed")
	}
	return r.goalLoopTestRepo.Update(ctx, id, updates)
}

type lostStartClaimRepo struct {
	*goalLoopTestRepo
}

func (r *lostStartClaimRepo) TransitionStatus(
	context.Context, int64, []string, map[string]any,
) (bool, error) {
	return false, nil
}

type cancelDuringStartRepo struct {
	*goalLoopTestRepo
	transitionCalls int
}

func (r *cancelDuringStartRepo) TransitionStatus(
	ctx context.Context,
	id int64,
	from []string,
	updates map[string]any,
) (bool, error) {
	r.transitionCalls++
	if r.transitionCalls == 2 {
		r.loops[id].Status = domain.StatusCancelled
		return false, nil
	}
	return r.goalLoopTestRepo.TransitionStatus(ctx, id, from, updates)
}

type loopPodCreator struct{}

func (loopPodCreator) CreatePod(
	context.Context, *agentpodsvc.OrchestrateCreatePodRequest,
) (*agentpodsvc.OrchestrateCreatePodResult, error) {
	return &agentpodsvc.OrchestrateCreatePodResult{Pod: &agentpod.Pod{
		OrganizationID: 1, PodKey: "loop-pod", RunnerID: 7, Status: agentpod.StatusRunning,
	}}, nil
}

type countingLoopPodCreator struct {
	calls int
}

func (c *countingLoopPodCreator) CreatePod(
	context.Context, *agentpodsvc.OrchestrateCreatePodRequest,
) (*agentpodsvc.OrchestrateCreatePodResult, error) {
	c.calls++
	return (&loopPodCreator{}).CreatePod(context.Background(), nil)
}

type loopVerificationDispatcher struct{}

func (loopVerificationDispatcher) SendRunVerification(
	context.Context, int64, *runnerv1.RunVerificationCommand,
) error {
	return nil
}
