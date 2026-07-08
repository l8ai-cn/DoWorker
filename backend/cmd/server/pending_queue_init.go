package main

import (
	"context"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/infra/logger"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"gorm.io/gorm"
)

type pendingQueueWiring struct {
	queue   *runner.PendingCommandQueue
	drainer *runner.PendingCommandDrainer
}

func initializePendingQueue(
	cfg *config.Config,
	db *gorm.DB,
	podSvc *agentpod.PodService,
	podCoordinator *runner.PodCoordinator,
	runnerConnMgr *runner.RunnerConnectionManager,
	eventBus *eventbus.EventBus,
	appLogger *logger.Logger,
) *pendingQueueWiring {
	pqCfg := cfg.PendingQueue
	repo := infra.NewPendingCommandRepository(db)
	queue := runner.NewPendingCommandQueue(
		repo,
		eventBus,
		pqCfg.MaxPerRunner,
		pqCfg.DefaultTTL,
		pqCfg.Enabled,
		appLogger.Logger,
	)
	drainer := runner.NewPendingCommandDrainer(
		repo,
		podSvc,
		podCoordinator.GetRunnerRepo(),
		nil,
		runnerConnMgr,
		podCoordinator,
		podSvc,
		eventBus,
		pqCfg.SweepInterval,
		appLogger.Logger,
	)
	queue.SetDrainer(drainer)
	queue.SetConnectionChecker(runnerConnMgr)
	podCoordinator.SetPendingQueue(queue)
	podCoordinator.SetPendingDrainer(drainer)
	podCoordinator.SetConnectionChecker(runnerConnMgr)
	drainer.SetQueueExpiredNotifier(func(ctx context.Context, podKey string) {
		podCoordinator.NotifyPodStatus(podKey, podDomain.StatusError, "")
	})

	origInit := runnerConnMgr.GetInitializedCallback()
	runnerConnMgr.SetInitializedCallback(func(runnerID int64, agents []string) {
		if origInit != nil {
			origInit(runnerID, agents)
		}
		drainer.DrainRunner(runnerID)
	})

	drainer.StartExpirySweeper(context.Background())
	slog.Info("Pending command queue initialized", "enabled", pqCfg.Enabled, "max_per_runner", pqCfg.MaxPerRunner)
	return &pendingQueueWiring{queue: queue, drainer: drainer}
}

func wirePendingQueueSender(w *pendingQueueWiring, sender runner.ServerMessageSender) {
	if w == nil || w.drainer == nil || sender == nil {
		return
	}
	w.drainer.SetMessageSender(sender)
}
