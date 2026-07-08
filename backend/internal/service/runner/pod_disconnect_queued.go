package runner

import (
	"context"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func (pc *PodCoordinator) failQueuedPodsForRunner(ctx context.Context, runnerID int64) {
	pods, err := pc.podStore.ListActiveByRunner(ctx, runnerID)
	if err != nil {
		pc.logger.Error("failed to list active pods for disconnected runner",
			"runner_id", runnerID, "error", err)
		return
	}

	now := time.Now()
	for _, pod := range pods {
		if pod.Status != agentpod.StatusQueued {
			continue
		}
		pc.cancelPendingForPod(ctx, pod.PodKey)
		rowsAffected, err := pc.podStore.UpdateByKeyAndStatusCounted(ctx, pod.PodKey, agentpod.StatusQueued, map[string]interface{}{
			"status":        agentpod.StatusError,
			"error_code":    ErrCodeRunnerDisconnected,
			"error_message": "Runner disconnected while pod was queued.",
			"finished_at":   now,
		})
		if err != nil {
			pc.logger.Error("failed to fail queued pod on disconnect",
				"pod_key", pod.PodKey, "error", err)
			continue
		}
		if rowsAffected > 0 {
			pc.emitPodReleased(ctx, pod.PodKey, runnerID, agentpod.StatusError)
			pc.logger.Warn("queued pod failed due to runner disconnect",
				"pod_key", pod.PodKey, "runner_id", runnerID)
		}
	}
}
