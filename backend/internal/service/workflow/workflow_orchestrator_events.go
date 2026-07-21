package workflow

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

func (o *WorkflowOrchestrator) HandlePodTerminated(ctx context.Context, podKey string, podStatus string, podFinishedAt *time.Time) {
	run, err := o.workflowRunService.FindActiveRunByPodKey(ctx, podKey)
	if err != nil {
		return
	}

	o.logger.Info("handling pod terminated for workflow run",
		"pod_key", podKey, "pod_status", podStatus, "run_id", run.ID, "workflow_id", run.WorkflowID)

	autopilotPhase := ""
	if run.AutopilotControllerKey != nil {
		autopilotPhase = o.workflowRunService.GetAutopilotPhase(ctx, *run.AutopilotControllerKey)
	}
	effectiveStatus := DeriveRunStatus(podStatus, autopilotPhase)

	if effectiveStatus == workflowDomain.RunStatusRunning {
		return
	}

	o.HandleRunCompleted(ctx, run, effectiveStatus)
}

func (o *WorkflowOrchestrator) HandleAutopilotTerminated(ctx context.Context, autopilotKey string, phase string) {
	if !agentpod.IsAutopilotPhaseTerminal(phase) {
		return
	}

	run, err := o.workflowRunService.FindActiveRunByAutopilotKey(ctx, autopilotKey)
	if err != nil {
		return
	}

	o.logger.Info("handling autopilot terminated for workflow run",
		"autopilot_key", autopilotKey, "phase", phase, "run_id", run.ID, "workflow_id", run.WorkflowID)

	effectiveStatus := DeriveRunStatus("", phase)

	o.HandleRunCompleted(ctx, run, effectiveStatus)
}
