package agentpod

import (
	"context"
	"testing"

	poddomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveExecutionSource(t *testing.T) {
	snapshotID := int64(91)
	tests := []struct {
		name string
		req  OrchestrateCreatePodRequest
		want ExecutionSource
	}{
		{
			name: "plan",
			req: OrchestrateCreatePodRequest{
				WorkerSpecDraft: &workercreation.Draft{},
			},
			want: ExecutionSourcePlan,
		},
		{
			name: "snapshot",
			req: OrchestrateCreatePodRequest{
				WorkerSpecSnapshotID: &snapshotID,
			},
			want: ExecutionSourceSnapshot,
		},
		{
			name: "lineage",
			req: OrchestrateCreatePodRequest{
				SourcePodKey: "source-pod",
			},
			want: ExecutionSourceLineage,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveExecutionSource(&test.req)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestResolveExecutionSourceRejectsMultipleSources(t *testing.T) {
	snapshotID := int64(91)
	tests := []struct {
		name string
		req  OrchestrateCreatePodRequest
	}{
		{
			name: "plan and snapshot",
			req: OrchestrateCreatePodRequest{
				WorkerSpecDraft:      &workercreation.Draft{},
				WorkerSpecSnapshotID: &snapshotID,
			},
		},
		{
			name: "plan and lineage",
			req: OrchestrateCreatePodRequest{
				WorkerSpecDraft: &workercreation.Draft{},
				SourcePodKey:    "source-pod",
			},
		},
		{
			name: "snapshot and lineage",
			req: OrchestrateCreatePodRequest{
				WorkerSpecSnapshotID: &snapshotID,
				SourcePodKey:         "source-pod",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := resolveExecutionSource(&test.req)
			require.ErrorIs(t, err, ErrConflictingWorkerCreateInput)
		})
	}
}

func TestResolveExecutionSourceRejectsMissingSource(t *testing.T) {
	_, err := resolveExecutionSource(&OrchestrateCreatePodRequest{
		AgentSlug: "codex-cli",
	})

	require.ErrorIs(t, err, ErrConflictingWorkerCreateInput)
}

func TestResumeAppendsInvocationPromptToPinnedWorkerSpec(t *testing.T) {
	db := setupOrchestratorTestDB(t)
	podService := NewPodService(infra.NewPodRepository(db))
	snapshotID := int64(91)
	spec := normalizedSnapshotWorkerSpec(t)
	modelResourceID := spec.Runtime.ModelBinding.ResourceID
	sourceLayer := "MODE acp\nPROMPT \"First run.\"\n"
	source, err := podService.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:       1,
		RunnerID:             1,
		AgentSlug:            spec.Runtime.WorkerType.Slug.String(),
		CreatedByID:          1,
		ModelResourceID:      &modelResourceID,
		InteractionMode:      string(spec.TypeConfig.InteractionMode),
		AutomationLevel:      string(spec.TypeConfig.AutomationLevel),
		AgentfileLayer:       sourceLayer,
		WorkerSpecSnapshotID: &snapshotID,
	})
	require.NoError(t, err)
	require.NoError(t, podService.UpdatePodStatus(
		context.Background(),
		source.PodKey,
		poddomain.StatusTerminated,
	))
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:     podService,
		WorkerCreation: &workerCreationPreparer{},
		WorkerSpecs: &workerSpecSnapshotLoader{snapshot: specdomain.Snapshot{
			ID:             snapshotID,
			OrganizationID: 1,
			Spec:           spec,
		}},
		WorkerDependencies: snapshotDependencyLoader(t, 1, spec),
	})
	override := "Continue with the next run."
	req := &OrchestrateCreatePodRequest{
		OrganizationID:           1,
		UserID:                   1,
		SourcePodKey:             source.PodKey,
		WorkerSpecPromptOverride: &override,
	}

	_, _, err = orchestrator.handleResumeMode(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, req.AgentfileLayer)
	resolved, err := extractFromAgentfileLayer(
		baseAgentfileSrc,
		*req.AgentfileLayer,
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, override, resolved.Prompt)
	require.NotNil(t, req.workerSpecSnapshotID)
	assert.Equal(t, snapshotID, *req.workerSpecSnapshotID)
}
