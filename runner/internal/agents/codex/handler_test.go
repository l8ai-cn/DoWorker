package codex

import (
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func TestHandler_AgentMessageDelta(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "Hello "})
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "world!"})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(f.Chunks))
	}
	if f.Chunks[0].Text != "Hello " || f.Chunks[0].Role != "assistant" {
		t.Errorf("chunk[0] = %+v", f.Chunks[0])
	}
}

func TestHandler_ReasoningDelta(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/reasoning/summaryTextDelta", reasoningDelta{Delta: "hmm"})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ThinkingTexts) != 1 || f.ThinkingTexts[0] != "hmm" {
		t.Errorf("thinking = %v", f.ThinkingTexts)
	}
}

func TestHandler_ReasoningTextDelta(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/reasoning/textDelta", reasoningDelta{Delta: "deep thought"})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ThinkingTexts) != 1 || f.ThinkingTexts[0] != "deep thought" {
		t.Errorf("thinking = %v", f.ThinkingTexts)
	}
}

func TestHandler_PlanDelta(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/plan/delta", planDelta{Delta: "Step 1: Read files"})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 1 {
		t.Fatalf("expected 1 plan chunk, got %d", len(f.Chunks))
	}
	if f.Chunks[0].Text != "Step 1: Read files" || f.Chunks[0].Role != "plan" {
		t.Errorf("chunk = %+v", f.Chunks[0])
	}
}

func TestHandler_ItemStarted_ToolCall(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{"id": "tc1", "type": "toolCall", "toolName": "Read"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolUpdates) != 1 {
		t.Fatalf("expected 1 tool update, got %d", len(f.ToolUpdates))
	}
	if f.ToolUpdates[0].Status != "running" || f.ToolUpdates[0].ToolName != "Read" {
		t.Errorf("update = %+v", f.ToolUpdates[0])
	}
}

func TestHandler_ItemStarted_CommandExecution(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{"id": "ce1", "type": "commandExecution"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolUpdates) != 1 || f.ToolUpdates[0].ToolName != "shell" {
		t.Errorf("updates = %+v", f.ToolUpdates)
	}
}

func TestHandler_ItemCompleted_CommandExecution(t *testing.T) {
	f := newFixture()
	defer f.Close()
	exitCode := 0
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id": "ce1", "type": "commandExecution",
			"exitCode": exitCode, "aggregatedOutput": "file list",
		},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolResults) != 1 {
		t.Fatalf("expected 1 tool result, got %d", len(f.ToolResults))
	}
	if !f.ToolResults[0].Success || f.ToolResults[0].ResultText != "file list" {
		t.Errorf("result = %+v", f.ToolResults[0])
	}
}

func TestHandler_ItemCompleted_CommandExecution_Failure(t *testing.T) {
	f := newFixture()
	defer f.Close()
	exitCode := 1
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id": "ce2", "type": "commandExecution",
			"exitCode": exitCode, "aggregatedOutput": "error",
		},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolResults) != 1 || f.ToolResults[0].Success {
		t.Errorf("expected failure, got %+v", f.ToolResults)
	}
}

func TestHandler_ItemCompleted_ToolCall(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "tc1", "type": "toolCall", "toolName": "Write"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolUpdates) != 1 || f.ToolUpdates[0].Status != "completed" {
		t.Errorf("updates = %+v", f.ToolUpdates)
	}
}

func TestHandler_ItemCompleted_AgentMessage(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "x", "type": "agentMessage", "text": "你好，我在线。"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(f.Chunks))
	}
	if f.Chunks[0].Text != "你好，我在线。" || f.Chunks[0].Role != "assistant" {
		t.Errorf("chunk = %+v", f.Chunks[0])
	}
	if len(f.StateChanges) != 1 || f.StateChanges[0] != acp.StateIdle {
		t.Errorf("states = %v", f.StateChanges)
	}
}

func TestHandler_AgentMessageDeltaThenCompleted_NoDuplicateAndIdle(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{ItemID: "msg-1", Delta: "Hello "})
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{ItemID: "msg-1", Delta: "world"})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "msg-1", "type": "agentMessage", "text": "Hello world"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 2 {
		t.Fatalf("expected 2 delta chunks, got %d: %+v", len(f.Chunks), f.Chunks)
	}
	if f.Chunks[0].Text != "Hello " || f.Chunks[1].Text != "world" {
		t.Errorf("chunks = %+v", f.Chunks)
	}
	if len(f.StateChanges) != 1 || f.StateChanges[0] != acp.StateIdle {
		t.Errorf("states = %v", f.StateChanges)
	}
}

func TestHandler_AgentMessageDeltaNoItemIDThenCompleted_NoDuplicate(t *testing.T) {
	// Codex builds that stream deltas WITHOUT an itemId but complete with an id:
	// the id-keyed guard misses, so without the per-turn fallback the full text
	// gets re-emitted, doubling every assistant message.
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "你好，"})
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "我在线。"})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "codex-msg-7", "type": "agentMessage", "text": "你好，我在线。"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 2 {
		t.Fatalf("expected 2 delta chunks (no re-emit), got %d: %+v", len(f.Chunks), f.Chunks)
	}
	if f.Chunks[0].Text != "你好，" || f.Chunks[1].Text != "我在线。" {
		t.Errorf("chunks = %+v", f.Chunks)
	}
}

func TestHandler_TwoMessagesStreamedNoItemID_NoDuplicate(t *testing.T) {
	// Two agentMessages in one turn, both streamed without itemId. Each completed
	// must suppress only its own stream, not leak suppression across messages.
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "第一句。"})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "m1", "type": "agentMessage", "text": "第一句。"},
	})
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "第二句。"})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "m2", "type": "agentMessage", "text": "第二句。"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 2 {
		t.Fatalf("expected 2 chunks (one per message), got %d: %+v", len(f.Chunks), f.Chunks)
	}
	if f.Chunks[0].Text != "第一句。" || f.Chunks[1].Text != "\n\n第二句。" {
		t.Errorf("chunks = %+v", f.Chunks)
	}
}

func TestHandler_AgentMessageCompletedOnly_StillEmits(t *testing.T) {
	// No deltas at all: item/completed must still emit the full text once.
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "only-1", "type": "agentMessage", "text": "只有完成事件。"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 1 || f.Chunks[0].Text != "只有完成事件。" {
		t.Fatalf("expected 1 emitted chunk, got %+v", f.Chunks)
	}
}

func TestHandler_PreambleThenToolsThenFinal_SingleIdle(t *testing.T) {
	// The premature-idle bug: a preamble agentMessage before tool calls used to
	// end the turn immediately, so the agent's real work (file writes) never
	// reached the session. The turn must stay active until turn/completed.
	f := newFixture()
	defer f.Close()
	// Preamble message (Codex streams without itemId here).
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{Delta: "我先看下目录。"})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "pre-1", "type": "agentMessage", "text": "我先看下目录。"},
	})
	// Tool activity begins well within the (short, test-only) fallback window.
	time.Sleep(10 * time.Millisecond)
	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{"id": "fc1", "type": "fileChange", "filePath": "index.html"},
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "fc1", "type": "fileChange", "status": "completed", "filePath": "index.html"},
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "final-1", "type": "agentMessage", "text": "已创建 index.html。"},
	})
	writeNotification(f.PW, "turn/completed", map[string]any{})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.StateChanges) != 1 || f.StateChanges[0] != acp.StateIdle {
		t.Fatalf("expected exactly 1 idle (from turn/completed), got %v", f.StateChanges)
	}
	if len(f.ToolResults) != 1 || !f.ToolResults[0].Success {
		t.Errorf("expected fileChange tool result, got %+v", f.ToolResults)
	}
}

func TestHandler_ThreadStatusDrivesIdle_NoPrematureDebounce(t *testing.T) {
	// New Codex app-server build: thread/status/changed is authoritative. A
	// preamble message during a long silent generation gap must NOT end the turn
	// via the debounce; only thread/status idle ends it.
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "thread/status/changed", map[string]any{
		"status": map[string]any{"type": "active"},
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{"id": "pre", "type": "agentMessage", "text": "我先看下目录。"},
	})
	// Wait well past the (short, test-only) debounce window; no idle must fire.
	time.Sleep(80 * time.Millisecond)
	f.mu.Lock()
	if len(f.StateChanges) != 0 {
		f.mu.Unlock()
		t.Fatalf("debounce fired despite thread lifecycle signal: %v", f.StateChanges)
	}
	f.mu.Unlock()
	writeNotification(f.PW, "thread/status/changed", map[string]any{
		"status": map[string]any{"type": "idle"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.StateChanges) != 1 || f.StateChanges[0] != acp.StateIdle {
		t.Fatalf("expected exactly 1 idle from thread/status, got %v", f.StateChanges)
	}
}

func TestHandler_ItemStarted_CommandAsString(t *testing.T) {
	// Codex builds send item.command as a plain string on commandExecution; the
	// item/started must not be dropped on a type mismatch.
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{"id": "ce9", "type": "commandExecution", "command": "ls -la"},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolUpdates) != 1 || f.ToolUpdates[0].ToolName != "shell" {
		t.Errorf("expected shell tool update, got %+v", f.ToolUpdates)
	}
}

func TestHandler_TurnCompleted(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "turn/completed", map[string]any{})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.StateChanges) != 1 || f.StateChanges[0] != acp.StateIdle {
		t.Errorf("states = %v", f.StateChanges)
	}
}

func TestHandler_ErrorNotification(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "error", map[string]any{"message": "HTTP error: 401 Unauthorized"})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.LogMessages) != 1 {
		t.Fatalf("expected 1 log message, got %d", len(f.LogMessages))
	}
	if f.LogMessages[0] != "error:HTTP error: 401 Unauthorized" {
		t.Errorf("log = %q", f.LogMessages[0])
	}
}

func TestHandler_TurnCompleted_Failed(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "turn/completed", map[string]any{
		"turn": map[string]any{
			"status": "failed",
			"error":  map[string]any{"message": "API error"},
		},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.StateChanges) != 1 || f.StateChanges[0] != acp.StateIdle {
		t.Errorf("states = %v", f.StateChanges)
	}
}

func TestHandler_UnknownNotification(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "some/future/method", map[string]any{})
	f.Drain()
}

func TestHandler_InvalidParams(t *testing.T) {
	f := newFixture()
	defer f.Close()
	for _, method := range []string{
		"item/agentMessage/delta",
		"item/reasoning/summaryTextDelta",
		"item/plan/delta",
		"item/started",
		"item/completed",
	} {
		writeNotification(f.PW, method, "invalid")
	}
	f.Drain()
}
