package main

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	blockstoreinfra "github.com/l8ai-cn/agentcloud/backend/internal/infra/blockstore"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/database"
	adminservice "github.com/l8ai-cn/agentcloud/backend/internal/service/admin"
	apikeyservice "github.com/l8ai-cn/agentcloud/backend/internal/service/apikey"
	blockstoreservice "github.com/l8ai-cn/agentcloud/backend/internal/service/blockstore"
	goalloop "github.com/l8ai-cn/agentcloud/backend/internal/service/goalloop"
	knowledgebaseservice "github.com/l8ai-cn/agentcloud/backend/internal/service/knowledgebase"
	notifservice "github.com/l8ai-cn/agentcloud/backend/internal/service/notification"
	permissionpolicysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/permissionpolicy"
	podsessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionusage"
	tokenusagesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tokenusage"
	workflow "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func initializePlatformServices(services *serviceContainer, cfg *config.Config, db *gorm.DB, redisClient *redis.Client) {
	services.file = initializeFileService(cfg)
	services.supportTicket = initializeSupportTicketService(cfg, db)
	services.apikey = apikeyservice.NewService(infra.NewAPIKeyRepository(db), redisClient)
	services.apikeyAdapter = apikeyservice.NewMiddlewareAdapter(services.apikey)
	services.workflow = workflow.NewWorkflowService(infra.NewWorkflowRepository(db))
	services.workflowRun = workflow.NewWorkflowRunService(infra.NewWorkflowRunRepository(db))
	services.goalLoop = goalloop.NewService(infra.NewGoalLoopRepository(db))
	services.license = initializeLicenseService(cfg, db)
	services.extension, services.extensionRepo, services.marketplaceWorker = initializeExtensionServices(cfg, db)
	services.knowledgeBase = initializeKnowledgeBaseService(cfg, db)
	if services.knowledgeBase != nil {
		services.kbSyncWorker = knowledgebaseservice.NewSyncWorker(services.knowledgeBase, cfg.KnowledgeBase.SyncInterval)
	}

	services.notifPrefStore = notifservice.NewPreferenceStore(infra.NewNotificationPreferenceRepository(db))
	services.tokenUsage = tokenusagesvc.NewService(infra.NewTokenUsageRepository(db), slog.Default())
	services.podSessionUsage = podsessionsvc.NewService(db)
	services.permissionPolicy = permissionpolicysvc.NewService(db)
	services.blockstore = blockstoreservice.NewService(blockstoreinfra.NewRepository(db), slog.Default())
	if embedder := selectEmbedder(); embedder != nil {
		services.blockstore.SetEmbedder(embedder)
	}
	services.ticket.SetBlockstore(services.blockstore)

	// REST and Connect admin handlers must share one audit-log pipeline.
	services.adminDB = database.NewGormWrapper(db)
	services.admin = adminservice.NewService(services.adminDB)
}
