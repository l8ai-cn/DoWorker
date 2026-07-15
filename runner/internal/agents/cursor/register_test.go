package cursor

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
	"github.com/stretchr/testify/assert"
)

func TestCursorRegistered(t *testing.T) {
	assert.True(t, tokenusage.IsParserOptOut("cursor-cli"), "slug cursor-cli should be opt-out from token-usage fixture contract")
	assert.True(t, tokenusage.IsParserOptOut("agent"), "launch_command agent should be opt-out from token-usage fixture contract")
	assert.Nil(t, tokenusage.GetParser("cursor-cli"), "opt-out agents must not register a parser (slug key)")
	assert.Nil(t, tokenusage.GetParser("agent"), "opt-out agents must not register a parser (runtime launch_command key)")

	transport := acp.NewTransport("cursor-acp", acp.EventCallbacks{}, nil)
	assert.NotNil(t, transport)

	assert.True(t, agentkit.IsAgentProcess("agent"), "Cursor CLI process name agent must be registered")
	assert.False(t, agentkit.IsAgentProcess("cursor"), "bare 'cursor' must NOT be registered (collides with Cursor IDE)")
	assert.False(t, agentkit.IsAgentProcess("cursor-agent"), "retired Cursor CLI executable must not be registered")
	assert.False(t, agentkit.IsAgentProcess("cursor-cli"), "the DB slug must NOT be registered as a process name")
}
