package codex

import (
	"encoding/json"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func (t *transport) getSessionID() string {
	t.sessionMu.RLock()
	defer t.sessionMu.RUnlock()
	return t.sessionID
}

func (t *transport) handleNotification(method string, params json.RawMessage) {
	sid := t.getSessionID()

	switch method {
	case "turn/started":
		t.resetTurnIdleGate()
		t.cancelIdleFallback()

	case "thread/status/changed":
		t.handleThreadStatusChanged(params)

	case "turn/completed":
		t.handleTurnCompleted(params)

	case "item/agentMessage/delta":
		var d agentMessageDelta
		if err := json.Unmarshal(params, &d); err != nil {
			t.logger.Warn("failed to parse agentMessage/delta", "error", err)
			return
		}
		t.cancelIdleFallback()
		t.markAgentMessageStreamed(d.ItemID)
		if t.callbacks.OnContentChunk != nil {
			t.callbacks.OnContentChunk(sid, acp.ContentChunk{Text: d.Delta, Role: "assistant"})
		}

	case "item/reasoning/summaryTextDelta", "item/reasoning/textDelta":
		var d reasoningDelta
		if err := json.Unmarshal(params, &d); err != nil {
			t.logger.Warn("failed to parse reasoning delta", "error", err)
			return
		}
		t.cancelIdleFallback()
		if t.callbacks.OnThinkingUpdate != nil {
			t.callbacks.OnThinkingUpdate(sid, acp.ThinkingUpdate{Text: d.Delta})
		}

	case "item/plan/delta":
		var d planDelta
		if err := json.Unmarshal(params, &d); err != nil {
			t.logger.Warn("failed to parse plan/delta", "error", err)
			return
		}
		if t.callbacks.OnContentChunk != nil {
			t.callbacks.OnContentChunk(sid, acp.ContentChunk{Text: d.Delta, Role: "plan"})
		}

	case "item/started":
		t.handleItemStarted(sid, params)

	case "item/completed":
		t.handleItemCompleted(sid, params)

	case "rawResponseItem/completed":
		t.handleRawResponseItemCompleted(sid, params)

	case "configWarning", "warning":
		t.handleConfigWarning(params)

	case "error":
		t.handleErrorNotification(params)

	default:
		t.logger.Debug("unhandled codex notification", "method", method)
	}
}

func (t *transport) handleItemStarted(sid string, params json.RawMessage) {
	var is itemStartedParams
	if err := json.Unmarshal(params, &is); err != nil {
		t.logger.Warn("failed to parse item/started", "error", err)
		return
	}
	t.cancelIdleFallback()
	switch is.Item.Type {
	case "toolCall":
		if t.callbacks.OnToolCallUpdate != nil {
			t.callbacks.OnToolCallUpdate(sid, acp.ToolCallUpdate{
				ToolCallID: is.Item.ID, ToolName: is.Item.ToolName, Status: "running",
			})
		}
	case "commandExecution":
		if t.callbacks.OnToolCallUpdate != nil {
			t.callbacks.OnToolCallUpdate(sid, acp.ToolCallUpdate{
				ToolCallID: is.Item.ID, ToolName: "shell", Status: "running",
			})
		}
	case "fileChange":
		if t.callbacks.OnToolCallUpdate != nil {
			t.callbacks.OnToolCallUpdate(sid, acp.ToolCallUpdate{
				ToolCallID: is.Item.ID, ToolName: "fileChange", Status: "running",
			})
		}
	}
}
