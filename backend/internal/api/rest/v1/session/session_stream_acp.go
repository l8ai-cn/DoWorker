package sessionapi

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	domainitem "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
)

type toolCallState struct {
	itemID     string
	responseID string
}

type streamScratch struct {
	mu        sync.Mutex
	toolCalls map[string]toolCallState
	reasoning map[string]string
}

func (p *SessionStreamPublisher) scratch(sessionID string) *streamScratch {
	if p != nil && p.Hub != nil {
		return p.Hub.scratchFor(sessionID)
	}
	return &streamScratch{toolCalls: map[string]toolCallState{}, reasoning: map[string]string{}}
}

func (p *SessionStreamPublisher) handleToolCallUpdate(ctx context.Context, sessionID, payloadJSON string) {
	if p == nil || p.Items == nil || p.Hub == nil {
		return
	}
	var body struct {
		ToolCallID    string `json:"toolCallId"`
		ToolName      string `json:"toolName"`
		Status        string `json:"status"`
		ArgumentsJSON string `json:"argumentsJson"`
	}
	if json.Unmarshal([]byte(payloadJSON), &body) != nil || body.ToolCallID == "" {
		return
	}
	if body.Status != "running" && body.Status != "in_progress" && body.Status != "pending" {
		return
	}
	sc := p.scratch(sessionID)
	sc.mu.Lock()
	if _, ok := sc.toolCalls[body.ToolCallID]; ok {
		sc.mu.Unlock()
		return
	}
	sc.mu.Unlock()

	turnID := p.ensureTurn(sessionID, "")
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := p.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	args := body.ArgumentsJSON
	if args == "" {
		args = "{}"
	}
	item := map[string]any{
		"id": itemID, "type": "function_call", "response_id": turnID,
		"status": "in_progress", "name": body.ToolName, "arguments": args, "call_id": body.ToolCallID,
	}
	payload, _ := json.Marshal(item)
	_ = p.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "function_call",
		ResponseID: turnID, Status: "in_progress", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
	sc.mu.Lock()
	sc.toolCalls[body.ToolCallID] = toolCallState{itemID: itemID, responseID: turnID}
	sc.mu.Unlock()
	p.Hub.Publish(sessionID, formatSSE(sseTurnItemDone, map[string]any{"item": item}))
}

func (p *SessionStreamPublisher) handleToolCallResult(ctx context.Context, sessionID, payloadJSON string) {
	if p == nil || p.Items == nil || p.Hub == nil {
		return
	}
	var body struct {
		ToolCallID   string `json:"toolCallId"`
		Success      bool   `json:"success"`
		ResultText   string `json:"resultText"`
		ErrorMessage string `json:"errorMessage"`
	}
	if json.Unmarshal([]byte(payloadJSON), &body) != nil || body.ToolCallID == "" {
		return
	}
	sc := p.scratch(sessionID)
	sc.mu.Lock()
	state, ok := sc.toolCalls[body.ToolCallID]
	sc.mu.Unlock()
	turnID := p.ensureTurn(sessionID, "")
	if ok && state.responseID != "" {
		turnID = state.responseID
	}
	output := body.ResultText
	if output == "" && !body.Success {
		output = body.ErrorMessage
	}
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := p.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	item := map[string]any{
		"id": itemID, "type": "function_call_output", "response_id": turnID,
		"status": "completed", "call_id": body.ToolCallID, "output": output,
	}
	payload, _ := json.Marshal(item)
	_ = p.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "function_call_output",
		ResponseID: turnID, Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
	p.Hub.Publish(sessionID, formatSSE(sseTurnItemDone, map[string]any{"item": item}))
	p.touchSession(ctx, sessionID)
}

func (p *SessionStreamPublisher) handleThinkingUpdate(ctx context.Context, sessionID, payloadJSON string) {
	if p == nil || p.Hub == nil {
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(payloadJSON), &body) != nil || body.Text == "" {
		return
	}
	turnID := p.ensureTurn(sessionID, "")
	sc := p.scratch(sessionID)
	sc.mu.Lock()
	if sc.reasoning[turnID] == "" {
		p.Hub.Publish(sessionID, formatSSE(sseTurnReasoningStarted, map[string]any{"response_id": turnID}))
	}
	sc.reasoning[turnID] += body.Text
	sc.mu.Unlock()
	p.Hub.Publish(sessionID, formatSSE(sseTurnReasoningDelta, map[string]any{
		"delta": body.Text, "response_id": turnID,
	}))
}

func (p *SessionStreamPublisher) flushReasoningItem(ctx context.Context, sessionID, turnID string) {
	if p == nil || p.Items == nil || turnID == "" {
		return
	}
	sc := p.scratch(sessionID)
	sc.mu.Lock()
	text := sc.reasoning[turnID]
	delete(sc.reasoning, turnID)
	sc.mu.Unlock()
	if text == "" {
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
	item := map[string]any{
		"id": itemID, "type": "reasoning", "response_id": turnID,
		"status": "completed", "text": text,
	}
	payload, _ := json.Marshal(item)
	_ = p.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "reasoning",
		ResponseID: turnID, Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
}
