package cursor

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

func init() {
	acp.RegisterTransport("cursor-acp", func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransport(cb, l)
	})
	tokenusage.RegisterParserOptOut([]string{"cursor-cli", "agent"})
	agentkit.RegisterProcessNames("agent")
}
