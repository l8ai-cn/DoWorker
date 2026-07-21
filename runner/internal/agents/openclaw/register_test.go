package openclaw

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
	"github.com/stretchr/testify/assert"
)

func TestOpenClawRegistersRuntimeContracts(t *testing.T) {
	assert.True(t, tokenusage.IsParserOptOut("openclaw"))
	assert.True(t, agentkit.IsAgentProcess("openclaw"))
}
