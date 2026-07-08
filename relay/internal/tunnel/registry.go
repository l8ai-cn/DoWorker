package tunnel

import (
	"context"
	"sync"
	"time"
)

// Registry maps runnerID -> active Tunnel and coordinates reconnect grace.
type Registry struct {
	mu      sync.RWMutex
	tunnels map[int64]*Tunnel
	waiters map[int64][]chan *Tunnel
}

// RegistryStats is a point-in-time snapshot for heartbeat/otel.
type RegistryStats struct {
	ActiveTunnels int
	ActiveStreams int
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		tunnels: make(map[int64]*Tunnel),
		waiters: make(map[int64][]chan *Tunnel),
	}
}

// Register installs a tunnel for its runnerID. If one already exists it is
// closed (reconnect takeover) and replaced. Any waiters are woken with the new
// tunnel.
func (r *Registry) Register(t *Tunnel) {
	r.mu.Lock()
	old := r.tunnels[t.RunnerID]
	r.tunnels[t.RunnerID] = t
	waiters := r.waiters[t.RunnerID]
	delete(r.waiters, t.RunnerID)
	r.mu.Unlock()

	if old != nil && old != t {
		old.Close()
	}
	for _, ch := range waiters {
		select {
		case ch <- t:
		default:
		}
	}
}

// Get returns the tunnel for runnerID, or nil.
func (r *Registry) Get(runnerID int64) *Tunnel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tunnels[runnerID]
}

// Unregister removes the tunnel only if the currently-registered instance is
// the given one (avoids removing a newer reconnect).
func (r *Registry) Unregister(t *Tunnel) {
	r.mu.Lock()
	if r.tunnels[t.RunnerID] == t {
		delete(r.tunnels, t.RunnerID)
	}
	r.mu.Unlock()
}

// WaitForTunnel returns the tunnel for runnerID immediately if present,
// otherwise polls for up to grace (covering a runner reconnect window). Returns
// nil if the grace elapses or ctx is cancelled.
func (r *Registry) WaitForTunnel(ctx context.Context, runnerID int64, grace time.Duration) *Tunnel {
	if t := r.Get(runnerID); t != nil {
		return t
	}
	if grace <= 0 {
		return nil
	}
	deadline := time.After(grace)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-deadline:
			return r.Get(runnerID)
		case <-ticker.C:
			if t := r.Get(runnerID); t != nil {
				return t
			}
		}
	}
}

// Stats returns a snapshot of tunnel/stream counts.
func (r *Registry) Stats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	streams := 0
	for _, t := range r.tunnels {
		streams += t.StreamCount()
	}
	return RegistryStats{ActiveTunnels: len(r.tunnels), ActiveStreams: streams}
}
