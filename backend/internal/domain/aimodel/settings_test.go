package aimodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoAgentSettings_MiniMax(t *testing.T) {
	s := DoAgentSettings("minimax", "MiniMax-M3", "https://api.minimax.chat/v1", map[string]string{
		"api_key": "sk-test",
	})
	provider, ok := s["provider"].(map[string]interface{})
	assert.True(t, ok)
	minimax, ok := provider["minimax"].(map[string]interface{})
	assert.True(t, ok)
	opts, ok := minimax["options"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "sk-test", opts["apiKey"])
	assert.Equal(t, "anthropic", opts["kind"])
	assert.Equal(t, "minimax/MiniMax-M3", s["model"])
}
