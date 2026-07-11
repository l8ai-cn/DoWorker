package grok

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func TestGrokBuildRegistered(t *testing.T) {
	assert.Nil(t, tokenusage.GetParser("grok"))
	assert.True(t, tokenusage.IsParserOptOut("grok"))
	assert.True(t, tokenusage.IsParserOptOut("grok-build"))
	assert.True(t, agentkit.IsAgentProcess("grok"))
	assert.Equal(t, TransportType, acp.TransportTypeForCommand("grok"))
}
