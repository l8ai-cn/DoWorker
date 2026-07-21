package codex

import (
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func (t *transport) handleTurnStarted(params json.RawMessage) {
	var started turnStartedParams
	if err := json.Unmarshal(params, &started); err != nil {
		t.logger.Warn("failed to parse turn/started", "error", err)
		return
	}
	if started.Turn.ID == "" {
		t.logger.Warn("turn/started missing turn id")
		return
	}
	t.turnMu.Lock()
	t.turnID = started.Turn.ID
	t.turnMu.Unlock()
	t.markLifecycleSignal()
	if t.callbacks.OnStateChange != nil {
		t.callbacks.OnStateChange(acp.StateProcessing)
	}
}

func (t *transport) getActiveTurnID() string {
	t.turnMu.RLock()
	defer t.turnMu.RUnlock()
	return t.turnID
}

func (t *transport) clearActiveTurn(completedTurnID string) {
	t.turnMu.Lock()
	defer t.turnMu.Unlock()
	if completedTurnID == "" || t.turnID == completedTurnID {
		t.turnID = ""
	}
}
