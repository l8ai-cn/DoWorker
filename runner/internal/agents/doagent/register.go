package doagent

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

func init() {
	acp.RegisterTransport(TransportType, func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return newTransport(cb, l)
	})

	tokenusage.RegisterParser([]string{"do-agent"}, &doagentParser{})
	agentkit.RegisterProcessNames("do-agent")

	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "DO_AGENT_HOME",
		UserDirName: ".agent",
	})
}
