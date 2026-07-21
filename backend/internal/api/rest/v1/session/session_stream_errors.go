package sessionapi

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	domainitem "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
)

const sseResponseRetry = "response.retry"

var reconnectLogPattern = regexp.MustCompile(`(?i)^reconnecting\.\.\.\s+([1-9][0-9]*)/([1-9][0-9]*)$`)

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
	if attempt, maxAttempts, ok := parseReconnectLog(msg); ok {
		p.publishRetry(sessionID, msg, attempt, maxAttempts)
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
	p.touchSession(ctx, sessionID)
}

func parseReconnectLog(message string) (int, int, bool) {
	matches := reconnectLogPattern.FindStringSubmatch(message)
	if len(matches) != 3 {
		return 0, 0, false
	}
	attempt, attemptErr := strconv.Atoi(matches[1])
	maxAttempts, maxErr := strconv.Atoi(matches[2])
	if attemptErr != nil || maxErr != nil || attempt > maxAttempts {
		return 0, 0, false
	}
	return attempt, maxAttempts, true
}

func (p *SessionStreamPublisher) publishRetry(
	sessionID string,
	message string,
	attempt int,
	maxAttempts int,
) {
	p.Hub.Publish(sessionID, formatSSE(sseResponseRetry, map[string]any{
		"source":        "agent",
		"attempt":       attempt,
		"max_attempts":  maxAttempts,
		"delay_seconds": 0,
		"error":         map[string]any{"code": "agent_reconnecting", "message": message},
	}))
}
