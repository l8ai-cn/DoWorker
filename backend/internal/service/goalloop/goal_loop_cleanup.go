package goalloop

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
)

func (s *Service) retryPendingPodCleanup(
	ctx context.Context,
	loop *domain.GoalLoop,
) error {
	pending, ok := loop.PendingPodCleanup()
	if !ok {
		return nil
	}
	if err := s.stopPod(ctx, loop); err != nil {
		persistErr := s.persistPendingPodCleanup(
			ctx, loop.ID, *loop.PodKey, pending.TargetStatus, pending.Reason, err,
		)
		return errors.Join(fmt.Errorf("%s: %w", pending.Reason, err), persistErr)
	}
	if loop.IsTerminal() {
		return s.clearTerminalCleanupMarker(ctx, loop, pending)
	}
	updates := cleanupTerminalUpdates(pending)
	_, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusDraft,
		domain.StatusActive,
		domain.StatusPaused,
		domain.StatusVerifying,
	}, updates)
	return err
}

func (s *Service) persistPendingPodCleanup(
	ctx context.Context,
	loopID int64,
	podKey, targetStatus, reason string,
	stopErr error,
) error {
	return s.repo.Update(ctx, loopID, map[string]any{
		"pod_key": podKey,
		"verification_error": domain.EncodePendingPodCleanup(
			targetStatus, reason, stopErr.Error(),
		),
	})
}

func (s *Service) clearTerminalCleanupMarker(
	ctx context.Context,
	loop *domain.GoalLoop,
	pending domain.PendingPodCleanup,
) error {
	var verificationError any
	if loop.Status == domain.StatusFailed && pending.TargetStatus == domain.StatusFailed {
		verificationError = pending.Reason
	}
	return s.repo.Update(ctx, loop.ID, map[string]any{
		"verification_error": verificationError,
	})
}

func cleanupTerminalUpdates(pending domain.PendingPodCleanup) map[string]any {
	updates := map[string]any{
		"status":             pending.TargetStatus,
		"verification_error": pending.Reason,
	}
	switch pending.TargetStatus {
	case domain.StatusCompleted:
		updates["verification_error"] = nil
		updates["completed_at"] = time.Now()
	case domain.StatusCancelled:
		updates["verification_error"] = nil
		updates["completed_at"] = time.Now()
	case domain.StatusFailed:
		updates["completed_at"] = time.Now()
	}
	return updates
}
