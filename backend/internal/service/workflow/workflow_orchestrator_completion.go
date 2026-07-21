package workflow

import (
	"context"
	"time"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
)

func (o *WorkflowOrchestrator) SetRunPodKey(ctx context.Context, runID int64, podKey string, autopilotKey string) error {
	bound, err := o.workflowRunService.BindPod(ctx, runID, podKey, autopilotKey)
	if err != nil {
		o.logger.Error("failed to set run pod key", "run_id", runID, "pod_key", podKey, "error", err)
		return err
	}
	if !bound {
		return ErrRunStartLost
	}
	o.logger.Info("run pod key set", "run_id", runID, "pod_key", podKey, "autopilot_key", autopilotKey)
	return nil
}

// MarkRunFailed is the no-Pod fallback path — bypasses Pod SSOT.
func (o *WorkflowOrchestrator) MarkRunFailed(ctx context.Context, runID int64, errorMessage string) error {
	o.logger.Warn("marking run as failed", "run_id", runID, "error_message", errorMessage)
	return o.markRunTerminal(ctx, runID, workflowDomain.RunStatusFailed, errorMessage)
}

func (o *WorkflowOrchestrator) MarkRunCancelled(ctx context.Context, runID int64, reason string) error {
	o.logger.Info("marking run as cancelled", "run_id", runID, "reason", reason)
	return o.markRunTerminal(ctx, runID, workflowDomain.RunStatusCancelled, reason)
}

// markRunTerminal uses FinishRun's WHERE finished_at IS NULL guard for idempotency under concurrent calls.
func (o *WorkflowOrchestrator) markRunTerminal(ctx context.Context, runID int64, status string, errorMessage string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        status,
		"finished_at":   now,
		"error_message": errorMessage,
	}
	updated, err := o.workflowRunService.FinishRun(ctx, runID, updates)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}

	run, _ := o.workflowRunService.GetByID(ctx, runID)
	if run != nil {
		o.publishRunEvent(run.OrganizationID, eventbus.EventWorkflowRunFailed, run)
		_ = o.workflowService.UpdateRunStats(ctx, run.WorkflowID, status, now)
	}
	return nil
}

func (o *WorkflowOrchestrator) HandleRunCompleted(ctx context.Context, run *workflowDomain.WorkflowRun, effectiveStatus string) {
	now := time.Now()

	// FinishRun's WHERE finished_at IS NULL is the atomic guard against double-counting
	// when concurrent events both try to complete the same run.
	runUpdates := map[string]interface{}{
		"status":      effectiveStatus,
		"finished_at": now,
	}
	if run.StartedAt != nil {
		durationSec := int(now.Sub(*run.StartedAt).Seconds())
		runUpdates["duration_sec"] = durationSec
	}
	updated, err := o.workflowRunService.FinishRun(ctx, run.ID, runUpdates)
	if err != nil {
		o.logger.Error("failed to mark run as finished",
			"run_id", run.ID, "error", err)
		return
	}
	if !updated {
		o.logger.Debug("run already finished, skipping duplicate completion",
			"run_id", run.ID)
		return
	}

	run.Status = effectiveStatus
	run.FinishedAt = &now

	if err := o.workflowService.UpdateRunStats(ctx, run.WorkflowID, effectiveStatus, now); err != nil {
		o.logger.Error("failed to update workflow run stats",
			"workflow_id", run.WorkflowID, "run_id", run.ID, "error", err)
	}

	eventType := eventbus.EventWorkflowRunCompleted
	if effectiveStatus == workflowDomain.RunStatusFailed || effectiveStatus == workflowDomain.RunStatusTimeout || effectiveStatus == workflowDomain.RunStatusCancelled {
		eventType = eventbus.EventWorkflowRunFailed
	}
	o.publishRunEvent(run.OrganizationID, eventType, run)

	manifest, err := run.PinnedExecution()
	if err != nil {
		o.logger.Error(
			"workflow run execution manifest is unavailable",
			"workflow_id",
			run.WorkflowID,
			"run_id",
			run.ID,
			"error",
			err,
		)
		return
	}
	o.applyRunCompletionEffects(ctx, run, effectiveStatus, manifest)

	o.logger.Info("workflow run completed",
		"workflow_id", run.WorkflowID,
		"run_id", run.ID,
		"effective_status", effectiveStatus,
		"pod_key", run.PodKey,
	)
}
