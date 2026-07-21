package infra

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
)

var _ expertmarket.Repository = (*expertMarketRepo)(nil)

type expertMarketRepo struct {
	db *gorm.DB
}

func NewExpertMarketRepository(db *gorm.DB) expertmarket.Repository {
	return &expertMarketRepo{db: db}
}
