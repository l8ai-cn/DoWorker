package codex

import (
	"strings"
	"testing"
)

func TestHandlerSeparatesPreambleFromFinalMarkdown(t *testing.T) {
	f := newFixture()
	defer f.Close()

	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{
		Delta: "先检查工作区。",
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id":   "preamble",
			"type": "agentMessage",
			"text": "先检查工作区。",
		},
	})
	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{
			"id":      "shell-1",
			"type":    "commandExecution",
			"command": "pwd",
		},
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id":               "shell-1",
			"type":             "commandExecution",
			"status":           "completed",
			"exitCode":         0,
			"aggregatedOutput": "/workspace\n",
		},
	})
	writeNotification(f.PW, "item/agentMessage/delta", agentMessageDelta{
		Delta: "# Verification\n\nDone.",
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id":   "final",
			"type": "agentMessage",
			"text": "# Verification\n\nDone.",
		},
	})
	f.Drain()

	f.mu.Lock()
	defer f.mu.Unlock()

	var text strings.Builder
	for _, chunk := range f.Chunks {
		text.WriteString(chunk.Text)
	}
	if got := text.String(); !strings.Contains(got, "先检查工作区。\n\n# Verification") {
		t.Fatalf("assistant messages were not separated: %q", got)
	}
}
