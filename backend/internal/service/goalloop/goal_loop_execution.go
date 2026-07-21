package goalloop

import (
	"context"
	"errors"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
)

func (s *Service) Cancel(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error) {
	loop, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if loop.IsTerminal() {
		if _, pending := loop.PendingPodCleanup(); pending {
			if err := s.retryPendingPodCleanup(ctx, loop); err != nil {
				return nil, err
			}
			return s.GetBySlug(ctx, orgID, slug)
		}
		return nil, ErrInvalidState
	}
	if err := s.stopPod(ctx, loop); err != nil {
		persistErr := s.persistPendingPodCleanup(
			ctx, loop.ID, stringValueOrEmpty(loop.PodKey), domain.StatusCancelled,
			"loop cancellation requested", err,
		)
		if persistErr != nil {
			return nil, errors.Join(err, persistErr)
		}
		return nil, err
	}
	transitioned, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusDraft,
		domain.StatusActive,
		domain.StatusPaused,
		domain.StatusVerifying,
	}, map[string]any{
		"status": domain.StatusCancelled, "completed_at": time.Now(),
		"verification_request_id": nil, "retry_prompt_command_id": nil,
		"retry_prompt_created_at": nil,
	})
	if err != nil {
		return nil, err
	}
	if !transitioned {
		return nil, ErrInvalidState
	}
	return s.GetBySlug(ctx, orgID, slug)
}

func (s *Service) HandlePodStatus(ctx context.Context, podKey, status string) error {
	loop, err := s.repo.GetByPodKey(ctx, podKey)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if loop.IsTerminal() || loop.Status == domain.StatusPaused {
		return nil
	}
	if status == agentpod.StatusCompleted {
		return s.beginVerification(ctx, loop)
	}
	if status == agentpod.StatusError || status == agentpod.StatusTerminated {
		return s.escalate(ctx, loop, "pod stopped before verification", nil)
	}
	return nil
}

func (s *Service) HandleAutopilotStatus(ctx context.Context, autopilotKey, phase string) error {
	loop, err := s.repo.GetByAutopilotControllerKey(ctx, autopilotKey)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if loop.IsTerminal() || loop.Status == domain.StatusPaused || loop.Status == domain.StatusVerifying {
		return nil
	}
	switch phase {
	case agentpod.AutopilotPhaseCompleted:
		return s.beginVerification(ctx, loop)
	case agentpod.AutopilotPhasePaused, agentpod.AutopilotPhaseWaitingApproval,
		agentpod.AutopilotPhaseUserTakeover, agentpod.AutopilotPhaseMaxIterations,
		agentpod.AutopilotPhaseFailed, agentpod.AutopilotPhaseStopped:
		return s.escalate(ctx, loop, "autopilot stopped: "+phase, nil)
	default:
		return nil
	}
}

func (s *Service) ExpireTimedOut(ctx context.Context, now time.Time) error {
	loops, err := s.repo.ListTimedOut(ctx, now)
	if err != nil {
		return err
	}
	var sweepErr error
	for _, loop := range loops {
		if _, pending := loop.PendingPodCleanup(); pending {
			if err := s.retryPendingPodCleanup(ctx, loop); err != nil {
				sweepErr = errors.Join(sweepErr, err)
			}
			continue
		}
		var err error
		if loop.Status == domain.StatusVerifying && persistedVerificationSucceeded(loop) {
			err = s.completeVerifiedLoop(ctx, loop, nil)
		} else {
			err = s.escalate(ctx, loop, "runtime budget exhausted", nil)
		}
		if err != nil {
			sweepErr = errors.Join(sweepErr, err)
		}
	}
	return errors.Join(
		sweepErr,
		s.RecoverPendingVerifications(ctx),
		s.RecoverPendingRetryPrompts(ctx),
	)
}

func (s *Service) stopPod(ctx context.Context, loop *domain.GoalLoop) error {
	if s.podLookup == nil || s.podTerminator == nil || loop.PodKey == nil {
		return nil
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	if !pod.IsActive() {
		return nil
	}
	return s.podTerminator.TerminatePod(ctx, pod.PodKey)
}

func (s *Service) verificationReady() bool {
	return s.podLookup != nil && s.verificationSender != nil
}

func stringValueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
