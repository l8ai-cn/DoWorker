package workbench

import (
	"encoding/json"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"google.golang.org/protobuf/proto"
)

func (m *Mapper) ToolUpdate(
	sessionID string,
	update acp.ToolCallUpdate,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setExternalSessionLocked(sessionID)
	identity, category, ok := resolveToolIdentity(m.sourceProtocol, update.ToolName)
	if !ok {
		return m.batchLocked(
			update,
			m.unsupportedMutationLocked("tool.unknown", stringPayload(update)),
		)
	}
	phase, ok := toolPhase(update.Status)
	if !ok {
		return m.batchLocked(
			update,
			m.unsupportedMutationLocked("tool.phase", stringPayload(update)),
		)
	}
	operation := agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE
	execution := m.tools[update.ToolCallID]
	if execution == nil {
		execution = &agentworkbenchv2.ToolExecution{
			ExecutionId: update.ToolCallID,
			Identity:    identity,
			Category:    stringPointer(category),
			Input:       rawPayload("application/json", arguments(update.ArgumentsJSON)),
			Title:       stringPointer(update.ToolName),
		}
		m.tools[update.ToolCallID] = execution
		operation = agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND
	}
	execution.Phase = phase
	content := toolTimelineContent(execution)
	return m.batchLocked(
		update,
		timelineMutation(operation, toolItemID(update.ToolCallID), content),
	)
}

func (m *Mapper) ToolResult(
	sessionID string,
	result acp.ToolCallResult,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setExternalSessionLocked(sessionID)
	execution := m.tools[result.ToolCallID]
	if execution == nil {
		identity, category, ok := resolveToolIdentity(m.sourceProtocol, result.ToolName)
		if !ok {
			return m.batchLocked(
				result,
				m.unsupportedMutationLocked("tool.unknown", stringPayload(result)),
			)
		}
		execution = &agentworkbenchv2.ToolExecution{
			ExecutionId: result.ToolCallID,
			Identity:    identity,
			Category:    stringPointer(category),
			Input:       rawPayload("application/json", "{}"),
			Title:       stringPointer(result.ToolName),
		}
		m.tools[result.ToolCallID] = execution
	}
	output := result.ResultText
	if !result.Success && output == "" {
		output = result.ErrorMessage
	}
	if output != "" {
		execution.Results = []*agentworkbenchv2.ToolResult{{
			ResultId: result.ToolCallID + ":result",
			Primary:  true,
			Blocks: []*agentworkbenchv2.ContentBlock{
				textBlock(result.ToolCallID, "tool-result", output),
			},
		}}
	}
	if result.Success {
		execution.Phase = agentworkbenchv2.ToolPhase_TOOL_PHASE_COMPLETED
		execution.Failure = nil
	} else {
		execution.Phase = agentworkbenchv2.ToolPhase_TOOL_PHASE_FAILED
		execution.Failure = &agentworkbenchv2.ToolFailure{
			Code:    "tool_execution_failed",
			Message: result.ErrorMessage,
		}
	}
	return m.batchLocked(
		result,
		timelineMutation(
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
			toolItemID(result.ToolCallID),
			toolTimelineContent(execution),
		),
	)
}

func toolTimelineContent(
	execution *agentworkbenchv2.ToolExecution,
) *agentworkbenchv2.TimelineItemContent {
	return &agentworkbenchv2.TimelineItemContent{
		Content: &agentworkbenchv2.TimelineItemContent_ToolExecution{
			ToolExecution: proto.Clone(execution).(*agentworkbenchv2.ToolExecution),
		},
	}
}

func toolItemID(toolCallID string) string {
	return "tool:" + toolCallID
}

func arguments(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

func toolPhase(status string) (agentworkbenchv2.ToolPhase, bool) {
	switch status {
	case "pending":
		return agentworkbenchv2.ToolPhase_TOOL_PHASE_QUEUED, true
	case "running", "in_progress":
		return agentworkbenchv2.ToolPhase_TOOL_PHASE_RUNNING, true
	case "completed":
		return agentworkbenchv2.ToolPhase_TOOL_PHASE_COMPLETED, true
	case "failed":
		return agentworkbenchv2.ToolPhase_TOOL_PHASE_FAILED, true
	case "cancelled":
		return agentworkbenchv2.ToolPhase_TOOL_PHASE_CANCELLED, true
	default:
		return agentworkbenchv2.ToolPhase_TOOL_PHASE_UNSPECIFIED, false
	}
}

func stringPayload(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
