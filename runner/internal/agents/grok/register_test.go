package grok

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

func TestGrokBuildRegistered(t *testing.T) {
	assert.Nil(t, tokenusage.GetParser("grok"))
	assert.True(t, tokenusage.IsParserOptOut("grok"))
	assert.True(t, tokenusage.IsParserOptOut("grok-build"))
	assert.True(t, agentkit.IsAgentProcess("grok"))
	transport, err := acp.NewTransport(TransportType, acp.EventCallbacks{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, transport)
}
