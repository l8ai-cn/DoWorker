package goalloop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (s *Service) Verify(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error) {
	loop, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if loop.Status != domain.StatusActive && loop.Status != domain.StatusVerifying {
		return nil, ErrInvalidState
	}
	if _, pending := loop.PendingPodCleanup(); pending {
		if err := s.retryPendingPodCleanup(ctx, loop); err != nil {
			return nil, err
		}
		return s.GetBySlug(ctx, orgID, slug)
	}
	if loop.Status == domain.StatusVerifying && loop.VerificationExitCode != nil {
		if err := s.finishPersistedVerification(ctx, loop); err != nil {
			return nil, err
		}
		return s.GetBySlug(ctx, orgID, slug)
	}
	if err := s.beginVerification(ctx, loop); err != nil {
		return nil, err
	}
	return s.GetBySlug(ctx, orgID, slug)
}

func (s *Service) finishPersistedVerification(
	ctx context.Context,
	loop *domain.GoalLoop,
) error {
	if persistedVerificationSucceeded(loop) {
		return s.completeVerifiedLoop(ctx, loop, nil)
	}
	reason := fmt.Sprintf("verification exited with code %d", *loop.VerificationExitCode)
	if loop.VerificationError != nil && strings.TrimSpace(*loop.VerificationError) != "" {
		reason = strings.TrimSpace(*loop.VerificationError)
	}
	return s.escalate(ctx, loop, reason, nil)
}

func persistedVerificationSucceeded(loop *domain.GoalLoop) bool {
	return loop.VerificationExitCode != nil &&
		*loop.VerificationExitCode == 0 &&
		(loop.VerificationError == nil || strings.TrimSpace(*loop.VerificationError) == "")
}

func (s *Service) HandleVerificationResult(
	ctx context.Context,
	runnerID int64,
	result *runnerv1.VerificationResultEvent,
) error {
	if result == nil || strings.TrimSpace(result.GetRequestId()) == "" {
		return ErrInvalidInput
	}
	loop, err := s.repo.GetByVerificationRequestID(ctx, result.GetRequestId())
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if loop.Status != domain.StatusVerifying || loop.PodKey == nil || *loop.PodKey != result.GetPodKey() {
		return nil
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	if pod.OrganizationID != loop.OrganizationID || pod.RunnerID != runnerID {
		return ErrInvalidInput
	}
	now := time.Now()
	output := truncateOutput(result.GetOutput())
	updates := map[string]any{
		"verification_exit_code": int(result.GetExitCode()), "verification_output": output,
		"verification_output_truncated": result.GetOutputTruncated() || len(result.GetOutput()) > len(output),
		"verification_error":            nullableString(result.GetError()), "verified_at": now,
	}
	if result.GetError() == "" && result.GetExitCode() == 0 {
		return s.completeVerifiedLoop(ctx, loop, updates)
	}
	return s.escalate(ctx, loop, verificationFailureReason(result), updates)
}

func (s *Service) beginVerification(ctx context.Context, loop *domain.GoalLoop) error {
	if loop.Status == domain.StatusVerifying || loop.IsTerminal() || loop.Status == domain.StatusPaused {
		return nil
	}
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
	requestID := uuid.NewString()
	claimed, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusActive,
	}, map[string]any{
		"status": domain.StatusVerifying, "verification_request_id": requestID,
		"verification_error": nil,
	})
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if err := s.verificationSender.SendRunVerification(ctx, pod.RunnerID, &runnerv1.RunVerificationCommand{
		RequestId: requestID, PodKey: pod.PodKey, Command: loop.VerificationCommand,
		TimeoutSeconds: int32(min(loop.TimeoutMinutes*60, 900)),
	}); err != nil {
		return s.escalate(ctx, loop, fmt.Sprintf("verification dispatch failed: %v", err), nil)
	}
	return nil
}
