package main

import (
	"context"
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
	"gorm.io/gorm"
)

func setupPerpetualPodCallbacks(db *gorm.DB, podCoordinator *runner.PodCoordinator, eventBus *eventbus.EventBus) {
	podCoordinator.SetPodRestartingCallback(func(podKey string, exitCode, restartCount int32) {
		var pod struct {
			OrganizationID int64 `gorm:"column:organization_id"`
		}
		if err := db.Table("pods").Where("pod_key = ?", podKey).First(&pod).Error; err != nil {
			slog.Error("failed to get pod for restarting event", "pod_key", podKey, "error", err)
			return
		}

		data := &eventsv1.PodRestartingEventData{
			PodKey:       podKey,
			ExitCode:     exitCode,
			RestartCount: restartCount,
		}
		event, err := eventbus.NewEntityEvent(eventbus.EventPodRestarting, pod.OrganizationID, "pod", podKey, data)
		if err != nil {
			slog.Error("failed to create pod restarting event", "error", err)
			return
		}
		if err := eventBus.Publish(context.Background(), event); err != nil {
			slog.Error("failed to publish pod restarting event", "error", err)
		}
	})
}
