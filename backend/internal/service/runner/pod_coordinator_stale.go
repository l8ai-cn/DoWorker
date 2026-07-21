package runner

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func (pc *PodCoordinator) MarkStaleAsDisconnected(ctx context.Context, threshold time.Time) (int64, error) {
	keys, err := pc.podStore.ListStaleActivePodKeys(ctx, threshold)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := pc.podStore.MarkStaleAsDisconnected(ctx, threshold)
	if err != nil {
		return rowsAffected, err
	}

	for _, podKey := range keys {
		pc.podRouter.UnregisterPod(podKey)
		pc.notifyStatusChange(podKey, agentpod.StatusDisconnected, "")
	}

	if rowsAffected > 0 {
		pc.logger.Info("marked stale pods as disconnected", "count", rowsAffected)
	}
	return rowsAffected, nil
}

func (pc *PodCoordinator) CleanupStaleTerminal(ctx context.Context, threshold time.Time) (int64, error) {
	keys, err := pc.podStore.ListStaleRecoverablePodKeys(ctx, threshold)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := pc.podStore.CleanupStale(ctx, threshold)
	if err != nil {
		return rowsAffected, err
	}

	for _, podKey := range keys {
		pc.emitPodReleased(ctx, podKey, 0, agentpod.StatusTerminated)
	}

	if rowsAffected > 0 {
		pc.logger.Info("terminated stale disconnected/orphaned pods", "count", rowsAffected)
	}
	return rowsAffected, nil
}
