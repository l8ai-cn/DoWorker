package main

import (
	"path/filepath"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeWorkerServicesValidatesDefinitionCatalog(t *testing.T) {
	t.Run("rejects an incomplete catalog", func(t *testing.T) {
		_, err := initializeWorkerServices(
			&config.Config{WorkerDefinitionsDir: t.TempDir()},
			nil, nil, nil, nil, nil, nil,
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "read worker definition schema")
	})

	t.Run("loads the formal catalog", func(t *testing.T) {
		root, err := filepath.Abs(filepath.Join("..", "..", "..", "config", "worker-types"))
		require.NoError(t, err)

		services, err := initializeWorkerServices(
			&config.Config{WorkerDefinitionsDir: root},
			nil, nil, nil, nil, nil, nil,
		)

		require.NoError(t, err)
		assert.ElementsMatch(t, []string{
			"aider", "claude-code", "codex-cli", "cursor-cli", "do-agent",
			"e2e-echo", "gemini-cli", "grok-build", "hermes", "loopal",
			"minimax-cli", "openclaw", "opencode", "pattern-designer",
			"seedance-expert", "video-studio",
		}, services.workerDefinitions.Slugs())
	})
}
