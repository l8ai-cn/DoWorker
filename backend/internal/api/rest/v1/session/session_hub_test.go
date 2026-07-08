package sessionapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionHub_RemoveSession(t *testing.T) {
	hub := NewSessionHub()
	sessionID := "conv_remove"

	hub.StartTurn(sessionID, "resp_1")
	hub.scratchFor(sessionID).toolCalls["tc1"] = toolCallState{}

	hub.RemoveSession(sessionID)

	_, ok := hub.ActiveResponse(sessionID)
	require.False(t, ok)

	hub.mu.RLock()
	_, hasScratch := hub.scratch[sessionID]
	hub.mu.RUnlock()
	require.False(t, hasScratch)
}

func TestElicitationStore_RemoveSession(t *testing.T) {
	store := NewElicitationStore()
	store.Add("sess-1", &ElicitationRecord{ID: "e1", Status: "pending"})
	store.RemoveSession("sess-1")
	require.Equal(t, 0, store.PendingCount("sess-1"))
}

func TestSessionHub_PublishIdleOnDelete(t *testing.T) {
	hub := NewSessionHub()
	sessionID := "conv_delete_idle"

	ch := hub.Subscribe(sessionID)
	defer hub.Unsubscribe(sessionID, ch)

	hub.Publish(sessionID, formatSSE("session.status", map[string]any{
		"conversation_id": sessionID, "status": "idle",
	}))
	hub.RemoveSession(sessionID)

	msg := <-ch
	require.True(t, strings.Contains(msg, "session.status"))
	require.True(t, strings.Contains(msg, `"status":"idle"`))
}
