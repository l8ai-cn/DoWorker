package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func TestPrepareSnapshotWorkerCreateProjectsImmutableSnapshot(t *testing.T) {
	snapshotID := int64(91)
	spec := podServiceWorkerSpec()
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
		WorkerCreation: preparer,
		WorkerSpecs:    loader,
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
	})

	err := orchestrator.prepareSnapshotWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:       7,
			UserID:               5,
			RunnerID:             88,
			WorkerSpecSnapshotID: &snapshotID,
		},
	)

	require.ErrorIs(t, err, ErrConflictingWorkerCreateInput)
	assert.Zero(t, preparer.snapshotCalls)
}

type snapshotWorkerCreationPreparer struct {
	prepared      workercreation.PreparedSnapshot
	err           error
	snapshotCalls int
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
