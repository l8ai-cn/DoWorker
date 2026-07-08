package infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"gorm.io/gorm"
)

var _ agentpod.PendingCommandRepository = (*pendingCommandRepo)(nil)

type pendingCommandRepo struct{ db *gorm.DB }

func NewPendingCommandRepository(db *gorm.DB) agentpod.PendingCommandRepository {
	return &pendingCommandRepo{db: db}
}

// isUniqueViolation matches without gorm.Config{TranslateError} — the postgres
// driver only maps 23505 to gorm.ErrDuplicatedKey when that option is on,
// and this codebase does not enable it (same pattern as blockstore/repo.go).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "UNIQUE constraint failed")
}

func (r *pendingCommandRepo) Enqueue(ctx context.Context, cmd *agentpod.PendingCommand) error {
	err := r.db.WithContext(ctx).Create(cmd).Error
	if err != nil {
		if isUniqueViolation(err) {
			return agentpod.ErrDuplicateCommand
		}
		return err
	}
	return nil
}

func (r *pendingCommandRepo) CountByRunner(ctx context.Context, runnerID int64) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&agentpod.PendingCommand{}).
		Where("runner_id = ?", runnerID).Count(&n).Error
	return int(n), err
}

func (r *pendingCommandRepo) ListByRunnerFIFO(ctx context.Context, runnerID int64, limit int) ([]*agentpod.PendingCommand, error) {
	var rows []*agentpod.PendingCommand
	q := r.db.WithContext(ctx).Where("runner_id = ?", runnerID).Order("id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return rows, q.Find(&rows).Error
}

func (r *pendingCommandRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&agentpod.PendingCommand{}, id).Error
}

func (r *pendingCommandRepo) DeleteByPodKey(ctx context.Context, podKey string) (int64, error) {
	res := r.db.WithContext(ctx).Where("pod_key = ?", podKey).Delete(&agentpod.PendingCommand{})
	return res.RowsAffected, res.Error
}

func (r *pendingCommandRepo) ListExpired(ctx context.Context, now time.Time, limit int) ([]*agentpod.PendingCommand, error) {
	var rows []*agentpod.PendingCommand
	q := r.db.WithContext(ctx).Where("expires_at <= ?", now).Order("id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return rows, q.Find(&rows).Error
}

func (r *pendingCommandRepo) ListRunnerIDsWithPending(ctx context.Context, limit int) ([]int64, error) {
	var ids []int64
	q := r.db.WithContext(ctx).Model(&agentpod.PendingCommand{}).
		Distinct("runner_id")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return ids, q.Pluck("runner_id", &ids).Error
}

func (r *pendingCommandRepo) PositionByPodKey(ctx context.Context, runnerID int64, podKey string) (int, error) {
	var target agentpod.PendingCommand
	if err := r.db.WithContext(ctx).
		Where("runner_id = ? AND pod_key = ?", runnerID, podKey).
		Order("id ASC").First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	var ahead int64
	if err := r.db.WithContext(ctx).Model(&agentpod.PendingCommand{}).
		Where("runner_id = ? AND id < ?", runnerID, target.ID).
		Count(&ahead).Error; err != nil {
		return 0, err
	}
	return int(ahead) + 1, nil
}

func (r *pendingCommandRepo) GetCreatePodByPodKey(ctx context.Context, podKey string) (*agentpod.PendingCommand, error) {
	var cmd agentpod.PendingCommand
	err := r.db.WithContext(ctx).
		Where("pod_key = ? AND command_type = ?", podKey, agentpod.CommandTypeCreatePod).
		Order("id ASC").
		First(&cmd).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}
