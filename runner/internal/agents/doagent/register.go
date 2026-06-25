package doagent

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	acp.RegisterAgent("do-agent", TransportType, func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return newTransport(cb, l)
	})

	tokenusage.RegisterParser([]string{"do-agent"}, &doagentParser{})
	agentkit.RegisterProcessNames("do-agent")

	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "DO_AGENT_HOME",
		UserDirName: ".agent",
	})
}
