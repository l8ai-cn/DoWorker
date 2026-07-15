package runner

import (
	"context"
	"time"
)

const (
	tunnelStateConnected    = "connected"
	tunnelStateDisconnected = "disconnected"
	tunnelConnectionFailed  = "tunnel_connection_failed"
)

func (s *Service) UpdateTunnelConnection(
	ctx context.Context,
	runnerID int64,
	connected bool,
	errorCode string,
) error {
	now := time.Now()
	state := tunnelStateConnected
	var lastError *string
	if !connected {
		state = tunnelStateDisconnected
		code := tunnelErrorCode(errorCode)
		lastError = &code
	}
	if err := s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"tunnel_state":        state,
		"tunnel_last_seen_at": now,
		"tunnel_last_error":   lastError,
	}); err != nil {
		return err
	}
	s.updateActiveTunnelConnection(runnerID, state, now, lastError)
	return nil
}

func tunnelErrorCode(errorCode string) string {
	if errorCode == "" {
		return tunnelConnectionFailed
	}
	if len(errorCode) <= 255 {
		return errorCode
	}
	return errorCode[:255]
}

func (s *Service) updateActiveTunnelConnection(
	runnerID int64,
	state string,
	lastSeenAt time.Time,
	lastError *string,
) {
	s.activeMu.Lock()
	defer s.activeMu.Unlock()

	active, ok := s.activeRunners.Load(runnerID)
	if !ok {
		return
	}
	ar, ok := active.(*ActiveRunner)
	if !ok || ar.Runner == nil {
		return
	}
	updated := *ar.Runner
	updated.TunnelState = state
	updated.TunnelLastSeenAt = &lastSeenAt
	updated.TunnelLastError = lastError
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   &updated,
		LastPing: ar.LastPing,
		PodCount: ar.PodCount,
	})
}
