package goalloop

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func TestVerificationSuccessKeepsLoopNonTerminalWhenPodStopFails(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	service := verificationServiceWithStopError(loop)

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  0,
	})

	require.ErrorContains(t, err, "runner unavailable")
	require.Equal(t, domain.StatusVerifying, loop.Status)
	pending, ok := loop.PendingPodCleanup()
	require.True(t, ok)
	require.Equal(t, domain.StatusCompleted, pending.TargetStatus)
	require.Nil(t, loop.CompletedAt)
}

func TestVerificationFailureKeepsLoopNonTerminalWhenPodStopFails(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	service := verificationServiceWithStopError(loop)

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  1,
	})

	require.ErrorContains(t, err, "runner unavailable")
	require.Equal(t, domain.StatusVerifying, loop.Status)
	pending, ok := loop.PendingPodCleanup()
	require.True(t, ok)
	require.Equal(t, domain.StatusPaused, pending.TargetStatus)
}

func TestVerifyRetriesCleanupAfterSuccessfulEvidenceWasPersisted(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	repo := newGoalLoopTestRepo(loop)
	terminator := &goalLoopTerminator{err: errors.New("runner unavailable")}
	service := NewService(repo)
	service.podLookup = &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	service.podTerminator = terminator

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  0,
		Output:    "tests passed",
	})
	require.Error(t, err)
	require.Equal(t, 0, *loop.VerificationExitCode)
	require.Equal(t, "tests passed", stringValue(loop.VerificationOutput))

	terminator.err = nil
	updated, err := service.Verify(context.Background(), 1, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusCompleted, updated.Status)
	require.Empty(t, stringValue(updated.VerificationError))
}

func TestVerifyRetriesCleanupAfterFailedEvidenceWasPersisted(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	repo := newGoalLoopTestRepo(loop)
	terminator := &goalLoopTerminator{err: errors.New("runner unavailable")}
	service := NewService(repo)
	service.podLookup = &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	service.podTerminator = terminator

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  1,
		Output:    "tests failed",
	})
	require.Error(t, err)
	require.Equal(t, 1, *loop.VerificationExitCode)

	terminator.err = nil
	updated, err := service.Verify(context.Background(), 1, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusPaused, updated.Status)
	require.Equal(t, "verification exited with code 1", stringValue(updated.VerificationError))
}

func TestVerifyDoesNotCompleteRunnerErrorWithZeroExitCode(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	repo := newGoalLoopTestRepo(loop)
	terminator := &goalLoopTerminator{err: errors.New("runner unavailable")}
	service := NewService(repo)
	service.podLookup = &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	service.podTerminator = terminator

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  0,
		Error:     "verification process crashed",
	})
	require.Error(t, err)

	terminator.err = nil
	updated, err := service.Verify(context.Background(), 1, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusPaused, updated.Status)
	require.Equal(
		t,
		"verification failed: verification process crashed",
		stringValue(updated.VerificationError),
	)
}

func TestVerifyFinishesPersistedFailedEvidence(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	exitCode := 1
	reason := "verification exited with code 1"
	loop.VerificationExitCode = &exitCode
	loop.VerificationError = &reason
	service := NewService(newGoalLoopTestRepo(loop))
	service.podLookup = &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	service.podTerminator = &goalLoopTerminator{}

	updated, err := service.Verify(context.Background(), 1, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusPaused, updated.Status)
	require.Equal(t, reason, stringValue(updated.VerificationError))
}

func TestVerificationCompletionDoesNotOverwriteConcurrentCancel(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	repo := &cancelBeforeTerminalRepo{goalLoopTestRepo: newGoalLoopTestRepo(loop)}
	service := NewService(repo)
	service.podLookup = &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	service.podTerminator = &goalLoopTerminator{}

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  0,
	})

	require.NoError(t, err)
	require.Equal(t, domain.StatusCancelled, loop.Status)
}

func TestTruncateOutputPreservesUTF8(t *testing.T) {
	output := strings.Repeat("a", (64<<10)-1) + "界"

	truncated := truncateOutput(output)

	require.True(t, utf8.ValidString(truncated))
	require.LessOrEqual(t, len(truncated), 64<<10)
}

func TestTruncateOutputRepairsInvalidUTF8AtExactLimit(t *testing.T) {
	output := strings.Repeat("a", (64<<10)-1) + string([]byte{0xe7})

	truncated := truncateOutput(output)

	require.True(t, utf8.ValidString(truncated))
	require.Len(t, truncated, (64<<10)-1)
}

func TestBeginVerificationDoesNotDispatchWhenStatusClaimIsLost(t *testing.T) {
	podKey := "goal-loop-pod"
	loop := &domain.GoalLoop{
		ID: 1, OrganizationID: 1, Status: domain.StatusActive,
		PodKey: &podKey, VerificationCommand: "go test ./...", TimeoutMinutes: 60,
	}
	repo := &lostVerificationClaimRepo{goalLoopTestRepo: newGoalLoopTestRepo(loop)}
	dispatcher := &countingVerificationDispatcher{}
	service := NewService(repo)
	service.podLookup = &goalLoopPodStore{pod: runningPod(podKey)}
	service.verificationSender = dispatcher

	err := service.beginVerification(context.Background(), loop)

	require.NoError(t, err)
	require.Zero(t, dispatcher.calls)
}

func TestBeginVerificationClearsPreviousAttemptEvidence(t *testing.T) {
	podKey := "goal-loop-pod"
	exitCode := 1
	output := "previous failure"
	verifiedAt := time.Now()
	loop := &domain.GoalLoop{
		ID: 1, OrganizationID: 1, Status: domain.StatusActive,
		PodKey: &podKey, VerificationCommand: "go test ./...", TimeoutMinutes: 60,
		VerificationExitCode: &exitCode, VerificationOutput: &output,
		VerificationOutputTruncated: true, VerifiedAt: &verifiedAt,
	}
	dispatcher := &countingVerificationDispatcher{}
	service := NewService(newGoalLoopTestRepo(loop))
	service.podLookup = &goalLoopPodStore{pod: runningPod(podKey)}
	service.verificationSender = dispatcher

	err := service.beginVerification(context.Background(), loop)

	require.NoError(t, err)
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.Nil(t, loop.VerificationExitCode)
	require.Nil(t, loop.VerificationOutput)
	require.False(t, loop.VerificationOutputTruncated)
	require.Nil(t, loop.VerifiedAt)
	require.Equal(t, 1, dispatcher.calls)
}

func TestVerifyRedispatchesPersistedVerificationRequest(t *testing.T) {
	podKey := "goal-loop-pod"
	requestID := "verify-persisted"
	loop := &domain.GoalLoop{
		ID: 1, OrganizationID: 1, Slug: "verify-recovery",
		Status: domain.StatusVerifying, PodKey: &podKey,
		VerificationRequestID: &requestID,
		VerificationCommand:   "go test ./...",
		TimeoutMinutes:        60,
	}
	dispatcher := &countingVerificationDispatcher{}
	service := NewService(newGoalLoopTestRepo(loop))
	service.podLookup = &goalLoopPodStore{pod: runningPod(podKey)}
	service.verificationSender = dispatcher

	updated, err := service.Verify(context.Background(), loop.OrganizationID, loop.Slug)

	require.NoError(t, err)
	require.Equal(t, domain.StatusVerifying, updated.Status)
	require.Len(t, dispatcher.commands, 1)
	require.Equal(t, requestID, dispatcher.commands[0].GetRequestId())
}

func TestVerificationDispatchFailureRemainsRecoverable(t *testing.T) {
	podKey := "goal-loop-pod"
	loop := &domain.GoalLoop{
		ID: 1, OrganizationID: 1, Slug: "verify-recovery",
		Status: domain.StatusActive, PodKey: &podKey,
		VerificationCommand: "go test ./...", TimeoutMinutes: 60,
	}
	dispatcher := &countingVerificationDispatcher{err: errors.New("runner disconnected")}
	service := NewService(newGoalLoopTestRepo(loop))
	service.podLookup = &goalLoopPodStore{pod: runningPod(podKey)}
	service.verificationSender = dispatcher

	err := service.beginVerification(context.Background(), loop)

	require.ErrorContains(t, err, "runner disconnected")
	require.Equal(t, domain.StatusVerifying, loop.Status)
	require.NotNil(t, loop.VerificationRequestID)
	require.Equal(t, 1, dispatcher.calls)

	dispatcher.err = nil
	require.NoError(t, service.RecoverPendingVerifications(context.Background()))
	require.Equal(t, 2, dispatcher.calls)
	require.Equal(t, dispatcher.commands[0].GetRequestId(), dispatcher.commands[1].GetRequestId())
}

func verificationServiceWithStopError(loop *domain.GoalLoop) *Service {
	service := NewService(newGoalLoopTestRepo(loop))
	service.podLookup = &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	service.podTerminator = &goalLoopTerminator{err: errors.New("runner unavailable")}
	return service
}

type lostVerificationClaimRepo struct {
	*goalLoopTestRepo
}

func (r *lostVerificationClaimRepo) TransitionStatus(
	context.Context, int64, []string, map[string]any,
) (bool, error) {
	return false, nil
}

type countingVerificationDispatcher struct {
	calls    int
	commands []*runnerv1.RunVerificationCommand
	err      error
}

func (d *countingVerificationDispatcher) SendRunVerification(
	_ context.Context,
	_ int64,
	command *runnerv1.RunVerificationCommand,
) error {
	d.calls++
	d.commands = append(d.commands, command)
	return d.err
}

type cancelBeforeTerminalRepo struct {
	*goalLoopTestRepo
}

func (r *cancelBeforeTerminalRepo) TransitionStatus(
	ctx context.Context,
	id int64,
	from []string,
	updates map[string]any,
) (bool, error) {
	if updates["status"] == domain.StatusCompleted {
		r.loops[id].Status = domain.StatusCancelled
		return false, nil
	}
	return r.goalLoopTestRepo.TransitionStatus(ctx, id, from, updates)
}
