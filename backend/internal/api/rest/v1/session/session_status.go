package sessionapi

import podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"

func mapSessionStatus(pod *podDomain.Pod) string {
	if pod == nil {
		return "idle"
	}
	return mapPodSessionStatus(pod.Status, pod.AgentStatus)
}

func mapPodSessionStatus(podStatus, agentStatus string) string {
	switch podStatus {
	case podDomain.StatusInitializing, podDomain.StatusQueued:
		return "launching"
	case podDomain.StatusError, podDomain.StatusTerminated, podDomain.StatusOrphaned:
		return "failed"
	case podDomain.StatusCompleted, podDomain.StatusDisconnected:
		return "idle"
	case podDomain.StatusRunning:
		switch agentStatus {
		case podDomain.AgentStatusExecuting:
			return "running"
		case podDomain.AgentStatusWaiting:
			return "waiting"
		default:
			return "idle"
		}
	default:
		if agentStatus == podDomain.AgentStatusExecuting {
			return "running"
		}
		return "idle"
	}
}
