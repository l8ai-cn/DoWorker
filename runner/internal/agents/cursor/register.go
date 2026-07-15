package cursor

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	acp.RegisterTransport("cursor-acp", func(cb acp.EventCallbacks, log *slog.Logger) acp.Transport {
		return acp.NewACPTransport(cb, log)
	})
	tokenusage.RegisterParserOptOut([]string{"cursor-cli", "agent"})
	agentkit.RegisterProcessNames("agent")
}
