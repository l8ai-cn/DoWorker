package runner

import "context"

func (pc *PodCoordinator) cancelPendingForPod(ctx context.Context, podKey string) {
	if pc.pendingQueue != nil {
		_ = pc.pendingQueue.CancelByPodKey(ctx, podKey)
	}
}
