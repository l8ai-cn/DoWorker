package main

import (
	"testing"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeOrchestrationControlBuildsPhase2AService(t *testing.T) {
	creation := workercreation.NewService(workercreation.Deps{
		Catalog: workerruntime.DefaultCatalog(),
	})

	service, err := initializeOrchestrationControl(
		testkit.SetupTestDB(t),
		organization.NewService(nil),
		creation,
	)

	require.NoError(t, err)
	assert.NotNil(t, service)
}

func TestInitializeOrchestrationControlRejectsMissingDependencies(t *testing.T) {
	_, err := initializeOrchestrationControl(nil, nil, nil)
	assert.Error(t, err)
}

func TestAttachOrchestrationControlStoresServiceOnContainer(t *testing.T) {
	services := &serviceContainer{
		org: organization.NewService(nil),
		workerServices: workerServices{
			workerCreation: workercreation.NewService(workercreation.Deps{
				Catalog: workerruntime.DefaultCatalog(),
			}),
		},
	}

	err := attachOrchestrationControl(
		services,
		testkit.SetupTestDB(t),
	)

	require.NoError(t, err)
	assert.NotNil(t, services.orchestration)
}
