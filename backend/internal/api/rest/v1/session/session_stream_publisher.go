package sessionapi

import (
	"context"
	"time"

	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
)

type SessionStreamPublisher struct {
	Items        *itemsvc.Service
	Sessions     *sessionsvc.Service
	Hub          *SessionHub
	Elicitations *ElicitationStore
	Updates      *SessionUpdatesHub
	Usage        *sessionusagesvc.Service
	Pods         PodUpdater
}

func NewSessionStreamPublisher(
	hub *SessionHub,
	items *itemsvc.Service,
	sessions *sessionsvc.Service,
	elicitations *ElicitationStore,
) *SessionStreamPublisher {
	return &SessionStreamPublisher{
		Hub: hub, Items: items, Sessions: sessions, Elicitations: elicitations,
	}
}

func itemsvcNewTurnID() (string, error) {
	return itemsvc.NewResponseID()
}

func (p *SessionStreamPublisher) sessionForPod(ctx context.Context, podKey string) (string, bool) {
	if p == nil || p.Sessions == nil {
		return "", false
	}
	row, err := p.Sessions.GetByPodKey(ctx, podKey)
	if err != nil || row == nil {
		return "", false
	}
	return row.ID, true
}

func (p *SessionStreamPublisher) ensureTurn(sessionID, turnID string) string {
	if p == nil || p.Hub == nil {
		return turnID
	}
	if turnID != "" {
		if active, ok := p.Hub.ActiveResponse(sessionID); !ok || active != turnID {
			p.Hub.StartTurn(sessionID, turnID)
		}
		return turnID
	}
	if active, ok := p.Hub.ActiveResponse(sessionID); ok {
		return active
	}
	id, err := itemsvcNewTurnID()
	if err != nil {
		return ""
	}
	p.publishTurnStarted(sessionID, id)
	return id
}

func (p *SessionStreamPublisher) publishTurnStarted(sessionID, turnID string) {
	if p == nil || p.Hub == nil || turnID == "" {
		return
	}
	p.Hub.StartTurn(sessionID, turnID)
	now := time.Now().Unix()
	p.Hub.Publish(sessionID, formatSSE(sseTurnStarted, map[string]any{
		"id": turnID, "status": "in_progress", "model": "", "created_at": now,
		"conversation": map[string]any{"id": sessionID},
	}))
}

func (p *SessionStreamPublisher) touchSession(ctx context.Context, sessionID string) {
	if p.Sessions != nil {
		_ = p.Sessions.TouchUpdatedAt(ctx, sessionID)
	}
	if p.Updates != nil {
		p.Updates.NotifyChanged(sessionID)
	}
}
