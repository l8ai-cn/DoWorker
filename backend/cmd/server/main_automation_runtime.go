package main

import (
	"context"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	notificationdomain "github.com/anthropics/agentsmesh/backend/internal/domain/notification"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	goalloop "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	"github.com/anthropics/agentsmesh/backend/internal/service/instance"
	notifservice "github.com/anthropics/agentsmesh/backend/internal/service/notification"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	workflow "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func configurePodNotificationRouting(
	podRouter *runner.PodRouter,
	notifications *notifservice.Dispatcher,
) {
	podRouter.SetNotifyFunc(func(
		ctx context.Context,
		orgID int64,
		source, entityID, title, body, link, resolver string,
	) {
		err := notifications.Dispatch(ctx, &notificationdomain.NotificationRequest{
			OrganizationID:    orgID,
			Source:            source,
			SourceEntityID:    entityID,
			Title:             title,
			Body:              body,
			Link:              link,
			RecipientResolver: resolver,
		})
		if err != nil {
			slog.Error(
				"failed to dispatch notification",
				"source",
				source,
				"error",
				err,
			)
		}
		RecordOSCNotification(entityID)
	})
}

func configurePodRuntimeEvents(
	db *gorm.DB,
	runnerConnections *runner.RunnerConnectionManager,
	podCoordinator *runner.PodCoordinator,
	eventBus *eventbus.EventBus,
	notifications *notifservice.Dispatcher,
) {
	setupRunnerEventCallbacks(db, runnerConnections, eventBus)
	setupPodEventCallbacks(db, podCoordinator, eventBus, notifications)
	setupPerpetualPodCallbacks(db, podCoordinator, eventBus)
	startOSCDedupCleanup()
}

func initializeAutomationRuntime(
	cfg *config.Config,
	services *serviceContainer,
	db *gorm.DB,
	runnerConnections *runner.RunnerConnectionManager,
	redisClient *redis.Client,
	eventBus *eventbus.EventBus,
	podOrchestrator *agentpod.PodOrchestrator,
	podCoordinator *runner.PodCoordinator,
	promptDispatcher goalloop.PromptDispatcher,
	logger *slog.Logger,
) (
	*instance.OrgAwarenessService,
	*workflow.WorkflowOrchestrator,
	*workflow.WorkflowScheduler,
	func(),
) {
	services.autopilot.SetCommandSender(podCoordinator)
	orgAwareness := instance.NewOrgAwarenessService(
		infra.NewRunnerOrgQuerier(db),
		runnerConnections,
		redisClient,
		cfg.Server.Address,
		logger,
	)
	orgAwareness.Start()
	setupOrgAwarenessRefresh(eventBus, orgAwareness)

	orchestrator := workflow.NewWorkflowOrchestrator(
		services.workflow,
		services.workflowRun,
		eventBus,
		logger,
	)
	orchestrator.SetPodDependencies(
		podOrchestrator,
		services.autopilot,
		podCoordinator,
		services.ticket,
		services.repository,
	)
	scheduler := workflow.NewWorkflowScheduler(
		services.workflow,
		orchestrator,
		orgAwareness,
		logger,
	)
	scheduler.Start()
	setupWorkflowEventSubscriptions(eventBus, orchestrator)

	monitor := configureGoalLoopService(
		services,
		podOrchestrator,
		podCoordinator,
		promptDispatcher,
		eventBus,
		logger,
	)
	return orgAwareness, orchestrator, scheduler, monitor.Stop
}
