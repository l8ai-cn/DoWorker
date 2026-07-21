package main

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func initializeServices(
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
) (*serviceContainer, error) {
	encryptor := crypto.NewEncryptor(cfg.JWT.Secret)
	services, err := initializeIdentityServices(cfg, db, redisClient, encryptor)
	if err != nil {
		return nil, err
	}
	if err := initializeWorkspaceServices(services, cfg, db, encryptor); err != nil {
		return nil, err
	}
	initializePlatformServices(services, cfg, db, redisClient)
	return services, nil
}
