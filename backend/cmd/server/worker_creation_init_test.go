package main

import (
	"testing"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeWorkerCreationServiceUsesDefaultCatalog(t *testing.T) {
	service := initializeWorkerCreationService(nil, nil, nil, nil)

	require.NotNil(t, service)
	assert.Equal(t, workerruntime.DefaultCatalogRevision, service.Revision())
}
