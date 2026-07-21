package grok

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

const TransportType = "grok-build-acp"

func init() {
	acp.RegisterTransport(TransportType, func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransportWithHandshakeHook(cb, l, authenticate)
	})

	tokenusage.RegisterParserOptOut([]string{"grok", "grok-build"})
	agentkit.RegisterProcessNames("grok")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "GROK_HOME",
		UserDirName: ".grok",
	})
}
