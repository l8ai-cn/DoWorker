package runner

import (
	"sync"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
)

type acpBackendBridge struct {
	mu      sync.Mutex
	usage   map[string]*acp.TurnUsage
	sawReal map[string]bool
}

func newAcpBackendBridge() *acpBackendBridge {
	return &acpBackendBridge{
		usage:   make(map[string]*acp.TurnUsage),
		sawReal: make(map[string]bool),
	}
}

func (b *acpBackendBridge) onUsage(podKey string, u acp.TurnUsage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := b.usage[podKey]
	if t == nil {
		t = &acp.TurnUsage{}
		b.usage[podKey] = t
	}
	b.sawReal[podKey] = true
	if u.Model != "" {
		t.Model = u.Model
	}
	t.InputTokens += u.InputTokens
	t.OutputTokens += u.OutputTokens
	t.CacheReadTokens += u.CacheReadTokens
	t.CacheCreationTokens += u.CacheCreationTokens
}

func (b *acpBackendBridge) usageForTurn(podKey, turnText string) acp.TurnUsage {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := b.usage[podKey]
	if t == nil {
		t = &acp.TurnUsage{}
		b.usage[podKey] = t
	}
	if !b.sawReal[podKey] {
		out := int64(len(turnText) / 4)
		if out < 1 {
			out = 1
		}
		t.InputTokens += out + 10
		t.OutputTokens += out
	}
	return *t
}

func (b *acpBackendBridge) onStateIdle(h *RunnerMessageHandler, conn client.ConnectionSender, podKey, turnText string) {
	h.reportTurnUsage(conn, podKey, b.usageForTurn(podKey, turnText))
}
