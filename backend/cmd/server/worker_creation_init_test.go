package main

import (
	"path/filepath"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeWorkerServicesValidatesDefinitionCatalog(t *testing.T) {
	t.Run("rejects incomplete catalog", func(t *testing.T) {
		_, err := initializeWorkerServices(
			&config.Config{WorkerDefinitionsDir: t.TempDir()},
			nil, nil, nil, nil,
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "read worker definition schema")
	})

	t.Run("loads formal catalog", func(t *testing.T) {
		root, err := filepath.Abs(filepath.Join("..", "..", "..", "config", "worker-types"))
		require.NoError(t, err)

		services, err := initializeWorkerServices(
			&config.Config{WorkerDefinitionsDir: root},
			nil, nil, nil, nil,
		)

		require.NoError(t, err)
		assert.Len(t, services.workerDefinitions.Slugs(), 12)
	})
}
