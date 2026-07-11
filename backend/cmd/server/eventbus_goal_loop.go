package main

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	goalloop "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	eventsv1 "github.com/anthropics/agentsmesh/proto/gen/go/events/v1"
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
