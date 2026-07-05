package omnigent

import (
	"context"

	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

var (
	usageSvc   *sessionusagesvc.Service
	usageHub   *SessionHub
	podUpdater PodUpdater
)

func SetSessionUsage(svc *sessionusagesvc.Service, hub *SessionHub) {
	usageSvc, usageHub = svc, hub
}

func SetPodUpdater(updater PodUpdater) { podUpdater = updater }

func ForwardPodUsage(ctx context.Context, evt *runnerv1.PodUsageEvent) {
	if evt == nil || evt.GetPodKey() == "" {
		return
	}
	if usageSvc == nil {
		return
	}
	_ = usageSvc.Upsert(ctx, evt.GetPodKey(), evt.GetModel(),
		evt.GetInputTokens(), evt.GetOutputTokens(),
		evt.GetCacheReadTokens(), evt.GetCacheCreationTokens())
	if usageHub == nil || globalBridge == nil {
		return
	}
	sessionID, ok := globalBridge.sessionForPod(ctx, evt.GetPodKey())
	if !ok {
		return
	}
	agg, err := usageSvc.Aggregate(ctx, evt.GetPodKey())
	if err != nil {
		return
	}
	payload := map[string]any{"conversation_id": sessionID}
	if agg.TotalCostUSD != nil {
		payload["total_cost_usd"] = *agg.TotalCostUSD
	}
	if len(agg.UsageByModel) > 0 {
		payload["usage_by_model"] = agg.UsageByModel
	}
	usageHub.Publish(sessionID, formatSSE("session.usage", payload))
}

func ForwardExternalSession(ctx context.Context, podKey, externalID string) {
	if podUpdater == nil || podKey == "" || externalID == "" {
		return
	}
	_ = podUpdater.UpdateExternalSessionID(ctx, podKey, externalID)
}

type PodUpdater interface {
	UpdateExternalSessionID(ctx context.Context, podKey, externalID string) error
}
