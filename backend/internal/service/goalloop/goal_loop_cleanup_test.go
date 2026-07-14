package goalloop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
)

func TestExpireTimedOutRetriesPendingCleanupBeforeApplyingTimeoutPolicy(t *testing.T) {
	podKey := "goal-loop-pod"
	cleanup := domain.EncodePendingPodCleanup(
		domain.StatusCompleted,
		"verification succeeded",
		"runner unavailable",
	)
	loop := &domain.GoalLoop{
		ID: 1, OrganizationID: 1, Slug: "repair-checkout-total",
		Status: domain.StatusVerifying, PodKey: &podKey, VerificationError: &cleanup,
		EscalationPolicy: domain.EscalationFail,
	}
	repo := &timedOutGoalLoopRepo{
		goalLoopTestRepo: newGoalLoopTestRepo(loop),
		timedOut:         []*domain.GoalLoop{loop},
	}
	service := NewService(repo)
	service.podLookup = &goalLoopPodStore{pod: runningPod(podKey)}
	service.podTerminator = &goalLoopTerminator{}

	err := service.ExpireTimedOut(context.Background(), time.Now())

	require.NoError(t, err)
	require.Equal(t, domain.StatusCompleted, loop.Status)
	require.Empty(t, stringValue(loop.VerificationError))
}

func TestExpireTimedOutContinuesAfterCleanupFailure(t *testing.T) {
	first := pendingCleanupLoop(1, "first-pod")
	second := pendingCleanupLoop(2, "second-pod")
	repo := &timedOutGoalLoopRepo{
		goalLoopTestRepo: newGoalLoopTestRepo(first, second),
		timedOut:         []*domain.GoalLoop{first, second},
	}
	service := NewService(repo)
	service.podLookup = cleanupPodStore{
		"first-pod":  runningPod("first-pod"),
		"second-pod": runningPod("second-pod"),
	}
	service.podTerminator = &selectiveCleanupTerminator{failedKey: "first-pod"}

	err := service.ExpireTimedOut(context.Background(), time.Now())

	require.ErrorContains(t, err, "runner unavailable")
	require.Equal(t, domain.StatusVerifying, first.Status)
	require.Equal(t, domain.StatusCompleted, second.Status)
}

func pendingCleanupLoop(id int64, podKey string) *domain.GoalLoop {
	cleanup := domain.EncodePendingPodCleanup(
		domain.StatusCompleted,
		"verification succeeded",
		"runner unavailable",
	)
	return &domain.GoalLoop{
		ID: id, OrganizationID: 1, Slug: "repair-checkout-total",
		Status: domain.StatusVerifying, PodKey: &podKey, VerificationError: &cleanup,
		EscalationPolicy: domain.EscalationFail,
	}
}

type timedOutGoalLoopRepo struct {
	*goalLoopTestRepo
	timedOut []*domain.GoalLoop
}

func (r *timedOutGoalLoopRepo) ListTimedOut(
	context.Context,
	time.Time,
) ([]*domain.GoalLoop, error) {
	return r.timedOut, nil
}

type cleanupPodStore map[string]*agentpod.Pod

func (s cleanupPodStore) GetPod(_ context.Context, podKey string) (*agentpod.Pod, error) {
	return s[podKey], nil
}

type selectiveCleanupTerminator struct {
	failedKey string
}

func (t *selectiveCleanupTerminator) TerminatePod(_ context.Context, podKey string) error {
	if podKey == t.failedKey {
		return errors.New("runner unavailable")
	}
	return nil
}
