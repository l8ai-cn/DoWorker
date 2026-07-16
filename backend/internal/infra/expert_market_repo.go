package infra

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
)

var _ expertmarket.Repository = (*expertMarketRepo)(nil)

type expertMarketRepo struct {
	db *gorm.DB
}

func NewExpertMarketRepository(db *gorm.DB) expertmarket.Repository {
	return &expertMarketRepo{db: db}
}
