package runner

import (
	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
)

// defaultUsageModel prices estimated (non-agent-reported) usage. Agents that
// report real usage carry their own model name in TurnUsage.
const defaultUsageModel = "gpt-4o-mini"

func (h *RunnerMessageHandler) reportTurnUsage(conn client.ConnectionSender, podKey string, usage acp.TurnUsage) {
	if conn == nil || podKey == "" {
		return
	}
	if usage.InputTokens == 0 && usage.OutputTokens == 0 {
		return
	}
	model := usage.Model
	if model == "" {
		model = defaultUsageModel
	}
	_ = conn.SendPodUsageEvent(podKey, model,
		usage.InputTokens, usage.OutputTokens,
		usage.CacheReadTokens, usage.CacheCreationTokens)
}
