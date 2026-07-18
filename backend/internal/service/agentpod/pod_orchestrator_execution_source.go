package agentpod

import (
	"strings"

	"github.com/anthropics/agentsmesh/agentfile"
)

type ExecutionSource string

const (
	ExecutionSourcePlan     ExecutionSource = "plan"
	ExecutionSourceSnapshot ExecutionSource = "snapshot"
	ExecutionSourceLineage  ExecutionSource = "lineage"
)

func resolveExecutionSource(
	req *OrchestrateCreatePodRequest,
) (ExecutionSource, error) {
	var source ExecutionSource
	count := 0
	if req.WorkerSpecDraft != nil {
		source = ExecutionSourcePlan
		count++
	}
	if req.WorkerSpecSnapshotID != nil {
		source = ExecutionSourceSnapshot
		count++
	}
	if req.SourcePodKey != "" {
		source = ExecutionSourceLineage
		count++
	}
	if count > 1 {
		return "", ErrConflictingWorkerCreateInput
	}
	return source, nil
}

func appendWorkerSpecPromptOverride(req *OrchestrateCreatePodRequest) {
	if override := strings.TrimSpace(workerSpecStringValue(
		req.WorkerSpecPromptOverride,
	)); override != "" {
		appendAgentfileLayer(
			&req.AgentfileLayer,
			"PROMPT "+agentfile.FormatStringLiteral(override),
		)
	}
}
