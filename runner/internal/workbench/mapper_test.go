package workbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/stretchr/testify/require"
)

func TestContentChunkRejectsUnknownRole(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")

	batch := mapper.ContentChunk("", acp.ContentChunk{
		Role: "plan",
		Text: "not a chat message",
	})

	require.Len(t, batch.GetMutations(), 1)
	require.NotNil(t, batch.GetMutations()[0].GetUnsupported())
	require.Nil(t, batch.GetMutations()[0].GetTimeline())
}

func TestToolUpdateRejectsUnknownPhase(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")

	batch := mapper.ToolUpdate("", acp.ToolCallUpdate{
		ToolCallID: "tool-1",
		ToolName:   "shell",
		Status:     "mystery",
	})

	require.Len(t, batch.GetMutations(), 1)
	require.NotNil(t, batch.GetMutations()[0].GetUnsupported())
	require.Nil(t, batch.GetMutations()[0].GetTimeline())
}

func TestPlanRejectsUnknownStepStatus(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")

	batch := mapper.Plan("", acp.PlanUpdate{Steps: []acp.PlanStep{{
		Title:  "Render",
		Status: "paused",
	}}})

	require.Len(t, batch.GetMutations(), 1)
	require.NotNil(t, batch.GetMutations()[0].GetUnsupported())
	require.Nil(t, batch.GetMutations()[0].GetTimeline())
}

func TestArtifactsProducesTypedMutations(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")
	artifact := &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "workspace:output/demo.mp4",
		Revision:   1,
		Filename:   "demo.mp4",
		MediaType:  "video/mp4",
		Status:     agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
	}

	batch := mapper.Artifacts([]*agentworkbenchv2.ArtifactDescriptor{artifact})

	require.Len(t, batch.GetMutations(), 1)
	require.Equal(t, artifact, batch.GetMutations()[0].GetArtifact())
	require.NotEmpty(t, batch.GetMutations()[0].GetSource().GetStableEventId())
}

func TestArtifactsLinkToExplicitToolResult(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")
	mapper.ToolUpdate("", acp.ToolCallUpdate{
		ToolCallID: "tool-1", ToolName: "shell", Status: "running",
	})
	mapper.ToolResult("", acp.ToolCallResult{
		ToolCallID: "tool-1", ToolName: "shell", Success: true,
		ResultText: "created output",
	})
	revision := uint64(2)
	artifact := &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "rendered-page", Revision: revision,
		Filename: "index.html", MediaType: "text/html",
		Role:   stringPointer("preview"),
		Status: agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		Revisions: []*agentworkbenchv2.ArtifactRevision{{
			Revision: revision,
			Provenance: &agentworkbenchv2.ArtifactProvenance{
				ToolExecutionId: stringPointer("tool-1"),
			},
		}},
	}

	batch := mapper.Artifacts([]*agentworkbenchv2.ArtifactDescriptor{artifact})

	require.Len(t, batch.GetMutations(), 2)
	require.Equal(t, artifact, batch.GetMutations()[0].GetArtifact())
	tool := batch.GetMutations()[1].GetTimeline().GetContent().GetToolExecution()
	require.Empty(t, tool.GetArtifacts())
	require.Len(t, tool.GetResults(), 1)
	require.Len(t, tool.GetResults()[0].GetArtifacts(), 1)
	reference := tool.GetResults()[0].GetArtifacts()[0]
	require.Equal(t, "rendered-page", reference.GetArtifactId())
	require.Equal(t, revision, reference.GetRevision())
}

func TestArtifactsReportUnknownExplicitToolResult(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")
	artifact := &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "orphan", Revision: 1,
		Revisions: []*agentworkbenchv2.ArtifactRevision{{
			Revision: 1,
			Provenance: &agentworkbenchv2.ArtifactProvenance{
				ToolExecutionId: stringPointer("missing-tool"),
			},
		}},
	}

	batch := mapper.Artifacts([]*agentworkbenchv2.ArtifactDescriptor{artifact})

	require.Len(t, batch.GetMutations(), 2)
	require.Equal(t, artifact, batch.GetMutations()[0].GetArtifact())
	require.Equal(t,
		"artifact.tool_execution_missing",
		batch.GetMutations()[1].GetUnsupported().GetValue().GetIdentity().GetSemanticKey())
}

func TestSessionInitializedPublishesCapabilitiesAndCurrentConfiguration(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")

	batch := mapper.SessionInitialized(acp.Configuration{
		Model:                    "gpt-5.4",
		SupportedModels:          []string{"gpt-5.4", "gpt-5.3-codex"},
		PermissionMode:           "ask_dangerous",
		SupportedPermissionModes: []string{"ask_dangerous", "bypass"},
		SupportedArtifactActions: []string{"image.edit", "presentation.export"},
	})

	require.Len(t, batch.GetMutations(), 2)
	capabilities := batch.GetMutations()[0].GetCapabilities()
	require.Equal(t, "2", capabilities.GetProtocolVersion())
	require.Equal(t, []string{"gpt-5.4", "gpt-5.3-codex"}, capabilities.GetModels())
	require.Equal(t, []string{"ask_dangerous", "bypass"}, capabilities.GetPermissionModes())
	require.Equal(t,
		[]string{"artifact.download", "image.edit", "presentation.export"},
		capabilities.GetArtifactOperations())
	configuration := batch.GetMutations()[1].GetConfiguration()
	require.Equal(t, "gpt-5.4", configuration.GetModel())
	require.Equal(t, "ask_dangerous", configuration.GetPermissionMode())
}

func TestConfigurationChangedPublishesCurrentValues(t *testing.T) {
	mapper := NewMapper("pod-1", "codex")

	batch := mapper.ConfigurationChanged(acp.ConfigUpdate{
		Model: "gpt-5.4", PermissionMode: "bypass",
	})

	require.Len(t, batch.GetMutations(), 1)
	require.Equal(t, "gpt-5.4", batch.GetMutations()[0].GetConfiguration().GetModel())
	require.Equal(t, "bypass", batch.GetMutations()[0].GetConfiguration().GetPermissionMode())
}
