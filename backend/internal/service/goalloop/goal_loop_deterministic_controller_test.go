package goalloop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestWaitingWorkerDispatchesVerification(t *testing.T) {
	loop, service, verification, _ := deterministicLoopService(3, 3, 3)

	err := service.HandlePodAgentStatus(
		context.Background(), *loop.PodKey, agentpod.AgentStatusWaiting, time.Now(),
	)

	require.NoError(t, err)
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.Len(t, verification.commands, 1)
	require.Equal(t, *loop.PodKey, verification.commands[0].GetPodKey())
	require.NotEmpty(t, verification.commands[0].GetRequestId())
}

func TestFailedVerificationPromptsSameWorker(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(3, 3, 3)
	require.NoError(t, service.HandlePodAgentStatus(
		context.Background(), *loop.PodKey, agentpod.AgentStatusWaiting, time.Now(),
	))

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: verification.commands[0].GetRequestId(),
		PodKey:    *loop.PodKey,
		ExitCode:  1,
		Output:    "checkout test failed",
	})

	require.NoError(t, err)
	require.Equal(t, 1, loop.CurrentIteration)
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.Nil(t, loop.VerificationRequestID)
	require.Len(t, prompts.calls, 1)
	require.Equal(t, int64(7), prompts.calls[0].runnerID)
	require.Equal(t, *loop.PodKey, prompts.calls[0].podKey)
	require.Contains(t, prompts.calls[0].prompt, "checkout test failed")

	require.NoError(t, handleCurrentAgentStatus(
		service, *loop.PodKey, agentpod.AgentStatusExecuting,
		time.Now().Add(time.Second),
	))
	require.Equal(t, domain.StatusActive, loop.Status)
	require.Equal(t, 1, prompts.attempts)
}

func TestRepeatedWaitingAfterFailureDoesNotRunVerifierAgain(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(5, 5, 5)
	require.NoError(t, verificationFailureCycle(t, service, verification, loop))

	err := service.HandlePodAgentStatus(
		context.Background(), *loop.PodKey, agentpod.AgentStatusWaiting, time.Now(),
	)

	require.NoError(t, err)
	require.Len(t, verification.commands, 1)
	require.Len(t, prompts.calls, 1)
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.NotNil(t, loop.RetryPromptCommandID)
}

func TestStaleExecutingEventDoesNotActivateRetryIteration(t *testing.T) {
	loop, service, verification, _ := deterministicLoopService(5, 5, 5)
	require.NoError(t, verificationFailureCycle(t, service, verification, loop))
	require.NotNil(t, loop.RetryPromptCreatedAt)

	err := handleCurrentAgentStatus(
		service, *loop.PodKey, agentpod.AgentStatusExecuting,
		loop.RetryPromptCreatedAt.Add(-time.Millisecond),
	)

	require.NoError(t, err)
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.NotNil(t, loop.RetryPromptCommandID)
}

func TestStaleWaitingEventDoesNotVerifyActiveRetryWorker(t *testing.T) {
	loop, service, verification, _ := deterministicLoopService(5, 5, 5)
	require.NoError(t, verificationFailureCycle(t, service, verification, loop))
	retryCreatedAt := *loop.RetryPromptCreatedAt
	podStore := service.podLookup.(*goalLoopPodStore)
	podStore.pod.AgentStatus = agentpod.AgentStatusExecuting
	require.NoError(t, handleCurrentAgentStatus(
		service, *loop.PodKey, agentpod.AgentStatusExecuting,
		retryCreatedAt.Add(time.Second),
	))

	err := service.HandlePodAgentStatus(
		context.Background(), *loop.PodKey, agentpod.AgentStatusWaiting,
		retryCreatedAt.Add(-time.Millisecond),
	)

	require.NoError(t, err)
	require.Equal(t, domain.StatusActive, loop.Status)
	require.Len(t, verification.commands, 1)
}

func TestRepeatedVerificationErrorEscalatesWithoutAnotherPrompt(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(5, 5, 2)
	first := verificationFailureCycle(t, service, verification, loop)
	require.NoError(t, first)
	require.Len(t, prompts.calls, 1)
	require.NoError(t, handleCurrentAgentStatus(
		service, *loop.PodKey, agentpod.AgentStatusExecuting,
		time.Now().Add(time.Second),
	))

	second := verificationFailureCycle(t, service, verification, loop)

	require.NoError(t, second)
	require.Equal(t, domain.StatusPaused, loop.Status)
	require.Equal(t, 2, loop.CurrentIteration)
	require.Len(t, prompts.calls, 1)
	require.Contains(t, stringValue(loop.VerificationError), "same verification error")
}

func TestMaxIterationsEscalatesWithoutPrompt(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(1, 5, 5)

	err := verificationFailureCycle(t, service, verification, loop)

	require.NoError(t, err)
	require.Equal(t, domain.StatusPaused, loop.Status)
	require.Equal(t, 1, loop.CurrentIteration)
	require.Empty(t, prompts.calls)
	require.Contains(t, stringValue(loop.VerificationError), "maximum iterations")
}

func TestNoVerificationProgressEscalates(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(5, 2, 5)
	require.NoError(t, verificationFailureCycle(t, service, verification, loop))
	require.NoError(t, handleCurrentAgentStatus(
		service, *loop.PodKey, agentpod.AgentStatusExecuting,
		time.Now().Add(time.Second),
	))

	err := verificationFailureCycle(t, service, verification, loop)

	require.NoError(t, err)
	require.Equal(t, domain.StatusPaused, loop.Status)
	require.Len(t, prompts.calls, 1)
	require.Contains(t, stringValue(loop.VerificationError), "no verification progress")
}

func TestDuplicateVerificationResultDoesNotPromptTwice(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(5, 5, 5)
	require.NoError(t, service.HandlePodAgentStatus(
		context.Background(), *loop.PodKey, agentpod.AgentStatusWaiting, time.Now(),
	))
	result := &runnerv1.VerificationResultEvent{
		RequestId: verification.commands[0].GetRequestId(),
		PodKey:    *loop.PodKey,
		ExitCode:  1,
		Output:    "checkout test failed",
	}

	require.NoError(t, service.HandleVerificationResult(context.Background(), 7, result))
	require.NoError(t, service.HandleVerificationResult(context.Background(), 7, result))

	require.Len(t, prompts.calls, 1)
	require.Equal(t, 1, loop.CurrentIteration)
}

func TestConcurrentDuplicateVerificationResultQueuesOnePrompt(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(5, 5, 5)
	repo := newConcurrentVerificationRepo(loop)
	service.repo = repo
	require.NoError(t, service.HandlePodAgentStatus(
		context.Background(), *loop.PodKey, agentpod.AgentStatusWaiting, time.Now(),
	))
	result := &runnerv1.VerificationResultEvent{
		RequestId: verification.commands[0].GetRequestId(),
		PodKey:    *loop.PodKey,
		ExitCode:  1,
		Output:    "checkout test failed",
	}
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			errs <- service.HandleVerificationResult(context.Background(), 7, result)
		}()
	}
	repo.waitForReaders()

	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	require.Len(t, prompts.calls, 1)
	require.Equal(t, 1, loop.CurrentIteration)
}

func TestPromptQueueFailureRemainsRecoverable(t *testing.T) {
	loop, service, verification, prompts := deterministicLoopService(3, 3, 3)
	prompts.err = errors.New("runner disconnected")

	err := verificationFailureCycle(t, service, verification, loop)

	require.ErrorContains(t, err, "runner disconnected")
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.NotNil(t, loop.RetryPromptCommandID)

	prompts.err = nil
	require.NoError(t, service.RecoverPendingRetryPrompts(context.Background()))
	require.Len(t, prompts.calls, 1)
	require.Equal(t, 2, prompts.attempts)
}

func TestRetryPromptRecoveryRequiresExecutionDependencies(t *testing.T) {
	loop, service, verification, _ := deterministicLoopService(3, 3, 3)
	require.NoError(t, verificationFailureCycle(t, service, verification, loop))
	service.podLookup = nil

	err := service.RecoverPendingRetryPrompts(context.Background())

	require.ErrorIs(t, err, ErrExecutionUnavailable)
}

func TestCancelClearsPendingIterationState(t *testing.T) {
	loop, service, verification, _ := deterministicLoopService(3, 3, 3)
	require.NoError(t, verificationFailureCycle(t, service, verification, loop))
	require.NotNil(t, loop.RetryPromptCommandID)

	cancelled, err := service.Cancel(context.Background(), loop.OrganizationID, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusCancelled, cancelled.Status)
	require.Nil(t, cancelled.VerificationRequestID)
	require.Nil(t, cancelled.RetryPromptCommandID)
	require.Nil(t, cancelled.RetryPromptCreatedAt)
}
