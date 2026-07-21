package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestProjectRunnerBatchPreservesRichArtifacts(t *testing.T) {
	sourceID := "image-source"
	resultID := "image-result"
	batch := runnerBatch(
		artifactMutation("source:1", 1, imageArtifact(sourceID, resultID)),
		artifactMutation("source:2", 2, videoArtifact()),
	)

	projected, err := ProjectRunnerBatch(nil, "conv_1", "stream-1", batch)
	require.NoError(t, err)
	require.Equal(t, uint64(1), projected.Snapshot.Revision)
	require.Equal(t, uint64(2), projected.Snapshot.LatestSequence)
	require.Len(t, projected.Delta.Events, 2)
	require.Len(t, projected.Snapshot.Artifacts, 2)

	image := projected.Snapshot.Artifacts[0]
	require.Equal(t, uint64(7), image.Revision)
	require.Len(t, image.Representations, 2)
	require.Equal(t, sourceID, image.GetManifest().GetImageEdit().GetSourceRepresentationId())
	require.Equal(t, resultID, image.GetManifest().GetImageEdit().GetResultRepresentationId())

	video := projected.Snapshot.Artifacts[1]
	require.Equal(
		t,
		agentworkbenchv2.VideoStage_VIDEO_STAGE_RENDERING,
		video.GetManifest().GetVideo().GetStage(),
	)
	require.Equal(t, 0.4, video.GetManifest().GetVideo().GetProgressFraction())
	require.NotEmpty(t, projected.Delta.Digest)
	require.NotEmpty(t, projected.Snapshot.Digest)
}

func TestProjectRunnerBatchAppliesTimelineUpdatesAndStatus(t *testing.T) {
	current, err := ProjectRunnerBatch(nil, "conv_1", "stream-1", runnerBatch(
		timelineMutation(
			"source:1",
			1,
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND,
			"tool:1",
			toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_RUNNING),
		),
	))
	require.NoError(t, err)

	projected, err := ProjectRunnerBatch(
		current.Snapshot,
		"conv_1",
		"stream-1",
		runnerBatch(
			timelineMutation(
				"source:2",
				2,
				agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
				"tool:1",
				toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_COMPLETED),
			),
			statusMutation("source:3", 3, agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE),
		),
	)
	require.NoError(t, err)
	require.Equal(t, uint64(2), projected.Snapshot.Revision)
	require.Equal(t, uint64(3), projected.Snapshot.LatestSequence)
	require.Len(t, projected.Snapshot.History, 1)
	require.Equal(
		t,
		agentworkbenchv2.ToolPhase_TOOL_PHASE_COMPLETED,
		projected.Snapshot.History[0].GetContent().GetToolExecution().GetPhase(),
	)
	require.Equal(t, agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE, projected.Snapshot.Status)
}

func TestProjectRunnerBatchAppliesConfiguration(t *testing.T) {
	model := "gpt-5.4"
	mode := "ask_dangerous"
	batch := runnerBatch(&agentworkbenchv2.RunnerWorkbenchMutation{
		Source: source("source:1", 1),
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Configuration{
			Configuration: &agentworkbenchv2.SessionConfiguration{
				Model: &model, PermissionMode: &mode,
			},
		},
	})

	projected, err := ProjectRunnerBatch(nil, "conv_1", "stream-1", batch)

	require.NoError(t, err)
	require.Equal(t, model, projected.Snapshot.GetConfiguration().GetModel())
	require.Equal(t, mode, projected.Snapshot.GetConfiguration().GetPermissionMode())
	require.Equal(
		t,
		model,
		projected.Delta.GetEvents()[0].GetConfigurationChanged().GetConfiguration().GetModel(),
	)
}

func TestProjectRunnerBatchRejectsInvalidSourceAndMissingUpdate(t *testing.T) {
	tests := []struct {
		name  string
		batch *agentworkbenchv2.RunnerWorkbenchEventBatch
	}{
		{
			name: "duplicate source identity",
			batch: runnerBatch(
				statusMutation("same", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
				statusMutation("same", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE),
			),
		},
		{
			name: "missing timeline item",
			batch: runnerBatch(timelineMutation(
				"source:1",
				1,
				agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
				"missing",
				toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_COMPLETED),
			)),
		},
		{
			name: "unspecified operation",
			batch: runnerBatch(timelineMutation(
				"source:1",
				1,
				agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UNSPECIFIED,
				"tool:1",
				toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_RUNNING),
			)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ProjectRunnerBatch(nil, "conv_1", "stream-1", test.batch)
			require.ErrorIs(t, err, ErrInvalidBatch)
		})
	}
}

func TestProjectRunnerBatchValidatesToolArtifactReferences(t *testing.T) {
	current, err := ProjectRunnerBatch(nil, "conv_1", "stream-1", runnerBatch(
		timelineMutation(
			"source:1",
			1,
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND,
			"tool:1",
			toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_RUNNING),
		),
	))
	require.NoError(t, err)

	artifact := imageArtifact("image-source", "image-result")
	artifact.Revisions = []*agentworkbenchv2.ArtifactRevision{{
		Revision: artifact.Revision,
		Provenance: &agentworkbenchv2.ArtifactProvenance{
			ToolExecutionId: stringPointer("execution-1"),
		},
	}}
	valid := runnerBatch(
		artifactMutation("source:2", 2, artifact),
		timelineMutation(
			"source:3",
			3,
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
			"tool:1",
			toolContentWithArtifact(artifact.ArtifactId, artifact.Revision),
		),
	)

	projected, err := ProjectRunnerBatch(
		current.Snapshot,
		"conv_1",
		"stream-1",
		valid,
	)

	require.NoError(t, err)
	require.Len(t,
		projected.Snapshot.History[0].GetContent().GetToolExecution().
			GetResults()[0].GetArtifacts(),
		1,
	)
}

func TestProjectRunnerBatchRejectsBrokenToolArtifactReferences(t *testing.T) {
	current, err := ProjectRunnerBatch(nil, "conv_1", "stream-1", runnerBatch(
		timelineMutation(
			"source:1",
			1,
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND,
			"tool:1",
			toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_RUNNING),
		),
	))
	require.NoError(t, err)
	artifact := imageArtifact("image-source", "image-result")
	artifact.Revisions = []*agentworkbenchv2.ArtifactRevision{{
		Revision: artifact.Revision,
		Provenance: &agentworkbenchv2.ArtifactProvenance{
			ToolExecutionId: stringPointer("different-execution"),
		},
	}}

	tests := []struct {
		name      string
		artifact  *agentworkbenchv2.ArtifactDescriptor
		reference *agentworkbenchv2.ArtifactReference
	}{
		{
			name: "missing artifact",
			reference: &agentworkbenchv2.ArtifactReference{
				ArtifactId: "missing", Revision: uint64Pointer(1),
			},
		},
		{
			name:     "revision mismatch",
			artifact: artifact,
			reference: &agentworkbenchv2.ArtifactReference{
				ArtifactId: artifact.ArtifactId,
				Revision:   uint64Pointer(artifact.Revision + 1),
			},
		},
		{
			name:     "provenance mismatch",
			artifact: artifact,
			reference: &agentworkbenchv2.ArtifactReference{
				ArtifactId: artifact.ArtifactId,
				Revision:   uint64Pointer(artifact.Revision),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutations := []*agentworkbenchv2.RunnerWorkbenchMutation{}
			sequence := uint64(2)
			if test.artifact != nil {
				mutations = append(
					mutations,
					artifactMutation("source:2", sequence, test.artifact),
				)
				sequence++
			}
			mutations = append(mutations, timelineMutation(
				"source:3",
				sequence,
				agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
				"tool:1",
				toolContentWithReference(test.reference),
			))

			_, err := ProjectRunnerBatch(
				current.Snapshot,
				"conv_1",
				"stream-1",
				runnerBatch(mutations...),
			)

			require.ErrorIs(t, err, ErrInvalidBatch)
		})
	}
}
