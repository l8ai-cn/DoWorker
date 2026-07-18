package main

import (
	"context"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	knowledgebaseservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerlogservice "github.com/anthropics/agentsmesh/backend/internal/service/runnerlog"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"gorm.io/gorm"
)

func initLogUploadService(
	cfg *config.Config,
	db *gorm.DB,
	runnerConnMgr *runner.RunnerConnectionManager,
) *runnerlogservice.Service {
	if cfg.Storage.AccessKey == "" || cfg.Storage.SecretKey == "" {
		slog.Info("Runner log upload service disabled: storage not configured")
		return nil
	}
	logUploadStorage := initializeLogUploadStorage(cfg)
	if logUploadStorage == nil {
		return nil
	}
	logUploadRepo := infra.NewRunnerLogRepository(db)
	logUploadSvc := runnerlogservice.NewService(logUploadRepo, logUploadStorage)

	runnerConnMgr.SetLogUploadStatusCallback(func(runnerID int64, data *runnerv1.LogUploadStatusEvent) {
		logUploadSvc.HandleUploadStatus(runnerID, data.RequestId, data.Phase, data.Progress, data.Message, data.Error, data.SizeBytes)
	})
	slog.Info("Runner log upload service initialized")
	return logUploadSvc
}

func knowledgeBaseResolverOrNil(svc *knowledgebaseservice.Service) agentpod.KnowledgeBaseResolverForOrchestrator {
	if svc == nil {
		return nil
	}
	return svc
}

func createPodOrchestrator(
	services *serviceContainer,
	podCoordinator *runner.PodCoordinator,
) *agentpod.PodOrchestrator {
	if services.envBundle == nil {
		panic("createPodOrchestrator: services.envBundle is required for ConfigBuilder")
	}
	configBuilder := agent.NewConfigBuilder(services.agentSvc, services.envBundle)
	if services.extension != nil {
		configBuilder.SetExtensionProvider(services.extension)
		slog.Info("ExtensionProvider connected to ConfigBuilder")
	}
	slog.Info("EnvBundleService connected to ConfigBuilder")
	orchestrator := agentpod.NewPodOrchestrator(&agentpod.PodOrchestratorDeps{
		PodService:         services.pod,
		ConfigBuilder:      configBuilder,
		PodCoordinator:     podCoordinator,
		BillingService:     services.billing,
		UserService:        services.user,
		RepoService:        services.repository,
		TicketService:      services.ticket,
		RunnerSelector:     services.runner,
		AgentResolver:      services.agentSvc,
		RunnerQuery:        services.runner,
		UserConfigQuery:    services.userConfig,
		PodRepo:            services.podRepo,
		PermissionPolicy:   services.permissionPolicy,
		KnowledgeBases:     knowledgeBaseResolverOrNil(services.knowledgeBase),
		ModelResources:     services.aiResource,
		WorkerCreation:     services.workerCreation,
		WorkerSpecs:        services.workerSpecs,
		WorkerDependencies: services.workerDependencies,
	})
	slog.Info("PodOrchestrator created")
	return orchestrator
}

func startMarketplaceWorker(services *serviceContainer) func() {
	if services.marketplaceWorker != nil {
		services.marketplaceWorker.Start(context.Background())
		slog.Info("MarketplaceWorker started")
	}
	return func() {
		if services.marketplaceWorker != nil {
			services.marketplaceWorker.Stop()
			slog.Info("MarketplaceWorker stopped")
		}
	}
}

func startKnowledgeBaseSyncWorker(services *serviceContainer) func() {
	worker := services.kbSyncWorker
	if worker != nil {
		worker.Start(context.Background())
		slog.Info("KnowledgeBase SyncWorker started")
	}
	return func() {
		if worker != nil {
			worker.Stop()
			slog.Info("KnowledgeBase SyncWorker stopped")
		}
	}
}
