package sessionapi

import (
	"context"
	"encoding/json"
	"time"

	domainitem "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
)

func (p *SessionStreamPublisher) handlePermissionRequest(sessionID, payloadJSON string) {
	if p == nil || p.Elicitations == nil || p.Hub == nil {
		return
	}
	var body struct {
		RequestID   string `json:"requestId"`
		ToolName    string `json:"toolName"`
		Description string `json:"description"`
	}
	if json.Unmarshal([]byte(payloadJSON), &body) != nil || body.RequestID == "" {
		return
	}
	elicitID, err := NewElicitID()
	if err != nil {
		return
	}
	msg := body.Description
	if msg == "" {
		msg = body.ToolName
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
	p.Elicitations.Add(sessionID, &ElicitationRecord{
		ID: elicitID, RequestID: body.RequestID, Message: msg,
		Phase: "tool_call_approval", Status: "pending", RawEvent: raw,
	})
	p.Hub.Publish(sessionID, formatSSE(sseElicitationRequest, raw))
	if p.Updates != nil {
		p.Updates.NotifyChanged(sessionID)
	}
}

func (p *SessionStreamPublisher) finishAssistantMessage(ctx context.Context, sessionID, turnID, text string) {
	if p == nil || p.Items == nil || p.Hub == nil {
		return
	}
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := p.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"id": itemID, "type": "message", "response_id": turnID, "status": "completed",
		"role":    "assistant",
		"content": []map[string]any{{"type": "output_text", "text": text}},
	})
	_ = p.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "message",
		ResponseID: turnID, Status: "completed", Position: pos, Payload: payload,
		CreatedAt: time.Now(),
	})
	item := map[string]any{
		"id": itemID, "type": "message", "response_id": turnID, "status": "completed",
		"role":    "assistant",
		"content": []map[string]any{{"type": "output_text", "text": text}},
	}
	p.Hub.Publish(sessionID, formatSSE(sseTurnItemDone, map[string]any{"item": item}))
	p.Hub.Publish(sessionID, formatSSE(sseTurnCompleted, map[string]any{
		"id": turnID, "status": "completed", "model": "", "created_at": time.Now().Unix(),
		"completed_at": time.Now().Unix(),
	}))
	p.touchSession(ctx, sessionID)
}
