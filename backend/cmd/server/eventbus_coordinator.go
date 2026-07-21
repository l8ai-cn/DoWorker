package main

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	coordinatorsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/coordinator"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
)

// setupCoordinatorEventSubscriptions wires pod terminal events to the
// coordinator so it can post feedback to the external task source and advance
// the materialized ticket. Non-coordinator pods early-return inside the handler
// (no execution row keyed by pod_key).
func setupCoordinatorEventSubscriptions(eventBus *eventbus.EventBus, svc *coordinatorsvc.Service) {
	eventBus.Subscribe(eventbus.EventPodTerminated, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			slog.Error("failed to unmarshal pod terminated event for coordinator", "error", err)
			return
		}
		svc.HandlePodTerminated(context.Background(), data.PodKey, data.Status)
	})

	eventBus.Subscribe(eventbus.EventPodStatusChanged, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		switch data.Status {
		case agentpod.StatusCompleted, agentpod.StatusError:
			svc.HandlePodTerminated(context.Background(), data.PodKey, data.Status)
		}
	})

	slog.Info("Coordinator event subscriptions registered")
}
