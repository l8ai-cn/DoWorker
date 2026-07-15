package gemini

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
)

func init() {
	acp.RegisterTransport("gemini-acp", func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransport(cb, l)
	})
	agentkit.RegisterProcessNames("gemini")
}
