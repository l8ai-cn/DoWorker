package opencode

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	acp.RegisterTransport("opencode-acp", func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransport(cb, l)
	})
	tokenusage.RegisterParser([]string{"opencode"}, &opencodeParser{})
	agentkit.RegisterProcessNames("opencode")
}
