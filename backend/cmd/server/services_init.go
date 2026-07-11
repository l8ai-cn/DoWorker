package main

import (
	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func initializeServices(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *serviceContainer {
	encryptor := crypto.NewEncryptor(cfg.JWT.Secret)
	services := initializeIdentityServices(cfg, db, redisClient, encryptor)
	initializeWorkspaceServices(services, cfg, db, encryptor)
	initializePlatformServices(services, cfg, db, redisClient)
	return services
}
