package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"gorm.io/gorm"
)

func (repo *agentWorkbenchRepository) GetSnapshot(
	ctx context.Context,
	sessionID string,
) (*agentworkbench.SessionState, error) {
	if sessionID == "" {
		return nil, agentworkbench.ErrInvalidArgument
	}
	var record agentWorkbenchStateRecord
	err := repo.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record.domain(), nil
}

func (repo *agentWorkbenchRepository) ListAfter(
	ctx context.Context,
	sessionID string,
	streamEpoch string,
	sequence uint64,
	limit int,
) ([]agentworkbench.Event, error) {
	if sessionID == "" || streamEpoch == "" || limit <= 0 {
		return nil, agentworkbench.ErrInvalidArgument
	}
	var records []agentWorkbenchEventRecord
	err := repo.db.WithContext(ctx).
		Where(
			"session_id = ? AND stream_epoch = ? AND sequence > ?",
			sessionID,
			streamEpoch,
			sequence,
		).
		Order("sequence ASC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	events := make([]agentworkbench.Event, len(records))
	for index := range records {
		events[index] = records[index].domain()
	}
	return events, nil
}
