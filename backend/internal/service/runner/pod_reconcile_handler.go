package runner

import "context"

func (pc *PodCoordinator) reconcilePods(ctx context.Context, runnerID int64, reportedPods map[string]bool) {
	pc.reconciler.reconcile(ctx, runnerID, reportedPods)
}
