package doagent

import (
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	doagentfixture "github.com/anthropics/agentsmesh/runner/internal/agents/doagent/testsupport"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoAgentRegistered(t *testing.T) {
	assert.NotNil(t, tokenusage.GetParser("do-agent"))
	assert.True(t, agentkit.IsAgentProcess("do-agent"))
	transport, err := acp.NewTransport(TransportType, acp.EventCallbacks{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestDoAgentParser_Fixture(t *testing.T) {
	sandbox := doagentfixture.BuildFixtureSandbox(t)
	usage, err := (&doagentParser{}).Parse(sandbox, time.Unix(0, 0))
	require.NoError(t, err)
	require.NotNil(t, usage)

	m := usage.Models["deepseek/deepseek-v4-pro"]
	require.NotNil(t, m)
	assert.Equal(t, int64(2000), m.InputTokens)
	assert.Equal(t, int64(460), m.OutputTokens)
}
