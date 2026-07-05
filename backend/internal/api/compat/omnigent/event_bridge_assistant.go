package omnigent

import (
	"context"
	"encoding/json"
	"time"

	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

func (b *EventBridge) handlePermissionRequest(sessionID, payloadJSON string) {
	if b.Elicitations == nil {
		return
	}
	var p struct {
		RequestID   string `json:"requestId"`
		ToolName    string `json:"toolName"`
		Description string `json:"description"`
	}
	if json.Unmarshal([]byte(payloadJSON), &p) != nil || p.RequestID == "" {
		return
	}
	elicitID, err := NewElicitID()
	if err != nil {
		return
	}
	msg := p.Description
	if msg == "" {
		msg = p.ToolName
	}
	if msg == "" {
		msg = "Approve tool execution?"
	}
	raw := map[string]any{
		"elicitation_id": elicitID,
		"params": map[string]any{
			"mode": "form", "message": msg, "phase": "tool_call_approval",
			"policy_name": "tool_call_approval", "content_preview": "",
		},
	}
	b.Elicitations.Add(sessionID, &ElicitationRecord{
		ID: elicitID, RequestID: p.RequestID, Message: msg,
		Phase: "tool_call_approval", Status: "pending", RawEvent: raw,
	})
	b.Hub.Publish(sessionID, formatSSE("response.elicitation_request", raw))
	if globalUpdatesHub != nil {
		globalUpdatesHub.NotifyChanged(sessionID)
	}
}

func (b *EventBridge) finishAssistantMessage(ctx context.Context, sessionID, responseID, text string) {
	if b.Items == nil {
		return
	}
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := b.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"id": itemID, "type": "message", "response_id": responseID, "status": "completed",
		"role": "assistant",
		"content": []map[string]any{{"type": "output_text", "text": text}},
	})
	_ = b.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "message",
		ResponseID: responseID, Status: "completed", Position: pos, Payload: payload,
		CreatedAt: time.Now(),
	})
	b.Hub.Publish(sessionID, formatSSE("response.output_item.done", map[string]any{
		"item": map[string]any{
			"id": itemID, "type": "message", "response_id": responseID, "status": "completed",
			"role": "assistant",
			"content": []map[string]any{{"type": "output_text", "text": text}},
		},
	}))
	b.Hub.Publish(sessionID, formatSSE("response.completed", map[string]any{
		"id": responseID, "status": "completed", "model": "", "created_at": time.Now().Unix(),
		"completed_at": time.Now().Unix(),
	}))
	if b.Sessions != nil {
		_ = b.Sessions.TouchUpdatedAt(ctx, sessionID)
	}
	if globalUpdatesHub != nil {
		globalUpdatesHub.NotifyChanged(sessionID)
	}
}
