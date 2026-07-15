package goalloop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (s *Service) escalate(
	ctx context.Context, loop *domain.GoalLoop, reason string, extra map[string]any,
) error {
	updates := copyUpdates(extra)
	updates["verification_error"] = reason
	if loop.EscalationPolicy == domain.EscalationPause {
		updates["status"] = domain.StatusPaused
	} else {
		updates["status"] = domain.StatusFailed
		updates["completed_at"] = time.Now()
	}
	if err := s.stopPod(ctx, loop); err != nil {
		return s.recordStopFailure(
			ctx, loop, updates["status"].(string), reason, err, extra,
		)
	}
	_, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusActive,
		domain.StatusVerifying,
	}, updates)
	return err
}

func (s *Service) recordStopFailure(
	ctx context.Context,
	loop *domain.GoalLoop,
	targetStatus string,
	reason string,
	stopErr error,
	extra map[string]any,
) error {
	updates := copyUpdates(extra)
	updates["verification_error"] = domain.EncodePendingPodCleanup(
		targetStatus, reason, stopErr.Error(),
	)
	err := s.repo.Update(ctx, loop.ID, updates)
	if err != nil {
		return errors.Join(stopErr, err)
	}
	return fmt.Errorf("%s: %w", reason, stopErr)
}

func (s *Service) completeVerifiedLoop(
	ctx context.Context,
	loop *domain.GoalLoop,
	evidence map[string]any,
) error {
	if err := s.stopPod(ctx, loop); err != nil {
		return s.recordStopFailure(
			ctx, loop, domain.StatusCompleted, "verification succeeded", err, evidence,
		)
	}
	updates := copyUpdates(evidence)
	updates["verification_error"] = nil
	updates["status"] = domain.StatusCompleted
	updates["completed_at"] = time.Now()
	_, err := s.repo.TransitionStatus(
		ctx,
		loop.ID,
		[]string{domain.StatusVerifying},
		updates,
	)
	return err
}

func copyUpdates(values map[string]any) map[string]any {
	updates := make(map[string]any, len(values)+3)
	for key, value := range values {
		updates[key] = value
	}
	return updates
}

func verificationFailureReason(result *runnerv1.VerificationResultEvent) string {
	if result.GetError() != "" {
		return "verification failed: " + result.GetError()
	}
	return fmt.Sprintf("verification exited with code %d", result.GetExitCode())
}

func truncateOutput(output string) string {
	const maxBytes = 64 << 10
	end := min(len(output), maxBytes)
	for end > 0 && !utf8.ValidString(output[:end]) {
		end--
	}
	return output[:end]
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
