package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"gorm.io/gorm"
)

func (r *podRepo) GetByOrchestrationWorkerLaunchID(
	ctx context.Context,
	organizationID int64,
	launchID int64,
) (*agentpod.Pod, error) {
	var pod agentpod.Pod
	err := r.db.WithContext(ctx).
		Preload("ActiveConfigRevision").
		Where(
			"organization_id = ? AND orchestration_worker_launch_id = ?",
			organizationID,
			launchID,
		).
		First(&pod).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pod, nil
}
