package goalloop

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func (s *Service) beginVerification(ctx context.Context, loop *domain.GoalLoop) error {
	if loop.Status == domain.StatusVerifying || loop.IsTerminal() || loop.Status == domain.StatusPaused {
		return nil
	}
	if !s.verificationReady() || loop.PodKey == nil {
		return ErrVerificationPending
	}
	requestID := uuid.NewString()
	claimed, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusActive,
	}, map[string]any{
		"status": domain.StatusVerifying, "verification_request_id": requestID,
		"verification_exit_code": nil, "verification_output": nil,
		"verification_output_truncated": false, "verification_error": nil,
		"verified_at": nil,
	})
	if err != nil || !claimed {
		return err
	}
	return s.dispatchVerification(ctx, loop, requestID)
}

func (s *Service) dispatchVerification(
	ctx context.Context,
	loop *domain.GoalLoop,
	requestID string,
) error {
	if !s.verificationReady() || loop.PodKey == nil {
		return ErrVerificationPending
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	if pod.OrganizationID != loop.OrganizationID || pod.RunnerID <= 0 {
		return ErrInvalidInput
	}
	return s.verificationSender.SendRunVerification(ctx, pod.RunnerID, &runnerv1.RunVerificationCommand{
		RequestId: requestID, PodKey: pod.PodKey, Command: loop.VerificationCommand,
		TimeoutSeconds: int32(min(loop.TimeoutMinutes*60, 900)),
	})
}

func (s *Service) RecoverPendingVerifications(ctx context.Context) error {
	loops, err := s.repo.ListVerificationPending(ctx)
	if err != nil {
		return err
	}
	var recoveryErr error
	for _, loop := range loops {
		if loop.VerificationRequestID == nil {
			continue
		}
		if err := s.dispatchVerification(ctx, loop, *loop.VerificationRequestID); err != nil {
			recoveryErr = errors.Join(recoveryErr, err)
		}
	}
	return recoveryErr
}
