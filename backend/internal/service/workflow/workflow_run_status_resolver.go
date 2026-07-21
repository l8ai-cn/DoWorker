package workflow

import (
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

// ResolveRunStatus derives effective run.Status from Pod (SSOT) + Autopilot phase.
func ResolveRunStatus(run *workflowDomain.WorkflowRun, podStatus string, autopilotPhase string, podFinishedAt *time.Time) {
	if run.PodKey == nil {
		return
	}

	run.Status = DeriveRunStatus(podStatus, autopilotPhase)

	if podFinishedAt != nil {
		run.FinishedAt = podFinishedAt
		if run.StartedAt != nil {
			d := int(podFinishedAt.Sub(*run.StartedAt).Seconds())
			run.DurationSec = &d
		}
	}
}

func DeriveRunStatus(podStatus string, autopilotPhase string) string {
	if autopilotPhase != "" {
		switch autopilotPhase {
		case agentpod.AutopilotPhaseCompleted:
			return workflowDomain.RunStatusCompleted
		case agentpod.AutopilotPhaseFailed:
			return workflowDomain.RunStatusFailed
		case agentpod.AutopilotPhaseStopped:
			return workflowDomain.RunStatusCancelled
		case agentpod.AutopilotPhaseMaxIterations:
			// MaxIterations = "best-effort within iteration quota" → still counts as completed.
			return workflowDomain.RunStatusCompleted
		default:
			// Pod status is the SSOT — overrides any non-terminal autopilot phase.
			if isPodDoneForLoop(podStatus) {
				return deriveFromPodStatus(podStatus)
			}
			return workflowDomain.RunStatusRunning
		}
	}

	if isPodDoneForLoop(podStatus) {
		return deriveFromPodStatus(podStatus)
	}
	return workflowDomain.RunStatusRunning
}

func isPodDoneForLoop(podStatus string) bool {
	return podStatus == agentpod.StatusCompleted ||
		podStatus == agentpod.StatusTerminated ||
		podStatus == agentpod.StatusError
}

func deriveFromPodStatus(podStatus string) string {
	switch podStatus {
	case agentpod.StatusCompleted:
		return workflowDomain.RunStatusCompleted
	case agentpod.StatusTerminated:
		return workflowDomain.RunStatusCancelled
	case agentpod.StatusError:
		return workflowDomain.RunStatusFailed
	default:
		return workflowDomain.RunStatusFailed
	}
}
