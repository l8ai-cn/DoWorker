package codex

import (
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func (t *transport) markAgentMessageStreamed(itemID string) {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()
	// Track that streaming happened this turn even when the build omits itemId,
	// so item/completed can suppress the redundant full-text re-emit.
	t.streamedSinceMsg = true
	if itemID == "" {
		return
	}
	if t.streamedAgentMsgIDs == nil {
		t.streamedAgentMsgIDs = make(map[string]struct{})
	}
	t.streamedAgentMsgIDs[itemID] = struct{}{}
}

// agentMessageAlreadyEmitted reports whether the just-completed agentMessage was
// already delivered via streaming deltas — so item/completed must not re-emit it
// (that re-emit is what doubled every assistant message). Matches by itemId when
// present, else falls back to the per-turn streamed flag for Codex builds that
// stream deltas without an itemId. Consumes the tracking either way.
func (t *transport) agentMessageAlreadyEmitted(itemID string) bool {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()
	streamed := t.streamedSinceMsg
	t.streamedSinceMsg = false
	if itemID != "" {
		if _, ok := t.streamedAgentMsgIDs[itemID]; ok {
			delete(t.streamedAgentMsgIDs, itemID)
			return true
		}
	}
	return streamed
}

func (t *transport) applyAgentMessageBoundary(text string) string {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()
	if !t.messageBoundaryDue {
		return text
	}
	t.messageBoundaryDue = false
	return "\n\n" + text
}

func (t *transport) markAgentMessageCompleted() {
	t.streamMu.Lock()
	t.messageBoundaryDue = true
	t.streamMu.Unlock()
}

func (t *transport) resetAgentMessageBoundary() {
	t.streamMu.Lock()
	t.messageBoundaryDue = false
	t.streamMu.Unlock()
}

func (t *transport) notifyTurnIdle() {
	if t.callbacks.OnStateChange != nil {
		t.callbacks.OnStateChange(acp.StateIdle)
	}
}

// scheduleIdleAfterMessage arms the end-of-turn fallback for builds that omit
// turn/completed. cancelIdleFallback (called on any further turn activity or on
// an authoritative turn/completed) prevents a preamble message from ending the
// turn while the agent is still working.
func (t *transport) markLifecycleSignal() {
	t.idleMu.Lock()
	t.hasLifecycleSignal = true
	t.idleMu.Unlock()
}

func (t *transport) scheduleIdleAfterMessage() {
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	// Builds that emit thread/status/changed or turn/completed drive idle
	// authoritatively; arming the debounce there would end the turn during a
	// long silent generation gap.
	if t.hasLifecycleSignal {
		return
	}
	if t.idleTimer != nil {
		t.idleTimer.Stop()
	}
	t.idleTimer = time.AfterFunc(t.idleFallback, t.notifyTurnIdle)
}

func (t *transport) cancelIdleFallback() {
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	if t.idleTimer != nil {
		t.idleTimer.Stop()
		t.idleTimer = nil
	}
}
