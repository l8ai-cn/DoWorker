package agentsession

import (
	"context"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessionDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	itemService "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
	"gorm.io/gorm"
)

type DeferredCommitter struct {
	db *gorm.DB
}

func NewDeferredCommitter(db *gorm.DB) *DeferredCommitter {
	return &DeferredCommitter{db: db}
}

func (c *DeferredCommitter) CommitCreate(
	ctx context.Context,
	session *sessionDomain.Session,
	command *podDomain.PendingCommand,
	maxPerRunner int,
	writeItems func(*itemService.Service) error,
) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureDeferredCommandCapacity(tx, command.RunnerID, maxPerRunner); err != nil {
			return err
		}
		if err := tx.Create(session).Error; err != nil {
			return err
		}
		if writeItems != nil {
			if err := writeItems(itemService.NewService(tx)); err != nil {
				return err
			}
		}
		return tx.Create(command).Error
	})
}

func ensureDeferredCommandCapacity(tx *gorm.DB, runnerID int64, maxPerRunner int) error {
	if tx.Name() == "postgres" {
		if err := tx.Exec(
			"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
			podDomain.PendingCommandRunnerLockName(runnerID),
		).Error; err != nil {
			return err
		}
	}
	var count int64
	if err := tx.Model(&podDomain.PendingCommand{}).
		Where("runner_id = ?", runnerID).
		Count(&count).Error; err != nil {
		return err
	}
	if int(count) >= maxPerRunner {
		return podDomain.ErrQueueFull
	}
	return nil
}
