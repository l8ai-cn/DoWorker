package hermes

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
	"github.com/stretchr/testify/assert"
)

func TestHermesRegistersRuntimeContracts(t *testing.T) {
	assert.True(t, tokenusage.IsParserOptOut("hermes"))
	assert.True(t, tokenusage.IsParserOptOut("hermes-acp"))
	assert.True(t, agentkit.IsAgentProcess("hermes"))
	assert.True(t, agentkit.IsAgentProcess("hermes-acp"))
}
