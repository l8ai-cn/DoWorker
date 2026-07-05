package omnigent

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/require"
)

func TestEnsureResponseTurnAutoStarts(t *testing.T) {
	hub := NewSessionHub()
	b := &EventBridge{Hub: hub}
	id := b.ensureResponseTurn("conv_test", "")
	require.NotEmpty(t, id)
	got, ok := hub.ActiveResponse("conv_test")
	require.True(t, ok)
	require.Equal(t, id, got)
}

func TestMapPodSessionStatus(t *testing.T) {
	require.Equal(t, "launching", mapPodSessionStatus(podDomain.StatusInitializing, ""))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusRunning, podDomain.AgentStatusIdle))
	require.Equal(t, "running", mapPodSessionStatus(podDomain.StatusRunning, podDomain.AgentStatusExecuting))
}

func TestHandleAcpSessionMessageDoneUsesBuffer(t *testing.T) {
	hub := NewSessionHub()
	b := &EventBridge{Hub: hub, Sessions: nil}
	// No session lookup — should not panic when session missing.
	b.HandleAcpSession(context.Background(), "missing-pod", "message_done", `{"text":"hi"}`)
}
