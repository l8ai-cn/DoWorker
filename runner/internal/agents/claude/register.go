package claude

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

const TransportType = "claude-stream-json"

func init() {
	acp.RegisterTransport(TransportType, func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return newTransport(cb, l)
	})
	tokenusage.RegisterParser([]string{"claude", "claude-code"}, &claudeParser{})
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "CLAUDE_CONFIG_DIR",
		UserDirName: ".claude",
	})
	agentkit.RegisterProcessNames("claude")
}
