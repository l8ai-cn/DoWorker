package infra

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/audit"
	"gorm.io/gorm"
)

type aiResourceMutationRunner struct{ db *gorm.DB }
type aiResourceAuditRecorder struct{ db *gorm.DB }

func NewAIResourceMutationRunner(db *gorm.DB) service.MutationRunner {
	return &aiResourceMutationRunner{db: db}
}

func (runner *aiResourceMutationRunner) Run(ctx context.Context, mutation func(airesource.Repository, service.AuditRecorder) error) error {
	return runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return mutation(&aiResourceRepo{db: tx}, &aiResourceAuditRecorder{db: tx})
	})
}

func (recorder *aiResourceAuditRecorder) Record(ctx context.Context, entry *audit.Log) error {
	return recorder.db.WithContext(ctx).Create(entry).Error
}
