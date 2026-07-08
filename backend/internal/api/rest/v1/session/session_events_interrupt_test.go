package sessionapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishSessionInterrupted_NestedEnvelope(t *testing.T) {
	hub := NewSessionHub()
	stream := &SessionStreamPublisher{Hub: hub}
	sessionID := "conv_interrupt"

	ch := hub.Subscribe(sessionID)
	defer hub.Unsubscribe(sessionID, ch)

	hub.StartTurn(sessionID, "resp_codex")
	stream.PublishSessionInterrupted(sessionID, "resp_codex")

	msg := <-ch
	require.True(t, strings.Contains(msg, "event: session.interrupted"))
	require.True(t, strings.Contains(msg, `"type":"session.interrupted"`))
	require.True(t, strings.Contains(msg, `"response_id":"resp_codex"`))
}
