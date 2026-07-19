package agentpod

import (
	"context"
	"testing"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func TestPrepareSnapshotWorkerCreateProjectsImmutableSnapshot(t *testing.T) {
	snapshotID := int64(91)
	spec := normalizedSnapshotWorkerSpec(t)
	preparer := &snapshotWorkerCreationPreparer{
		prepared: workercreation.PreparedSnapshot{
			Spec:           spec,
			AgentfileLayer: "MODE acp\nPROMPT \"Run checks.\"\n",
		},
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 7,
			Spec:           spec,
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 7, spec),
	})
	alias := "override-alias"
	prompt := "Override task."
	req := &OrchestrateCreatePodRequest{
		OrganizationID:           7,
		UserID:                   5,
		WorkerSpecSnapshotID:     &snapshotID,
		WorkerSpecPromptOverride: &prompt,
		Alias:                    &alias,
	}

	err := orchestrator.prepareSnapshotWorkerCreate(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, spec.Runtime.WorkerType.Slug.String(), req.AgentSlug)
	require.NotNil(t, req.ModelResourceID)
	assert.Equal(t, spec.Runtime.ModelBinding.ResourceID, *req.ModelResourceID)
	assert.Equal(t, &alias, req.Alias)
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `PROMPT "Run checks."`)
	assert.Contains(t, *req.AgentfileLayer, `PROMPT "Override task."`)
	resolved, err := extractFromAgentfileLayer(
		baseAgentfileSrc,
		*req.AgentfileLayer,
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, prompt, resolved.Prompt)
	require.NotNil(t, req.workerSpecSnapshotID)
	assert.Equal(t, snapshotID, *req.workerSpecSnapshotID)
	assert.Nil(t, req.resolvedWorkerSpec)
	require.NotNil(t, req.preparedWorkerSpec)
	assert.Equal(t, spec, *req.preparedWorkerSpec)
	assert.Equal(t, 1, preparer.snapshotCalls)
}

func TestPrepareSnapshotWorkerCreateRejectsSnapshotScopeMismatch(t *testing.T) {
	snapshotID := int64(91)
	preparer := &snapshotWorkerCreationPreparer{}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 8,
			Spec:           podServiceWorkerSpec(),
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 7, podServiceWorkerSpec()),
	})

	err := orchestrator.prepareSnapshotWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:       7,
			UserID:               5,
			WorkerSpecSnapshotID: &snapshotID,
		},
	)

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
	assert.Zero(t, preparer.snapshotCalls)
}

func TestPrepareSnapshotWorkerCreateRejectsInvalidSnapshotID(t *testing.T) {
	snapshotID := int64(0)
	preparer := &snapshotWorkerCreationPreparer{}
	loader := &workerSpecSnapshotLoader{
		snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 7,
			Spec:           podServiceWorkerSpec(),
		},
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation:     preparer,
		WorkerSpecs:        loader,
		WorkerDependencies: snapshotDependencyLoader(t, 7, podServiceWorkerSpec()),
	})

	err := orchestrator.prepareSnapshotWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:       7,
			UserID:               5,
			WorkerSpecSnapshotID: &snapshotID,
		},
	)

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
	assert.Zero(t, preparer.snapshotCalls)
	assert.Zero(t, loader.snapshotID)
}

func TestPrepareSnapshotWorkerCreateRejectsLegacyRuntimeOverrides(t *testing.T) {
	snapshotID := int64(91)
	preparer := &snapshotWorkerCreationPreparer{}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 7,
			Spec:           podServiceWorkerSpec(),
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 7, podServiceWorkerSpec()),
	})

	err := orchestrator.prepareSnapshotWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:       7,
			UserID:               5,
			AgentSlug:            "codex-cli",
			WorkerSpecSnapshotID: &snapshotID,
		},
	)

	require.ErrorIs(t, err, ErrConflictingWorkerCreateInput)
	assert.Zero(t, preparer.snapshotCalls)
}

func TestPrepareSnapshotWorkerCreateAllowsTicketAssociation(t *testing.T) {
	snapshotID := int64(91)
	ticketID := int64(42)
	ticketSlug := "am-42"
	spec := normalizedSnapshotWorkerSpec(t)
	preparer := &snapshotWorkerCreationPreparer{
		prepared: workercreation.PreparedSnapshot{
			Spec:           spec,
			AgentfileLayer: "MODE acp\n",
		},
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 7,
			Spec:           spec,
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 7, spec),
	})
	req := &OrchestrateCreatePodRequest{
		OrganizationID:       7,
		UserID:               5,
		WorkerSpecSnapshotID: &snapshotID,
		TicketID:             &ticketID,
		TicketSlug:           &ticketSlug,
	}

	err := orchestrator.prepareSnapshotWorkerCreate(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, &ticketID, req.TicketID)
	assert.Equal(t, &ticketSlug, req.TicketSlug)
	assert.Equal(t, 1, preparer.snapshotCalls)
}

func TestCreatePodReplaysWorkerLaunchWithSamePodAndCommand(t *testing.T) {
	db := setupOrchestratorTestDB(t)
	spec := normalizedSnapshotWorkerSpec(t)
	snapshotID := int64(91)
	launchID := int64(71)
	preparer := &snapshotWorkerCreationPreparer{
		prepared: workercreation.PreparedSnapshot{
			Spec:           spec,
			AgentfileLayer: "MODE acp\nPROMPT \"Run checks.\"\n",
		},
	}
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Connection.ProviderKey = spec.Runtime.ModelBinding.ProviderKey
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID
	provider := newCodexTestProvider()
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:     NewPodService(infra.NewPodRepository(db)),
		ConfigBuilder:  agent.NewConfigBuilder(provider, noopBundleLoader{}),
		AgentResolver:  &mockAgentResolver{agentDef: provider.agentDef},
		WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID: snapshotID, OrganizationID: 1, Spec: spec,
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 1, spec),
		ModelResources:     &recordingModelResourceResolver{resource: resource},
		RunnerSelector: &mockRunnerSelector{
			runner: &runnerDomain.Runner{ID: 1},
		},
	})
	create := func() (*OrchestrateCreatePodResult, error) {
		return orchestrator.CreatePod(
			context.Background(),
			&OrchestrateCreatePodRequest{
				OrganizationID:              1,
				UserID:                      1,
				WorkerSpecSnapshotID:        &snapshotID,
				OrchestrationWorkerLaunchID: &launchID,
				DeferRunnerDispatch:         true,
				QueueIfUnavailable:          true,
			},
		)
	}

	first, err := create()
	require.NoError(t, err)
	second, err := create()
	require.NoError(t, err)

	assert.Equal(t, first.Pod.ID, second.Pod.ID)
	assert.Equal(t, first.Pod.PodKey, second.Pod.PodKey)
	assert.Equal(
		t,
		first.DeferredCreateCommand.PodKey,
		second.DeferredCreateCommand.PodKey,
	)
	var count int64
	require.NoError(t, db.Model(&poddomain.Pod{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestPrepareSnapshotWorkerCreateRejectsMissingDependencyArtifact(t *testing.T) {
	snapshotID := int64(91)
	spec := normalizedSnapshotWorkerSpec(t)
	preparer := &snapshotWorkerCreationPreparer{}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID: snapshotID, OrganizationID: 7, Spec: spec,
		}},
	})

	err := orchestrator.prepareSnapshotWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID: 7, UserID: 5, WorkerSpecSnapshotID: &snapshotID,
		},
	)

	require.ErrorIs(t, err, ErrWorkerSpecDependencyUnavailable)
	assert.Zero(t, preparer.snapshotCalls)
}

type snapshotWorkerCreationPreparer struct {
	prepared      workercreation.PreparedSnapshot
	err           error
	snapshotCalls int
}

func normalizedSnapshotWorkerSpec(t *testing.T) specdomain.Spec {
	t.Helper()
	spec := podServiceWorkerSpec()
	definition := formalWorkerDefinitionForPlanTest(t, spec.Runtime.WorkerType.Slug.String())
	spec.Runtime.WorkerType.DefinitionHash = definition.DefinitionHash
	spec, err := specdomain.NormalizeAndValidate(spec)
	require.NoError(t, err)
	return spec
}

func snapshotDependencyLoader(
	t *testing.T,
	organizationID int64,
	spec specdomain.Spec,
) *workerSpecDependencyLoader {
	t.Helper()
	_, document := planArtifactForTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, organizationID),
			ctxKeyUserID,
			int64(1),
		),
		&spec,
		"MODE acp\n",
		resolvedModelResourceFromSpecForArtifactTest(t, spec),
		nil,
	)
	return &workerSpecDependencyLoader{
		document: *document,
	}
}

func snapshotDependencyLoaderWithDigest(
	organizationID int64,
	digest string,
) *workerSpecDependencyLoader {
	return &workerSpecDependencyLoader{
		document: workerdependency.Document{
			OrganizationID: organizationID,
			Worker: workerdependency.Worker{
				SpecDigest: digest,
			},
		},
	}
}

func (*snapshotWorkerCreationPreparer) Prepare(
	context.Context,
	specservice.Scope,
	workercreation.Draft,
) (workercreation.Prepared, error) {
	return workercreation.Prepared{}, nil
}

func (*snapshotWorkerCreationPreparer) ValidateWorkerTypeSnapshot(
	context.Context,
	specservice.Scope,
	specdomain.WorkerType,
) error {
	return nil
}

func (preparer *snapshotWorkerCreationPreparer) PrepareSnapshot(
	_ context.Context,
	_ specservice.Scope,
	_ specdomain.Snapshot,
) (workercreation.PreparedSnapshot, error) {
	preparer.snapshotCalls++
	return preparer.prepared, preparer.err
}

func (preparer *snapshotWorkerCreationPreparer) PrepareSnapshotWithDependencies(
	_ context.Context,
	_ specservice.Scope,
	_ specdomain.Snapshot,
	dependencies workerdependency.Document,
) (workercreation.PreparedSnapshot, error) {
	preparer.snapshotCalls++
	prepared := preparer.prepared
	if prepared.Dependencies == nil {
		prepared.Dependencies = &dependencies
	}
	return prepared, preparer.err
}

type workerSpecDependencyLoader struct {
	document workerdependency.Document
	err      error
}

func (loader *workerSpecDependencyLoader) GetBySnapshotID(
	context.Context,
	int64,
	int64,
) (workerdependency.Document, error) {
	return loader.document, loader.err
}
