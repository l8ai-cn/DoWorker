package sessionapi

import (
	"context"
	"errors"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
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
	if err == nil && row != nil {
		return row.ID, true
	}
	if !errors.Is(err, sessionsvc.ErrNotFound) {
		return "", false
	}
	return p.createSessionForPod(ctx, podKey)
}

func (p *SessionStreamPublisher) createSessionForPod(
	ctx context.Context,
	podKey string,
) (string, bool) {
	if p.Pods == nil {
		return "", false
	}
	pod, err := p.Pods.GetByKey(ctx, podKey)
	if err != nil || pod == nil || pod.InteractionMode != podDomain.InteractionModeACP {
		return "", false
	}
	id, err := sessionsvc.NewID()
	if err != nil {
		return "", false
	}
	title := pod.Alias
	if title == nil {
		title = pod.Title
	}
	now := time.Now()
	row := &sessionDomain.Session{
		ID: id, OrganizationID: pod.OrganizationID, UserID: pod.CreatedByID,
		PodKey: pod.PodKey, AgentSlug: pod.AgentSlug, Title: title,
		Status: "idle", CreatedAt: now, UpdatedAt: now,
	}
	if err := p.Sessions.Create(ctx, row); err == nil {
		return row.ID, true
	}
	existing, err := p.Sessions.GetByPodKey(ctx, podKey)
	if err != nil || existing == nil {
		return "", false
	}
	return existing.ID, true
}

func (p *SessionStreamPublisher) sessionForRunnerPod(
	ctx context.Context,
	runnerID int64,
	podKey string,
) (string, bool) {
	if !p.runnerOwnsPod(ctx, runnerID, podKey) {
		return "", false
	}
	return p.sessionForPod(ctx, podKey)
}

func (p *SessionStreamPublisher) runnerOwnsPod(
	ctx context.Context,
	runnerID int64,
	podKey string,
) bool {
	if p == nil || p.Pods == nil {
		return false
	}
	_, err := p.Pods.GetByKeyAndRunner(ctx, podKey, runnerID)
	return err == nil
}

func (p *SessionStreamPublisher) sessionForRunnerPod(
	ctx context.Context,
	runnerID int64,
	podKey string,
) (string, bool) {
	if !p.runnerOwnsPod(ctx, runnerID, podKey) {
		return "", false
	}
	return p.sessionForPod(ctx, podKey)
}

func (p *SessionStreamPublisher) runnerOwnsPod(
	ctx context.Context,
	runnerID int64,
	podKey string,
) bool {
	if p == nil || p.Pods == nil {
		return false
	}
	_, err := p.Pods.GetByKeyAndRunner(ctx, podKey, runnerID)
	return err == nil
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
