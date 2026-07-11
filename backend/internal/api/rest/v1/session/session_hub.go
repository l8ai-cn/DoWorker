package sessionapi

import (
	"sync"
)

type SessionHub struct {
	mu       sync.RWMutex
	channels map[string]map[chan string]struct{}
	turns    map[string]*activeTurn
	scratch  map[string]*streamScratch
}

type activeTurn struct {
	ResponseID string
	Buffer     string
}

func NewSessionHub() *SessionHub {
	return &SessionHub{
		channels: make(map[string]map[chan string]struct{}),
		turns:    make(map[string]*activeTurn),
		scratch:  make(map[string]*streamScratch),
	}
}

func (h *SessionHub) scratchFor(sessionID string) *streamScratch {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.scratch[sessionID] == nil {
		h.scratch[sessionID] = &streamScratch{
			toolCalls: make(map[string]toolCallState),
			reasoning: make(map[string]string),
		}
	}
	return h.scratch[sessionID]
}

func (h *SessionHub) Subscribe(sessionID string) chan string {
	ch := make(chan string, 64)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels[sessionID] == nil {
		h.channels[sessionID] = make(map[chan string]struct{})
	}
	h.channels[sessionID][ch] = struct{}{}
	return ch
}

func (h *SessionHub) Unsubscribe(sessionID string, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.channels[sessionID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(h.channels, sessionID)
		}
	}
	close(ch)
}

func (h *SessionHub) Publish(sessionID, frame string) {
	h.mu.RLock()
	subs := h.channels[sessionID]
	h.mu.RUnlock()
	for ch := range subs {
		select {
		case ch <- frame:
		default:
		}
	}
}

func (h *SessionHub) StartTurn(sessionID, responseID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.turns[sessionID] = &activeTurn{ResponseID: responseID}
}

func (h *SessionHub) ActiveResponse(sessionID string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t, ok := h.turns[sessionID]
	if !ok {
		return "", false
	}
	return t.ResponseID, true
}

func (h *SessionHub) AppendDelta(sessionID, delta string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if t, ok := h.turns[sessionID]; ok {
		t.Buffer += delta
	}
}

func (h *SessionHub) FinishTurn(sessionID string) (responseID, text string, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	t, found := h.turns[sessionID]
	if !found {
		return "", "", false
	}
	delete(h.turns, sessionID)
	return t.ResponseID, t.Buffer, true
}

// RemoveSession drops in-memory turn/scratch state after the session row is
// deleted. Live SSE subscribers remain until they disconnect and Unsubscribe.
func (h *SessionHub) RemoveSession(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.turns, sessionID)
	delete(h.scratch, sessionID)
}
