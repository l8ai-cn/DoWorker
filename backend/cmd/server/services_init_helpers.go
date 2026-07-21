package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/extension"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/gitea"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/storage"
	extensionservice "github.com/l8ai-cn/agentcloud/backend/internal/service/extension"
	fileservice "github.com/l8ai-cn/agentcloud/backend/internal/service/file"
	knowledgebaseservice "github.com/l8ai-cn/agentcloud/backend/internal/service/knowledgebase"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/license"
	supportticketservice "github.com/l8ai-cn/agentcloud/backend/internal/service/supportticket"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"gorm.io/gorm"
)

func initializeFileService(cfg *config.Config) *fileservice.Service {
	if cfg.Storage.AccessKey == "" || cfg.Storage.SecretKey == "" {
		slog.Warn("Storage not configured, file upload disabled")
		return nil
	}

	s3Storage, err := storage.NewS3Storage(storageS3Config(cfg))
	if err != nil {
		slog.Error("Failed to initialize storage", "error", err)
		return nil
	}

	if err := s3Storage.EnsureBucket(context.Background()); err != nil {
		slog.Warn("Failed to ensure bucket exists", "bucket", cfg.Storage.Bucket, "error", err)
	}

	slog.Info("Storage initialized", "endpoint", cfg.Storage.Endpoint, "bucket", cfg.Storage.Bucket)
	return fileservice.NewService(s3Storage, cfg.Storage)
}

func initializeKnowledgeBaseService(cfg *config.Config, db *gorm.DB) *knowledgebaseservice.Service {
	if !cfg.KnowledgeBase.Enabled() {
		slog.Warn("Internal Gitea control plane or SSH clone URL not configured, knowledge bases disabled")
		return nil
	}
	giteaClient := gitea.NewClient(gitea.Config{
		BaseURL:         cfg.KnowledgeBase.GiteaURL,
		AdminToken:      cfg.KnowledgeBase.GiteaToken,
		Namespace:       cfg.KnowledgeBase.GiteaOrg,
		CloneBaseURL:    cfg.KnowledgeBase.CloneBaseURL,
		SSHCloneBaseURL: cfg.KnowledgeBase.SSHCloneBaseURL,
		SSHKnownHosts:   cfg.KnowledgeBase.SSHKnownHosts,
	})
	kbRepo := infra.NewKnowledgeBaseRepository(db)
	slog.Info("Knowledge base service initialized",
		"gitea", cfg.KnowledgeBase.GiteaURL, "namespace", cfg.KnowledgeBase.GiteaOrg)
	svc := knowledgebaseservice.NewService(kbRepo, giteaClient, slog.Default())
	svc.SetSecretsEncryptor(crypto.NewEncryptor(cfg.JWT.Secret))
	return svc
}

func newGiteaClientForNamespace(cfg *config.Config, namespace string) *gitea.Client {
	if !cfg.KnowledgeBase.GitopsEnabled() {
		return nil
	}
	return gitea.NewClient(gitea.Config{
		BaseURL:         cfg.KnowledgeBase.GiteaURL,
		AdminToken:      cfg.KnowledgeBase.GiteaToken,
		Namespace:       namespace,
		CloneBaseURL:    cfg.KnowledgeBase.CloneBaseURL,
		SSHCloneBaseURL: cfg.KnowledgeBase.SSHCloneBaseURL,
		SSHKnownHosts:   cfg.KnowledgeBase.SSHKnownHosts,
	})
}

func initializeLicenseService(cfg *config.Config, db *gorm.DB) *license.Service {
	if !cfg.Payment.IsOnPremise() && cfg.Payment.License.PublicKeyPath == "" {
		return nil
	}

	licenseRepo := infra.NewLicenseRepository(db)
	licenseSvc, err := license.NewService(licenseRepo, &cfg.Payment.License, slog.Default())
	if err != nil {
		slog.Warn("Failed to initialize license service", "error", err)
		return nil
	}

	slog.Info("License service initialized")
	return licenseSvc
}

func initializeExtensionServices(cfg *config.Config, db *gorm.DB) (*extensionservice.Service, extension.Repository, *extensionservice.MarketplaceWorker) {
	if cfg.Storage.AccessKey == "" || cfg.Storage.SecretKey == "" {
		slog.Warn("Storage not configured, extension services disabled")
		return nil, nil, nil
	}

	s3Storage, err := storage.NewS3Storage(storageS3Config(cfg))
	if err != nil {
		slog.Error("Failed to initialize storage for extensions", "error", err)
		return nil, nil, nil
	}

	extRepo := infra.NewExtensionRepository(db)
	encryptor := crypto.NewEncryptor(cfg.JWT.Secret)
	extSvc := extensionservice.NewService(extRepo, s3Storage, encryptor)
	skillPkg := extensionservice.NewSkillPackager(extRepo, s3Storage)
	extSvc.SetSkillPackager(skillPkg)
	extSvc.SetSkillCatalog(infra.NewSkillCatalogRepository(db))

	var mcpRegistrySyncer *extensionservice.McpRegistrySyncer
	if cfg.Marketplace.RegistryEnabled {
		mcpRegistryClient := extensionservice.NewMcpRegistryClient(cfg.Marketplace.RegistryURL)
		mcpRegistrySyncer = extensionservice.NewMcpRegistrySyncer(mcpRegistryClient, extRepo)
		slog.Info("MCP Registry syncer enabled", "url", cfg.Marketplace.RegistryURL)
	}

	syncInterval := cfg.Marketplace.SyncInterval
	if syncInterval == 0 {
		syncInterval = 1 * time.Hour
	}
	mktWorker := extensionservice.NewMarketplaceWorker(mcpRegistrySyncer, syncInterval)
	slog.Info("MarketplaceWorker configured", "interval", syncInterval)

	slog.Info("Extension services initialized")
	return extSvc, extRepo, mktWorker
}

func initializeSupportTicketService(cfg *config.Config, db *gorm.DB) *supportticketservice.Service {
	supportTicketRepo := infra.NewSupportTicketRepository(db)

	if cfg.Storage.AccessKey == "" || cfg.Storage.SecretKey == "" {
		slog.Warn("Storage not configured, support ticket attachments disabled")
		return supportticketservice.NewService(supportTicketRepo, nil, cfg.Storage)
	}

	s3Storage, err := storage.NewS3Storage(storageS3Config(cfg))
	if err != nil {
		slog.Error("Failed to initialize storage for support tickets", "error", err)
		return supportticketservice.NewService(supportTicketRepo, nil, cfg.Storage)
	}

	slog.Info("Support ticket service initialized")
	return supportticketservice.NewService(supportTicketRepo, s3Storage, cfg.Storage)
}

func initializeLogUploadStorage(cfg *config.Config) storage.Storage {
	s3Storage, err := storage.NewS3Storage(storageS3Config(cfg))
	if err != nil {
		slog.Error("Failed to initialize storage for runner logs", "error", err)
		return nil
	}

	if err := s3Storage.EnsureBucket(context.Background()); err != nil {
		slog.Warn("Failed to ensure bucket for runner logs", "error", err)
	}
	return s3Storage
}
