package gemini

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/stretchr/testify/assert"
)

func TestGeminiRegistered(t *testing.T) {
	assert.True(t, agentkit.IsAgentProcess("gemini"))
}
