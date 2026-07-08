package runner

import "context"

func (pc *PodCoordinator) notifyStatusChange(podKey, status, agentStatus string) {
	if pc.statusBroadcast != nil {
		pc.statusBroadcast.notify(podKey, status, agentStatus)
	}
}

func (pc *PodCoordinator) NotifyPodStatus(podKey, status, agentStatus string) {
	pc.notifyStatusChange(podKey, status, agentStatus)
}

func (pc *PodCoordinator) emitPodReleased(ctx context.Context, podKey string, runnerID int64, status string) {
	pc.podRouter.UnregisterPod(podKey)
	pc.clearMissCount(podKey)
	if status != "" {
		pc.notifyStatusChange(podKey, status, "")
	}
	if runnerID != 0 {
		pc.triggerPendingDrain(runnerID)
	}
	_ = ctx
}
