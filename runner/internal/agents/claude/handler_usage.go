package claude

import (
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

type assistantUsagePayload struct {
	Model string `json:"model"`
	Usage struct {
		InputTokens              int64 `json:"input_tokens"`
		OutputTokens             int64 `json:"output_tokens"`
		CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

func (t *transport) emitAssistantUsage(sessionID string, raw json.RawMessage) {
	if t.callbacks.OnUsage == nil || len(raw) == 0 {
		return
	}
	var payload assistantUsagePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	u := payload.Usage
	if u.InputTokens == 0 && u.OutputTokens == 0 {
		return
	}
	t.callbacks.OnUsage(sessionID, acp.TurnUsage{
		Model:               payload.Model,
		InputTokens:         u.InputTokens,
		OutputTokens:        u.OutputTokens,
		CacheReadTokens:     u.CacheReadInputTokens,
		CacheCreationTokens: u.CacheCreationInputTokens,
	})
}
