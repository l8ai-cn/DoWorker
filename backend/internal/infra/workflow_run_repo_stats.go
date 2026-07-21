package infra

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

// ComputeLoopStats computes run statistics from Pod status (SSOT).
func (r *workflowRunRepo) ComputeLoopStats(ctx context.Context, workflowID int64) (total, successful, failed int, err error) {
	type finishedStats struct {
		Total      int `gorm:"column:total"`
		Successful int `gorm:"column:successful"`
		Failed     int `gorm:"column:failed"`
	}
	var fs finishedStats
	err = r.db.WithContext(ctx).
		Table("workflow_runs").
		Select(`
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = ?) as successful,
			COUNT(*) FILTER (WHERE status IN (?, ?, ?)) as failed
		`, workflow.RunStatusCompleted, workflow.RunStatusFailed, workflow.RunStatusTimeout, workflow.RunStatusCancelled).
		Where("workflow_id = ? AND finished_at IS NOT NULL", workflowID).
		Scan(&fs).Error
	if err != nil {
		return
	}
	total = fs.Total
	successful = fs.Successful
	failed = fs.Failed

	// Phase 2: Resolve active runs via Go-side SSOT
	total, successful, failed, err = r.resolveActiveRunStats(ctx, workflowID, total, successful, failed)
	return
}

func (r *workflowRunRepo) resolveActiveRunStats(ctx context.Context, workflowID int64, total, successful, failed int) (int, int, int, error) {
	type activeRunRow struct {
		Status         string  `gorm:"column:status"`
		PodKey         *string `gorm:"column:pod_key"`
		PodStatus      *string `gorm:"column:pod_status"`
		AutopilotPhase *string `gorm:"column:autopilot_phase"`
	}
	var activeRows []activeRunRow
	err := r.db.WithContext(ctx).
		Table("workflow_runs lr").
		Select("lr.status, lr.pod_key, p.status as pod_status, ac.phase as autopilot_phase").
		Joins("LEFT JOIN pods p ON p.pod_key = lr.pod_key").
		Joins("LEFT JOIN autopilot_controllers ac ON ac.autopilot_controller_key = lr.autopilot_controller_key").
		Where("lr.workflow_id = ? AND lr.finished_at IS NULL", workflowID).
		Find(&activeRows).Error
	if err != nil {
		return total, successful, failed, err
	}

	for _, row := range activeRows {
		total++

		var effectiveStatus string
		if row.PodKey == nil {
			effectiveStatus = row.Status
		} else {
			podStatus := ""
			if row.PodStatus != nil {
				podStatus = *row.PodStatus
			}
			autopilotPhase := ""
			if row.AutopilotPhase != nil {
				autopilotPhase = *row.AutopilotPhase
			}
			effectiveStatus = deriveWorkflowRunStatus(podStatus, autopilotPhase)
		}

		switch effectiveStatus {
		case workflow.RunStatusCompleted:
			successful++
		case workflow.RunStatusFailed, workflow.RunStatusTimeout, workflow.RunStatusCancelled:
			failed++
		}
	}
	return total, successful, failed, nil
}

func (r *workflowRunRepo) BatchGetPodStatuses(ctx context.Context, podKeys []string) ([]workflow.PodStatusInfo, error) {
	if len(podKeys) == 0 {
		return nil, nil
	}

	var results []workflow.PodStatusInfo
	err := r.db.WithContext(ctx).
		Table("pods").
		Select("pod_key, status, finished_at").
		Where("pod_key IN ?", podKeys).
		Find(&results).Error
	return results, err
}

func (r *workflowRunRepo) BatchGetAutopilotPhases(ctx context.Context, autopilotKeys []string) (map[string]string, error) {
	if len(autopilotKeys) == 0 {
		return nil, nil
	}

	type row struct {
		Key   string `gorm:"column:autopilot_controller_key"`
		Phase string `gorm:"column:phase"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Table("autopilot_controllers").
		Select("autopilot_controller_key, phase").
		Where("autopilot_controller_key IN ?", autopilotKeys).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.Key] = r.Phase
	}
	return result, nil
}

// CountActiveRunsByWorkflowIDs batch-counts active runs for multiple workflows using Pod status (SSOT).
func (r *workflowRunRepo) CountActiveRunsByWorkflowIDs(ctx context.Context, workflowIDs []int64) (map[int64]int64, error) {
	if len(workflowIDs) == 0 {
		return nil, nil
	}

	type countRow struct {
		WorkflowID int64 `gorm:"column:workflow_id"`
		Count      int64 `gorm:"column:count"`
	}
	var rows []countRow
	err := r.db.WithContext(ctx).
		Table("workflow_runs").
		Select("workflow_runs.workflow_id, COUNT(*) as count").
		Joins("LEFT JOIN pods ON pods.pod_key = workflow_runs.pod_key").
		Where("workflow_runs.workflow_id IN ?", workflowIDs).
		Where(
			"(workflow_runs.pod_key IS NULL AND workflow_runs.status = ?) OR "+
				"(workflow_runs.pod_key IS NOT NULL AND pods.status IN ?)",
			workflow.RunStatusPending,
			agentpod.ActiveStatuses(),
		).
		Group("workflow_runs.workflow_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.WorkflowID] = row.Count
	}
	return result, nil
}

func (r *workflowRunRepo) GetAvgDuration(ctx context.Context, workflowID int64) (*float64, error) {
	var avg *float64
	err := r.db.WithContext(ctx).
		Table("workflow_runs").
		Where("workflow_id = ? AND duration_sec IS NOT NULL AND finished_at IS NOT NULL", workflowID).
		Select("AVG(duration_sec)").
		Scan(&avg).Error
	return avg, err
}
