package codex

import (
	"encoding/json"
	"strings"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func (t *transport) handleErrorNotification(params json.RawMessage) {
	var ep errorNotificationParams
	if err := json.Unmarshal(params, &ep); err != nil {
		t.logger.Warn("failed to parse error notification", "error", err)
		return
	}
	msg := ep.Message
	if msg == "" && ep.Error != nil {
		msg = ep.Error.Message
	}
	if msg == "" {
		msg = "codex error"
	}
	if t.callbacks.OnLog != nil {
		t.callbacks.OnLog("error", msg)
	}
}

// handleThreadStatusChanged is the authoritative turn lifecycle for Codex
// app-server builds (cliVersion 0.14x): status.type "active" means the turn is
// running, "idle" means it ended. Seeing it disables the fragile agentMessage
// debounce, which otherwise ends the turn during long silent generation gaps.
func (t *transport) handleThreadStatusChanged(params json.RawMessage) {
	var p threadStatusChangedParams
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}
	t.markLifecycleSignal()
	switch p.Status.Type {
	case "idle":
		t.cancelIdleFallback()
		t.clearToolOutputs()
		t.notifyTurnIdle()
	case "active":
		t.cancelIdleFallback()
		if t.callbacks.OnStateChange != nil {
			t.callbacks.OnStateChange(acp.StateProcessing)
		}
	}
}

func (t *transport) handleTurnCompleted(params json.RawMessage) {
	t.markLifecycleSignal()
	t.cancelIdleFallback()
	defer t.clearToolOutputs()
	var tc turnCompletedParams
	if err := json.Unmarshal(params, &tc); err != nil {
		t.clearActiveTurn("")
		if t.callbacks.OnStateChange != nil {
			t.callbacks.OnStateChange(acp.StateIdle)
		}
		return
	}
	t.clearActiveTurn(tc.Turn.ID)
	if tc.Turn.Status == "failed" && t.callbacks.OnLog != nil {
		msg := "turn failed"
		if tc.Turn.Error != nil {
			msg = "turn failed: " + tc.Turn.Error.Message
		}
		t.callbacks.OnLog("error", msg)
	}
	if u := tc.Turn.Usage; u != nil && t.callbacks.OnUsage != nil &&
		(u.InputTokens > 0 || u.OutputTokens > 0) {
		t.callbacks.OnUsage(t.getSessionID(), acp.TurnUsage{
			InputTokens:     u.InputTokens,
			OutputTokens:    u.OutputTokens,
			CacheReadTokens: u.CachedInputTokens,
		})
	}
	if t.callbacks.OnStateChange != nil {
		t.callbacks.OnStateChange(acp.StateIdle)
	}
}

func (t *transport) handleItemCompleted(sid string, params json.RawMessage) {
	var ic itemCompletedParams
	if err := json.Unmarshal(params, &ic); err != nil {
		t.logger.Warn("failed to parse item/completed", "error", err)
		return
	}
	switch ic.Item.Type {
	case "agentMessage":
		text := agentMessageText(ic.Item.Text, ic.Item.Content)
		if !t.agentMessageAlreadyEmitted(ic.Item.ID) {
			emitAssistantChunk(t.callbacks, sid, t.applyAgentMessageBoundary(text))
		}
		t.markAgentMessageCompleted()
		// A completed agentMessage may be only a preamble before the agent runs
		// tools, so end the turn on a debounce that later turn activity cancels —
		// authoritative end comes from turn/completed.
		t.scheduleIdleAfterMessage()
	case "error":
		msg := strings.TrimSpace(ic.Item.Message)
		if msg == "" {
			msg = strings.TrimSpace(ic.Item.Text)
		}
		if msg != "" && t.callbacks.OnLog != nil {
			t.callbacks.OnLog("error", msg)
		}
	case "toolCall":
		if t.callbacks.OnToolCallUpdate != nil {
			t.callbacks.OnToolCallUpdate(sid, acp.ToolCallUpdate{
				ToolCallID: ic.Item.ID, ToolName: ic.Item.ToolName, Status: "completed",
			})
		}
	case "commandExecution":
		exitCode := 0
		if ic.Item.ExitCode != nil {
			exitCode = *ic.Item.ExitCode
		}
		if t.callbacks.OnToolCallResult != nil {
			output := t.commandOutput(ic.Item)
			t.callbacks.OnToolCallResult(sid, acp.ToolCallResult{
				ToolCallID: ic.Item.ID, ToolName: "shell",
				Success: exitCode == 0, ResultText: output,
			})
		}
	case "fileChange":
		success := ic.Item.Status == "" || ic.Item.Status == "completed"
		if t.callbacks.OnToolCallResult != nil {
			t.callbacks.OnToolCallResult(sid, acp.ToolCallResult{
				ToolCallID: ic.Item.ID, ToolName: "fileChange",
				Success: success, ResultText: t.fileChangeOutput(ic.Item),
			})
		}
	}
}
