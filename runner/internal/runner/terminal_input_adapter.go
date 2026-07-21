package runner

import "github.com/l8ai-cn/agentcloud/runner/internal/agentkit"

// adaptTerminalInput delegates to the agentkit registry.
func adaptTerminalInput(data []byte, agentType string) []byte {
	return agentkit.AdaptTerminalInput(data, agentType)
}
