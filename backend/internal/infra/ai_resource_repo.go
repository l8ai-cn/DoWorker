package infra

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

var _ airesource.Repository = (*aiResourceRepo)(nil)

type aiResourceRepo struct {
	db *gorm.DB
}

func NewAIResourceRepository(db *gorm.DB) airesource.Repository {
	return &aiResourceRepo{db: db}
}
