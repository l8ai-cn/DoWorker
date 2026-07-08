package aimodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHarnessEnvVars_CodexOpenAI(t *testing.T) {
	m := &AIModel{ProviderType: ProviderTypeOpenAI, Model: "gpt-5.5", BaseURL: "https://token.example.cn"}
	env := HarnessEnvVars("codex-cli", "", m, map[string]string{"api_key": "sk-test"})
	assert.Equal(t, "sk-test", env["OPENAI_API_KEY"])
	assert.Equal(t, "https://token.example.cn", env["OPENAI_BASE_URL"])
	assert.Equal(t, "gpt-5.5", env["OPENAI_MODEL"])
}

func TestHarnessEnvVars_CodexOverrideModel(t *testing.T) {
	m := &AIModel{ProviderType: ProviderTypeOpenAI, Model: "gpt-5.5"}
	env := HarnessEnvVars("codex-cli", "gpt-4.1", m, map[string]string{"api_key": "sk-test"})
	assert.Equal(t, "gpt-4.1", env["OPENAI_MODEL"])
}

func TestHarnessEnvVars_WrongProvider(t *testing.T) {
	m := &AIModel{ProviderType: ProviderTypeMiniMax, Model: "MiniMax-M3"}
	assert.Nil(t, HarnessEnvVars("codex-cli", "", m, map[string]string{"api_key": "sk-test"}))
}

func TestHarnessEnvVars_ClaudeAnthropic(t *testing.T) {
	m := &AIModel{ProviderType: ProviderTypeAnthropic, BaseURL: "https://api.anthropic.com"}
	env := HarnessEnvVars("claude-code", "", m, map[string]string{"api_key": "sk-ant-test"})
	assert.Equal(t, "sk-ant-test", env["ANTHROPIC_API_KEY"])
	assert.Equal(t, "https://api.anthropic.com", env["ANTHROPIC_BASE_URL"])
}

func TestPreferredProviders(t *testing.T) {
	assert.Equal(t, []string{ProviderTypeOpenAI}, PreferredProviders("codex-cli"))
	assert.Equal(t, []string{ProviderTypeAnthropic}, PreferredProviders("claude-code"))
	assert.Equal(t, []string{ProviderTypeGemini}, PreferredProviders("gemini-cli"))
	assert.Equal(t, []string{ProviderTypeAnthropic, ProviderTypeMiniMax}, PreferredProviders("do-agent"))
	assert.Nil(t, PreferredProviders("e2e-echo"))
}

func TestHarnessMountKindFor(t *testing.T) {
	assert.Equal(t, HarnessMountConfig, HarnessMountKindFor("do-agent", true))
	assert.Equal(t, HarnessMountEnv, HarnessMountKindFor("codex-cli", false))
	assert.Equal(t, HarnessMountEnv, HarnessMountKindFor("claude-code", false))
	assert.Equal(t, HarnessMountNone, HarnessMountKindFor("e2e-echo", false))
}
