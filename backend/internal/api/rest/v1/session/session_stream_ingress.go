package sessionapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (p *SessionStreamPublisher) StartAssistantTurn(sessionID string) {
	if p == nil {
		return
	}
	id, err := itemsvcNewTurnID()
	if err != nil {
		return
	}
	p.publishTurnStarted(sessionID, id)
}

func (p *SessionStreamPublisher) PublishSessionStatus(sessionID, status string) {
	if p == nil || p.Hub == nil {
		return
	}
	p.Hub.Publish(sessionID, formatSSE(sseSessionStatus, map[string]any{
		"conversation_id": sessionID, "status": status,
	}))
}

func (p *SessionStreamPublisher) PublishSessionInterrupted(sessionID, responseID string) {
	if p == nil || p.Hub == nil {
		return
	}
	data := map[string]any{"requested_at": time.Now().Unix()}
	if responseID != "" {
		data["response_id"] = responseID
	}
	p.Hub.Publish(sessionID, formatSSE(sseSessionInterrupted, map[string]any{
		"type": sseSessionInterrupted,
		"data": data,
	}))
}

func (p *SessionStreamPublisher) PublishInputConsumed(sessionID, itemID, author string, content []map[string]any) {
	if p == nil || p.Hub == nil {
		return
	}
	p.Hub.Publish(sessionID, formatSSE(sseSessionInputConsumed, map[string]any{
		"type": sseSessionInputConsumed,
		"data": map[string]any{
			"item_id": itemID, "type": "message", "created_by": author,
			"data": map[string]any{"role": "user", "content": content},
		},
	}))
}

func (p *SessionStreamPublisher) PublishElicitationResolved(sessionID, elicitID string) {
	if p == nil || p.Hub == nil {
		return
	}
	p.Hub.Publish(sessionID, formatSSE(sseElicitationResolved, map[string]any{
		"elicitation_id": elicitID,
	}))
}

func (p *SessionStreamPublisher) HandleAcpSession(
	ctx context.Context,
	runnerID int64,
	podKey, eventType, payloadJSON string,
) {
	if p == nil || p.Hub == nil {
		return
	}
	sessionID, ok := p.sessionForRunnerPod(ctx, runnerID, podKey)
	if !ok {
		slog.Debug("session stream: drop acp event, no session for pod", "pod_key", podKey, "event", eventType)
		return
	}
	switch eventType {
	case "contentChunk":
		var chunk struct {
			Text string `json:"text"`
			Role string `json:"role"`
		}
		if json.Unmarshal([]byte(payloadJSON), &chunk) != nil || chunk.Role != "assistant" || chunk.Text == "" {
			return
		}
		turnID := p.ensureTurn(sessionID, "")
		p.Hub.AppendDelta(sessionID, chunk.Text)
		p.Hub.Publish(sessionID, formatSSE(sseTurnTextDelta, map[string]any{
			"delta": chunk.Text, "response_id": turnID,
		}))
	case "sessionState":
		var body struct {
			State string `json:"state"`
		}
		if json.Unmarshal([]byte(payloadJSON), &body) != nil || body.State != "idle" {
			return
		}
		p.PublishSessionStatus(sessionID, "idle")
		turnID, text, ok := p.Hub.FinishTurn(sessionID)
		if !ok || text == "" {
			return
		}
		if turnID == "" {
			turnID = p.ensureTurn(sessionID, "")
		}
		p.flushReasoningItem(ctx, sessionID, turnID)
		p.finishAssistantMessage(ctx, sessionID, turnID, text)
	case "permissionRequest":
		p.handlePermissionRequest(sessionID, payloadJSON)
	case "toolCallUpdate":
		p.handleToolCallUpdate(ctx, sessionID, payloadJSON)
	case "toolCallResult":
		p.handleToolCallResult(ctx, sessionID, payloadJSON)
	case "thinkingUpdate":
		p.handleThinkingUpdate(ctx, sessionID, payloadJSON)
	case "log":
		p.handleAcpLog(ctx, sessionID, payloadJSON)
	}
}

func (p *SessionStreamPublisher) PublishPodStatus(ctx context.Context, podKey, podStatus, agentStatus string) {
	if p == nil || p.Hub == nil {
		return
	}
	sessionID, ok := p.sessionForPod(ctx, podKey)
	if !ok {
		return
	}
	status := mapPodSessionStatus(podStatus, agentStatus)
	p.PublishSessionStatus(sessionID, status)
	if p.Updates != nil {
		p.Updates.NotifyChanged(sessionID)
	}
}

func (p *SessionStreamPublisher) HandlePodUsage(
	ctx context.Context,
	runnerID int64,
	evt *runnerv1.PodUsageEvent,
) {
	if p == nil || evt == nil || evt.GetPodKey() == "" || p.Usage == nil {
		return
	}
	if !p.runnerOwnsPod(ctx, runnerID, evt.GetPodKey()) {
		return
	}
	_ = p.Usage.Upsert(ctx, evt.GetPodKey(), evt.GetModel(),
		evt.GetInputTokens(), evt.GetOutputTokens(),
		evt.GetCacheReadTokens(), evt.GetCacheCreationTokens())
	if p.Hub == nil {
		return
	}
	sessionID, ok := p.sessionForPod(ctx, evt.GetPodKey())
	if !ok {
		return
	}
	agg, err := p.Usage.Aggregate(ctx, evt.GetPodKey())
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
	p.Hub.Publish(sessionID, formatSSE(sseSessionUsage, payload))
}

func (p *SessionStreamPublisher) UpdateExternalSessionID(
	ctx context.Context,
	runnerID int64,
	podKey, externalID string,
) {
	if p == nil || p.Pods == nil || podKey == "" || externalID == "" {
		return
	}
	if !p.runnerOwnsPod(ctx, runnerID, podKey) {
		return
	}
	_ = p.Pods.UpdateExternalSessionID(ctx, podKey, externalID)
}
