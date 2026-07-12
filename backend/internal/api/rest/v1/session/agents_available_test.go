package sessionapi

import (
	"testing"
	"time"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
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
		{ID: "codex-cli", Name: "Codex", Builtin: true, CreatedAt: now.Unix(), Harness: stringPtr("codex-cli"), SupportedModes: []string{"acp", "pty"}},
	}, rows)
}

func stringPtr(value string) *string {
	return &value
}
