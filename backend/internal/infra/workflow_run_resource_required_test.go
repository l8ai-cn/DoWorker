package infra

import (
	"context"
	"fmt"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/stretchr/testify/require"
)

func TestTriggerRunAtomicRejectsWorkflowWithoutCompleteResourceBinding(t *testing.T) {
	resourceID := int64(90)
	tests := []struct {
		name       string
		resourceID *int64
	}{
		{name: "absent binding"},
		{name: "partial binding", resourceID: &resourceID},
	}

	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := setupLoopTestDB(t)
			runRepo := NewWorkflowRunRepository(db)
			workflowRepo := NewWorkflowRepository(db)
			ctx := context.Background()
			row := &workflow.Workflow{
				OrganizationID:          1,
				Name:                    "Legacy workflow",
				Slug:                    fmt.Sprintf("legacy-workflow-%d", index),
				PromptTemplate:          "Review authorization",
				ExecutionMode:           workflow.ExecutionModeDirect,
				Status:                  workflow.StatusEnabled,
				SandboxStrategy:         workflow.SandboxStrategyFresh,
				ConcurrencyPolicy:       workflow.ConcurrencyPolicySkip,
				MaxConcurrentRuns:       1,
				TimeoutMinutes:          60,
				AutopilotConfig:         []byte("{}"),
				ConfigOverrides:         []byte("{}"),
				CreatedByID:             1,
				OrchestrationResourceID: test.resourceID,
			}
			require.NoError(t, workflowRepo.Create(ctx, row))

			_, err := runRepo.TriggerRunAtomic(
				ctx,
				&workflow.TriggerRunAtomicParams{
					WorkflowID:    row.ID,
					TriggerType:   workflow.RunTriggerManual,
					TriggerSource: "test",
				},
			)

			require.ErrorIs(t, err, workflow.ErrWorkflowResourceRequired)
			var runCount int64
			require.NoError(t, db.Model(&workflow.WorkflowRun{}).
				Where("workflow_id = ?", row.ID).
				Count(&runCount).Error)
			require.Zero(t, runCount)
		})
	}
}
