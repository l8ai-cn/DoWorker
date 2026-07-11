package main

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/auth"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	ssoservice "github.com/anthropics/agentsmesh/backend/internal/service/sso"
	"github.com/anthropics/agentsmesh/backend/internal/service/user"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func initializeIdentityServices(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, encryptor *crypto.Encryptor) *serviceContainer {
	userSvc := user.NewServiceWithEncryption(infra.NewUserRepository(db), cfg.JWT.Secret)
	authCfg := &auth.Config{
		JWTSecret:         cfg.JWT.Secret,
		JWTExpiration:     time.Duration(cfg.JWT.ExpirationHours) * time.Hour,
		RefreshExpiration: time.Duration(cfg.JWT.ExpirationHours*7) * time.Hour,
		Issuer:            "agentsmesh",
	}
	authSvc := auth.NewServiceWithRedis(authCfg, userSvc, redisClient)
	ssoSvc := ssoservice.NewServiceWithRedis(infra.NewSSOConfigRepository(db), cfg.JWT.Secret, cfg, redisClient)
	authSvc.SetSSOChecker(ssoSvc)

	agentSvc := agent.NewAgentService(infra.NewAgentRepository(db))
	envBundleSvc := envbundleservice.NewService(infra.NewEnvBundleRepository(db), encryptor)
	userConfigSvc := agent.NewUserConfigService(infra.NewUserConfigRepository(db), agentSvc)

	return &serviceContainer{
		auth:       authSvc,
		user:       userSvc,
		sso:        ssoSvc,
		agentSvc:   agentSvc,
		envBundle:  envBundleSvc,
		userConfig: userConfigSvc,
	}
}
