package runner

import "sync"

type podStatusListener func(podKey, status, agentStatus string)

type podStatusBroadcaster struct {
	mu        sync.RWMutex
	listeners []podStatusListener
}

func newPodStatusBroadcaster() *podStatusBroadcaster {
	return &podStatusBroadcaster{}
}

func (b *podStatusBroadcaster) set(fn podStatusListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if fn == nil {
		b.listeners = nil
		return
	}
	b.listeners = []podStatusListener{fn}
}

func (b *podStatusBroadcaster) add(fn podStatusListener) {
	if fn == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners = append(b.listeners, fn)
}

func (b *podStatusBroadcaster) notify(podKey, status, agentStatus string) {
	b.mu.RLock()
	listeners := append([]podStatusListener(nil), b.listeners...)
	b.mu.RUnlock()
	for _, fn := range listeners {
		fn(podKey, status, agentStatus)
	}
}

func (b *podStatusBroadcaster) len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.listeners)
}
