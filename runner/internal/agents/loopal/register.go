package loopal

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	acp.RegisterTransport("loopal-acp", func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransport(cb, l)
	})
	// Loopal has no on-disk session format yet — opt out of the cross-agent
	// fixture contract until persistence is implemented (see parser.go).
	tokenusage.RegisterParserOptOut([]string{"loopal"})
	agentkit.RegisterProcessNames("loopal")
}
