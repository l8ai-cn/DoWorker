package openclaw

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
	"github.com/stretchr/testify/assert"
)

func TestOpenClawRegistersRuntimeContracts(t *testing.T) {
	assert.True(t, tokenusage.IsParserOptOut("openclaw"))
	assert.True(t, agentkit.IsAgentProcess("openclaw"))
}
