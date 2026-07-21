package gemini

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
)

func init() {
	acp.RegisterTransport("gemini-acp", func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransport(cb, l)
	})
	agentkit.RegisterProcessNames("gemini")
}
