package main

import (
	"testing"

	workerruntime "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/organization"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeOrchestrationControlBuildsRequiredPlanners(t *testing.T) {
	creation := workercreation.NewService(workercreation.Deps{
		Catalog: workerruntime.DefaultCatalog(),
	})

	services, err := initializeOrchestrationControl(
		testkit.SetupTestDB(t),
		organization.NewService(nil),
		creation,
	)

	require.NoError(t, err)
	assert.NotNil(t, services.control)
	assert.NotNil(t, services.bindingApply)
	assert.NotNil(t, services.workerTemplateApply)
	assert.NotNil(t, services.promptApply)
	assert.NotNil(t, services.expertApply)
	assert.NotNil(t, services.workflowApply)
	assert.NotNil(t, services.goalLoopApply)
	assert.NotNil(t, services.workerApplyRuntime.registry)
	assert.NotNil(t, services.workerApplyRuntime.repository)
	assert.NotNil(t, services.workerApplyRuntime.resolver)
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
	assert.NotNil(t, services.bindingApply)
	assert.NotNil(t, services.workerTemplateApply)
	assert.NotNil(t, services.promptApply)
	assert.NotNil(t, services.expertApply)
	assert.NotNil(t, services.workflowApply)
	assert.NotNil(t, services.goalLoopApply)
	assert.NotNil(t, services.workerApplyRuntime.registry)
	assert.NotNil(t, services.workerApplyRuntime.repository)
	assert.NotNil(t, services.workerApplyRuntime.resolver)
}

func TestAttachOrchestrationWorkerApplyStoresRuntimeService(t *testing.T) {
	services := &serviceContainer{
		org: organization.NewService(nil),
		workerServices: workerServices{
			workerCreation: workercreation.NewService(workercreation.Deps{
				Catalog: workerruntime.DefaultCatalog(),
			}),
		},
	}
	require.NoError(t, attachOrchestrationControl(
		services,
		testkit.SetupTestDB(t),
	))

	err := attachOrchestrationWorkerApply(
		services,
		&workerPodOrchestratorStub{},
		&workerDispatchQueueStub{},
	)

	require.NoError(t, err)
	assert.NotNil(t, services.workerApply)
}
