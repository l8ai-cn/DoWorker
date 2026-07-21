package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
)

type TriggerRunRequest struct {
	WorkflowID    int64
	TriggerType   string
	TriggerSource string
	TriggerParams json.RawMessage
}

type TriggerRunResult struct {
	Run      *workflowDomain.WorkflowRun
	Workflow *workflowDomain.Workflow
	Skipped  bool
	Reason   string
}

func (o *WorkflowOrchestrator) TriggerRun(ctx context.Context, req *TriggerRunRequest) (*TriggerRunResult, error) {
	atomicResult, err := o.workflowRunService.TriggerRunAtomic(ctx, &workflowDomain.TriggerRunAtomicParams{
		WorkflowID:    req.WorkflowID,
		TriggerType:   req.TriggerType,
		TriggerSource: req.TriggerSource,
		TriggerParams: req.TriggerParams,
	})
	if err != nil {
		if errors.Is(err, workflowDomain.ErrWorkflowDisabled) {
			return nil, ErrWorkflowDisabled
		}
		return nil, err
	}

	result := &TriggerRunResult{
		Run:      atomicResult.Run,
		Workflow: atomicResult.Workflow,
		Skipped:  atomicResult.Skipped,
		Reason:   atomicResult.Reason,
	}

	if result.Run != nil && atomicResult.Workflow != nil {
		if result.Skipped {
			// Skipped runs count toward total_runs so the denormalized counter stays in sync with ComputeLoopStats (SSOT).
			_ = o.workflowService.UpdateRunStats(ctx, atomicResult.Workflow.ID, workflowDomain.RunStatusSkipped, time.Now())
		} else {
			o.publishRunEvent(atomicResult.Workflow.OrganizationID, eventbus.EventWorkflowRunStarted, result.Run)
			o.logger.Info("workflow run triggered",
				"workflow_id", atomicResult.Workflow.ID,
				"workflow_slug", atomicResult.Workflow.Slug,
				"run_id", result.Run.ID,
				"run_number", result.Run.RunNumber,
				"trigger_type", req.TriggerType,
			)
		}
	}

	return result, nil
}
