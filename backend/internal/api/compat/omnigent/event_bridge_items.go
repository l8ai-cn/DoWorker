package omnigent

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

type toolCallState struct {
	itemID     string
	responseID string
}

type eventBridgeScratch struct {
	mu        sync.Mutex
	toolCalls map[string]toolCallState
	reasoning map[string]string
}

func (b *EventBridge) scratch(sessionID string) *eventBridgeScratch {
	if b != nil && b.Hub != nil {
		return b.Hub.scratchFor(sessionID)
	}
	return &eventBridgeScratch{toolCalls: map[string]toolCallState{}, reasoning: map[string]string{}}
}

func (b *EventBridge) handleToolCallUpdate(ctx context.Context, sessionID, payloadJSON string) {
	if b == nil || b.Items == nil {
		return
	}
	var p struct {
		ToolCallID    string `json:"toolCallId"`
		ToolName      string `json:"toolName"`
		Status        string `json:"status"`
		ArgumentsJSON string `json:"argumentsJson"`
	}
	if json.Unmarshal([]byte(payloadJSON), &p) != nil || p.ToolCallID == "" {
		return
	}
	if p.Status != "running" && p.Status != "in_progress" && p.Status != "pending" {
		return
	}
	sc := b.scratch(sessionID)
	sc.mu.Lock()
	if _, ok := sc.toolCalls[p.ToolCallID]; ok {
		sc.mu.Unlock()
		return
	}
	sc.mu.Unlock()

	respID := b.ensureResponseTurn(sessionID, "")
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := b.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	args := p.ArgumentsJSON
	if args == "" {
		args = "{}"
	}
	item := map[string]any{
		"id": itemID, "type": "function_call", "response_id": respID,
		"status": "in_progress", "name": p.ToolName, "arguments": args, "call_id": p.ToolCallID,
	}
	payload, _ := json.Marshal(item)
	_ = b.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "function_call",
		ResponseID: respID, Status: "in_progress", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
	sc.mu.Lock()
	sc.toolCalls[p.ToolCallID] = toolCallState{itemID: itemID, responseID: respID}
	sc.mu.Unlock()
	b.Hub.Publish(sessionID, formatSSE("response.output_item.done", map[string]any{"item": item}))
}

func (b *EventBridge) handleToolCallResult(ctx context.Context, sessionID, payloadJSON string) {
	if b == nil || b.Items == nil {
		return
	}
	var p struct {
		ToolCallID   string `json:"toolCallId"`
		ToolName     string `json:"toolName"`
		Success      bool   `json:"success"`
		ResultText   string `json:"resultText"`
		ErrorMessage string `json:"errorMessage"`
	}
	if json.Unmarshal([]byte(payloadJSON), &p) != nil || p.ToolCallID == "" {
		return
	}
	sc := b.scratch(sessionID)
	sc.mu.Lock()
	state, ok := sc.toolCalls[p.ToolCallID]
	sc.mu.Unlock()
	respID := b.ensureResponseTurn(sessionID, "")
	if ok && state.responseID != "" {
		respID = state.responseID
	}
	output := p.ResultText
	if output == "" && !p.Success {
		output = p.ErrorMessage
	}
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		return
	}
	pos, err := b.Items.NextPosition(ctx, sessionID)
	if err != nil {
		return
	}
	item := map[string]any{
		"id": itemID, "type": "function_call_output", "response_id": respID,
		"status": "completed", "call_id": p.ToolCallID, "output": output,
	}
	payload, _ := json.Marshal(item)
	_ = b.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "function_call_output",
		ResponseID: respID, Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
	b.Hub.Publish(sessionID, formatSSE("response.output_item.done", map[string]any{"item": item}))
	if b.Sessions != nil {
		_ = b.Sessions.TouchUpdatedAt(ctx, sessionID)
	}
}

func (b *EventBridge) handleThinkingUpdate(ctx context.Context, sessionID, payloadJSON string) {
	if b == nil || b.Hub == nil {
		return
	}
	var p struct {
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(payloadJSON), &p) != nil || p.Text == "" {
		return
	}
	respID := b.ensureResponseTurn(sessionID, "")
	sc := b.scratch(sessionID)
	sc.mu.Lock()
	if sc.reasoning[respID] == "" {
		b.Hub.Publish(sessionID, formatSSE("response.reasoning.started", map[string]any{"response_id": respID}))
	}
	sc.reasoning[respID] += p.Text
	sc.mu.Unlock()
	b.Hub.Publish(sessionID, formatSSE("response.reasoning_text.delta", map[string]any{
		"delta": p.Text, "response_id": respID,
	}))
}

func (b *EventBridge) flushReasoningItem(ctx context.Context, sessionID, responseID string) {
	if b == nil || b.Items == nil || responseID == "" {
		return
	}
	sc := b.scratch(sessionID)
	sc.mu.Lock()
	text := sc.reasoning[responseID]
	delete(sc.reasoning, responseID)
	sc.mu.Unlock()
	if text == "" {
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
	item := map[string]any{
		"id": itemID, "type": "reasoning", "response_id": responseID,
		"status": "completed", "text": text,
	}
	payload, _ := json.Marshal(item)
	_ = b.Items.Append(ctx, &domainitem.Item{
		ID: itemID, SessionID: sessionID, ItemType: "reasoning",
		ResponseID: responseID, Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
	})
}
