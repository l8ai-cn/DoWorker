package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	grpcserver "github.com/anthropics/agentsmesh/backend/internal/api/grpc"
	"github.com/anthropics/agentsmesh/backend/internal/api/rest"
	"github.com/anthropics/agentsmesh/backend/internal/config"
	notifDomain "github.com/anthropics/agentsmesh/backend/internal/domain/notification"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	blockstoreinfra "github.com/anthropics/agentsmesh/backend/internal/infra/blockstore"
	"github.com/anthropics/agentsmesh/backend/internal/infra/database"
	"github.com/anthropics/agentsmesh/backend/internal/infra/logger"
	otelinit "github.com/anthropics/agentsmesh/backend/internal/infra/otel"
	"github.com/anthropics/agentsmesh/backend/internal/infra/websocket"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	channelService "github.com/anthropics/agentsmesh/backend/internal/service/channel"
	coordinatorsvc "github.com/anthropics/agentsmesh/backend/internal/service/coordinator"
	expertSvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/anthropics/agentsmesh/backend/internal/service/instance"
	notifService "github.com/anthropics/agentsmesh/backend/internal/service/notification"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	skillSvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	workflow "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrate(os.Args[2:])
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg.WarnInsecureDefaults()

	appLogger, err := logger.New(logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		FilePath:   cfg.Log.FilePath,
		MaxSizeMB:  cfg.Log.MaxSizeMB,
		MaxBackups: cfg.Log.MaxBackups,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()
	appLogger.SetDefault()
	slog.Info("Logger initialized", "level", cfg.Log.Level, "file", cfg.Log.FilePath)

	otelProvider, err := otelinit.InitProvider(context.Background(), "do-worker-backend", "1.0.0")
	if err != nil {
		slog.Warn("OpenTelemetry initialization failed, continuing without tracing", "error", err)
		otelProvider = &otelinit.Provider{}
	}
	defer otelProvider.Shutdown(context.Background())

	slog.SetDefault(slog.New(otelinit.NewTraceContextHandler(slog.Default().Handler())))

	db, err := database.New(cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}

	hub, eventBus, redisClient := initializeInfrastructure(cfg, appLogger)
	services, err := initializeServices(cfg, db, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	setupEventBusHub(eventBus, hub)

	ticketEventPublisher := ticket.NewEventBusPublisher(eventBus, appLogger.Logger)
	services.ticket.SetEventPublisher(ticketEventPublisher)
	podEventPublisher := agentpod.NewEventBusPublisher(eventBus, appLogger.Logger)
	services.pod.SetEventPublisher(podEventPublisher)
	services.channel.SetEventBus(eventBus)
	services.channel.SetPodCreatorResolver(services.pod)
	services.blockstore.SetPublisher(blockstoreinfra.NewOpPublisher(eventBus))

	notifRelay := websocket.NewNotificationRelay(hub, redisClient, appLogger.Logger)
	notifRelay.StartSubscriber(context.Background())

	notifDispatcher := notifService.NewDispatcher(notifRelay, services.notifPrefStore)
	notifDispatcher.RegisterResolver("pod_creator", notifService.NewPodCreatorResolver(services.pod))
	notifDispatcher.RegisterResolver("channel_members", notifService.NewChannelMemberResolver(services.channel))
	services.notifDispatcher = notifDispatcher

	channelRepo := infra.NewChannelRepository(db)
	userLookup := infra.NewChannelUserLookup(db)
	podLookup := infra.NewChannelPodLookup(db)
	channelUserNames := infra.NewChannelUserNameResolver(db)
	services.channel.SetUserLookup(userLookup)
	services.channel.AddPostSendHook(channelService.NewMentionValidatorHook(userLookup, podLookup, channelRepo))
	services.channel.AddPostSendHook(channelService.NewEventPublishHook(eventBus, channelUserNames, services.channel))
	services.channel.AddPostSendHook(channelService.NewNotificationHook(notifDispatcher, channelUserNames))
	if services.imBridge != nil {
		services.channel.AddPostSendHook(services.imBridge.OutboundHook())
		services.imBridge.StartMonitor(context.Background())
	}
	slog.Info("Channel PostSendHooks registered")

	services.user.AddPreDeleteHook(func(ctx context.Context, userID int64) error {
		return services.channel.CleanupUserReferences(ctx, userID)
	})

	if redisClient != nil {
		eventBus.StartRedisSubscriber(context.Background())
	}

	runnerConnMgr, podCoordinator, podRouter, heartbeatBatcher, sandboxQuerySvc, sandboxFsSvc := initializeRunnerComponents(services.pod, services.runnerRepo, redisClient, appLogger, services.agentSvc)

	pendingQueueWiring := initializePendingQueue(cfg, db, services.pod, podCoordinator, runnerConnMgr, eventBus, appLogger)

	podCoordinator.SetAutopilotRepo(services.autopilotRepo)

	podCoordinator.SetTokenUsageService(services.tokenUsage)

	relayManager := relay.NewManagerWithOptions()
	relayTokenGenerator := relay.NewTokenGenerator(cfg.JWT.Secret, "agentsmesh-relay")
	relayDNSService, relayACMEManager := initializeRelayServices(cfg)
	slog.Info("Relay services initialized")

	geoResolver := initializeGeoResolver()
	defer geoResolver.Close()

	podRouter.SetEventBus(eventBus)
	podRouter.SetPodInfoGetter(services.pod)

	podRouter.SetNotifyFunc(func(ctx context.Context, orgID int64, source, entityID, title, body, link, resolver string) {
		if err := notifDispatcher.Dispatch(ctx, &notifDomain.NotificationRequest{
			OrganizationID:    orgID,
			Source:            source,
			SourceEntityID:    entityID,
			Title:             title,
			Body:              body,
			Link:              link,
			RecipientResolver: resolver,
		}); err != nil {
			slog.Error("failed to dispatch notification", "source", source, "error", err)
		}
		RecordOSCNotification(entityID)
	})

	podOrchestrator := createPodOrchestrator(services, podCoordinator)
	if err := attachOrchestrationWorkerApply(
		services,
		podOrchestrator,
		pendingQueueWiring.queue,
	); err != nil {
		log.Fatalf("Failed to initialize orchestration Worker apply: %v", err)
	}
	services.mesh.SetPodCreator(podOrchestrator)

	services.channel.AddPostSendHook(channelService.NewPodPromptHook(podRouter, channelRepo, runner.NewChannelPromptQueuer(services.pod, pendingQueueWiring.queue)))
	slog.Info("PodPromptHook registered with PodRouter")

	setupRunnerEventCallbacks(db, runnerConnMgr, eventBus)
	setupPodEventCallbacks(db, podCoordinator, eventBus, notifDispatcher)
	setupPerpetualPodCallbacks(db, podCoordinator, eventBus)
	startOSCDedupCleanup()

	services.autopilot.SetCommandSender(podCoordinator)

	runnerOrgQuerier := infra.NewRunnerOrgQuerier(db)
	orgAwareness := instance.NewOrgAwarenessService(runnerOrgQuerier, runnerConnMgr, redisClient, cfg.Server.Address, appLogger.Logger)
	orgAwareness.Start()
	setupOrgAwarenessRefresh(eventBus, orgAwareness)
	slog.Info("OrgAwarenessService started")

	workflowOrchestrator := workflow.NewWorkflowOrchestrator(services.workflow, services.workflowRun, eventBus, appLogger.Logger)
	workflowOrchestrator.SetPodDependencies(podOrchestrator, services.autopilot, podCoordinator, services.ticket, services.repository)
	workflowScheduler := workflow.NewWorkflowScheduler(services.workflow, workflowOrchestrator, orgAwareness, appLogger.Logger)
	workflowScheduler.Start()
	setupWorkflowEventSubscriptions(eventBus, workflowOrchestrator)
	slog.Info("Workflow orchestrator and scheduler created")

	goalLoopTimeoutMonitor := configureGoalLoopService(services, podOrchestrator, podCoordinator, pendingQueueWiring.queue, eventBus, appLogger.Logger)
	defer goalLoopTimeoutMonitor.Stop()

	coordinatorEnsurer := coordinatorsvc.NewRunnerEnsurer(services.runner, nil, appLogger.Logger)
	if launcher, kind, err := coordinatorsvc.NewRunnerLauncherFromEnv(
		services.workerRuntimeCatalog,
		services.workerDefinitions.Slugs(),
		appLogger.Logger,
	); err != nil {
		slog.Error("Coordinator runner launcher config invalid", "error", err)
	} else if launcher != nil {
		coordinatorEnsurer = coordinatorsvc.NewRunnerEnsurer(services.runner, launcher, appLogger.Logger)
		slog.Info("Coordinator runner auto-provision enabled", "launcher", kind)
	}
	coordinatorSvc := coordinatorsvc.NewService(coordinatorsvc.Deps{
		Store:         infra.NewCoordinatorRepository(db),
		Tickets:       services.ticket,
		Dispatch:      podOrchestrator,
		Platform:      coordinatorsvc.NewPlatformFactory(services.repository, services.user),
		RunnerEnsurer: coordinatorEnsurer,
		Logger:        appLogger.Logger,
	})
	coordinatorScheduler := coordinatorsvc.NewScheduler(coordinatorSvc, appLogger.Logger)
	coordinatorScheduler.Start()
	defer coordinatorScheduler.Stop()
	setupCoordinatorEventSubscriptions(eventBus, coordinatorSvc)
	slog.Info("Coordinator service and scheduler created")

	grpcResult := initPKIAndGRPCWiring(cfg, services, runnerConnMgr, podCoordinator, podRouter, sandboxQuerySvc, sandboxFsSvc, podOrchestrator, workflowOrchestrator, services.workflowRun, appLogger, relayTokenGenerator, db)

	if grpcResult.server != nil {
		wirePendingQueueSender(pendingQueueWiring, grpcserver.NewGRPCCommandSender(grpcResult.server.RunnerAdapter()))
	}

	versionChecker := runner.NewVersionChecker(redisClient)
	if versionChecker != nil {
		versionChecker.Start(context.Background())
	}

	logUploadSvc := initLogUploadService(cfg, db, runnerConnMgr)

	svc := buildServicesContainer(services, runnerConnMgr, podCoordinator, podOrchestrator, hub, eventBus,
		grpcResult, sandboxQuerySvc, sandboxFsSvc, logUploadSvc, relayManager, relayTokenGenerator, relayDNSService,
		relayACMEManager, geoResolver, versionChecker, workflowOrchestrator, workflowScheduler, redisClient, pendingQueueWiring)
	svc.Coordinator = coordinatorSvc
	expertGitops := gitops.NewService(newGiteaClientForNamespace(cfg, "am-experts"), appLogger.Logger)
	skillStore := infra.NewSkillCatalogRepository(db)
	svc.Expert = expertSvc.NewService(expertSvc.Deps{
		Store:       infra.NewExpertRepository(db),
		Pods:        services.pod,
		Dispatch:    podOrchestrator,
		Repos:       services.repository,
		WorkerSpecs: services.workerSpecs,
		Skills:      skillStore,
		Gitops:      expertGitops,
		Logger:      appLogger.Logger,
	})

	// Unified skill service (namespace am-skills): one internal git repo per
	// skill, for both in-platform authoring and external imports. It reuses
	// the extension packager for the package->object-storage pipeline.
	// NewService returns nil (routes no-op) when gitea or the packager is not
	// configured.
	skillGitops := gitops.NewService(newGiteaClientForNamespace(cfg, "am-skills"), appLogger.Logger)
	var skillPackager skillSvc.SkillPackagerBridge
	if services.extension != nil {
		if pkg := services.extension.SkillPackager(); pkg != nil {
			skillPackager = pkg
		}
	}
	svc.Skill = skillSvc.NewService(skillSvc.Deps{
		Store:    skillStore,
		Gitops:   skillGitops,
		Packager: skillPackager,
		Logger:   appLogger.Logger,
	})

	router := rest.NewRouter(cfg, svc, db, appLogger.Logger, redisClient)
	grpcResult.server = startGRPCServer(cfg, grpcResult.server)

	cleanupMkt := startMarketplaceWorker(services)
	defer cleanupMkt()

	cleanupKbSync := startKnowledgeBaseSyncWorker(services)
	defer cleanupKbSync()

	subscriptionScheduler := startSubscriptionJobs(db, cfg, services.email, appLogger.Logger)
	taskManager := startTaskManager(podCoordinator, services.message, redisClient, appLogger.Logger)

	// Start HTTP server (Connect-RPC handlers wrap the Gin router)
	srv := startHTTPServer(cfg, wrapWithConnect(cfg, services, svc, router))

	waitForShutdown(srv, grpcResult.server, eventBus, heartbeatBatcher, subscriptionScheduler, taskManager, workflowScheduler, orgAwareness, relayManager, services, db, redisClient)
}
