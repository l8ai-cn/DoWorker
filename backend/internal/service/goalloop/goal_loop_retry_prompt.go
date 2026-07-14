package goalloop

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (s *Service) queueRetryPrompt(ctx context.Context, loop *domain.GoalLoop) error {
	if loop.RetryPromptCommandID == nil || loop.PodKey == nil {
		return nil
	}
	if s.podLookup == nil || s.promptSender == nil || !s.promptSender.Enabled() {
		return ErrExecutionUnavailable
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	result := &runnerv1.VerificationResultEvent{
		ExitCode: int32Value(loop.VerificationExitCode),
		Output:   stringValueOrEmpty(loop.VerificationOutput),
	}
	err = s.promptSender.EnqueueSendPrompt(
		ctx,
		loop.OrganizationID,
		pod.RunnerID,
		*loop.PodKey,
		*loop.RetryPromptCommandID,
		buildVerificationRetryPrompt(
			loop.CurrentIteration+1,
			loop.MaxIterations,
			result,
		),
		time.Duration(loop.TimeoutMinutes)*time.Minute,
	)
	if errors.Is(err, agentpod.ErrDuplicateCommand) {
		return nil
	}
	return err
}

func (s *Service) RecoverPendingRetryPrompts(ctx context.Context) error {
	loops, err := s.repo.ListRetryPromptPending(ctx)
	if err != nil {
		return err
	}
	var recoveryErr error
	for _, loop := range loops {
		if err := s.queueRetryPrompt(ctx, loop); err != nil {
			recoveryErr = errors.Join(recoveryErr, err)
		}
	}
	return recoveryErr
}

func int32Value(value *int) int32 {
	if value == nil {
		return 0
	}
	return int32(*value)
}
