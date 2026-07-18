package infra

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
)

func deriveWorkflowRunStatus(podStatus string, autopilotPhase string) string {
	if autopilotPhase != "" {
		switch autopilotPhase {
		case agentpod.AutopilotPhaseCompleted,
			agentpod.AutopilotPhaseMaxIterations:
			return workflow.RunStatusCompleted
		case agentpod.AutopilotPhaseFailed:
			return workflow.RunStatusFailed
		case agentpod.AutopilotPhaseStopped:
			return workflow.RunStatusCancelled
		default:
			if isPodDone(podStatus) {
				return podToRunStatus(podStatus)
			}
			return workflow.RunStatusRunning
		}
	}
	if isPodDone(podStatus) {
		return podToRunStatus(podStatus)
	}
	return workflow.RunStatusRunning
}

func isPodDone(podStatus string) bool {
	return podStatus == agentpod.StatusCompleted ||
		podStatus == agentpod.StatusTerminated ||
		podStatus == agentpod.StatusError
}

func podToRunStatus(podStatus string) string {
	switch podStatus {
	case agentpod.StatusCompleted:
		return workflow.RunStatusCompleted
	case agentpod.StatusTerminated:
		return workflow.RunStatusCancelled
	case agentpod.StatusError:
		return workflow.RunStatusFailed
	default:
		return workflow.RunStatusFailed
	}
}
