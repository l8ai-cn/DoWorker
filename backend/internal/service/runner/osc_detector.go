package runner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
)

type NotifyFunc func(ctx context.Context, orgID int64, source, entityID, title, body, link, resolver string)

type OSCDetector struct {
	eventBus      *eventbus.EventBus
	podInfoGetter PodInfoGetter
	notifyFunc    NotifyFunc
}

func NewOSCDetector(eventBus *eventbus.EventBus, podInfoGetter PodInfoGetter) *OSCDetector {
	return &OSCDetector{
		eventBus:      eventBus,
		podInfoGetter: podInfoGetter,
	}
}

// PublishNotification routes via NotificationDispatcher (preference-aware) when notifyFunc is set,
// else falls back to direct EventBus publish.
func (d *OSCDetector) PublishNotification(ctx context.Context, podKey, title, body string) bool {
	if d.podInfoGetter == nil {
		return false
	}

	orgID, _, err := d.podInfoGetter.GetPodOrganizationAndCreator(ctx, podKey)
	if err != nil {
		return false
	}

	if d.notifyFunc != nil {
		resolver := fmt.Sprintf("pod_creator:%s", podKey)
		link := fmt.Sprintf("/workspace?pod=%s", podKey)
		d.notifyFunc(ctx, orgID, "terminal:osc", podKey, title, body, link, resolver)
		return true
	}

	slog.WarnContext(ctx, "OSC notification dropped: notifyFunc not configured", "pod_key", podKey)
	return false
}

func (d *OSCDetector) PublishTitle(ctx context.Context, podKey, title string) bool {
	if d.eventBus == nil || d.podInfoGetter == nil {
		return false
	}

	orgID, _, err := d.podInfoGetter.GetPodOrganizationAndCreator(ctx, podKey)
	if err != nil {
		return false
	}

	if err := d.podInfoGetter.UpdatePodTitle(ctx, podKey, title); err != nil {
	}

	titleData, err := eventbus.MarshalEventData(&eventsv1.PodTitleChangedEventData{
		PodKey: podKey,
		Title:  title,
	})
	if err != nil {
		slog.Error("failed to marshal pod title event", "error", err)
		return false
	}
	d.eventBus.Publish(ctx, &eventbus.Event{
		Type:           eventbus.EventPodTitleChanged,
		Category:       eventbus.CategoryEntity,
		OrganizationID: orgID,
		EntityType:     "pod",
		EntityID:       podKey,
		Data:           titleData,
	})

	return true
}
