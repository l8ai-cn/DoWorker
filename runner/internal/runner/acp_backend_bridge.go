package runner

import (
	"encoding/json"
	"sync"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/client"
)

type acpBackendBridge struct {
	mu      sync.Mutex
	buffers map[string]string
	states  map[string]string
	usage   map[string]*acp.TurnUsage
	sawReal map[string]bool
}

func newAcpBackendBridge() *acpBackendBridge {
	return &acpBackendBridge{
		buffers: make(map[string]string),
		states:  make(map[string]string),
		usage:   make(map[string]*acp.TurnUsage),
		sawReal: make(map[string]bool),
	}
}

// onUsage folds agent-reported per-turn usage into the pod's cumulative
// totals. PodUsageEvent has SET semantics on the backend, so reports must
// always carry cumulative values.
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

// usageForTurn returns cumulative usage for the pod. When the agent never
// reported real usage, the turn's text length feeds a rough estimate so
// PTY-adjacent mock agents still produce non-zero cost.
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

func (b *acpBackendBridge) onChunk(conn client.ConnectionSender, podKey string, chunk acp.ContentChunk) {
	if chunk.Role != "assistant" || chunk.Text == "" {
		return
	}
	b.mu.Lock()
	b.buffers[podKey] += chunk.Text
	b.mu.Unlock()
	payload, _ := json.Marshal(map[string]string{"delta": chunk.Text})
	_ = conn.SendAcpSessionEvent(podKey, "content_delta", string(payload))
}

func (b *acpBackendBridge) onPermission(conn client.ConnectionSender, podKey string, req acp.PermissionRequest) {
	payload, _ := json.Marshal(req)
	_ = conn.SendAcpSessionEvent(podKey, "permission_request", string(payload))
}

func (b *acpBackendBridge) onStateChange(h *RunnerMessageHandler, pod *Pod, conn client.ConnectionSender, podKey, newState string) {
	b.mu.Lock()
	prev := b.states[podKey]
	b.states[podKey] = newState
	text := b.buffers[podKey]
	if prev == acp.StateProcessing && newState == acp.StateIdle && text != "" {
		b.buffers[podKey] = ""
		b.mu.Unlock()
		payload, _ := json.Marshal(map[string]string{"text": text})
		_ = conn.SendAcpSessionEvent(podKey, "message_done", string(payload))
		h.reportTurnUsage(conn, podKey, b.usageForTurn(podKey, text))
		return
	}
	b.mu.Unlock()
}
