package sessionapi

import (
	"testing"
	"time"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailableAgentRowsOnlyReturnsAgentsBackedByEligibleRunner(t *testing.T) {
	now := time.Now()
	rows, err := availableAgentRows(
		[]*agentDomain.Agent{
			{Slug: "aider", Name: "Aider", IsBuiltin: true, IsActive: true, SupportedModes: "pty", CreatedAt: now},
			{Slug: "codex-cli", Name: "Codex", IsBuiltin: true, IsActive: true, SupportedModes: "acp,pty", CreatedAt: now},
			{Slug: "gemini-cli", Name: "Gemini", IsBuiltin: true, IsActive: true, SupportedModes: "acp", CreatedAt: now},
		},
		[]string{"aider", "codex-cli"},
		false,
	)

	require.NoError(t, err)
	assert.Equal(t, []agentWire{
		{ID: "aider", Name: "Aider", Builtin: true, CreatedAt: now.Unix(), Harness: stringPtr("aider"), SupportedModes: []string{"pty"}},
		{ID: "codex-cli", Name: "Codex", Builtin: true, CreatedAt: now.Unix(), Harness: stringPtr("codex-cli"), SupportedModes: []string{"acp", "pty"}, RequiresModelResource: true},
	}, rows)
}

func TestAvailableAgentRowsProjectsModelResourceRequirement(t *testing.T) {
	now := time.Now()
	rows, err := availableAgentRows(
		[]*agentDomain.Agent{
			{
				Slug: "custom-codex", Name: "Custom Codex", IsBuiltin: true, IsActive: true,
				Executable: "codex", SupportedModes: "acp,pty", CreatedAt: now,
			},
			{
				Slug: "cursor-cli", Name: "Cursor", IsBuiltin: true, IsActive: true,
				Executable: "agent", SupportedModes: "acp,pty", CreatedAt: now,
			},
		},
		[]string{"custom-codex", "cursor-cli"},
		false,
	)

	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "cursor-cli", rows[0].ID)
	assert.False(t, rows[0].RequiresModelResource)
	require.Equal(t, "custom-codex", rows[1].ID)
	assert.True(t, rows[1].RequiresModelResource)
}

func stringPtr(value string) *string {
	return &value
}
