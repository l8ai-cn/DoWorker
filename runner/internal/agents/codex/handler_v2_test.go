package codex

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func TestHandler_ItemCompleted_AgentMessageContentArray(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id":   "msg-v2",
			"type": "agentMessage",
			"content": []map[string]any{
				{"type": "text", "text": "Hello from v2"},
			},
		},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(f.Chunks))
	}
	if f.Chunks[0].Text != "Hello from v2" || f.Chunks[0].Role != "assistant" {
		t.Errorf("chunk = %+v", f.Chunks[0])
	}
}

func TestHandler_ConfigWarning(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "configWarning", map[string]any{
		"summary": "missing API key",
		"details": "set OPENAI_API_KEY",
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.LogMessages) != 1 || !strings.Contains(f.LogMessages[0], "missing API key") {
		t.Errorf("logs = %v", f.LogMessages)
	}
}

func TestHandler_RawResponseItemCompleted(t *testing.T) {
	f := newFixture()
	defer f.Close()
	writeNotification(f.PW, "rawResponseItem/completed", map[string]any{
		"item": map[string]any{
			"type": "agentMessage",
			"content": []map[string]any{
				{"type": "text", "text": "raw item text"},
			},
		},
	})
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Chunks) != 1 || f.Chunks[0].Text != "raw item text" {
		t.Errorf("chunks = %+v", f.Chunks)
	}
}

func TestHandler_PermissionsApprovalRequest(t *testing.T) {
	f := newFixture()
	defer f.Close()
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      200,
		"method":  "item/permissions/requestApproval",
		"params":  map[string]any{"reason": "network access"},
	}
	data, _ := json.Marshal(msg)
	f.PW.Write(append(data, '\n'))
	f.Drain()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.PermissionReqs) != 1 {
		t.Fatalf("expected 1 permission req, got %d", len(f.PermissionReqs))
	}
	if f.PermissionReqs[0].ToolName != "permissions" {
		t.Errorf("tool = %q", f.PermissionReqs[0].ToolName)
	}
	foundWaiting := false
	for _, s := range f.StateChanges {
		if s == acp.StateWaitingPermission {
			foundWaiting = true
		}
	}
	if !foundWaiting {
		t.Errorf("states = %v", f.StateChanges)
	}
}

func TestAgentMessageText_ContentArray(t *testing.T) {
	got := agentMessageText("", []agentMessageContentPart{
		{Type: "text", Text: "a"},
		{Type: "text", Text: "b"},
	})
	if got != "a\nb" {
		t.Fatalf("got %q", got)
	}
}
