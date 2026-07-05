package omnigent

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
)

type EventBridge struct {
	Items        *itemsvc.Service
	Sessions     *sessionsvc.Service
	Hub          *SessionHub
	Elicitations *ElicitationStore
}

func (b *EventBridge) sessionForPod(ctx context.Context, podKey string) (string, bool) {
	if b == nil || b.Sessions == nil {
		return "", false
	}
	row, err := b.Sessions.GetByPodKey(ctx, podKey)
	if err != nil || row == nil {
		return "", false
	}
	return row.ID, true
}

func (b *EventBridge) ensureResponseTurn(sessionID, responseID string) string {
	if responseID != "" {
		if active, ok := b.Hub.ActiveResponse(sessionID); !ok || active != responseID {
			b.Hub.StartTurn(sessionID, responseID)
		}
		return responseID
	}
	if active, ok := b.Hub.ActiveResponse(sessionID); ok {
		return active
	}
	respID, err := itemsvc.NewResponseID()
	if err != nil {
		return ""
	}
	b.Hub.StartTurn(sessionID, respID)
	now := time.Now().Unix()
	b.Hub.Publish(sessionID, formatSSE("response.created", map[string]any{
		"id": respID, "status": "in_progress", "model": "", "created_at": now,
		"conversation": map[string]any{"id": sessionID},
	}))
	b.Hub.Publish(sessionID, formatSSE("response.in_progress", map[string]any{
		"id": respID, "status": "in_progress", "model": "", "created_at": now,
	}))
	return respID
}

func (b *EventBridge) HandleAcpSession(ctx context.Context, podKey, eventType, payloadJSON string) {
	if b == nil || b.Hub == nil {
		return
	}
	sessionID, ok := b.sessionForPod(ctx, podKey)
	if !ok {
		slog.Debug("omnigent: drop acp event, no session for pod", "pod_key", podKey, "event", eventType)
		return
	}
	switch eventType {
	case "content_delta":
		var p struct {
			Delta      string `json:"delta"`
			ResponseID string `json:"response_id"`
		}
		if json.Unmarshal([]byte(payloadJSON), &p) != nil || p.Delta == "" {
			return
		}
		respID := b.ensureResponseTurn(sessionID, p.ResponseID)
		b.Hub.AppendDelta(sessionID, p.Delta)
		b.Hub.Publish(sessionID, formatSSE("response.output_text.delta", map[string]any{
			"delta": p.Delta, "response_id": respID,
		}))
	case "message_done":
		var p struct {
			ResponseID string `json:"response_id"`
			Text       string `json:"text"`
		}
		if json.Unmarshal([]byte(payloadJSON), &p) != nil {
			return
		}
		text := p.Text
		if bufID, buf, ok := b.Hub.FinishTurn(sessionID); ok && text == "" {
			text = buf
			if p.ResponseID == "" {
				p.ResponseID = bufID
			}
		}
		if text == "" {
			return
		}
		if p.ResponseID == "" {
			p.ResponseID = b.ensureResponseTurn(sessionID, "")
		}
		b.flushReasoningItem(ctx, sessionID, p.ResponseID)
		b.finishAssistantMessage(ctx, sessionID, p.ResponseID, text)
	case "permission_request":
		b.handlePermissionRequest(sessionID, payloadJSON)
	case "tool_call_update":
		b.handleToolCallUpdate(ctx, sessionID, payloadJSON)
	case "tool_call_result":
		b.handleToolCallResult(ctx, sessionID, payloadJSON)
	case "thinking_update":
		b.handleThinkingUpdate(ctx, sessionID, payloadJSON)
	}
}

var globalBridge *EventBridge

func SetEventBridge(b *EventBridge) { globalBridge = b }

func ForwardAcpSession(ctx context.Context, podKey, eventType, payload string) {
	if globalBridge != nil {
		globalBridge.HandleAcpSession(ctx, podKey, eventType, payload)
	}
}

func ForwardAgentStatus(ctx context.Context, podKey, agentStatus string) {
	if globalBridge == nil || globalBridge.Hub == nil {
		return
	}
	sessionID, ok := globalBridge.sessionForPod(ctx, podKey)
	if !ok {
		return
	}
	status := "idle"
	switch agentStatus {
	case podDomain.AgentStatusExecuting:
		status = "running"
	case podDomain.AgentStatusWaiting:
		status = "waiting"
	}
	globalBridge.Hub.Publish(sessionID, formatSSE("session.status", map[string]any{
		"conversation_id": sessionID, "status": status,
	}))
	if globalUpdatesHub != nil {
		globalUpdatesHub.NotifyChanged(sessionID)
	}
}
