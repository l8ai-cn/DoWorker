package runner

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func (pc *PodCoordinator) terminateQueuedPod(ctx context.Context, pod *agentpod.Pod, podKey string) error {
	pc.cancelPendingForPod(ctx, podKey)
	now := time.Now()
	rowsAffected, err := pc.podStore.UpdateByKeyAndStatusCounted(ctx, podKey, agentpod.StatusQueued, map[string]interface{}{
		"status":      agentpod.StatusCompleted,
		"finished_at": now,
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPodAlreadyTerminated
	}
	pc.logger.Info("queued pod cancelled", "pod_key", podKey, "runner_id", pod.RunnerID)
	pc.emitPodReleased(ctx, podKey, pod.RunnerID, agentpod.StatusCompleted)
	return nil
}
