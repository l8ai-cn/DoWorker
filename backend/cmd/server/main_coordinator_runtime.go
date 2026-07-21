package main

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	coordinatorsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/coordinator"
	"gorm.io/gorm"
)

func initializeCoordinatorRuntime(
	services *serviceContainer,
	db *gorm.DB,
	podOrchestrator *agentpod.PodOrchestrator,
	eventBus *eventbus.EventBus,
	logger *slog.Logger,
) (*coordinatorsvc.Service, *coordinatorsvc.Scheduler) {
	runnerEnsurer := coordinatorsvc.NewRunnerEnsurer(
		services.runner,
		nil,
		logger,
	)
	if launcher, kind, err := coordinatorsvc.NewRunnerLauncherFromEnv(
		services.workerRuntimeCatalog,
		services.workerDefinitions.Slugs(),
		logger,
	); err != nil {
		slog.Error("Coordinator runner launcher config invalid", "error", err)
	} else if launcher != nil {
		runnerEnsurer = coordinatorsvc.NewRunnerEnsurer(
			services.runner,
			launcher,
			logger,
		)
		slog.Info("Coordinator runner auto-provision enabled", "launcher", kind)
	}
	service := coordinatorsvc.NewService(coordinatorsvc.Deps{
		Store:         infra.NewCoordinatorRepository(db),
		Tickets:       services.ticket,
		Dispatch:      podOrchestrator,
		Platform:      coordinatorsvc.NewPlatformFactory(services.repository, services.user),
		RunnerEnsurer: runnerEnsurer,
		Snapshots:     infra.NewWorkerSpecSnapshotRepository(db),
		Artifacts:     infra.NewWorkerSpecDependencyArtifactRepository(db),
		Logger:        logger,
	})
	scheduler := coordinatorsvc.NewScheduler(service, logger)
	scheduler.Start()
	setupCoordinatorEventSubscriptions(eventBus, service)
	slog.Info("Coordinator service and scheduler created")
	return service, scheduler
}
