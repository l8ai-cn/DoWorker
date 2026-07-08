package sessionapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

const sseResponseError = "response.error"

func (p *SessionStreamPublisher) handleAcpLog(ctx context.Context, sessionID, payloadJSON string) {
	if p == nil || p.Hub == nil {
		return
	}
	var body struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}
	if json.Unmarshal([]byte(payloadJSON), &body) != nil {
		return
	}
	if body.Level != "error" && body.Level != "warn" {
		return
	}
	msg := strings.TrimSpace(body.Message)
	if msg == "" {
		return
	}
	code := "agent_log"
	if body.Level == "error" {
		code = "turn_error"
	}
	p.publishTurnError(ctx, sessionID, code, msg)
}

func (p *SessionStreamPublisher) publishTurnError(ctx context.Context, sessionID, code, message string) {
	if p == nil || p.Items == nil || p.Hub == nil {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	if code == "" {
		code = "agent_failed"
	}
	turnID := p.ensureTurn(sessionID, "")
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := p.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	item := map[string]any{
		"id": itemID, "type": "error", "response_id": turnID,
		"code": code, "message": message, "source": "agent",
	}
	payload, _ := json.Marshal(item)
	_ = p.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "error",
		ResponseID: turnID, Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
	p.Hub.Publish(sessionID, formatSSE(sseTurnItemDone, map[string]any{"item": item}))
	p.Hub.Publish(sessionID, formatSSE(sseResponseError, map[string]any{
		"source":      "agent",
		"response_id": turnID,
		"error":       map[string]any{"code": code, "message": message},
	}))
	p.touchSession(ctx, sessionID)
}
