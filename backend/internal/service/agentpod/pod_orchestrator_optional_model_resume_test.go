package agentpod

import (
	"context"
	"testing"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResumeInheritsWorkerSpecWithoutMainModel(t *testing.T) {
	db := setupOrchestratorTestDB(t)
	podService := NewPodService(infra.NewPodRepository(db))
	snapshotID := int64(91)
	spec := normalizedSnapshotWorkerSpec(t)
	spec.Runtime.ModelBinding = specdomain.ModelBinding{}
	spec.Runtime.WorkerType.Slug = slugkit.MustNewForTest("cursor-cli")
	definition := formalWorkerDefinitionForPlanTest(t, spec.Runtime.WorkerType.Slug.String())
	spec.Runtime.WorkerType.DefinitionHash = definition.DefinitionHash
	spec, err := specdomain.NormalizeAndValidate(spec)
	require.NoError(t, err)
	agentfileLayer := "MODE acp\n"
	source, err := podService.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1, RunnerID: 1, CreatedByID: 1,
		AgentSlug:       spec.Runtime.WorkerType.Slug.String(),
		InteractionMode: string(spec.TypeConfig.InteractionMode),
		AutomationLevel: string(spec.TypeConfig.AutomationLevel),
		AgentfileLayer:  agentfileLayer, WorkerSpecSnapshotID: &snapshotID,
	})
	require.NoError(t, err)
	require.NoError(t, podService.UpdatePodStatus(
		context.Background(), source.PodKey, poddomain.StatusTerminated,
	))
	preparer := &workerCreationPreparer{}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService: podService, WorkerCreation: preparer,
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID: snapshotID, OrganizationID: 1, Spec: spec,
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 1, spec),
	})
	req := &OrchestrateCreatePodRequest{
		OrganizationID: 1, UserID: 1, SourcePodKey: source.PodKey,
	}

	_, _, err = orchestrator.handleResumeMode(context.Background(), req)

	require.NoError(t, err)
	assert.Nil(t, source.ModelResourceID)
	assert.True(t, sourceMatchesWorkerSpec(source, spec))
	assert.Zero(t, preparer.calls)
}
