package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcserver "github.com/l8ai-cn/agentcloud/backend/internal/api/grpc"
	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/database"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/email"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/job"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/instance"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	tasksvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tasks"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func startHTTPServer(cfg *config.Config, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Starting server", "address", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return srv
}

func startSubscriptionJobs(db *gorm.DB, appConfig *config.Config, emailSvc email.Service, logger *slog.Logger) *job.SubscriptionScheduler {
	scheduler := job.NewSubscriptionScheduler(db, appConfig, emailSvc, logger)
	scheduler.Start()
	slog.Info("subscription scheduler started")
	return scheduler
}

type WorkflowSchedulerStopper interface {
	Stop()
}

func waitForShutdown(
	srv *http.Server,
	grpcServer *grpcserver.Server,
	eventBus *eventbus.EventBus,
	heartbeatBatcher *runner.HeartbeatBatcher,
	subscriptionScheduler *job.SubscriptionScheduler,
	taskManager *tasksvc.Manager,
	workflowScheduler WorkflowSchedulerStopper,
	orgAwareness *instance.OrgAwarenessService,
	relayManager *relay.Manager,
	services *serviceContainer,
	db *gorm.DB,
	redisClient *redis.Client,
) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if grpcServer != nil {
		grpcServer.Stop()
	}

	if subscriptionScheduler != nil {
		subscriptionScheduler.Stop()
	}

	if taskManager != nil {
		taskManager.Stop()
	}

	if workflowScheduler != nil {
		workflowScheduler.Stop()
	}

	if orgAwareness != nil {
		orgAwareness.Stop()
	}

	if heartbeatBatcher != nil {
		heartbeatBatcher.Stop()
	}

	if relayManager != nil {
		relayManager.Stop()
	}

	if services != nil {
		services.Close()
	}

	eventBus.Close()

	if db != nil {
		if err := database.Close(db); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}

	if redisClient != nil {
		redisClient.Close()
	}

	slog.Info("Server exited")
}
