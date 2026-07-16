package doagent

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapControlToRPC_SetModel(t *testing.T) {
	method, params, ok := mapControlToRPC("sess-1", "set_model", map[string]any{"model": "gpt-4"})
	require.True(t, ok)
	assert.Equal(t, "session/setModel", method)
	assert.Equal(t, "sess-1", params["sessionId"])
	assert.Equal(t, "gpt-4", params["model"])
}

func TestMapControlToRPC_DoAgentRPC(t *testing.T) {
	method, params, ok := mapControlToRPC("sess-1", "doagent.rpc", map[string]any{
		"method": "goal/list",
		"params": map[string]any{"cwd": "/tmp"},
	})
	require.True(t, ok)
	assert.Equal(t, "goal/list", method)
	assert.Equal(t, "sess-1", params["sessionId"])
	assert.Equal(t, "/tmp", params["cwd"])
}

func TestMapControlToRPC_GoalPassthrough(t *testing.T) {
	method, params, ok := mapControlToRPC("sess-1", "goal/pause", map[string]any{"goalId": "g1"})
	require.True(t, ok)
	assert.Equal(t, "goal/pause", method)
	assert.Equal(t, "g1", params["goalId"])
}

func TestMapControlToRPC_Unsupported(t *testing.T) {
	_, _, ok := mapControlToRPC("sess-1", "set_permission_mode", map[string]any{"mode": "allow"})
	assert.False(t, ok)
}

func TestDoAgentTransportRegistered(t *testing.T) {
	transport, err := acp.NewTransport(TransportType, acp.EventCallbacks{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}
