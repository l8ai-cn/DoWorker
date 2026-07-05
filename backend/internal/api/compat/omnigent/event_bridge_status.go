package omnigent

import (
	"context"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func ForwardPodStatus(ctx context.Context, podKey, podStatus, agentStatus string) {
	if globalBridge == nil || globalBridge.Hub == nil {
		return
	}
	sessionID, ok := globalBridge.sessionForPod(ctx, podKey)
	if !ok {
		return
	}
	status := mapPodSessionStatus(podStatus, agentStatus)
	globalBridge.Hub.Publish(sessionID, formatSSE("session.status", map[string]any{
		"conversation_id": sessionID, "status": status,
	}))
	if globalUpdatesHub != nil {
		globalUpdatesHub.NotifyChanged(sessionID)
	}
}

var globalUpdatesHub *SessionUpdatesHub

func SetUpdatesHub(h *SessionUpdatesHub) { globalUpdatesHub = h }

func mapPodSessionStatus(podStatus, agentStatus string) string {
	switch podStatus {
	case podDomain.StatusInitializing:
		return "launching"
	case podDomain.StatusError, podDomain.StatusTerminated:
		return "failed"
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
