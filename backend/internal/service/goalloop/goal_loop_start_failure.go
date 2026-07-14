package goalloop

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
)

func (s *Service) abortCancelledStart(
	ctx context.Context,
	loop *domain.GoalLoop,
	podKey string,
) (*domain.GoalLoop, error) {
	if err := s.podTerminator.TerminatePod(ctx, podKey); err != nil {
		persistErr := s.persistPendingPodCleanup(
			ctx, loop.ID, podKey, domain.StatusCancelled,
			"start cancelled after pod creation", err,
		)
		return nil, errors.Join(
			ErrInvalidState,
			fmt.Errorf("terminate pod: %w", err),
			persistErr,
		)
	}
	return nil, ErrInvalidState
}

func (s *Service) failAfterPod(
	ctx context.Context, loop *domain.GoalLoop, podKey, message string, cause error,
) (*domain.GoalLoop, error) {
	startErr := fmt.Errorf("%s: %w", message, cause)
	if terminateErr := s.podTerminator.TerminatePod(ctx, podKey); terminateErr != nil {
		cleanupErr := s.failStartWithPendingCleanup(
			ctx, loop, podKey, startErr, terminateErr,
		)
		return nil, errors.Join(
			startErr,
			fmt.Errorf("terminate pod: %w", terminateErr),
			cleanupErr,
		)
	}
	return s.failStart(ctx, loop, startErr)
}

func (s *Service) failStartWithPendingCleanup(
	ctx context.Context,
	loop *domain.GoalLoop,
	podKey string,
	startErr, terminateErr error,
) error {
	updates := map[string]any{
		"status":  domain.StatusFailed,
		"pod_key": podKey,
		"verification_error": domain.EncodePendingPodCleanup(
			domain.StatusFailed, startErr.Error(), terminateErr.Error(),
		),
		"completed_at": time.Now(),
	}
	transitioned, transitionErr := s.repo.TransitionStatus(
		ctx, loop.ID, []string{domain.StatusActive}, updates,
	)
	if transitioned && transitionErr == nil {
		return nil
	}
	persistErr := s.persistPendingPodCleanup(
		ctx, loop.ID, podKey, domain.StatusFailed, startErr.Error(), terminateErr,
	)
	return errors.Join(transitionErr, persistErr)
}

func (s *Service) failStart(
	ctx context.Context, loop *domain.GoalLoop, cause error,
) (*domain.GoalLoop, error) {
	transitioned, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusActive,
	}, map[string]any{
		"status": domain.StatusFailed, "verification_error": cause.Error(),
		"completed_at": time.Now(),
	})
	if err != nil {
		return nil, errors.Join(cause, err)
	}
	if !transitioned {
		return nil, cause
	}
	return nil, cause
}
