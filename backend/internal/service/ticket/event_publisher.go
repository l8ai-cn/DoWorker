package ticket

import (
	"context"
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
)

type EventBusPublisher struct {
	eventBus *eventbus.EventBus
	logger   *slog.Logger
}

func NewEventBusPublisher(eventBus *eventbus.EventBus, logger *slog.Logger) *EventBusPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &EventBusPublisher{
		eventBus: eventBus,
		logger:   logger.With("component", "ticket_event_publisher"),
	}
}

func (p *EventBusPublisher) PublishTicketEvent(ctx context.Context, eventType TicketEventType, orgID int64, slug, status, previousStatus string) {
	if p.eventBus == nil {
		return
	}

	var et eventbus.EventType
	switch eventType {
	case TicketEventCreated:
		et = eventbus.EventTicketCreated
	case TicketEventUpdated:
		et = eventbus.EventTicketUpdated
	case TicketEventStatusChanged:
		et = eventbus.EventTicketStatusChanged
	case TicketEventMoved:
		et = eventbus.EventTicketMoved
	case TicketEventDeleted:
		et = eventbus.EventTicketDeleted
	default:
		p.logger.Warn("unknown ticket event type", "type", eventType)
		return
	}

	data := &eventsv1.TicketStatusChangedEventData{
		Slug:           slug,
		Status:         status,
		PreviousStatus: previousStatus,
	}

	event, err := eventbus.NewEntityEvent(et, orgID, "ticket", slug, data)
	if err != nil {
		p.logger.Error("failed to create ticket event",
			"error", err,
			"type", eventType,
			"slug", slug,
		)
		return
	}

	if err := p.eventBus.Publish(ctx, event); err != nil {
		p.logger.Error("failed to publish ticket event",
			"error", err,
			"type", eventType,
			"slug", slug,
		)
	}
}
