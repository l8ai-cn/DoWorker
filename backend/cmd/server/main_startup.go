package main

import (
	"log/slog"

	grpcserver "github.com/anthropics/agentsmesh/backend/internal/api/grpc"
	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra/acme"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/infra/logger"
	"github.com/anthropics/agentsmesh/backend/internal/infra/websocket"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/geo"
	previewservice "github.com/anthropics/agentsmesh/backend/internal/service/preview"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerlogservice "github.com/anthropics/agentsmesh/backend/internal/service/runnerlog"
	workflow "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type grpcWiringResult struct {
	handler              *v1.GRPCRunnerHandler
	server               *grpcserver.Server
	runnerAdapter        *grpcserver.GRPCRunnerAdapter
	upgradeCommandSender runner.UpgradeCommandSender
	logUploadSender      runner.LogUploadCommandSender
}

func initPKIAndGRPCWiring(
	cfg *config.Config,
	services *serviceContainer,
	runnerConnMgr *runner.RunnerConnectionManager,
	podCoordinator *runner.PodCoordinator,
	podRouter *runner.PodRouter,
	sandboxQuerySvc *runner.SandboxQueryService,
	sandboxFsSvc *runner.SandboxFsService,
	podOrchestrator *agentpod.PodOrchestrator,
	workflowOrchestrator *workflow.WorkflowOrchestrator,
	workflowRunSvc *workflow.WorkflowRunService,
	appLogger *logger.Logger,
	relayTokenGenerator *relay.TokenGenerator,
	db *gorm.DB,
) *grpcWiringResult {
	if cfg.PKI.CACertFile == "" || cfg.PKI.CAKeyFile == "" {
		slog.Warn("PKI CA files not configured, gRPC/mTLS disabled")
		return &grpcWiringResult{
			handler: v1.NewGRPCRunnerHandler(services.runner, nil, cfg),
		}
	}

	mcpDeps := &grpcserver.MCPDependencies{
		PodService:           services.pod,
		PodOrchestrator:      podOrchestrator,
		ChannelService:       services.channel,
		BindingService:       services.binding,
		TicketService:        services.ticket,
		RepositoryService:    services.repository,
		RunnerService:        services.runner,
		AgentSvc:             services.agentSvc,
		UserConfigSvc:        services.userConfig,
		PodRouter:            podRouter,
		WorkflowService:      services.workflow,
		WorkflowRunService:   services.workflowRun,
		WorkflowOrchestrator: workflowOrchestrator,
		GoalLoopService:      services.goalLoop,
		BlockstoreService:    services.blockstore,
		KnowledgebaseService: services.knowledgeBase,
		ResourceControl:      services.orchestration,
		WorkerPlanApply:      services.workerApply,
		WorkflowPlanApply:    services.workflowApply,
	}
	grpcServer, grpcRunnerHandler := initializePKIAndGRPC(cfg, services.runner, services.org, services.agentSvc, runnerConnMgr, appLogger, mcpDeps)

	result := &grpcWiringResult{handler: grpcRunnerHandler, server: grpcServer}
	if grpcServer != nil {
		result.runnerAdapter = grpcServer.RunnerAdapter()
		services.goalLoop.SetVerificationDispatcher(result.runnerAdapter)
		grpcCommandSender := grpcserver.NewGRPCCommandSender(result.runnerAdapter)
		podCoordinator.SetCommandSender(grpcCommandSender)
		podRouter.SetCommandSender(grpcCommandSender)
		sandboxQuerySvc.SetSender(grpcCommandSender)
		if sandboxFsSvc != nil {
			sandboxFsSvc.SetSender(grpcCommandSender)
		}
		result.upgradeCommandSender = grpcCommandSender
		result.logUploadSender = grpcCommandSender
		slog.Info("PodCoordinator and PodRouter connected to gRPC Server")
		setupRelayTokenRefreshCallback(db, runnerConnMgr, relayTokenGenerator, grpcCommandSender)
		setupTunnelConnectCallback(cfg, db, runnerConnMgr, relayTokenGenerator, grpcCommandSender)
	}
	return result
}

func buildServicesContainer(
	services *serviceContainer,
	runnerConnMgr *runner.RunnerConnectionManager,
	podCoordinator *runner.PodCoordinator,
	podOrchestrator *agentpod.PodOrchestrator,
	hub *websocket.Hub,
	eventBus *eventbus.EventBus,
	grpcResult *grpcWiringResult,
	sandboxQuerySvc *runner.SandboxQueryService,
	sandboxFsSvc *runner.SandboxFsService,
	logUploadSvc *runnerlogservice.Service,
	relayManager *relay.Manager,
	relayTokenGenerator *relay.TokenGenerator,
	relayDNSService *relay.DNSService,
	relayACMEManager *acme.Manager,
	geoResolver geo.Resolver,
	versionChecker *runner.VersionChecker,
	workflowOrchestrator *workflow.WorkflowOrchestrator,
	workflowScheduler *workflow.WorkflowScheduler,
	redisClient *redis.Client,
	pendingQueueWiring *pendingQueueWiring,
) *v1.Services {
	svc := &v1.Services{
		Auth:                 services.auth,
		User:                 services.user,
		Org:                  services.org,
		AgentSvc:             services.agentSvc,
		UserConfig:           services.userConfig,
		Repository:           services.repository,
		Webhook:              services.webhook,
		Runner:               services.runner,
		RunnerConnMgr:        runnerConnMgr,
		PodCoordinator:       podCoordinator,
		Pod:                  services.pod,
		PodOrchestrator:      podOrchestrator,
		Autopilot:            services.autopilot,
		Channel:              services.channel,
		Ticket:               services.ticket,
		MRSync:               services.mrSync,
		Billing:              services.billing,
		Hub:                  hub,
		EventBus:             eventBus,
		Invitation:           services.invitation,
		PromoCode:            services.promoCode,
		AgentPodSettings:     services.agentpodSettings,
		AgentPodAIProvider:   services.agentpodAIProvider,
		AIResource:           services.aiResource,
		VirtualKey:           services.virtualKey,
		TokenQuota:           services.tokenQuota,
		EnvBundle:            services.envBundle,
		APIKey:               services.apikey,
		APIKeyAdapter:        services.apikeyAdapter,
		File:                 services.file,
		GRPCRunnerHandler:    grpcResult.handler,
		RunnerGRPCAdapter:    grpcResult.runnerAdapter,
		SandboxQueryService:  sandboxQuerySvc,
		SandboxFsService:     sandboxFsSvc,
		UpgradeCommandSender: grpcResult.upgradeCommandSender,
		LogUploadSender:      grpcResult.logUploadSender,
		LogUploadService:     logUploadSvc,
		RelayManager:         relayManager,
		RelayTokenGenerator:  relayTokenGenerator,
		RelayDNSService:      relayDNSService,
		RelayACMEManager:     relayACMEManager,
		GeoResolver:          geoResolver,
		VersionChecker:       versionChecker,
		Extension:            services.extension,
		Workflow:             services.workflow,
		WorkflowRun:          services.workflowRun,
		WorkflowOrchestrator: workflowOrchestrator,
		WorkflowScheduler:    workflowScheduler,
		SSO:                  services.sso,
		SupportTicket:        services.supportTicket,
		TokenUsage:           services.tokenUsage,
		Grant:                services.grant,
		PreviewSessions: previewservice.NewService(
			previewservice.NewSessionStore(redisClient),
			services.user,
			services.org,
			services.pod,
			services.grant,
		),
		Message:  services.message,
		IMBridge: services.imBridge,
		Redis:    redisClient,
	}
	if pendingQueueWiring != nil {
		svc.PendingQueue = pendingQueueWiring.queue
	}
	return svc
}
