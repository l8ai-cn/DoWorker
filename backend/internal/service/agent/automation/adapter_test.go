package automation

import (
	"strings"
	"testing"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/assert"
)

func TestLayerLinesFor_ClaudeLevels(t *testing.T) {
	assert.Equal(t, `CONFIG permission_mode = "default"`,
		LayerLinesFor("claude-code", podDomain.AutomationLevelInteractive, true))
	assert.Equal(t, `CONFIG permission_mode = "acceptEdits"`,
		LayerLinesFor("claude-code", podDomain.AutomationLevelAutoEdit, true))
	assert.Equal(t, "CONFIG permission_mode = \"bypassPermissions\"\nMODE acp",
		LayerLinesFor("claude-code", podDomain.AutomationLevelAutonomous, true))
}

func TestLayerLinesFor_CodexLevels(t *testing.T) {
	for _, agentSlug := range []string{"codex-cli", "video-studio"} {
		assert.Equal(t, `CONFIG approval_mode = "untrusted"`,
			LayerLinesFor(agentSlug, podDomain.AutomationLevelInteractive, true))
		assert.Equal(t, `CONFIG approval_mode = "on-request"`,
			LayerLinesFor(agentSlug, podDomain.AutomationLevelAutoEdit, true))
		autonomous := LayerLinesFor(agentSlug, podDomain.AutomationLevelAutonomous, true)
		assert.Contains(t, autonomous, `CONFIG approval_mode = "never"`)
		assert.Contains(t, autonomous, "MODE acp")
	}
}

func TestLayerLinesFor_LoopalLevels(t *testing.T) {
	assert.Equal(t, `CONFIG permission_mode = "supervised"`,
		LayerLinesFor("loopal", podDomain.AutomationLevelInteractive, true))
	assert.Equal(t, `CONFIG permission_mode = "auto"`,
		LayerLinesFor("loopal", podDomain.AutomationLevelAutoEdit, true))
	assert.Contains(t, LayerLinesFor("loopal", podDomain.AutomationLevelAutonomous, true),
		`CONFIG permission_mode = "bypass"`)
}

// Autonomous forces ACP for every agent when the agent supports it, so the
// non-interactive automation guarantee holds cross-agent.
func TestLayerLinesFor_AutonomousForcesACPWhenSupported(t *testing.T) {
	for _, slug := range []string{"claude-code", "codex-cli", "loopal", "gemini-cli", "do-agent", "aider", "unknown-agent"} {
		lines := LayerLinesFor(slug, podDomain.AutomationLevelAutonomous, true)
		assert.True(t, strings.Contains(lines, "MODE acp"),
			"agent %s autonomous should force MODE acp, got %q", slug, lines)
	}
}

// A pty-only agent (canForceMode=false) must not get a MODE line, but CONFIG
// overrides still apply so the tier is honored as far as the agent allows.
func TestLayerLinesFor_PtyOnlyAgentDropsMode(t *testing.T) {
	lines := LayerLinesFor("claude-code", podDomain.AutomationLevelAutonomous, false)
	assert.NotContains(t, lines, "MODE ")
	assert.Contains(t, lines, `CONFIG permission_mode = "bypassPermissions"`)
	assert.Equal(t, "", LayerLinesFor("gemini-cli", podDomain.AutomationLevelAutonomous, false))
}

// Empty/unknown level normalizes to the autonomous default.
func TestLayerLinesFor_EmptyDefaultsToAutonomous(t *testing.T) {
	assert.Equal(t,
		LayerLinesFor("claude-code", podDomain.AutomationLevelAutonomous, true),
		LayerLinesFor("claude-code", "", true))
	assert.Equal(t,
		LayerLinesFor("claude-code", podDomain.AutomationLevelAutonomous, true),
		LayerLinesFor("claude-code", "bogus", true))
}

// Non-autonomous levels must not force a MODE, leaving the user/base choice.
func TestLayerLinesFor_NonAutonomousLeavesMode(t *testing.T) {
	for _, lvl := range []string{podDomain.AutomationLevelInteractive, podDomain.AutomationLevelAutoEdit} {
		assert.NotContains(t, LayerLinesFor("claude-code", lvl, true), "MODE ")
		assert.Equal(t, "", LayerLinesFor("gemini-cli", lvl, true))
	}
}
