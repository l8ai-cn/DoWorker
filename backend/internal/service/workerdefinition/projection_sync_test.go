package workerdefinition

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncAgentProjectionsCreatesAndRepairsAllFormalWorkers(t *testing.T) {
	db := testkit.SetupTestDB(t)
	staleAgentFile := "AGENT claude\nEXECUTABLE claude\nMODE pty\n"
	require.NoError(t, db.Create(&agentdomain.Agent{
		Slug:              "claude-code",
		Name:              "Claude Code",
		LaunchCommand:     "claude",
		Executable:        "claude",
		AdapterID:         "legacy-adapter",
		AgentfileSource:   &staleAgentFile,
		IsBuiltin:         false,
		IsActive:          false,
		IsInternal:        true,
		SupportedModes:    "pty",
		UsesLegacyColumns: true,
	}).Error)
	catalog, err := Load(filepath.Join(repositoryRoot(t), "config", "worker-types"))
	require.NoError(t, err)

	synced, err := SyncAgentProjections(context.Background(), db, catalog)

	require.NoError(t, err)
	assert.Equal(t, len(catalog.Slugs()), synced)
	for _, slug := range catalog.Slugs() {
		definition, found := catalog.Get(slug)
		require.True(t, found)
		var projected agentdomain.Agent
		require.NoError(t, db.First(&projected, "slug = ?", slug).Error)
		assert.Equal(t, definition.Executable, projected.LaunchCommand, slug)
		assert.Equal(t, definition.Executable, projected.Executable, slug)
		assert.Equal(t, definition.AdapterID, projected.AdapterID, slug)
		require.NotNil(t, projected.AgentfileSource, slug)
		assert.Equal(t, definition.AgentFile, *projected.AgentfileSource, slug)
		assert.Equal(t, strings.Join(definition.Modes, ","), projected.SupportedModes, slug)
		assert.True(t, projected.IsBuiltin, slug)
		assert.True(t, projected.IsActive, slug)
		assert.False(t, projected.IsInternal, slug)
		assert.False(t, projected.UsesLegacyColumns, slug)
	}
}
