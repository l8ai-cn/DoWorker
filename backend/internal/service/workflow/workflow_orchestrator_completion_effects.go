package workflow

import (
	"context"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

func (o *WorkflowOrchestrator) applyRunCompletionEffects(
	ctx context.Context,
	run *workflowDomain.WorkflowRun,
	effectiveStatus string,
	manifest workflowDomain.WorkflowRunExecutionManifest,
) {
	if run.PodKey != nil &&
		manifest.SandboxStrategy == workflowDomain.SandboxStrategyPersistent &&
		run.OrchestrationResourceRevision != nil {
		o.applyPersistentRuntimeState(
			ctx,
			run,
			effectiveStatus,
			*run.OrchestrationResourceRevision,
		)
	}
	if manifest.CallbackURL != "" {
		go o.sendWebhookCallback(
			manifest.CallbackURL,
			manifest,
			run,
			effectiveStatus,
		)
	}
	if manifest.TicketID != nil && o.ticketService != nil {
		go o.postTicketComment(
			context.Background(),
			*manifest.TicketID,
			manifest.CreatedByID,
			manifest,
			run,
			effectiveStatus,
		)
	}
	if manifest.MaxRetainedRuns <= 0 {
		return
	}
	deleted, err := o.workflowRunService.DeleteOldFinishedRuns(
		ctx,
		run.WorkflowID,
		manifest.MaxRetainedRuns,
	)
	if err != nil {
		o.logger.Error(
			"failed to trim old workflow runs",
			"workflow_id",
			run.WorkflowID,
			"max_retained",
			manifest.MaxRetainedRuns,
			"error",
			err,
		)
		return
	}
	if deleted > 0 {
		o.logger.Info(
			"trimmed old workflow runs",
			"workflow_id",
			run.WorkflowID,
			"deleted",
			deleted,
			"max_retained",
			manifest.MaxRetainedRuns,
		)
	}
}

func (o *WorkflowOrchestrator) applyPersistentRuntimeState(
	ctx context.Context,
	run *workflowDomain.WorkflowRun,
	effectiveStatus string,
	resourceRevision int64,
) {
	switch effectiveStatus {
	case workflowDomain.RunStatusCompleted:
		_, err := o.workflowService.UpdateRuntimeStateForRevision(
			ctx,
			run.WorkflowID,
			resourceRevision,
			run.PodKey,
		)
		if err != nil {
			o.logger.Error(
				"failed to update workflow runtime state",
				"workflow_id",
				run.WorkflowID,
				"error",
				err,
			)
		}
	case workflowDomain.RunStatusFailed:
		_, err := o.workflowService.ClearRuntimeStateForRevision(
			ctx,
			run.WorkflowID,
			resourceRevision,
		)
		if err != nil {
			o.logger.Error(
				"failed to clear workflow runtime state after failure",
				"workflow_id",
				run.WorkflowID,
				"error",
				err,
			)
		}
	}
}
