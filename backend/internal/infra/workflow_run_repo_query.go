package infra

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"gorm.io/gorm"
)

// CountActiveRuns counts runs that are actually active, using Pod status as SSOT.
func (r *workflowRunRepo) CountActiveRuns(ctx context.Context, workflowID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("workflow_runs").
		Joins("LEFT JOIN pods ON pods.pod_key = workflow_runs.pod_key").
		Where("workflow_runs.workflow_id = ?", workflowID).
		Where(
			"(workflow_runs.pod_key IS NULL AND workflow_runs.status = ?) OR "+
				"(workflow_runs.pod_key IS NOT NULL AND pods.status IN ?)",
			workflow.RunStatusPending,
			agentpod.ActiveStatuses(),
		).
		Count(&count).Error
	return count, err
}

func (r *workflowRunRepo) GetActiveRunByPodKey(ctx context.Context, podKey string) (*workflow.WorkflowRun, error) {
	var run workflow.WorkflowRun
	err := r.db.WithContext(ctx).
		Where("pod_key = ? AND finished_at IS NULL", podKey).
		First(&run).Error
	if err != nil {
		if isNotFound(err) {
			return nil, workflow.ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (r *workflowRunRepo) GetTimedOutRuns(ctx context.Context, orgIDs []int64) ([]*workflow.WorkflowRun, error) {
	var runs []*workflow.WorkflowRun
	err := timedOutWorkflowRunsQuery(
		r.db.WithContext(ctx),
		orgIDs,
	).Find(&runs).Error
	return runs, err
}

func timedOutWorkflowRunsQuery(
	db *gorm.DB,
	orgIDs []int64,
) *gorm.DB {
	timedOutEligible := []string{agentpod.StatusInitializing, agentpod.StatusRunning, agentpod.StatusPaused}
	query := db.
		Table("workflow_runs").
		Joins("LEFT JOIN pods ON pods.pod_key = workflow_runs.pod_key").
		Where("workflow_runs.pod_key IS NOT NULL").
		Where("workflow_runs.finished_at IS NULL").
		Where("workflow_runs.execution_manifest IS NOT NULL").
		Where("pods.status IN ?", timedOutEligible).
		Where(
			"workflow_runs.started_at IS NOT NULL AND " +
				"workflow_runs.started_at < NOW() - " +
				"((workflow_runs.execution_manifest ->> 'timeout_minutes') || " +
				"' minutes')::INTERVAL",
		)
	if len(orgIDs) > 0 {
		query = query.Where("workflow_runs.organization_id IN ?", orgIDs)
	}
	return query
}

func (r *workflowRunRepo) GetLatestPodKey(ctx context.Context, workflowID int64) *string {
	type result struct {
		PodKey string `gorm:"column:pod_key"`
	}
	var res result
	err := r.db.WithContext(ctx).
		Table("workflow_runs").
		Select("workflow_runs.pod_key").
		Where("workflow_runs.workflow_id = ? AND workflow_runs.pod_key IS NOT NULL", workflowID).
		Order("workflow_runs.id DESC").
		Limit(1).
		Scan(&res).Error

	if err != nil || res.PodKey == "" {
		return nil
	}
	return &res.PodKey
}

// GetOrphanPendingRuns returns pending runs with no pod_key that are stuck for > 5 minutes.
func (r *workflowRunRepo) GetOrphanPendingRuns(ctx context.Context, orgIDs []int64) ([]*workflow.WorkflowRun, error) {
	var runs []*workflow.WorkflowRun
	query := r.db.WithContext(ctx).
		Where("pod_key IS NULL").
		Where("status = ?", workflow.RunStatusPending).
		Where("finished_at IS NULL").
		Where("created_at < NOW() - INTERVAL '5 minutes'")
	if len(orgIDs) > 0 {
		query = query.Where("organization_id IN ?", orgIDs)
	}
	err := query.Find(&runs).Error
	return runs, err
}

func (r *workflowRunRepo) GetIdleWorkflowPods(ctx context.Context, orgIDs []int64) ([]*workflow.WorkflowRun, error) {
	var runs []*workflow.WorkflowRun
	err := idleWorkflowPodsQuery(
		r.db.WithContext(ctx),
		orgIDs,
	).Find(&runs).Error
	return runs, err
}

func idleWorkflowPodsQuery(
	db *gorm.DB,
	orgIDs []int64,
) *gorm.DB {
	query := db.
		Table("workflow_runs").
		Joins("JOIN pods ON pods.pod_key = workflow_runs.pod_key").
		Where("workflow_runs.finished_at IS NULL").
		Where("workflow_runs.execution_manifest IS NOT NULL").
		Where("pods.status = ?", agentpod.StatusRunning).
		Where("pods.agent_status = ?", agentpod.AgentStatusWaiting).
		Where("pods.agent_waiting_since IS NOT NULL").
		Where(
			"(workflow_runs.execution_manifest ->> " +
				"'idle_timeout_seconds')::INTEGER > 0",
		).
		Where(
			"pods.agent_waiting_since < NOW() - " +
				"((workflow_runs.execution_manifest ->> " +
				"'idle_timeout_seconds') || ' seconds')::INTERVAL",
		)
	if len(orgIDs) > 0 {
		query = query.Where("workflow_runs.organization_id IN ?", orgIDs)
	}
	return query
}
