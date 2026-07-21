package main

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	goalloop "github.com/l8ai-cn/agentcloud/backend/internal/service/goalloop"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
)

func setupGoalLoopEventSubscriptions(eventBus *eventbus.EventBus, service *goalloop.Service) {
	eventBus.Subscribe(eventbus.EventPodTerminated, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			slog.Error("failed to decode goal loop pod event", "error", err)
			return
		}
		if err := service.HandlePodStatus(context.Background(), data.PodKey, data.Status); err != nil {
			slog.Error("failed to handle goal loop pod event", "pod_key", data.PodKey, "error", err)
		}
	})

	eventBus.Subscribe(eventbus.EventPodStatusChanged, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		switch data.Status {
		case agentpod.StatusCompleted, agentpod.StatusError, agentpod.StatusTerminated:
			if err := service.HandlePodStatus(context.Background(), data.PodKey, data.Status); err != nil {
				slog.Error("failed to handle goal loop pod status", "pod_key", data.PodKey, "error", err)
			}
		}
	})

	eventBus.Subscribe(eventbus.EventPodAgentChanged, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		if err := service.HandlePodAgentStatus(
			context.Background(),
			data.PodKey,
			data.AgentStatus,
			time.UnixMilli(event.Timestamp),
		); err != nil {
			slog.Error(
				"failed to handle goal loop worker status",
				"pod_key", data.PodKey,
				"agent_status", data.AgentStatus,
				"error", err,
			)
		}
	})

	eventBus.Subscribe(eventbus.EventAutopilotStatusChanged, func(event *eventbus.Event) {
		var data eventsv1.AutopilotStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		if err := service.HandleAutopilotStatus(context.Background(), data.AutopilotControllerKey, data.Phase); err != nil {
			slog.Error("failed to handle goal loop autopilot status", "autopilot_key", data.AutopilotControllerKey, "error", err)
		}
	})

	slog.Info("Goal Loop event subscriptions registered")
}

func configureGoalLoopService(
	services *serviceContainer,
	podCreator goalloop.PodCreator,
	podTerminator goalloop.PodTerminator,
	promptDispatcher goalloop.PromptDispatcher,
	eventBus *eventbus.EventBus,
	logger *slog.Logger,
) *goalloop.TimeoutMonitor {
	services.goalLoop.SetWorkerSpecSnapshotLoader(services.workerSpecs)
	services.goalLoop.SetWorkerTypeSnapshotValidator(services.workerCreation)
	services.goalLoop.SetExecutionDependencies(podCreator, services.pod, podTerminator)
	services.goalLoop.SetPromptDispatcher(promptDispatcher)
	setupGoalLoopEventSubscriptions(eventBus, services.goalLoop)
	monitor := goalloop.NewTimeoutMonitor(services.goalLoop, logger)
	monitor.Start()
	slog.Info("Goal Loop service configured")
	return monitor
}
