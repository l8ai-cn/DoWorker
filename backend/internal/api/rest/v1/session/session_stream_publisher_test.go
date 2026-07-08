package sessionapi

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/require"
)

func TestEnsureTurnAutoStarts(t *testing.T) {
	hub := NewSessionHub()
	p := &SessionStreamPublisher{Hub: hub}
	id := p.ensureTurn("conv_test", "")
	require.NotEmpty(t, id)
	got, ok := hub.ActiveResponse("conv_test")
	require.True(t, ok)
	require.Equal(t, id, got)
}

func TestMapPodSessionStatus(t *testing.T) {
	require.Equal(t, "launching", mapPodSessionStatus(podDomain.StatusInitializing, ""))
	require.Equal(t, "launching", mapPodSessionStatus(podDomain.StatusQueued, ""))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusRunning, podDomain.AgentStatusIdle))
	require.Equal(t, "running", mapPodSessionStatus(podDomain.StatusRunning, podDomain.AgentStatusExecuting))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusCompleted, ""))
	require.Equal(t, "failed", mapPodSessionStatus(podDomain.StatusOrphaned, ""))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusDisconnected, podDomain.AgentStatusIdle))
}

func TestMapSessionStatusDelegates(t *testing.T) {
	require.Equal(t, "idle", mapSessionStatus(nil))
	require.Equal(t, "launching", mapSessionStatus(&podDomain.Pod{Status: podDomain.StatusQueued}))
}

func TestHandleAcpSessionIdleFinalizesBuffer(t *testing.T) {
	hub := NewSessionHub()
	p := &SessionStreamPublisher{Hub: hub, Sessions: nil}
	p.HandleAcpSession(context.Background(), "missing-pod", "sessionState", `{"state":"idle"}`)
}
