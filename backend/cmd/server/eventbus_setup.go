package main

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/websocket"
)

func setupEventBusHub(eb *eventbus.EventBus, hub *websocket.Hub) {
	subscriber := websocket.NewHubEventSubscriber(hub, slog.Default())
	subscriber.Subscribe(eb)
}
