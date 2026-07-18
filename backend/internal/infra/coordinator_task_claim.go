package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	"gorm.io/gorm"
)

func (r *coordinatorRepo) WithinProjectDispatch(
	ctx context.Context,
	projectID int64,
	fn func(coordinator.Repository) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockKey := fmt.Sprintf("coordinator-project:%d", projectID)
		if err := tx.Exec(
			"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
			lockKey,
		).Error; err != nil {
			return err
		}
		return fn(&coordinatorRepo{db: tx})
	})
}

func (r *coordinatorRepo) GetActiveExecutionByProjectAndExternalID(
	ctx context.Context,
	projectID int64,
	externalID string,
) (*coordinator.Execution, error) {
	var execution coordinator.Execution
	err := r.db.WithContext(ctx).
		Where(
			"project_id = ? AND external_id = ? AND status IN ?",
			projectID,
			externalID,
			[]string{
				coordinator.ExecutionStatusPending,
				coordinator.ExecutionStatusClaimed,
				coordinator.ExecutionStatusRunning,
			},
		).
		Order("created_at DESC").
		First(&execution).Error
	if err == nil {
		return &execution, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, coordinator.ErrNotFound
	}
	return nil, err
}
