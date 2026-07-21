package main

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	tasksvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tasks"
	"github.com/redis/go-redis/v9"
)

func startTaskManager(
	podCoordinator *runner.PodCoordinator,
	messageSvc *agent.MessageService,
	redisClient *redis.Client,
	logger *slog.Logger,
) *tasksvc.Manager {
	if redisClient == nil {
		slog.Info("task manager skipped: redis unavailable")
		return nil
	}

	mgr := tasksvc.NewManager(podCoordinator, redisClient, logger, tasksvc.DefaultConfig())
	if messageSvc != nil {
		mgr.SetDeadLetterCleaner(messageSvc)
	}
	if err := mgr.Start(); err != nil {
		slog.Error("failed to start task manager", "error", err)
		return nil
	}
	slog.Info("task manager started")
	return mgr
}
