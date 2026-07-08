package codex

import (
	"strings"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

type agentMessageContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func agentMessageText(flatText string, content []agentMessageContentPart) string {
	if t := strings.TrimSpace(flatText); t != "" {
		return t
	}
	var b strings.Builder
	for _, part := range content {
		if part.Type != "" && part.Type != "text" {
			continue
		}
		if t := strings.TrimSpace(part.Text); t != "" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(t)
		}
	}
	return strings.TrimSpace(b.String())
}

func emitAssistantChunk(cb acp.EventCallbacks, sid, text string) {
	if text == "" || cb.OnContentChunk == nil {
		return
	}
	cb.OnContentChunk(sid, acp.ContentChunk{Text: text, Role: "assistant"})
}
