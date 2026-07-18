package infra

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

func (r *runnerRepository) IncrementPods(ctx context.Context, runnerID int64) error {
	result := r.db.WithContext(ctx).Exec(`
UPDATE runners
SET current_pods = current_pods + 1
WHERE id = ?
  AND current_pods < max_concurrent_pods
`, runnerID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return runner.ErrRunnerCapacityUnavailable
	}
	return nil
}

func (r *runnerRepository) DecrementPods(ctx context.Context, runnerID int64) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE runners SET current_pods = GREATEST(current_pods - 1, 0) WHERE id = ?", runnerID,
	).Error
}

func (r *runnerRepository) SetPodCount(ctx context.Context, runnerID int64, count int) error {
	return r.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("id = ?", runnerID).
		Update("current_pods", count).Error
}

func (r *runnerRepository) BatchUpdateHeartbeats(
	ctx context.Context,
	items []runner.HeartbeatUpdate,
) (int, error) {
	updated := 0
	for _, item := range items {
		updates := map[string]interface{}{
			"last_heartbeat": item.Timestamp,
			"current_pods":   item.CurrentPods,
			"status":         item.Status,
		}
		if item.Version != "" {
			updates["runner_version"] = item.Version
		}

		result := r.db.WithContext(ctx).Model(&runner.Runner{}).
			Where("id = ?", item.RunnerID).
			Updates(updates)
		if result.Error != nil {
			continue
		}
		if result.RowsAffected > 0 {
			updated++
		}
	}
	return updated, nil
}
