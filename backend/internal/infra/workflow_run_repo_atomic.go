package infra

import (
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"context"
)

// TriggerRunAtomic atomically creates a workflow run within a FOR UPDATE transaction.
func (r *workflowRunRepo) TriggerRunAtomic(ctx context.Context, params *workflow.TriggerRunAtomicParams) (*workflow.TriggerRunAtomicResult, error) {
	var result *workflow.TriggerRunAtomicResult

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var l workflow.Workflow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&l, params.WorkflowID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return workflow.ErrNotFound
			}
			return fmt.Errorf("failed to get workflow: %w", err)
		}

		if !l.IsEnabled() {
			return workflow.ErrWorkflowDisabled
		}

		// 2. Count active runs using Pod status (SSOT) — within the transaction
		var activeCount int64
		if err := tx.Table("workflow_runs").
			Joins("LEFT JOIN pods ON pods.pod_key = workflow_runs.pod_key").
			Where("workflow_runs.workflow_id = ?", l.ID).
			Where(
				"(workflow_runs.pod_key IS NULL AND workflow_runs.status = ?) OR "+
					"(workflow_runs.pod_key IS NOT NULL AND pods.status IN ?)",
				workflow.RunStatusPending,
				agentpod.ActiveStatuses(),
			).
			Count(&activeCount).Error; err != nil {
			return fmt.Errorf("failed to count active runs: %w", err)
		}

		if activeCount >= int64(l.MaxConcurrentRuns) {
			return r.handleConcurrencyPolicy(tx, &l, params, &result)
		}

		// 3. Get next run number atomically (inside transaction with lock)
		var maxNumber int
		if err := tx.Model(&workflow.WorkflowRun{}).
			Where("workflow_id = ?", l.ID).
			Select("COALESCE(MAX(run_number), 0)").
			Scan(&maxNumber).Error; err != nil {
			return fmt.Errorf("failed to get next run number: %w", err)
		}
		runNumber := maxNumber + 1

		resolvedPrompt := l.PromptTemplate
		now := time.Now()

		run := &workflow.WorkflowRun{
			OrganizationID:                l.OrganizationID,
			WorkflowID:                    l.ID,
			RunNumber:                     runNumber,
			Status:                        workflow.RunStatusPending,
			TriggerType:                   params.TriggerType,
			TriggerSource:                 &params.TriggerSource,
			TriggerParams:                 params.TriggerParams,
			ResolvedPrompt:                &resolvedPrompt,
			StartedAt:                     &now,
			OrchestrationResourceID:       l.OrchestrationResourceID,
			OrchestrationResourceRevision: l.OrchestrationResourceRevision,
			WorkerSpecSnapshotID:          l.WorkerSpecSnapshotID,
		}

		if err := tx.Create(run).Error; err != nil {
			return fmt.Errorf("failed to create workflow run: %w", err)
		}

		result = &workflow.TriggerRunAtomicResult{Run: run, Workflow: &l}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *workflowRunRepo) handleConcurrencyPolicy(tx *gorm.DB, l *workflow.Workflow, params *workflow.TriggerRunAtomicParams, result **workflow.TriggerRunAtomicResult) error {
	var maxNumber int
	tx.Model(&workflow.WorkflowRun{}).
		Where("workflow_id = ?", l.ID).
		Select("COALESCE(MAX(run_number), 0)").
		Scan(&maxNumber)

	now := time.Now()
	skippedRun := &workflow.WorkflowRun{
		OrganizationID:                l.OrganizationID,
		WorkflowID:                    l.ID,
		RunNumber:                     maxNumber + 1,
		Status:                        workflow.RunStatusSkipped,
		TriggerType:                   params.TriggerType,
		TriggerSource:                 &params.TriggerSource,
		FinishedAt:                    &now,
		OrchestrationResourceID:       l.OrchestrationResourceID,
		OrchestrationResourceRevision: l.OrchestrationResourceRevision,
		WorkerSpecSnapshotID:          l.WorkerSpecSnapshotID,
	}
	if err := tx.Create(skippedRun).Error; err != nil {
		return err
	}

	reason := "max concurrent runs reached"
	switch l.ConcurrencyPolicy {
	case workflow.ConcurrencyPolicyQueue:
		reason = "queued (not yet implemented)"
	case workflow.ConcurrencyPolicyReplace:
		reason = "replace (not yet implemented)"
	}

	*result = &workflow.TriggerRunAtomicResult{
		Run:      skippedRun,
		Workflow: l,
		Skipped:  true,
		Reason:   reason,
	}
	return nil
}
