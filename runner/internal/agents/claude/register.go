package claude

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
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
