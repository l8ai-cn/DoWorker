package agentpod

import (
	"context"
	"testing"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareStructuredWorkerCreateProjectsResolvedSpec(t *testing.T) {
	spec := podServiceWorkerSpec()
	layer := "MODE acp\nPROMPT \"Run checks.\"\n"
	prepared := preparedWorkerSpecForArtifactTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, int64(77)),
			ctxKeyUserID,
			int64(7),
		),
		spec,
		layer,
		nil,
	)
	spec = prepared.Spec
	resolved := prepared.Snapshot
	preparer := &workerCreationPreparer{
		prepared: prepared,
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
	})
	req := &OrchestrateCreatePodRequest{
		OrganizationID:  77,
		UserID:          7,
		WorkerSpecDraft: &workercreation.Draft{},
	}

	err := orchestrator.prepareStructuredWorkerCreate(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, spec.Runtime.WorkerType.Slug.String(), req.AgentSlug)
	require.NotNil(t, req.ModelResourceID)
	assert.Equal(t, spec.Runtime.ModelBinding.ResourceID, *req.ModelResourceID)
	assert.Equal(t, string(spec.TypeConfig.AutomationLevel), req.AutomationLevel)
	require.NotNil(t, req.Alias)
	assert.Equal(t, spec.Metadata.Alias, *req.Alias)
	require.NotNil(t, req.AgentfileLayer)
	assert.Equal(t, preparer.prepared.AgentfileLayer, *req.AgentfileLayer)
	require.NotNil(t, req.resolvedWorkerSpec)
	assert.Equal(t, resolved.SpecJSON(), req.resolvedWorkerSpec.SpecJSON())
	assert.Equal(t, spec, *req.preparedWorkerSpec)
	assert.Equal(t, specservice.Scope{OrgID: 77, UserID: 7}, preparer.scope)
}

func TestPrepareStructuredWorkerCreateRejectsConflictingLegacyInput(t *testing.T) {
	resolved := resolvedWorkerSpecForPodServiceTest(t, 77)
	preparer := &workerCreationPreparer{
		prepared: workercreation.Prepared{
			Snapshot:       resolved,
			Spec:           podServiceWorkerSpec(),
			AgentfileLayer: "MODE acp\n",
		},
	}
	layer := "MODE pty"
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
	})
	req := &OrchestrateCreatePodRequest{
		OrganizationID:  77,
		UserID:          7,
		AgentfileLayer:  &layer,
		WorkerSpecDraft: &workercreation.Draft{},
	}

	err := orchestrator.prepareStructuredWorkerCreate(context.Background(), req)

	require.ErrorIs(t, err, ErrConflictingWorkerCreateInput)
	assert.Zero(t, preparer.calls)
}

func TestPrepareStructuredWorkerCreateRequiresConfiguredService(t *testing.T) {
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{})

	err := orchestrator.prepareStructuredWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:  77,
			UserID:          7,
			WorkerSpecDraft: &workercreation.Draft{},
		},
	)

	require.ErrorIs(t, err, ErrWorkerCreationUnavailable)
}

func TestResumeInheritsWorkerSpecSnapshotWithoutPreparingDraft(t *testing.T) {
	db := setupOrchestratorTestDB(t)
	podService := NewPodService(infra.NewPodRepository(db))
	snapshotID := int64(91)
	spec := normalizedSnapshotWorkerSpec(t)
	modelResourceID := spec.Runtime.ModelBinding.ResourceID
	agentfileLayer := "MODE acp\n"
	source, err := podService.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:       1,
		RunnerID:             1,
		AgentSlug:            "codex-cli",
		CreatedByID:          1,
		ModelResourceID:      &modelResourceID,
		InteractionMode:      string(spec.TypeConfig.InteractionMode),
		AutomationLevel:      string(spec.TypeConfig.AutomationLevel),
		AgentfileLayer:       agentfileLayer,
		WorkerSpecSnapshotID: &snapshotID,
	})
	require.NoError(t, err)
	require.NoError(t, podService.UpdatePodStatus(
		context.Background(),
		source.PodKey,
		poddomain.StatusTerminated,
	))
	preparer := &workerCreationPreparer{}
	snapshots := &workerSpecSnapshotLoader{
		snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 1,
			Spec:           spec,
		},
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:         podService,
		WorkerCreation:     preparer,
		WorkerSpecs:        snapshots,
		WorkerDependencies: snapshotDependencyLoader(t, 1, spec),
	})
	req := &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   source.PodKey,
	}

	_, _, err = orchestrator.handleResumeMode(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, req.workerSpecSnapshotID)
	assert.Equal(t, snapshotID, *req.workerSpecSnapshotID)
	require.NotNil(t, req.preparedWorkerSpec)
	normalized, err := specdomain.NormalizeAndValidate(spec)
	require.NoError(t, err)
	assert.Equal(t, normalized, *req.preparedWorkerSpec)
	require.NotNil(t, req.AgentfileLayer)
	assert.Equal(t, agentfileLayer, *req.AgentfileLayer)
	assert.Equal(t, int64(1), snapshots.organizationID)
	assert.Equal(t, snapshotID, snapshots.snapshotID)
	assert.Zero(t, preparer.calls)
}

func TestResumeUsesPinnedWorkerSpecModelRevision(t *testing.T) {
	db := setupOrchestratorTestDB(t)
	podService := newTestPodService(db)
	spec := normalizedSnapshotWorkerSpec(t)
	snapshotID := int64(91)
	modelResourceID := spec.Runtime.ModelBinding.ResourceID
	agentfileLayer := "MODE acp\n"
	source, err := podService.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:       1,
		RunnerID:             1,
		AgentSlug:            spec.Runtime.WorkerType.Slug.String(),
		CreatedByID:          1,
		ModelResourceID:      &modelResourceID,
		InteractionMode:      string(spec.TypeConfig.InteractionMode),
		AutomationLevel:      string(spec.TypeConfig.AutomationLevel),
		AgentfileLayer:       agentfileLayer,
		WorkerSpecSnapshotID: &snapshotID,
	})
	require.NoError(t, err)
	require.NoError(t, podService.UpdatePodStatus(
		context.Background(),
		source.PodKey,
		poddomain.StatusTerminated,
	))
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Connection.ProviderKey = spec.Runtime.ModelBinding.ProviderKey
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision + 1
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID
	provider := newCodexTestProvider()
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:     podService,
		ConfigBuilder:  agent.NewConfigBuilder(provider, noopBundleLoader{}),
		AgentResolver:  &mockAgentResolver{agentDef: provider.agentDef},
		WorkerCreation: &workerCreationPreparer{},
		ModelResources: &recordingModelResourceResolver{
			resource: resource,
		},
		WorkerSpecs: &workerSpecSnapshotLoader{
			snapshot: specdomain.Snapshot{
				ID:             snapshotID,
				OrganizationID: 1,
				Spec:           spec,
			},
		},
		WorkerDependencies: snapshotDependencyLoader(t, 1, spec),
	})

	result, err := createPodWithPlanSourceForTest(t, orchestrator, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   source.PodKey,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	var podCount int64
	require.NoError(t, db.Model(&poddomain.Pod{}).Count(&podCount).Error)
	assert.Equal(t, int64(2), podCount)
}

func TestCreatePodUsesArtifactWithoutCurrentDefinitionCheck(t *testing.T) {
	spec := podServiceWorkerSpec()
	prepared := preparedWorkerSpecForArtifactTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, int64(1)),
			ctxKeyUserID,
			int64(1),
		),
		spec,
		"MODE acp\n",
		nil,
	)
	spec = prepared.Spec
	preparer := &workerCreationPreparer{
		prepared: prepared,
		validate: workercreation.ErrWorkerTypeDefinitionChanged,
	}
	provider := newCodexTestProvider()
	db := setupOrchestratorTestDB(t)
	ensureWorkerSpecSnapshotTable(t, db)
	resource := resolvedModelResourceFromSpecForArtifactTest(t, spec)
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		AgentResolver:  &mockAgentResolver{agentDef: provider.agentDef},
		WorkerCreation: preparer,
		ModelResources: &recordingModelResourceResolver{resource: resource},
		PodService:     newTestPodService(db),
		ConfigBuilder:  agent.NewConfigBuilder(provider, noopBundleLoader{}),
		PodCoordinator: &mockPodCoordinator{},
		RunnerSelector: &mockRunnerSelector{
			resolveRunner: &runnerDomain.Runner{ID: 1},
		},
	})

	_, err := createPodWithPlanSourceForTest(t, orchestrator, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		WorkerSpecDraft: &workercreation.Draft{},
	})

	require.NoError(t, err)
	assert.Zero(t, preparer.validateCalls)
}

func TestCreatePodPersistsWorkerSpecBeforeRunnerDispatch(t *testing.T) {
	db := setupOrchestratorTestDB(t)
	ensureWorkerSpecSnapshotTable(t, db)
	spec := podServiceWorkerSpec()
	prepared := preparedWorkerSpecForArtifactTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, int64(1)),
			ctxKeyUserID,
			int64(1),
		),
		spec,
		"MODE acp\nPROMPT \"Run checks.\"\n",
		nil,
	)
	spec = prepared.Spec
	preparer := &workerCreationPreparer{
		prepared: prepared,
	}
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID
	resource.Connection.ProviderKey = spec.Runtime.ModelBinding.ProviderKey
	provider := newCodexTestProvider()
	coordinator := &workerSpecDispatchObserver{db: db}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:     newTestPodService(db),
		ConfigBuilder:  agent.NewConfigBuilder(provider, noopBundleLoader{}),
		PodCoordinator: coordinator,
		RunnerSelector: &mockRunnerSelector{
			resolveRunner: &runnerDomain.Runner{ID: 1},
		},
		AgentResolver:  &mockAgentResolver{agentDef: provider.agentDef},
		ModelResources: &recordingModelResourceResolver{resource: resource},
		WorkerCreation: preparer,
	})

	result, err := createPodWithPlanSourceForTest(t, orchestrator, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		WorkerSpecDraft: &workercreation.Draft{},
	})

	require.NoError(t, err)
	require.NotNil(t, result.Pod.WorkerSpecSnapshotID)
	assert.Positive(t, *result.Pod.WorkerSpecSnapshotID)
	assert.True(t, coordinator.observedSnapshot)
	assert.NoError(t, coordinator.err)
	assert.Equal(t, 1, preparer.calls)
}
