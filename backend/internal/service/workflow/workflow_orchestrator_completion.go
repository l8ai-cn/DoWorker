package workflow

import (
	"context"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
)

func (o *WorkflowOrchestrator) SetRunPodKey(ctx context.Context, runID int64, podKey string, autopilotKey string) error {
	updates := map[string]interface{}{
		"pod_key": podKey,
	}
	if autopilotKey != "" {
		updates["autopilot_controller_key"] = autopilotKey
	}
	if err := o.workflowRunService.UpdateStatus(ctx, runID, updates); err != nil {
		o.logger.Error("failed to set run pod key", "run_id", runID, "pod_key", podKey, "error", err)
		return err
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

	workflow, _ := o.workflowService.GetByID(ctx, run.WorkflowID)
	if run.PodKey != nil && workflow != nil && workflow.IsPersistent() {
		switch effectiveStatus {
		case workflowDomain.RunStatusCompleted:
			if err := o.workflowService.UpdateRuntimeState(ctx, run.WorkflowID, nil, run.PodKey); err != nil {
				o.logger.Error("failed to update workflow runtime state",
					"workflow_id", run.WorkflowID, "error", err)
			}
		case workflowDomain.RunStatusFailed:
			if err := o.workflowService.ClearRuntimeState(ctx, run.WorkflowID); err != nil {
				o.logger.Error("failed to clear workflow runtime state after failure",
					"workflow_id", run.WorkflowID, "error", err)
			}
			o.logger.Info("cleared persistent sandbox resume chain after run failure",
				"workflow_id", run.WorkflowID, "run_id", run.ID, "pod_key", *run.PodKey)
		}
	}

	eventType := eventbus.EventWorkflowRunCompleted
	if effectiveStatus == workflowDomain.RunStatusFailed || effectiveStatus == workflowDomain.RunStatusTimeout || effectiveStatus == workflowDomain.RunStatusCancelled {
		eventType = eventbus.EventWorkflowRunFailed
	}
	o.publishRunEvent(run.OrganizationID, eventType, run)

	if workflow != nil && workflow.CallbackURL != nil && *workflow.CallbackURL != "" {
		go o.sendWebhookCallback(*workflow.CallbackURL, workflow, run, effectiveStatus)
	}

	if workflow != nil && workflow.TicketID != nil && o.ticketService != nil {
		go o.postTicketComment(context.Background(), *workflow.TicketID, workflow.CreatedByID, workflow, run, effectiveStatus)
	}

	// Trim by workflow.MaxRetainedRuns (data retention).
	if workflow != nil && workflow.MaxRetainedRuns > 0 {
		if deleted, err := o.workflowRunService.DeleteOldFinishedRuns(ctx, workflow.ID, workflow.MaxRetainedRuns); err != nil {
			o.logger.Error("failed to trim old workflow runs",
				"workflow_id", workflow.ID, "max_retained", workflow.MaxRetainedRuns, "error", err)
		} else if deleted > 0 {
			o.logger.Info("trimmed old workflow runs",
				"workflow_id", workflow.ID, "deleted", deleted, "max_retained", workflow.MaxRetainedRuns)
		}
	}

	o.logger.Info("workflow run completed",
		"workflow_id", run.WorkflowID,
		"run_id", run.ID,
		"effective_status", effectiveStatus,
		"pod_key", run.PodKey,
	)
}
