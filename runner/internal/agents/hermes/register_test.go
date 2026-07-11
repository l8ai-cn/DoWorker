package harn

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
	"github.com/stretchr/testify/assert"
)

func TestHarnRegistersRuntimeContracts(t *testing.T) {
	assert.True(t, tokenusage.IsParserOptOut("harn"))
	assert.True(t, agentkit.IsAgentProcess("harn"))
}
