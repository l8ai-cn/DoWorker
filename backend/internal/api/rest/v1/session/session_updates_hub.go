package sessionapi

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	domainrunner "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
)

const updatesHeartbeatS = 30

type SessionUpdatesHub struct {
	mu    sync.RWMutex
	peers map[*updatesPeer]struct{}
	deps  *Deps
}

type updatesPeer struct {
	hub     *SessionUpdatesHub
	userID  int64
	orgID   int64
	watched map[string]struct{}
	out     chan []byte
	stop    chan struct{}
}

func NewSessionUpdatesHub(d *Deps) *SessionUpdatesHub {
	return &SessionUpdatesHub{peers: make(map[*updatesPeer]struct{}), deps: d}
}

func (h *SessionUpdatesHub) Register(userID, orgID int64) *updatesPeer {
	p := &updatesPeer{
		hub: h, userID: userID, orgID: orgID,
		watched: make(map[string]struct{}),
		out:     make(chan []byte, 32),
		stop:    make(chan struct{}),
	}
	h.mu.Lock()
	h.peers[p] = struct{}{}
	h.mu.Unlock()
	go p.heartbeat()
	return p
}

func (h *SessionUpdatesHub) Unregister(p *updatesPeer) {
	h.mu.Lock()
	if _, ok := h.peers[p]; ok {
		delete(h.peers, p)
		close(p.stop)
	}
	h.mu.Unlock()
}

func (p *updatesPeer) Out() <-chan []byte { return p.out }

func (p *updatesPeer) SetWatch(ids []string) {
	p.hub.mu.Lock()
	p.watched = make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id != "" {
			p.watched[id] = struct{}{}
		}
	}
	p.hub.mu.Unlock()
	p.push(map[string]any{"type": "snapshot", "items": p.hub.itemsForWatch(p)})
}

func (p *updatesPeer) push(frame map[string]any) {
	body, err := json.Marshal(frame)
	if err != nil {
		return
	}
	select {
	case p.out <- body:
	default:
	}
}

func (p *updatesPeer) heartbeat() {
	ticker := time.NewTicker(updatesHeartbeatS * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-ticker.C:
			p.push(map[string]any{"type": "heartbeat"})
		}
	}
}

func (h *SessionUpdatesHub) itemsForWatch(p *updatesPeer) []conversationListItem {
	if h.deps == nil || h.deps.Sessions == nil {
		return nil
	}
	h.mu.RLock()
	ids := make([]string, 0, len(p.watched))
	for id := range p.watched {
		ids = append(ids, id)
	}
	h.mu.RUnlock()
	online := h.deps.runnerOnlineMap(p.orgID, p.userID)
	out := make([]conversationListItem, 0, len(ids))
	for _, id := range ids {
		row, err := h.deps.Sessions.Get(context.Background(), id)
		if err != nil || row == nil || row.UserID != p.userID || row.OrganizationID != p.orgID {
			continue
		}
		out = append(out, h.deps.listItemFrom(row, h.deps.loadPodCtx(row.PodKey), online))
	}
	return out
}

func (h *SessionUpdatesHub) NotifyChanged(sessionID string) {
	if h == nil || sessionID == "" || h.deps == nil || h.deps.Sessions == nil {
		return
	}
	row, err := h.deps.Sessions.Get(context.Background(), sessionID)
	if err != nil || row == nil {
		return
	}
	pod := h.deps.loadPodCtx(row.PodKey)
	h.mu.RLock()
	peers := make([]*updatesPeer, 0)
	for p := range h.peers {
		if _, ok := p.watched[sessionID]; ok {
			peers = append(peers, p)
		}
	}
	h.mu.RUnlock()
	for _, p := range peers {
		online := h.deps.runnerOnlineMap(p.orgID, p.userID)
		item := h.deps.listItemFrom(row, pod, online)
		p.push(map[string]any{"type": "changed", "items": []conversationListItem{item}})
	}
}

func (d *Deps) runnerOnlineMap(orgID, userID int64) map[string]bool {
	out := make(map[string]bool)
	if d.Runner == nil {
		return out
	}
	runners, err := d.Runner.ListRunners(context.Background(), orgID, userID)
	if err != nil {
		return out
	}
	for _, r := range runners {
		if r.IsEnabled && r.Status == domainrunner.RunnerStatusOnline {
			out[r.NodeID] = true
		}
	}
	return out
}

func (d *Deps) loadPodCtx(podKey string) *podDomain.Pod {
	if d.Pod == nil || podKey == "" {
		return nil
	}
	pod, _ := d.Pod.GetPod(context.Background(), podKey)
	return pod
}
