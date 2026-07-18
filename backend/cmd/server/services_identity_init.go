package main

import (
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/auth"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	ssoservice "github.com/anthropics/agentsmesh/backend/internal/service/sso"
	"github.com/anthropics/agentsmesh/backend/internal/service/user"
	authpkg "github.com/anthropics/agentsmesh/backend/pkg/auth"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func initializeIdentityServices(
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	encryptor *crypto.Encryptor,
) (*serviceContainer, error) {
	userSvc := user.NewServiceWithEncryption(infra.NewUserRepository(db), cfg.JWT.Secret)
	accessTokens, err := authpkg.LoadAccessTokenManager(authpkg.AccessTokenFileConfig{
		PrivateKeyFile: cfg.AccessToken.PrivateKeyFile,
		PublicKeyFile:  cfg.AccessToken.PublicKeyFile,
		KeyID:          cfg.AccessToken.KeyID,
		Issuer:         cfg.AccessToken.Issuer,
		Audiences:      cfg.AccessToken.Audiences,
		Duration:       time.Duration(cfg.AccessToken.ExpirationHours) * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize access tokens: %w", err)
	}
	authCfg := &auth.Config{
		JWTExpiration:       time.Duration(cfg.AccessToken.ExpirationHours) * time.Hour,
		RefreshExpiration:   time.Duration(cfg.AccessToken.ExpirationHours*7) * time.Hour,
		Issuer:              cfg.AccessToken.Issuer,
		AccessTokens:        accessTokens,
		AccessTokenAudience: cfg.AccessToken.CoreAudience,
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
	}, nil
}
