package grok

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

const TransportType = "grok-build"

func init() {
	acp.RegisterAgent("grok", TransportType, func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransportWithHandshakeHook(cb, l, authenticate)
	})

	tokenusage.RegisterParserOptOut([]string{"grok", "grok-build"})
	agentkit.RegisterProcessNames("grok")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "GROK_HOME",
		UserDirName: ".grok",
	})
}
