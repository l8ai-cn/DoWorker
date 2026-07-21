package websocket

import (
	"encoding/json"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
)

func TestHubEventSubscriber_EntityEvent(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	eb := eventbus.NewEventBus(nil, nil)
	defer eb.Close()

	NewHubEventSubscriber(hub, nil).Subscribe(eb)

	event, err := eventbus.NewEntityEvent(eventbus.EventPodCreated, 1, "pod", "pod-123",
		&eventsv1.PodCreatedEventData{PodKey: "pod-123", Status: "running"})
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	if err := eb.Publish(t.Context(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
}

func TestHubEventSubscriber_SystemEvent(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	eb := eventbus.NewEventBus(nil, nil)
	defer eb.Close()

	NewHubEventSubscriber(hub, nil).Subscribe(eb)

	data, _ := json.Marshal(map[string]string{"msg": "maintenance"})
	event := &eventbus.Event{
		Type:           eventbus.EventSystemMaintenance,
		Category:       eventbus.CategorySystem,
		OrganizationID: 1,
		Data:           data,
	}

	if err := eb.Publish(t.Context(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
}
