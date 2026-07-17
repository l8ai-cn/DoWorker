package goalloop

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (s *Service) HandlePodAgentStatus(
	ctx context.Context,
	podKey, status string,
	eventAt time.Time,
) error {
	loop, err := s.repo.GetByPodKey(ctx, podKey)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil || loop.IsTerminal() || loop.Status == domain.StatusPaused {
		return err
	}
	current, err := s.isCurrentPodAgentStatus(ctx, podKey, status)
	if err != nil || !current {
		return err
	}
	if loop.RetryPromptCommandID != nil {
		if status == agentpod.AgentStatusExecuting {
			return s.activateRetryWorker(ctx, loop, eventAt)
		}
		return s.queueRetryPrompt(ctx, loop)
	}
	if status == agentpod.AgentStatusWaiting && loop.Status == domain.StatusActive {
		return s.beginVerification(ctx, loop)
	}
	return nil
}

func (s *Service) isCurrentPodAgentStatus(
	ctx context.Context,
	podKey, status string,
) (bool, error) {
	if s.podLookup == nil {
		return false, ErrExecutionUnavailable
	}
	pod, err := s.podLookup.GetPod(ctx, podKey)
	if err != nil {
		return false, err
	}
	if pod == nil {
		return false, fmt.Errorf("goal loop pod %q not found", podKey)
	}
	return pod.AgentStatus == status, nil
}

func (s *Service) activateRetryWorker(
	ctx context.Context,
	loop *domain.GoalLoop,
	eventAt time.Time,
) error {
	if loop.RetryPromptCommandID == nil || loop.RetryPromptCreatedAt == nil ||
		eventAt.IsZero() || eventAt.Before(*loop.RetryPromptCreatedAt) {
		return nil
	}
	_, err := s.repo.TransitionRetryPrompt(
		ctx,
		loop.ID,
		*loop.RetryPromptCommandID,
		map[string]any{
			"status":                  domain.StatusActive,
			"retry_prompt_command_id": nil,
			"retry_prompt_created_at": nil,
		},
	)
	return err
}

func (s *Service) continueAfterVerificationFailure(
	ctx context.Context,
	loop *domain.GoalLoop,
	result *runnerv1.VerificationResultEvent,
	evidence map[string]any,
) error {
	iteration := loop.CurrentIteration + 1
	progressFingerprint, errorFingerprint := verificationFingerprints(result)
	noProgressCount := consecutiveFingerprintCount(
		loop.LastProgressFingerprint, progressFingerprint, loop.NoProgressCount,
	)
	sameErrorCount := consecutiveFingerprintCount(
		loop.LastErrorFingerprint, errorFingerprint, loop.SameErrorCount,
	)
	evidence["current_iteration"] = iteration
	evidence["no_progress_count"] = noProgressCount
	evidence["same_error_count"] = sameErrorCount
	evidence["last_progress_fingerprint"] = progressFingerprint
	evidence["last_error_fingerprint"] = errorFingerprint
	evidence["verification_request_id"] = nil
	evidence["verification_error"] = verificationFailureReason(result)

	if result.GetError() != "" {
		return s.consumeAndEscalate(ctx, loop, result.GetRequestId(), evidence)
	}
	if iteration >= effectiveLimit(loop.MaxIterations, 10) {
		evidence["verification_error"] = "maximum iterations reached"
		return s.consumeAndEscalate(ctx, loop, result.GetRequestId(), evidence)
	}
	if sameErrorCount >= effectiveLimit(loop.SameErrorLimit, 2) {
		evidence["verification_error"] = fmt.Sprintf(
			"same verification error repeated %d times", sameErrorCount,
		)
		return s.consumeAndEscalate(ctx, loop, result.GetRequestId(), evidence)
	}
	if noProgressCount >= effectiveLimit(loop.NoProgressLimit, 3) {
		evidence["verification_error"] = fmt.Sprintf(
			"no verification progress for %d iterations", noProgressCount,
		)
		return s.consumeAndEscalate(ctx, loop, result.GetRequestId(), evidence)
	}
	if s.promptSender == nil || !s.promptSender.Enabled() {
		return s.consumeAndEscalate(ctx, loop, result.GetRequestId(), evidence)
	}
	commandID := retryPromptCommandID(loop.ID, iteration+1)
	evidence["retry_prompt_command_id"] = commandID
	evidence["retry_prompt_created_at"] = time.Now().Truncate(time.Millisecond)
	claimed, err := s.repo.ConsumeVerificationResult(
		ctx,
		loop.ID,
		result.GetRequestId(),
		evidence,
	)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	updated, err := s.repo.GetByPodKey(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	return s.queueRetryPrompt(ctx, updated)
}

func (s *Service) consumeAndEscalate(
	ctx context.Context,
	loop *domain.GoalLoop,
	requestID string,
	evidence map[string]any,
) error {
	evidence["retry_prompt_command_id"] = nil
	evidence["retry_prompt_created_at"] = nil
	claimed, err := s.repo.ConsumeVerificationResult(ctx, loop.ID, requestID, evidence)
	if err != nil || !claimed {
		return err
	}
	return s.escalate(ctx, loop, evidence["verification_error"].(string), nil)
}

func retryPromptCommandID(loopID int64, iteration int) string {
	return fmt.Sprintf("goal-loop-%d-iteration-%d", loopID, iteration)
}
