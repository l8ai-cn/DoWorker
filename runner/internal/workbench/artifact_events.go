package workbench

import (
	"sort"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func (m *Mapper) Artifacts(
	artifacts []*agentworkbenchv2.ArtifactDescriptor,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	if len(artifacts) == 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	mutations := make(
		[]*agentworkbenchv2.RunnerWorkbenchMutation,
		0,
		len(artifacts)*2,
	)
	linked := make(map[string]*agentworkbenchv2.ToolExecution)
	for _, artifact := range artifacts {
		mutations = append(mutations, artifactMutation(artifact))
		toolExecutionID := artifactRevisionExecutionID(artifact)
		if toolExecutionID == "" {
			continue
		}
		execution := m.tools[toolExecutionID]
		if execution == nil {
			mutations = append(
				mutations,
				m.unsupportedMutationLocked(
					"artifact.tool_execution_missing",
					stringPayload(map[string]string{
						"artifact_id":       artifact.GetArtifactId(),
						"tool_execution_id": toolExecutionID,
					}),
				),
			)
			continue
		}
		upsertToolResultArtifact(execution, artifact)
		linked[toolExecutionID] = execution
	}
	ids := make([]string, 0, len(linked))
	for id := range linked {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		mutations = append(
			mutations,
			timelineMutation(
				agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
				toolItemID(id),
				toolTimelineContent(linked[id]),
			),
		)
	}
	return m.batchLocked(artifacts, mutations...)
}

func artifactRevisionExecutionID(
	artifact *agentworkbenchv2.ArtifactDescriptor,
) string {
	for _, revision := range artifact.GetRevisions() {
		if revision.GetRevision() == artifact.GetRevision() {
			return revision.GetProvenance().GetToolExecutionId()
		}
	}
	return ""
}

func artifactMutation(
	artifact *agentworkbenchv2.ArtifactDescriptor,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Artifact{
			Artifact: artifact,
		},
	}
}

func upsertToolResultArtifact(
	execution *agentworkbenchv2.ToolExecution,
	artifact *agentworkbenchv2.ArtifactDescriptor,
) {
	result := primaryToolResult(execution)
	reference := &agentworkbenchv2.ArtifactReference{
		ArtifactId: artifact.GetArtifactId(),
		Revision:   uint64Pointer(artifact.GetRevision()),
		Role:       optionalNonEmptyString(artifact.GetRole()),
	}
	for index, existing := range result.Artifacts {
		if existing.GetArtifactId() == artifact.GetArtifactId() {
			result.Artifacts[index] = reference
			return
		}
	}
	result.Artifacts = append(result.Artifacts, reference)
}

func primaryToolResult(
	execution *agentworkbenchv2.ToolExecution,
) *agentworkbenchv2.ToolResult {
	for _, result := range execution.Results {
		if result.GetPrimary() {
			return result
		}
	}
	if len(execution.Results) > 0 {
		return execution.Results[0]
	}
	result := &agentworkbenchv2.ToolResult{
		ResultId: execution.GetExecutionId() + ":result",
		Primary:  true,
	}
	execution.Results = append(execution.Results, result)
	return result
}

func uint64Pointer(value uint64) *uint64 {
	return &value
}
