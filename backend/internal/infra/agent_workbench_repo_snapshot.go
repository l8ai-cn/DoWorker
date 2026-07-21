package infra

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	"gorm.io/gorm"
)

func (repo *agentWorkbenchRepository) EnsureSnapshot(
	ctx context.Context,
	initial agentworkbench.SessionState,
) (*agentworkbench.SessionState, error) {
	if err := validateAgentWorkbenchInitialSnapshot(initial); err != nil {
		return nil, err
	}
	var stored *agentworkbench.SessionState
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockAgentWorkbenchSession(tx, initial.SessionID); err != nil {
			return err
		}
		current, err := loadAgentWorkbenchState(tx, initial.SessionID)
		if err != nil {
			return err
		}
		if current != nil {
			stored = current.domain()
			return nil
		}
		record := agentWorkbenchStateRecordFromDomain(initial, time.Now().UTC())
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		stored = record.domain()
		return nil
	})
	return stored, err
}

func validateAgentWorkbenchInitialSnapshot(
	initial agentworkbench.SessionState,
) error {
	if !validAgentWorkbenchText(initial.SessionID, 100) ||
		!validAgentWorkbenchText(initial.StreamEpoch, 100) ||
		initial.Revision != 0 ||
		initial.LatestSequence != 0 ||
		!validAgentWorkbenchDigest(initial.Digest) {
		return agentworkbench.ErrInvalidArgument
	}
	return nil
}
