package coordinator

import (
	"context"
	"errors"
	"fmt"
	"time"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
	agentpodSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
)

const coordinatorCompensationTimeout = 15 * time.Second

func (s *Service) attachExecutionPod(
	ctx context.Context,
	execution *coordinatordom.Execution,
	pod *agentpodSvc.OrchestrateCreatePodResult,
) error {
	err := s.store.UpdateExecution(ctx, execution.ID, map[string]any{
		"pod_id":     pod.Pod.ID,
		"pod_key":    pod.Pod.PodKey,
		"status":     coordinatordom.ExecutionStatusRunning,
		"stage":      "dispatched",
		"started_at": pod.Pod.CreatedAt,
	})
	if err == nil {
		return nil
	}

	cleanupCtx, cancel := coordinatorCompensationContext(ctx)
	defer cancel()
	terminateErr := s.podTerminator.TerminatePod(cleanupCtx, pod.Pod.PodKey)
	failErr := s.markExecutionFailed(cleanupCtx, execution.ID, "attachment_failed", err)
	return fmt.Errorf("attach execution pod: %w", errors.Join(err, terminateErr, failErr))
}

func (s *Service) failClaimedExecution(
	ctx context.Context,
	execution *coordinatordom.Execution,
	stage string,
	cause error,
) error {
	cleanupCtx, cancel := coordinatorCompensationContext(ctx)
	defer cancel()
	return errors.Join(cause, s.markExecutionFailed(cleanupCtx, execution.ID, stage, cause))
}

func (s *Service) markExecutionFailed(ctx context.Context, id int64, stage string, cause error) error {
	return s.store.UpdateExecution(ctx, id, map[string]any{
		"status":      coordinatordom.ExecutionStatusFailed,
		"stage":       stage,
		"error":       cause.Error(),
		"finished_at": time.Now(),
	})
}

func coordinatorCompensationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), coordinatorCompensationTimeout)
}
