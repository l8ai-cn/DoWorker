package agentworkbench

import agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"

func runnerBatch(
	mutations ...*agentworkbenchv2.RunnerWorkbenchMutation,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	return &agentworkbenchv2.RunnerWorkbenchEventBatch{
		PodKey:                "pod-1",
		AdapterId:             "codex-cli",
		SourceProtocolVersion: "acp/1",
		RunnerSessionEpoch:    "runner-epoch-1",
		Mutations:             mutations,
	}
}

func source(stableID string, sequence uint64) *agentworkbenchv2.RunnerSourceEvent {
	return &agentworkbenchv2.RunnerSourceEvent{
		StableEventId:  stableID,
		SourceSequence: sequence,
		OccurredAt:     "2026-07-16T10:00:00Z",
	}
}

func artifactMutation(
	stableID string,
	sequence uint64,
	artifact *agentworkbenchv2.ArtifactDescriptor,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Source: source(stableID, sequence),
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Artifact{
			Artifact: artifact,
		},
	}
}

func timelineMutation(
	stableID string,
	sequence uint64,
	operation agentworkbenchv2.RunnerTimelineOperation,
	itemID string,
	content *agentworkbenchv2.TimelineItemContent,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Source: source(stableID, sequence),
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Timeline{
			Timeline: &agentworkbenchv2.RunnerTimelineMutation{
				Operation: operation,
				ItemId:    itemID,
				Content:   content,
			},
		},
	}
}

func statusMutation(
	stableID string,
	sequence uint64,
	status agentworkbenchv2.SessionStatus,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Source: source(stableID, sequence),
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Status{
			Status: &agentworkbenchv2.RunnerStatusMutation{Status: status},
		},
	}
}

func toolContent(
	phase agentworkbenchv2.ToolPhase,
) *agentworkbenchv2.TimelineItemContent {
	return &agentworkbenchv2.TimelineItemContent{
		Content: &agentworkbenchv2.TimelineItemContent_ToolExecution{
			ToolExecution: &agentworkbenchv2.ToolExecution{
				ExecutionId: "execution-1",
				Identity: &agentworkbenchv2.ToolIdentity{
					Namespace:     "agentcloud",
					SemanticKey:   "shell.execute",
					SchemaVersion: "1",
				},
				Phase: phase,
				Input: &agentworkbenchv2.StructuredPayload{
					MediaType: "application/json",
					Data:      []byte(`{"command":"go test ./..."}`),
				},
			},
		},
	}
}

func toolContentWithArtifact(
	artifactID string,
	revision uint64,
) *agentworkbenchv2.TimelineItemContent {
	return toolContentWithReference(&agentworkbenchv2.ArtifactReference{
		ArtifactId: artifactID,
		Revision:   uint64Pointer(revision),
	})
}

func toolContentWithReference(
	reference *agentworkbenchv2.ArtifactReference,
) *agentworkbenchv2.TimelineItemContent {
	content := toolContent(agentworkbenchv2.ToolPhase_TOOL_PHASE_COMPLETED)
	tool := content.GetToolExecution()
	tool.Results = []*agentworkbenchv2.ToolResult{{
		ResultId: "execution-1:result",
		Primary:  true,
		Artifacts: []*agentworkbenchv2.ArtifactReference{
			reference,
		},
	}}
	return content
}

func imageArtifact(sourceID, resultID string) *agentworkbenchv2.ArtifactDescriptor {
	result := resultID
	return &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "image-edit-1",
		Revision:   7,
		Filename:   "edited.png",
		MediaType:  "image/png",
		Status:     agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		Representations: []*agentworkbenchv2.ArtifactRepresentation{
			representation(sourceID, "source.png", "image/png"),
			representation(resultID, "edited.png", "image/png"),
		},
		Manifest: &agentworkbenchv2.ArtifactManifest{
			Manifest: &agentworkbenchv2.ArtifactManifest_ImageEdit{
				ImageEdit: &agentworkbenchv2.ImageEditManifest{
					SourceRepresentationId: sourceID,
					ResultRepresentationId: &result,
					SourceWidth:            1024,
					SourceHeight:           768,
				},
			},
		},
	}
}

func videoArtifact() *agentworkbenchv2.ArtifactDescriptor {
	progress := 0.4
	return &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "video-1",
		Revision:   2,
		Filename:   "demo.mp4",
		MediaType:  "video/mp4",
		Status:     agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING,
		Representations: []*agentworkbenchv2.ArtifactRepresentation{
			representation("original", "demo.mp4", "video/mp4"),
		},
		Manifest: &agentworkbenchv2.ArtifactManifest{
			Manifest: &agentworkbenchv2.ArtifactManifest_Video{
				Video: &agentworkbenchv2.VideoManifest{
					Stage:                    agentworkbenchv2.VideoStage_VIDEO_STAGE_RENDERING,
					ProgressFraction:         &progress,
					OriginalRepresentationId: stringPointer("original"),
				},
			},
		},
	}
}

func representation(
	id string,
	filename string,
	mediaType string,
) *agentworkbenchv2.ArtifactRepresentation {
	return &agentworkbenchv2.ArtifactRepresentation{
		RepresentationId: id,
		Revision:         1,
		MediaType:        mediaType,
		Filename:         &filename,
		Status:           agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
	}
}
