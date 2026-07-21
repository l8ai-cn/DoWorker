package workflow

import (
	"context"
	"time"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

// CheckAndTriggerCronLoops uses FOR UPDATE SKIP LOCKED in per-workflow tx so multi-instance
// deployments never double-process a single workflow.
func (s *WorkflowScheduler) CheckAndTriggerCronLoops(ctx context.Context) error {
	orgIDs := s.getOrgIDs()

	dueLoops, err := s.workflowService.GetDueCronWorkflows(ctx, orgIDs)
	if err != nil {
		s.logger.Error("failed to get due cron workflows", "error", err)
		return err
	}

	if len(dueLoops) == 0 {
		return nil
	}

	s.logger.Info("found due cron workflows", "count", len(dueLoops))

	for _, workflow := range dueLoops {
		// Compute nextRunAt before claim so ClaimCronWorkflow advances it atomically with the claim.
		var nextRunAt *time.Time
		if workflow.CronExpression != nil {
			var calcErr error
			nextRunAt, calcErr = s.CalculateNextRun(*workflow.CronExpression)
			if calcErr != nil {
				s.logger.Error("invalid cron expression, skipping workflow",
					"workflow_id", workflow.ID, "cron", *workflow.CronExpression, "error", calcErr)
				continue
			}
		}

		claimed, err := s.workflowService.ClaimCronWorkflow(ctx, workflow.ID, nextRunAt)
		if err != nil {
			s.logger.Error("failed to claim cron workflow", "workflow_id", workflow.ID, "error", err)
			continue
		}
		if !claimed {
			continue
		}

		result, err := s.orchestrator.TriggerRun(ctx, &TriggerRunRequest{
			WorkflowID:    workflow.ID,
			TriggerType:   workflowDomain.RunTriggerCron,
			TriggerSource: "cron",
		})
		if err != nil {
			s.logger.Error("failed to trigger cron workflow", "workflow_id", workflow.ID, "error", err)
			continue
		}

		if !result.Skipped && result.Run != nil && result.Workflow != nil {
			go s.orchestrator.StartRun(context.Background(), result.Workflow, result.Run, result.Workflow.CreatedByID)
		}
	}

	return nil
}
